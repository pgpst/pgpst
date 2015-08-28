package smtpd

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"
)

type Server struct {
	Hostname       string
	WelcomeMessage string

	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	DataTimeout  time.Duration

	MaxConnections int
	MaxMessageSize int
	MaxRecipients  int

	WrapperChain   []Wrapper
	SenderChain    []Sender
	RecipientChain []Recipient
	DeliveryChain  []Delivery

	TLSConfig *tls.Config
	ForceTLS  bool

	extensions []string
}

func (s *Server) configureDefaults() error {
	if s.Hostname == "" {
		s.Hostname = "localhost"
	}

	if s.WelcomeMessage == "" {
		s.WelcomeMessage = fmt.Sprintf("%s ESMTP ready.", s.Hostname)
	}

	if s.ReadTimeout == 0 {
		s.ReadTimeout = time.Second * 60
	}

	if s.WriteTimeout == 0 {
		s.WriteTimeout = time.Second * 60
	}

	if s.DataTimeout == 0 {
		s.DataTimeout = time.Minute * 5
	}

	if s.MaxConnections == 0 {
		s.MaxConnections = 100
	}

	if s.MaxRecipients == 0 {
		s.MaxRecipients = 100
	}

	if s.MaxMessageSize == 0 {
		s.MaxMessageSize = 20 * 1024 * 1024 // 20MB
	}

	if s.WrapperChain == nil {
		s.WrapperChain = []Wrapper{}
	}

	if s.SenderChain == nil {
		s.SenderChain = []Sender{}
	}

	if s.RecipientChain == nil {
		s.RecipientChain = []Recipient{}
	}

	if s.DeliveryChain == nil {
		s.DeliveryChain = []Delivery{}
	}

	if s.ForceTLS && s.TLSConfig == nil {
		return errors.New("Cannot use ForceTLS with no TLSConfig")
	}

	s.extensions = []string{
		"SIZE " + strconv.Itoa(s.MaxMessageSize),
		"8BITMIME",
		"PIPELINING",
	}

	return nil
}

func (s *Server) ListenAndServe(addr string) error {
	if err := s.configureDefaults(); err != nil {
		return err
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return s.Serve(l)
}

func (s *Server) Serve(l net.Listener) error {
	if err := s.configureDefaults(); err != nil {
		return err
	}
	defer l.Close()

	var limiter chan struct{}
	if s.MaxConnections > 0 {
		limiter = make(chan struct{}, s.MaxConnections)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				time.Sleep(time.Second)
				continue
			}
			return err
		}

		// Prepare new bufio interfaces
		reader := bufio.NewReader(conn)
		writer := bufio.NewWriter(conn)
		scanner := bufio.NewScanner(reader)

		// Prepare a new Connection
		sc := &Connection{
			Server:  s,
			Addr:    conn.RemoteAddr(),
			conn:    conn,
			reader:  reader,
			writer:  writer,
			scanner: scanner,
		}

		// If there's no limiter, just serve
		if limiter == nil {
			go sc.serve()
		} else {
			go func() {
				// Try to push into buffered limiter
				select {
				case limiter <- struct{}{}:
					// Serve
					sc.serve()
					// Unlock the connection
					<-limiter
				default:
					// Reject the connection
					sc.reject()
				}
			}()
		}
	}
}
