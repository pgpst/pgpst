package smtpd_test

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"net/textproto"
	"strings"
	"testing"
	"time"

	"github.com/pgpst/pgpst/internal/github.com/pgpst/smtpd"
)

var localhostCert = []byte(`-----BEGIN CERTIFICATE-----
MIIBkzCCAT+gAwIBAgIQf4LO8+QzcbXRHJUo6MvX7zALBgkqhkiG9w0BAQswEjEQ
MA4GA1UEChMHQWNtZSBDbzAeFw03MDAxMDEwMDAwMDBaFw04MTA1MjkxNjAwMDBa
MBIxEDAOBgNVBAoTB0FjbWUgQ28wXDANBgkqhkiG9w0BAQEFAANLADBIAkEAx2Uj
2nl0ESnMMrdUOwQnpnIPQzQBX9MIYT87VxhHzImOukWcq5DrmN1ZB//diyrgiCLv
D0udX3YXNHMn1Ki8awIDAQABo3MwcTAOBgNVHQ8BAf8EBAMCAKQwEwYDVR0lBAww
CgYIKwYBBQUHAwEwDwYDVR0TAQH/BAUwAwEB/zA5BgNVHREEMjAwggtleGFtcGxl
LmNvbYIJbG9jYWxob3N0hwR/AAABhxAAAAAAAAAAAAAAAAAAAAABMAsGCSqGSIb3
DQEBCwNBAGcaB2Il0TIXFcJOdOLGPa6F8qZH1ZHBtVlCBnaJn4vZJGzID+V36Gn0
hA1AYfGAaF0c43oQofvv+XqQlTe4a+M=
-----END CERTIFICATE-----`)

var localhostKey = []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIBPAIBAAJBAMdlI9p5dBEpzDK3VDsEJ6ZyD0M0AV/TCGE/O1cYR8yJjrpFnKuQ
65jdWQf/3Ysq4Igi7w9LnV92FzRzJ9SovGsCAwEAAQJAVaFw2VWJbAmIQUuMJ+Ar
6wZW2aSO5okpsyHFqSyrQQIcAj/QOq8P83F8J10IreFWNlBlywJU9c7IlJtn/lqq
AQIhAOxHXOxrKPxqTIdIcNnWye/HRQ+5VD54QQr1+M77+bEBAiEA2AmsNNqj2fKj
j2xk+4vnBSY0vrb4q/O3WZ46oorawWsCIQDWdpfzx/i11E6OZMR6FinJSNh4w0Gi
SkjPiCBE0BX+AQIhAI/TiLk7YmBkQG3ovSYW0vvDntPlXpKj08ovJFw4U0D3AiEA
lGjGna4oaauI0CWI6pG0wg4zklTnrDWK7w9h/S/T4e0=
-----END RSA PRIVATE KEY-----`)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func cmd(c *textproto.Conn, expectedCode int, format string, args ...interface{}) error {
	id, err := c.Cmd(format, args...)
	if err != nil {
		return err
	}

	c.StartResponse(id)
	_, _, err = c.ReadResponse(expectedCode)
	c.EndResponse(id)

	return err
}

func runserver(t *testing.T, server *smtpd.Server) (addr string, closer func()) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Listen failed: %v", err)
	}

	go func() {
		server.Serve(ln)
	}()

	done := make(chan bool)

	go func() {
		<-done
		ln.Close()
	}()

	return ln.Addr().String(), func() {
		done <- true
	}
}

func runsslserver(t *testing.T, server *smtpd.Server) (addr string, closer func()) {
	cert, err := tls.X509KeyPair(localhostCert, localhostKey)
	if err != nil {
		t.Fatalf("Cert load failed: %v", err)
	}

	server.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	return runserver(t, server)
}

func TestSMTP(t *testing.T) {
	addr, closer := runserver(t, &smtpd.Server{})
	defer closer()

	c, err := smtp.Dial(addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}

	if err := c.Hello("localhost"); err != nil {
		t.Fatalf("HELO failed: %v", err)
	}

	if supported, _ := c.Extension("8BITMIME"); !supported {
		t.Fatal("8BITMIME not supported")
	}

	if supported, _ := c.Extension("STARTTLS"); supported {
		t.Fatal("STARTTLS supported")
	}

	if err := c.Mail("sender@example.org"); err != nil {
		t.Fatalf("Mail failed: %v", err)
	}

	if err := c.Rcpt("recipient@example.net"); err != nil {
		t.Fatalf("Rcpt failed: %v", err)
	}

	if err := c.Rcpt("recipient2@example.net"); err != nil {
		t.Fatalf("Rcpt2 failed: %v", err)
	}

	wc, err := c.Data()
	if err != nil {
		t.Fatalf("Data failed: %v", err)
	}

	_, err = fmt.Fprintf(wc, "This is the email body")
	if err != nil {
		t.Fatalf("Data body failed: %v", err)
	}

	err = wc.Close()
	if err != nil {
		t.Fatalf("Data close failed: %v", err)
	}

	if err := c.Reset(); err != nil {
		t.Fatalf("Reset failed: %v", err)
	}

	if err := c.Verify("foobar@example.net"); err == nil {
		t.Fatal("Unexpected support for VRFY")
	}

	if err := cmd(c.Text, 250, "NOOP"); err != nil {
		t.Fatalf("NOOP failed: %v", err)
	}

	if err := c.Quit(); err != nil {
		t.Fatalf("Quit failed: %v", err)
	}
}

func TestListenAndServe(t *testing.T) {
	addr, closer := runserver(t, &smtpd.Server{})
	closer()
	// Wait here for Windows to release the port.
	time.Sleep(100 * time.Millisecond)

	server := &smtpd.Server{}

	dial := make(chan struct{})
	go func() {
		close(dial)
		err := server.ListenAndServe(addr)
		if err != nil {
			t.Error(err)
		}
	}()

	<-dial
	time.Sleep(100 * time.Millisecond)

	c, err := smtp.Dial(addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}

	if err := c.Quit(); err != nil {
		t.Fatalf("Quit failed: %v", err)
	}
}

func TestSTARTTLS(t *testing.T) {
	addr, closer := runsslserver(t, &smtpd.Server{
		ForceTLS: true,
	})

	defer closer()

	c, err := smtp.Dial(addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}

	if err := c.Mail("sender@example.org"); err == nil {
		t.Fatal("Mail worked before TLS with ForceTLS")
	}

	if err := cmd(c.Text, 220, "STARTTLS"); err != nil {
		t.Fatalf("STARTTLS failed: %v", err)
	}

	if err := cmd(c.Text, 250, "foobar"); err == nil {
		t.Fatal("STARTTLS didn't fail with invalid handshake")
	}

	if err := c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		t.Fatalf("STARTTLS failed: %v", err)
	}

	if err := c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err == nil {
		t.Fatal("STARTTLS worked twice")
	}

	if err := c.Mail("sender@example.org"); err != nil {
		t.Fatalf("Mail failed: %v", err)
	}

	if err := c.Rcpt("recipient@example.net"); err != nil {
		t.Fatalf("Rcpt failed: %v", err)
	}

	if err := c.Rcpt("recipient2@example.net"); err != nil {
		t.Fatalf("Rcpt2 failed: %v", err)
	}

	wc, err := c.Data()
	if err != nil {
		t.Fatalf("Data failed: %v", err)
	}

	_, err = fmt.Fprintf(wc, "This is the email body")
	if err != nil {
		t.Fatalf("Data body failed: %v", err)
	}

	err = wc.Close()
	if err != nil {
		t.Fatalf("Data close failed: %v", err)
	}

	if err := c.Quit(); err != nil {
		t.Fatalf("Quit failed: %v", err)
	}
}

func TestSenderCheck(t *testing.T) {
	addr, closer := runserver(t, &smtpd.Server{
		SenderChain: []smtpd.Sender{
			smtpd.SenderFunc(func(x func(conn *smtpd.Connection)) func(*smtpd.Connection) {
				return func(conn *smtpd.Connection) {
					conn.Error(errors.New("Random error"))
				}
			}),
		},
	})
	defer closer()

	c, err := smtp.Dial(addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}

	if err := c.Mail("sender@example.org"); err == nil {
		t.Fatal("Unexpected MAIL success")
	}
}

func TestRecipientCheck(t *testing.T) {
	addr, closer := runserver(t, &smtpd.Server{
		RecipientChain: []smtpd.Recipient{
			smtpd.RecipientFunc(func(x func(conn *smtpd.Connection)) func(*smtpd.Connection) {
				return func(conn *smtpd.Connection) {
					conn.Error(errors.New("Random error"))
				}
			}),
		},
	})
	defer closer()

	c, err := smtp.Dial(addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}

	if err := c.Mail("sender@example.org"); err != nil {
		t.Fatalf("Mail failed: %v", err)
	}

	if err := c.Rcpt("recipient@example.net"); err == nil {
		t.Fatal("Unexpected RCPT success")
	}
}

func TestMaxMessageSize(t *testing.T) {
	addr, closer := runserver(t, &smtpd.Server{
		MaxMessageSize: 5,
	})
	defer closer()

	c, err := smtp.Dial(addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}

	if err := c.Mail("sender@example.org"); err != nil {
		t.Fatalf("MAIL failed: %v", err)
	}

	if err := c.Rcpt("recipient@example.net"); err != nil {
		t.Fatalf("RCPT failed: %v", err)
	}

	wc, err := c.Data()
	if err != nil {
		t.Fatalf("Data failed: %v", err)
	}

	_, err = fmt.Fprintf(wc, "This is the email body")
	if err != nil {
		t.Fatalf("Data body failed: %v", err)
	}

	err = wc.Close()
	if err == nil {
		t.Fatal("Allowed message larger than 5 bytes to pass.")
	}

	if err := c.Quit(); err != nil {
		t.Fatalf("QUIT failed: %v", err)
	}
}

func TestHandlerAndWrapper(t *testing.T) {
	addr, closer := runserver(t, &smtpd.Server{
		WrapperChain: []smtpd.Wrapper{
			smtpd.WrapperFunc(func(next func()) func() {
				return func() {
					log.Print("Hello from inside the wrapper!")
					next()
				}
			}),
		},
		DeliveryChain: []smtpd.Delivery{
			smtpd.DeliveryFunc(func(next func(conn *smtpd.Connection)) func(conn *smtpd.Connection) {
				return func(conn *smtpd.Connection) {
					if conn.Envelope.Sender != "sender@example.org" {
						t.Fatalf("Unknown sender: %v", conn.Envelope.Sender)
					}

					if len(conn.Envelope.Recipients) != 1 {
						t.Fatalf("Too many recipients: %d", len(conn.Envelope.Recipients))
					}

					if conn.Envelope.Recipients[0] != "recipient@example.net" {
						t.Fatalf("Unknown recipient: %v", conn.Envelope.Recipients[0])
					}

					if string(conn.Envelope.Data) != "This is the email body\n" {
						t.Fatalf("Wrong message body: %v", string(conn.Envelope.Data))
					}

					next(conn)
				}
			}),
		},
	})
	defer closer()

	c, err := smtp.Dial(addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}

	if err := c.Mail("sender@example.org"); err != nil {
		t.Fatalf("MAIL failed: %v", err)
	}

	if err := c.Rcpt("recipient@example.net"); err != nil {
		t.Fatalf("RCPT failed: %v", err)
	}

	wc, err := c.Data()
	if err != nil {
		t.Fatalf("Data failed: %v", err)
	}

	_, err = fmt.Fprintf(wc, "This is the email body")
	if err != nil {
		t.Fatalf("Data body failed: %v", err)
	}

	err = wc.Close()
	if err != nil {
		t.Fatalf("Data close failed: %v", err)
	}

	if err := c.Quit(); err != nil {
		t.Fatalf("QUIT failed: %v", err)
	}
}

func TestRejectHandler(t *testing.T) {
	addr, closer := runserver(t, &smtpd.Server{
		DeliveryChain: []smtpd.Delivery{
			smtpd.DeliveryFunc(func(next func(conn *smtpd.Connection)) func(conn *smtpd.Connection) {
				return func(conn *smtpd.Connection) {
					conn.Error(smtpd.ErrServerError)
				}
			}),
		},
	})
	defer closer()

	c, err := smtp.Dial(addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}

	if err := c.Mail("sender@example.org"); err != nil {
		t.Fatalf("MAIL failed: %v", err)
	}

	if err := c.Rcpt("recipient@example.net"); err != nil {
		t.Fatalf("RCPT failed: %v", err)
	}

	wc, err := c.Data()
	if err != nil {
		t.Fatalf("Data failed: %v", err)
	}

	_, err = fmt.Fprintf(wc, "This is the email body")
	if err != nil {
		t.Fatalf("Data body failed: %v", err)
	}

	err = wc.Close()
	if err == nil {
		t.Fatal("Unexpected accept of data")
	}

	if err := c.Quit(); err != nil {
		t.Fatalf("QUIT failed: %v", err)
	}
}

func TestMaxConnections(t *testing.T) {
	addr, closer := runserver(t, &smtpd.Server{
		MaxConnections: 1,
	})
	defer closer()

	c1, err := smtp.Dial(addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer c1.Close()

	_, err = smtp.Dial(addr)
	if err == nil {
		t.Fatal("Dial succeeded despite MaxConnections = 1")
	}
}

func TestNoMaxConnections(t *testing.T) {
	addr, closer := runserver(t, &smtpd.Server{
		MaxConnections: -1,
	})
	defer closer()

	c1, err := smtp.Dial(addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}

	c1.Close()
}

func TestMisconfiguredTLS(t *testing.T) {
	server := &smtpd.Server{
		ForceTLS: true,
	}

	if server.ListenAndServe(":123123123").Error() != "Cannot use ForceTLS with no TLSConfig" {
		t.Fatal("Unexpected configuration check success")
	}

	if err := server.Serve(nil); err != nil && err.Error() != "Cannot use ForceTLS with no TLSConfig" {
		t.Fatalf("Unexpected serving error; %s", err)
	}
}

func TestInvalidListenAddress(t *testing.T) {
	server := &smtpd.Server{}

	err := server.ListenAndServe("pleasedontwork")
	if err.Error() != "listen tcp: missing port in address pleasedontwork" {
		t.Fatalf("Unexpected port binding: %s", err)
	}
}

type tmpError struct{}

func (t *tmpError) Error() string {
	return "dumb error"
}
func (t *tmpError) Timeout() bool {
	return true
}
func (t *tmpError) Temporary() bool {
	return true
}

type tmpListener struct {
	ct int
}

func (t *tmpListener) Accept() (net.Conn, error) {
	if t.ct == 0 {
		t.ct = 1
		return nil, &tmpError{}
	}
	return nil, errors.New("not temporary error")
}
func (t *tmpListener) Close() error {
	return nil
}
func (t *tmpListener) Addr() net.Addr {
	x, _ := net.InterfaceAddrs()
	return x[0]
}

func TestTemporaryError(t *testing.T) {
	server := &smtpd.Server{}
	err := server.Serve(&tmpListener{})
	if err.Error() != "not temporary error" {
		t.Fatalf("Unexpected error; %s", err)
	}
}

func TestMaxRecipients(t *testing.T) {
	addr, closer := runserver(t, &smtpd.Server{
		MaxRecipients: 1,
	})
	defer closer()

	c, err := smtp.Dial(addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}

	if err := c.Mail("sender@example.org"); err != nil {
		t.Fatalf("MAIL failed: %v", err)
	}

	if err := c.Rcpt("recipient@example.net"); err != nil {
		t.Fatalf("RCPT failed: %v", err)
	}

	if err := c.Rcpt("recipient@example.net"); err == nil {
		t.Fatal("RCPT succeeded despite MaxRecipients = 1")
	}

	if err := c.Quit(); err != nil {
		t.Fatalf("QUIT failed: %v", err)
	}
}

func TestInvalidHelo(t *testing.T) {
	addr, closer := runserver(t, &smtpd.Server{})
	defer closer()

	c, err := smtp.Dial(addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}

	if err := c.Hello(""); err == nil {
		t.Fatal("Unexpected HELO success")
	}
}

func TestInvalidSender(t *testing.T) {
	addr, closer := runserver(t, &smtpd.Server{})
	defer closer()

	c, err := smtp.Dial(addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}

	if err := c.Mail("invalid@@example.org"); err == nil {
		t.Fatal("Unexpected MAIL success")
	}
}

func TestInvalidRecipient(t *testing.T) {
	addr, closer := runserver(t, &smtpd.Server{})
	defer closer()

	c, err := smtp.Dial(addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}

	if err := c.Mail("sender@example.org"); err != nil {
		t.Fatalf("Mail failed: %v", err)
	}

	if err := cmd(c.Text, 250, "MAIL"); err == nil {
		t.Fatal("Unexpected MAIL success")
	}

	if err := cmd(c.Text, 250, "MAIL ayylmao"); err == nil {
		t.Fatal("Unexpected MAIL success")
	}

	if err := c.Rcpt("invalid@@example.org"); err == nil {
		t.Fatal("Unexpected RCPT success")
	}

	if err := cmd(c.Text, 250, "RCPT"); err == nil {
		t.Fatal("Unexpected RCPT success")
	}

	if err := cmd(c.Text, 250, "RCPT ayylmao"); err == nil {
		t.Fatal("Unexpected RCPT success")
	}

	if err := c.Rcpt(""); err == nil {
		t.Fatal("Unexpected RCPT success")
	}
}

func TestRCPTbeforeMAIL(t *testing.T) {
	addr, closer := runserver(t, &smtpd.Server{})
	defer closer()

	c, err := smtp.Dial(addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}

	if err := c.Rcpt("recipient@example.net"); err == nil {
		t.Fatal("Unexpected RCPT success")
	}
}

func TestDATAbeforeRCPT(t *testing.T) {
	addr, closer := runserver(t, &smtpd.Server{})
	defer closer()

	c, err := smtp.Dial(addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}

	if err := c.Mail("sender@example.org"); err != nil {
		t.Fatalf("MAIL failed: %v", err)
	}

	if _, err := c.Data(); err == nil {
		t.Fatal("Data accepted despite no recipients")
	}

	if err := c.Quit(); err != nil {
		t.Fatalf("QUIT failed: %v", err)
	}
}

func TestInterruptedDATA(t *testing.T) {
	addr, closer := runserver(t, &smtpd.Server{
		DeliveryChain: []smtpd.Delivery{
			smtpd.DeliveryFunc(func(next func(conn *smtpd.Connection)) func(conn *smtpd.Connection) {
				return func(conn *smtpd.Connection) {
					t.Fatal("Accepted DATA despite disconnection")
					next(conn)
				}
			}),
		},
	})
	defer closer()

	c, err := smtp.Dial(addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer c.Close()

	if err := c.Mail("sender@example.org"); err != nil {
		t.Fatalf("MAIL failed: %v", err)
	}

	if err := c.Rcpt("recipient@example.net"); err != nil {
		t.Fatalf("RCPT failed: %v", err)
	}

	wc, err := c.Data()
	if err != nil {
		t.Fatalf("Data failed: %v", err)
	}

	_, err = fmt.Fprintf(wc, "This is the email body")
	if err != nil {
		t.Fatalf("Data body failed: %v", err)
	}
}

func TestTimeoutClose(t *testing.T) {
	addr, closer := runserver(t, &smtpd.Server{
		MaxConnections: 1,
		ReadTimeout:    time.Second,
		WriteTimeout:   time.Second,
	})
	defer closer()

	c1, err := smtp.Dial(addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}

	time.Sleep(time.Second * 2)

	c2, err := smtp.Dial(addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer c2.Close()

	if err := c1.Mail("sender@example.org"); err == nil {
		t.Fatal("MAIL succeeded despite being timed out.")
	}

	if err := c2.Mail("sender@example.org"); err != nil {
		t.Fatalf("MAIL failed: %v", err)
	}

	if err := c2.Quit(); err != nil {
		t.Fatalf("Quit failed: %v", err)
	}
}

func TestTLSTimeout(t *testing.T) {
	addr, closer := runsslserver(t, &smtpd.Server{
		ReadTimeout:  time.Second * 2,
		WriteTimeout: time.Second * 2,
	})
	defer closer()

	c, err := smtp.Dial(addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}

	if err := c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		t.Fatalf("STARTTLS failed: %v", err)
	}

	time.Sleep(time.Second)

	if err := c.Mail("sender@example.org"); err != nil {
		t.Fatalf("MAIL failed: %v", err)
	}

	time.Sleep(time.Second)

	if err := c.Rcpt("recipient@example.net"); err != nil {
		t.Fatalf("RCPT failed: %v", err)
	}

	time.Sleep(time.Second)

	if err := c.Rcpt("recipient@example.net"); err != nil {
		t.Fatalf("RCPT failed: %v", err)
	}

	time.Sleep(time.Second)

	if err := c.Quit(); err != nil {
		t.Fatalf("Quit failed: %v", err)
	}
}

func TestLongLine(t *testing.T) {
	addr, closer := runserver(t, &smtpd.Server{})
	defer closer()

	c, err := smtp.Dial(addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}

	if err := c.Mail(fmt.Sprintf("%s@example.org", strings.Repeat("x", 65*1024))); err == nil {
		t.Fatalf("MAIL failed: %v", err)
	}

	if err := c.Quit(); err != nil {
		t.Fatalf("Quit failed: %v", err)
	}
}

func TestEnvelopeReceived(t *testing.T) {
	addr, closer := runsslserver(t, &smtpd.Server{
		Hostname: "foobar.example.net",
		DeliveryChain: []smtpd.Delivery{
			smtpd.DeliveryFunc(func(next func(conn *smtpd.Connection)) func(conn *smtpd.Connection) {
				return func(conn *smtpd.Connection) {
					conn.Envelope.AddReceivedLine(conn)
					if !bytes.HasPrefix(conn.Envelope.Data, []byte("Received: from localhost [127.0.0.1] by foobar.example.net with ESMTP;")) {
						t.Fatal("Wrong received line.")
					}
					next(conn)
				}
			}),
		},
		ForceTLS: true,
	})
	defer closer()

	c, err := smtp.Dial(addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}

	if err := c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		t.Fatalf("STARTTLS failed: %v", err)
	}

	if err := c.Mail("sender@example.org"); err != nil {
		t.Fatalf("MAIL failed: %v", err)
	}

	if err := c.Rcpt("recipient@example.net"); err != nil {
		t.Fatalf("RCPT failed: %v", err)
	}

	wc, err := c.Data()
	if err != nil {
		t.Fatalf("Data failed: %v", err)
	}

	_, err = fmt.Fprintf(wc, "This is the email body")
	if err != nil {
		t.Fatalf("Data body failed: %v", err)
	}

	err = wc.Close()
	if err != nil {
		t.Fatalf("Data close failed: %v", err)
	}

	if err := c.Quit(); err != nil {
		t.Fatalf("QUIT failed: %v", err)
	}
}

func TestHELO(t *testing.T) {
	addr, closer := runserver(t, &smtpd.Server{})
	defer closer()

	c, err := smtp.Dial(addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}

	if err := cmd(c.Text, 502, "MAIL FROM:<christian@technobabble.dk>"); err != nil {
		t.Fatalf("MAIL didn't fail: %v", err)
	}

	if err := cmd(c.Text, 250, "HELO localhost"); err != nil {
		t.Fatalf("HELO failed: %v", err)
	}

	if err := cmd(c.Text, 502, "MAIL FROM:christian@technobabble.dk"); err != nil {
		t.Fatalf("MAIL didn't fail: %v", err)
	}

	if err := cmd(c.Text, 250, "HELO localhost"); err != nil {
		t.Fatalf("HELO failed: %v", err)
	}

	if err := c.Quit(); err != nil {
		t.Fatalf("Quit failed: %v", err)
	}
}

func TestErrors(t *testing.T) {
	cert, err := tls.X509KeyPair(localhostCert, localhostKey)
	if err != nil {
		t.Fatalf("Cert load failed: %v", err)
	}

	server := &smtpd.Server{}

	addr, closer := runserver(t, server)
	defer closer()

	c, err := smtp.Dial(addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}

	if err := c.Hello("localhost"); err != nil {
		t.Fatalf("HELO failed: %v", err)
	}

	if err := cmd(c.Text, 502, "MAIL FROM:christian@technobabble.dk"); err != nil {
		t.Fatalf("MAIL didn't fail: %v", err)
	}

	if err := c.Mail("sender@example.org"); err != nil {
		t.Fatalf("MAIL failed: %v", err)
	}

	if err := c.Mail("sender@example.org"); err == nil {
		t.Fatal("Duplicate MAIL didn't fail")
	}

	if err := cmd(c.Text, 502, "STARTTLS"); err != nil {
		t.Fatalf("STARTTLS didn't fail: %v", err)
	}

	server.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	if err := c.StartTLS(&tls.Config{InsecureSkipVerify: true}); err != nil {
		t.Fatalf("STARTTLS failed: %v", err)
	}

	if err := c.Quit(); err != nil {
		t.Fatalf("Quit failed: %v", err)
	}
}
