package smtpd

import (
	"bufio"
	"crypto/tls"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

type Protocol string

func (p Protocol) String() string {
	return string(p)
}

const (
	SMTP  Protocol = "SMTP"
	ESMTP          = "ESMTP"
)

type Connection struct {
	Server *Server

	HeloName string
	Protocol Protocol
	Addr     net.Addr
	TLS      *tls.ConnectionState

	conn    net.Conn
	reader  *bufio.Reader
	writer  *bufio.Writer
	scanner *bufio.Scanner

	Envelope    *Envelope
	Environment map[string]interface{}
}

func (c *Connection) serve() {
	defer c.close()

	ow := func() {
		// Send a Welcome message
		c.welcome()
		for {
			// Scan for new messages, will break when there's an error
			for c.scanner.Scan() {
				// Handle the textual version of the message
				c.handle(c.scanner.Text())
			}

			// Something is wrong!
			err := c.scanner.Err()
			if err != nil && err == bufio.ErrTooLong {
				c.reply(500, "Line too long")

				// Proceed to the next line and create a new scanner
				c.reader.ReadString('\n')
				c.scanner = bufio.NewScanner(c.reader)

				// Reset the context
				c.reset()
				continue
			} else if err != nil {
				log.Print(err)
			}

			break
		}
	}

	for _, wr := range c.Server.WrapperChain {
		ow = wr.Wrap(ow)
	}

	ow()
}

func (c *Connection) welcome() {
	// 220 Wilkommen!
	c.reply(220, c.Server.WelcomeMessage)
}

func (c *Connection) reset() {
	// Clear the current state
	c.Envelope = nil
}

func (c *Connection) reply(code int, message string) {
	// Write the string and flush the interface
	c.writer.WriteString(strconv.Itoa(code) + " " + message + "\r\n")
	c.flush()
}

func (c *Connection) flush() {
	c.conn.SetWriteDeadline(time.Now().Add(c.Server.WriteTimeout))
	c.writer.Flush()
	c.conn.SetReadDeadline(time.Now().Add(c.Server.ReadTimeout))
}

func (c *Connection) reject() {
	c.reply(421, "Maximum connections count exceeded. Try again later.")
	c.close()
}

func (c *Connection) Error(err error) {
	if se, ok := err.(Error); ok {
		c.reply(se.Code, se.Message)
	} else {
		c.reply(502, err.Error())
	}
}

func (c *Connection) close() {
	c.writer.Flush()
	time.Sleep(200 * time.Millisecond)
	c.conn.Close()
}

type command struct {
	Action string
	Fields []string
}

func (c *Connection) handle(line string) {
	cmd := &command{
		Fields: strings.Fields(line),
	}

	if len(cmd.Fields) > 0 {
		cmd.Action = strings.ToUpper(cmd.Fields[0])
	}

	switch cmd.Action {
	case "HELO":
		c.handleHELO(cmd)
	case "EHLO":
		c.handleEHLO(cmd)
	case "MAIL":
		c.handleMAIL(cmd)
	case "RCPT":
		c.handleRCPT(cmd)
	case "STARTTLS":
		c.handleSTARTTLS(cmd)
	case "DATA":
		c.handleDATA(cmd)
	case "RSET":
		c.handleRSET(cmd)
	case "NOOP":
		c.handleNOOP(cmd)
	case "QUIT":
		c.handleQUIT(cmd)
	default:
		c.reply(502, "Unsupported command.")
	}
}
