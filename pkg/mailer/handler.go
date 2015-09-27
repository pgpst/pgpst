package mailer

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"net/mail"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/pgpst/pgpst/internal/github.com/Sirupsen/logrus"
	"github.com/pgpst/pgpst/internal/github.com/codahale/chacha20"
	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pgpst/pgpst/internal/github.com/dchest/uniuri"
	"github.com/pgpst/pgpst/internal/github.com/lavab/go-spamc"
	"github.com/pgpst/pgpst/internal/github.com/pgpst/smtpd"
	"github.com/pgpst/pgpst/internal/golang.org/x/crypto/openpgp"
	"github.com/pgpst/pgpst/internal/golang.org/x/crypto/poly1305"

	"github.com/pgpst/pgpst/pkg/models"
	"github.com/pgpst/pgpst/pkg/utils"
)

func (m *Mailer) Wrap(x func()) func() {
	return func() {
		if m.Raven != nil {
			m.Raven.CapturePanic(x, nil)
		} else {
			x()
		}
	}
}

type recipient struct {
	Address *models.Address `gorethink:"address"`
	Account *models.Account `gorethink:"account"`
	Key     *models.Key     `gorethink:"key"`
	Labels  struct {
		Inbox string `gorethink:"inbox"`
		Spam  string `gorethink:"spam"`
	} `gorethink:"labels"`
}

func (m *Mailer) HandleRecipient(next func(conn *smtpd.Connection)) func(conn *smtpd.Connection) {
	return func(conn *smtpd.Connection) {
		// Prepare the context
		if conn.Environment == nil {
			conn.Environment = map[string]interface{}{}
		}
		if _, ok := conn.Environment["recipients"]; !ok {
			conn.Environment["recipients"] = []recipient{}
		}
		recipients := conn.Environment["recipients"].([]recipient)

		// Get the most recently added recipient and parse it
		addr, err := mail.ParseAddress(
			conn.Envelope.Recipients[len(conn.Envelope.Recipients)-1],
		)
		if err != nil {
			m.Error(conn, err)
			return
		}

		// Normalize the address
		addr.Address = utils.RemoveDots(utils.NormalizeAddress(addr.Address))

		// Fetch the address and account from database
		cursor, err := r.Table("addresses").Get(addr.Address).Default(map[string]interface{}{}).Do(func(address r.Term) map[string]interface{} {
			return map[string]interface{}{
				"address": address,
				"account": r.Branch(
					address.HasFields("id"),
					r.Table("accounts").Get(address.Field("owner")),
					nil,
				),
				"key": r.Branch(
					address.HasFields("public_key").And(address.Field("public_key").Ne("")),
					r.Table("keys").Get(address.Field("public_key")).Without("identities"),
					r.Branch(
						address.HasFields("id"),
						r.Table("keys").GetAllByIndex("owner", address.Field("owner")).OrderBy("date_created").CoerceTo("array").Do(func(keys r.Term) r.Term {
							return r.Branch(
								keys.Count().Gt(0),
								keys.Nth(-1).Without("identities"),
								nil,
							)
						}),
						nil,
					),
				),
			}
		}).Do(func(data r.Term) r.Term {
			return data.Merge(map[string]interface{}{
				"labels": r.Branch(
					data.Field("account").HasFields("id"),
					r.Table("labels").GetAllByIndex("nameOwnerSystem", []interface{}{
						"Inbox",
						data.Field("account").Field("id"),
						true,
					}, []interface{}{
						"Spam",
						data.Field("account").Field("id"),
						true,
					}).CoerceTo("array").Do(func(result r.Term) r.Term {
						return r.Branch(
							result.Count().Eq(2),
							map[string]interface{}{
								"inbox": result.Nth(0).Field("id"),
								"spam":  result.Nth(1).Field("id"),
							},
							nil,
						)
					}),
					nil,
				),
			})
		}).Run(m.Rethink)
		if err != nil {
			m.Error(conn, err)
			return
		}
		defer cursor.Close()
		var result recipient
		if err := cursor.One(&result); err != nil {
			m.Error(conn, err)
			return
		}

		// Check if anything got matched
		if result.Address == nil || result.Address.ID == "" || result.Account == nil || result.Account.ID == "" {
			conn.Error(errors.New("No such address"))
			return
		}
		if result.Key == nil || result.Labels.Inbox == "" || result.Labels.Spam == "" {
			conn.Error(errors.New("Account is not configured"))
			return
		}

		// Append the result to the recipients
		conn.Environment["recipients"] = append(recipients, result)

		// Run the next handler
		next(conn)
	}
}

type envelope struct {
	ID string

	From string
	To   []string
	CC   []string

	Subject    string
	MemoryHole []byte
	Body       []byte
}

func (m *Mailer) HandleDelivery(next func(conn *smtpd.Connection)) func(conn *smtpd.Connection) {
	return func(conn *smtpd.Connection) {
		// Context variables
		var (
			isSpam = false
			ctxID  = uniuri.NewLen(uniuri.UUIDLen)
		)

		// Check for spam
		spamReply, err := m.Spam.Report(string(conn.Envelope.Data))
		if err != nil {
			m.Log.WithFields(logrus.Fields{
				"ctx_id": ctxID,
				"err":    err,
			}).Error("Unable to check an email in spamd")
		}
		if spamReply != nil && spamReply.Code == spamc.EX_OK {
			if spam, ok := spamReply.Vars["isSpam"]; ok && spam.(bool) {
				isSpam = true
			}
		}

		// Trim the spaces from the input
		conn.Envelope.Data = bytes.TrimSpace(conn.Envelope.Data)

		// First run the analysis algorithm to generate an email description
		node := &models.EmailNode{}
		if err := analyzeEmail(node, conn.Envelope.Data); err != nil {
			m.Error(conn, err)
			return
		}

		// Calculate message ID
		messageID := node.Headers.Get("Message-ID")
		x1i := strings.Index(messageID, "<")
		if x1i != -1 {
			x2i := strings.Index(messageID[x1i+1:], ">")
			if x2i != -1 {
				messageID = messageID[x1i+1 : x1i+x2i+1]
			}
		}
		if messageID == "" {
			messageID = uniuri.NewLen(uniuri.UUIDLen) + "@invalid-incoming.pgp.st"
		}

		// Generate the members field
		fromHeader, err := mail.ParseAddress(node.Headers.Get("From"))
		if err != nil {
			m.Error(conn, err)
			return
		}
		members := []string{fromHeader.Address}

		toHeader, err := node.Headers.AddressList("From")
		if err != nil {
			m.Error(conn, err)
			return
		}
		for _, to := range toHeader {
			members = append(members, to.Address)
		}

		if x := node.Headers.Get("CC"); x != "" {
			ccHeader, err := node.Headers.AddressList("CC")
			if err != nil {
				m.Error(conn, err)
				return
			}
			for _, cc := range ccHeader {
				members = append(members, cc.Address)
			}
		}

		// Get the recipients list from the scope
		recipients := conn.Environment["recipients"].([]recipient)
		for _, recipient := range recipients {
			// Parse the found key
			keyring, err := openpgp.ReadKeyRing(bytes.NewReader(recipient.Key.Body))
			if err != nil {
				m.Error(conn, err)
				return
			}

			// Prepare a new email object
			email := &models.Email{
				ID:           uniuri.NewLen(uniuri.UUIDLen),
				DateCreated:  time.Now(),
				DateModified: time.Now(),
				Owner:        recipient.Account.ID,
				MessageID:    messageID,
				Status:       "received",
			}

			// Generate a new key
			key := make([]byte, 32)
			if _, err := io.ReadFull(rand.Reader, key); err != nil {
				m.Error(conn, err)
				return
			}
			akey := [32]byte{}
			copy(akey[:], key)

			// Generate a new nonce
			nonce := make([]byte, 8)
			if _, err = rand.Read(nonce); err != nil {
				m.Error(conn, err)
				return
			}

			// Set up a new stream
			stream, err := chacha20.New(key, nonce)
			if err != nil {
				m.Error(conn, err)
				return
			}

			// Prepare output
			ciphertext := make(
				[]byte,
				len(conn.Envelope.Data)+
					chacha20.NonceSize+
					(len(conn.Envelope.Data)/1024+1)*16,
			)

			// First write the nonce
			copy(ciphertext[:chacha20.NonceSize], nonce)

			// Then write the authenticated encrypted stream
			oi := chacha20.NonceSize
			for i := 0; i < len(conn.Envelope.Data); i += 1024 {
				// Get the chunk
				max := i + 1024
				if max > len(conn.Envelope.Data) {
					max = len(conn.Envelope.Data)
				}
				chunk := conn.Envelope.Data[i:max]

				// Encrypt the block
				stream.XORKeyStream(ciphertext[oi:oi+len(chunk)], chunk)

				// Authenticate it
				var out [16]byte
				poly1305.Sum(&out, ciphertext[oi:oi+len(chunk)], &akey)

				// Write it into the result
				copy(ciphertext[oi+len(chunk):oi+len(chunk)+poly1305.TagSize], out[:])

				// Increase the counter
				oi += 1024 + poly1305.TagSize
			}

			// Save it to the body
			email.Body = ciphertext

			// Create a manifest and encrypt it
			manifest, err := json.Marshal(models.Manifest{
				Key:         key,
				Nonce:       nonce,
				Description: node,
			})
			if err != nil {
				m.Error(conn, err)
				return
			}
			encryptedManifest, err := utils.PGPEncrypt(manifest, keyring)
			if err != nil {
				m.Error(conn, err)
				return
			}
			email.Manifest = encryptedManifest

			// Match the thread
			var thread *models.Thread

			// Get the References header
			var references string
			if x := node.Headers.Get("In-Reply-To"); x != "" {
				references = x
			} else if x := node.Headers.Get("References"); x != "" {
				references = x
			}

			// Match by Message-ID
			if references != "" {
				// As specified in http://www.jwz.org/doc/threading.html, first thing <> is the msg id
				// We support both <message-id> and message-id format.
				x1i := strings.Index(references, "<")
				if x1i != -1 {
					x2i := strings.Index(references[x1i+1:], ">")
					if x2i != -1 {
						references = references[x1i+1 : x1i+x2i+1]
					}
				}

				// Look up the message ID in the database
				cursor, err := r.Table("emails").GetAllByIndex("messageIDOwner", []interface{}{
					references,
					recipient.Account.ID,
				}).CoerceTo("array").Do(func(emails r.Term) r.Term {
					return r.Branch(
						emails.Count().Eq(1),
						r.Table("threads").Get(emails.Nth(0).Field("thread")).Default(map[string]interface{}{}),
						map[string]interface{}{},
					)
				}).Run(m.Rethink)
				if err != nil {
					m.Error(conn, err)
					return
				}
				defer cursor.Close()
				if err := cursor.One(&thread); err != nil {
					m.Error(conn, err)
					return
				}

				// Check if we've found it, clear it if it's invalid
				if thread.ID == "" {
					thread = nil
				}
			}

			// We can't match it by subject, so proceed to create a new thread
			if thread == nil {
				var secure string
				if findEncrypted(node) {
					secure = "all"
				} else {
					secure = "none"
				}

				labels := []string{recipient.Labels.Inbox}
				if isSpam {
					labels = append(labels, recipient.Labels.Spam)
				}

				thread = &models.Thread{
					ID:           uniuri.NewLen(uniuri.UUIDLen),
					DateCreated:  time.Now(),
					DateModified: time.Now(),
					Owner:        recipient.Account.ID,
					Labels:       labels,
					Members:      members,
					Secure:       secure,
				}

				if err := r.Table("threads").Insert(thread).Exec(m.Rethink); err != nil {
					m.Error(conn, err)
					return
				}
			} else {
				// Modify the existing thread
				foundInbox := false
				foundSpam := false
				for _, label := range thread.Labels {
					if label == recipient.Labels.Inbox {
						foundInbox = true
					}

					if isSpam {
						if label == recipient.Labels.Spam {
							foundSpam = true
						}

						if foundInbox && foundSpam {
							break
						}
					} else {
						if foundInbox {
							break
						}
					}
				}

				// Append to thread.Labels
				if !foundInbox {
					thread.Labels = append(thread.Labels, recipient.Labels.Inbox)
				}
				if !foundSpam && isSpam {
					thread.Labels = append(thread.Labels, recipient.Labels.Spam)
				}

				// Members update
				membersHash := map[string]struct{}{}
				for _, member := range thread.Members {
					membersHash[member] = struct{}{}
				}
				for _, member := range members {
					if _, ok := membersHash[member]; !ok {
						membersHash[member] = struct{}{}
					}
				}
				thread.Members = []string{}
				for member := range membersHash {
					thread.Members = append(thread.Members, member)
				}

				update := map[string]interface{}{
					"date_modified": time.Now(),
					"is_read":       false,
					"labels":        thread.Labels,
					"members":       thread.Members,
				}

				secure := findEncrypted(node)
				if (thread.Secure == "all" && !secure) ||
					(thread.Secure == "none" && secure) {
					thread.Secure = "some"
				}

				if err := r.Table("threads").Get(thread.ID).Update(update).Exec(m.Rethink); err != nil {
					m.Error(conn, err)
					return
				}
			}

			email.Thread = thread.ID
			if err := r.Table("emails").Insert(email).Exec(m.Rethink); err != nil {
				m.Error(conn, err)
				return
			}

			m.Log.WithFields(logrus.Fields{
				"address": recipient.Address.ID,
				"account": recipient.Account.MainAddress,
			}).Info("Email received")
		}

		next(conn)
	}
}

func (m *Mailer) Error(conn *smtpd.Connection, err error) {
	_, file, line, _ := runtime.Caller(1)
	m.Log.WithFields(logrus.Fields{
		"file": filepath.Base(file),
		"line": line,
	}).Error(err)
	conn.Error(err)
}
