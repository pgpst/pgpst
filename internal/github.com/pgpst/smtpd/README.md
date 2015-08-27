# lavab/smtpd [![Build Status](https://travis-ci.org/lavab/smtpd.svg?branch=master)](https://travis-ci.org/lavab/smtpd) [![GoDoc](https://godoc.org/github.com/lavab/smtpd?status.png)](https://godoc.org/github.com/lavab/smtpd)

smtpd is a simple library for implementing SMTP servers in Golang.
Supports STARTTLS, middleware chains for event handling and configurable
limits.

## Example

```go
package main

import (
	"log"

	"github.com/getsentry/raven-go"
	"github.com/lavab/smtpd"
)

func main() {
	// Connect to raven for sentry logging
	rc, err := raven.NewClient("some raven dsn", nil)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new SMTPD server
	server := &smtpd.Server{
		WelcomeMessage: "Hello world!",

		WrapperChain: []smtpd.Wrapper{
			smtpd.Wrapper(func(next smtpd.Wrapped) smtpd.Wrapped {
				return smtpd.Wrapped(func() {
					rc.CapturePanic(next, nil)
				})
			}),
		},
		DeliveryChain: []smtpd.Middleware{
			smtpd.Middleware(func(next smtpd.Handler) smtpd.Handler {
				return smtpd.Handler(func(conn *smtpd.Connection) {
					log.Printf("Sender: %s", conn.Envelope.Sender)
					log.Printf("Recipients: %s", strings.Join(conn.Envelope.Recipients, ", "))
					log.Printf("Body:\n%s", string(conn.Envelope.Body))
					
					next(conn)
				})
			}),
		},
	}

	// Bind it to an address
	if err := server.ListenAndServe(":25"); err != nil {
		log.Fatal(err)
	}
}
```