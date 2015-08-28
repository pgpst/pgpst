package smtpd

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/textproto"
	"strconv"
	"strings"
	"time"
)

func (c *Connection) handleHELO(cmd *command) {
	if len(cmd.Fields) < 2 {
		c.reply(502, "Missing parameter")
		return
	}

	if c.HeloName != "" {
		// Reset the state if there's a duplicate helo
		c.reset()
	}

	// Checking HELO and denying connections depending on it is forbidden
	// in RFC1123
	c.HeloName = cmd.Fields[1]
	c.Protocol = SMTP
	c.reply(250, "Go ahead.")

	return
}

func (c *Connection) handleEHLO(cmd *command) {
	if len(cmd.Fields) < 2 {
		c.reply(502, "Missing parameter")
		return
	}

	if c.HeloName != "" {
		// Same as above
		c.reset()
	}

	c.HeloName = cmd.Fields[1]
	c.Protocol = ESMTP

	// First line is the hostname
	c.writer.WriteString("250-" + c.Server.Hostname + "\r\n")

	// Can we send STARTTLS?
	if c.Server.TLSConfig != nil && c.TLS == nil {
		// Send all extensions in server's extensions list
		for _, extension := range c.Server.extensions {
			c.writer.WriteString("250-" + extension + "\r\n")
		}

		// And send STARTTLS without a dash
		if c.Server.TLSConfig != nil && c.TLS == nil {
			c.reply(250, "STARTTLS")
		}
	} else {
		// Send all extensions except the last one
		for _, extension := range c.Server.extensions[:len(c.Server.extensions)-1] {
			c.writer.WriteString("250-" + extension + "\r\n")
		}
		// Send the last one seperately, without a dash
		c.reply(250, c.Server.extensions[len(c.Server.extensions)-1])
	}

	return
}

func (c *Connection) handleMAIL(cmd *command) {
	if c.HeloName == "" {
		c.reply(502, "Please introduce yourself first.")
		return
	}

	if c.TLS == nil && c.Server.ForceTLS {
		c.reply(502, "Please turn on TLS by using STARTTLS to proceed.")
		return
	}

	if len(cmd.Fields) < 2 {
		c.reply(502, "Missing parameter.")
		return
	}

	// Parse the first field
	params := strings.Split(cmd.Fields[1], ":")
	if len(params) < 2 {
		c.reply(502, "Invalid second parameter.")
		return
	}

	if c.Envelope != nil {
		c.reply(502, "Duplicate MAIL. Please reset the envelope.")
		return
	}

	// Parse the address
	address, err := parseAddress(params[1])
	if err != nil {
		c.reply(502, err.Error())
		return
	}

	// Execute the sender checking chain
	oh := func(_ *Connection) {
		c.Envelope = &Envelope{
			Sender:     address,
			Recipients: []string{},
		}

		c.reply(250, "Go ahead.")
	}

	for _, ha := range c.Server.SenderChain {
		oh = ha.HandleSender(oh)
	}

	oh(c)

	return
}

func (c *Connection) handleRCPT(cmd *command) {
	if c.Envelope == nil {
		c.reply(502, "Missing MAIL FROM command.")
		return
	}

	if len(c.Envelope.Recipients) >= c.Server.MaxRecipients {
		c.reply(452, "Too many recipients")
		return
	}

	if len(cmd.Fields) < 2 {
		c.reply(502, "Missing parameter.")
		return
	}

	// Parse the first field
	params := strings.Split(cmd.Fields[1], ":")
	if len(params) < 2 {
		c.reply(502, "Invalid second parameter.")
		return
	}

	// Parse the address
	address, err := parseAddress(params[1])
	if err != nil {
		c.reply(502, err.Error())
		return
	}

	// Add it to the recipients list
	c.Envelope.Recipients = append(c.Envelope.Recipients, address)

	// Execute the recipient checking chain
	oh := func(_ *Connection) {
		c.reply(250, "Go ahead.")
	}

	for _, ha := range c.Server.RecipientChain {
		oh = ha.HandleRecipient(oh)
	}

	oh(c)

	return
}

func (c *Connection) handleSTARTTLS(cmd *command) {
	if c.TLS != nil {
		c.reply(502, "Already running in TLS")
		return
	}

	if c.Server.TLSConfig == nil {
		c.reply(502, "TLS not supported")
		return
	}

	tlsConn := tls.Server(c.conn, c.Server.TLSConfig)
	c.reply(220, "Go ahead")

	// Perform a handshake
	if err := tlsConn.Handshake(); err != nil {
		c.reply(550, "Handshake error")
		return
	}

	// Reset envelope, new EHLO/HELO is required after STARTTLS
	c.reset()

	// Reset deadlines on the old connection - zero it out
	c.conn.SetDeadline(time.Time{})

	// Replace connection with a TLS connection
	c.conn = tlsConn
	c.reader = bufio.NewReader(c.conn)
	c.writer = bufio.NewWriter(c.conn)
	c.scanner = bufio.NewScanner(c.reader)

	state := tlsConn.ConnectionState()
	c.TLS = &state

	// Flush the connection to set up new timeout deadlines
	c.flush()

	return
}

func (c *Connection) handleDATA(cmd *command) {
	if c.Envelope == nil || len(c.Envelope.Recipients) == 0 {
		c.reply(502, "Missing RCPT TO command.")
		return
	}

	c.reply(354, "Go ahead. End your data with <CR><LF>.<CR><LF>")
	c.conn.SetDeadline(time.Now().Add(c.Server.DataTimeout))

	data := &bytes.Buffer{}
	reader := textproto.NewReader(c.reader).DotReader()

	_, err := io.CopyN(data, reader, int64(c.Server.MaxMessageSize))

	if err == io.EOF {
		// Message was smaller than MaxMessageSize - deliver the message
		c.Envelope.Data = data.Bytes()

		// Execute the delivery chain
		oh := func(_ *Connection) {
			c.reply(250, "Thank you.")
			c.reset()
		}

		for _, ha := range c.Server.DeliveryChain {
			oh = ha.HandleDelivery(oh)
		}

		oh(c)
	}

	if err != nil {
		// Network error, ignore
		return
	}

	// Discard the rest and report an error.
	io.Copy(ioutil.Discard, reader)
	c.reply(552, "Message exceeded max message size of "+strconv.Itoa(c.Server.MaxMessageSize)+" bytes.")
	c.reset()
	return
}

func (c *Connection) handleRSET(cmd *command) {
	c.reset()
	c.reply(250, "Go ahead.")
	return
}

func (c *Connection) handleNOOP(cmd *command) {
	c.reply(250, "Go ahead")
	return
}

func (c *Connection) handleQUIT(cmd *command) {
	c.reply(221, "OK, bye")
	c.close()
	return
}
