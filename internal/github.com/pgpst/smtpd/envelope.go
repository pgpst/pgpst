package smtpd

import (
	"bytes"
	"crypto/tls"
	"strconv"
	"strings"
	"time"
)

type Envelope struct {
	Sender     string
	Recipients []string
	Data       []byte
}

var tlsVersions = map[uint16]string{
	tls.VersionSSL30: "SSL3.0",
	tls.VersionTLS10: "TLS1.0",
	tls.VersionTLS11: "TLS1.1",
	tls.VersionTLS12: "TLS1.2",
}

func (e *Envelope) AddReceivedLine(c *Connection) {
	var buf bytes.Buffer

	buf.WriteString("Received: from ")
	buf.WriteString(c.HeloName)
	buf.WriteString(" [")
	buf.WriteString(strings.Split(c.Addr.String(), ":")[0])
	buf.WriteString("] by ")
	buf.WriteString(c.Server.Hostname)
	buf.WriteString(" with ")
	buf.WriteString(c.Protocol.String())
	buf.WriteRune(';')

	if c.TLS != nil {
		buf.WriteString("\r\n\t(version=")
		buf.WriteString(tlsVersions[c.TLS.Version])
		buf.WriteString(" cipher=0x")
		buf.WriteString(strconv.FormatUint(uint64(c.TLS.CipherSuite), 16))
		buf.WriteString(");")
	}

	buf.WriteString("\r\n\t")
	buf.WriteString(time.Now().Format("Mon Jan 2 15:04:05 -0700 2006"))
	buf.WriteString("\r\n")

	// Add the Received line to the end of the byte slice
	line := wrap(buf.Bytes())
	e.Data = append(e.Data, line...)

	// Move the new Received line up front
	copy(e.Data[len(line):], e.Data[0:len(e.Data)-len(line)])
	copy(e.Data, line)
}
