package api

import (
	"bytes"
	"encoding/hex"
	"time"

	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
	"github.com/pgpst/pgpst/internal/github.com/gin-gonic/gin"
	"github.com/pgpst/pgpst/internal/golang.org/x/crypto/openpgp"

	"github.com/pgpst/pgpst/pkg/models"
)

func (a *API) createKey(c *gin.Context) {
	// Get token and account info from the context
	var (
		account = c.MustGet("account").(*models.Account)
		token   = c.MustGet("token").(*models.Token)
	)

	// Check the scope
	if !models.InScope(token.Scope, []string{"keys"}) {
		c.JSON(403, &gin.H{
			"code":  0,
			"error": "Your token has insufficient scope",
		})
		return
	}

	// Decode the input
	var input struct {
		Body []byte `json:"body"`
	}
	if err := c.Bind(&input); err != nil {
		c.JSON(422, &gin.H{
			"code":    0,
			"message": err.Error(),
		})
		return
	}

	// Validate the input
	if input.Body == nil || len(input.Body) == 0 {
		c.JSON(422, &gin.H{
			"code":    0,
			"message": "No key body in the input",
		})
		return
	}

	// Parse the key
	keyring, err := openpgp.ReadKeyRing(bytes.NewReader(input.Body))
	if err != nil {
		c.JSON(422, &gin.H{
			"code":    0,
			"message": "Invalid key format",
		})
		return
	}
	publicKey := keyring[0].PrimaryKey

	// Parse the identities
	identities := []*models.Identity{}
	for _, identity := range keyring[0].Identities {
		id := &models.Identity{
			Name: identity.Name,
		}

		if identity.SelfSignature != nil {
			sig := identity.SelfSignature
			id.SelfSignature = &models.Signature{
				Type:                 uint8(sig.SigType),
				Algorithm:            uint8(sig.PubKeyAlgo),
				Hash:                 uint(sig.Hash),
				CreationTime:         sig.CreationTime,
				SigLifetimeSecs:      sig.SigLifetimeSecs,
				KeyLifetimeSecs:      sig.KeyLifetimeSecs,
				IssuerKeyID:          sig.IssuerKeyId,
				IsPrimaryID:          sig.IsPrimaryId,
				RevocationReason:     sig.RevocationReason,
				RevocationReasonText: sig.RevocationReasonText,
			}
		}

		if identity.Signatures != nil {
			id.Signatures = []*models.Signature{}
			for _, sig := range identity.Signatures {
				id.Signatures = append(id.Signatures, &models.Signature{
					Type:                 uint8(sig.SigType),
					Algorithm:            uint8(sig.PubKeyAlgo),
					Hash:                 uint(sig.Hash),
					CreationTime:         sig.CreationTime,
					SigLifetimeSecs:      sig.SigLifetimeSecs,
					KeyLifetimeSecs:      sig.KeyLifetimeSecs,
					IssuerKeyID:          sig.IssuerKeyId,
					IsPrimaryID:          sig.IsPrimaryId,
					RevocationReason:     sig.RevocationReason,
					RevocationReasonText: sig.RevocationReasonText,
				})
			}
		}

		identities = append(identities, id)
	}

	// Acquire key's length
	length, err := publicKey.BitLength()
	if err != nil {
		c.JSON(422, &gin.H{
			"code":    0,
			"message": "Couldn't calculate bit length",
		})
		return
	}

	// Generate a new key struct
	key := &models.Key{
		ID:           hex.EncodeToString(publicKey.Fingerprint[:]),
		DateCreated:  time.Now(),
		DateModified: time.Now(),
		Owner:        account.ID,

		Algorithm:        uint8(publicKey.PubKeyAlgo),
		Length:           length,
		Body:             input.Body,
		KeyID:            publicKey.KeyId,
		KeyIDString:      publicKey.KeyIdString(),
		KeyIDShortString: publicKey.KeyIdShortString(),

		Identities: identities,
	}

	// Insert it into database
	if err := r.Table("keys").Insert(key).Exec(a.Rethink); err != nil {
		c.JSON(500, &gin.H{
			"code":    0,
			"message": err.Error(),
		})
		return
	}

	c.JSON(201, key)
}

func (a *API) readKey(c *gin.Context) {
	// Resolve the ID param
	id := c.Param("id")
	if len(id) != 40 {
		// Look up by email
		cursor, err := r.Table("addresses").Get(id).Default(map[string]interface{}{}).Run(a.Rethink)
		if err != nil {
			c.JSON(500, &gin.H{
				"code":    0,
				"message": err.Error(),
			})
			return
		}
		defer cursor.Close()
		var address *models.Address
		if err := cursor.One(&address); err != nil {
			c.JSON(500, &gin.H{
				"code":    0,
				"message": err.Error(),
			})
			return
		}

		// Check if we've found that address
		if address.ID == "" {
			c.JSON(404, &gin.H{
				"code":    0,
				"message": "Address not found",
			})
			return
		}

		// Set the ID accordingly
		id = address.PublicKey
	}

	// Fetch the key from database
	cursor, err := r.Table("keys").Get(id).Default(map[string]interface{}{}).Run(a.Rethink)
	if err != nil {
		c.JSON(500, &gin.H{
			"code":    0,
			"message": err.Error(),
		})
		return
	}
	defer cursor.Close()
	var key *models.Key
	if err := cursor.One(&key); err != nil {
		c.JSON(500, &gin.H{
			"code":    0,
			"message": err.Error(),
		})
		return
	}

	// Ensure that it exists, write the response
	if key.ID == "" {
		c.JSON(404, &gin.H{
			"code":    0,
			"message": "Key not found",
		})
		return
	}
	c.JSON(200, key)
}
