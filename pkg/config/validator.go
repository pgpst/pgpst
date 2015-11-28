package config

import (
	"errors"
	"net"
	"net/url"
	"os"
	"path/filepath"

	log "gopkg.in/inconshreveable/log15.v2"
)

type Validator struct{}

func (v *Validator) Validate(is interface{}) error {
	s := is.(*Config)

	// Check general settings
	if _, err := log.LvlFromString(s.LogLevel); err != nil {
		return err
	}

	// Check API config
	if s.API.Enabled {
		if _, err := net.ResolveTCPAddr("tcp", s.API.Address); err != nil {
			return err
		}
	}

	// Check Mailer config
	if s.Mailer.Enabled {
		if _, err := net.ResolveTCPAddr("tcp", s.Mailer.Address); err != nil {
			return err
		}
		if s.Mailer.SenderConcurrency < 1 {
			return errors.New("Invalid sender concurrency. It must be at least 1.")
		}
		if s.Mailer.TLSCert != "" {
			if _, err := os.Stat(s.Mailer.TLSCert); err != nil {
				return err
			}
			if _, err := os.Stat(s.Mailer.TLSKey); err != nil {
				return err
			}
		}
		if s.Mailer.WelcomeMessage == "" {
			return errors.New("Missing SMTP welcome message.")
		}
		if s.Mailer.ReadTimeout < 1 || s.Mailer.WriteTimeout < 1 || s.Mailer.DataTimeout < 1 {
			return errors.New("Invalid timeout value. It must be at least 1.")
		}
		if s.Mailer.MaxConnections < 0 {
			return errors.New("Invalid max connections count. It must not be negative.")
		}
		if s.Mailer.MaxRecipients < 1 {
			return errors.New("Invalid max recipients count. It must be at least 1.")
		}
	}

	// Check database config
	if s.Database == Postgres {
		if len(s.Postgres.ConnectionString) == 0 {
			return errors.New("Invalid Postgres connection string.")
		}
	} else if s.Database == SQLite {
		if len(s.SQLite.ConnectionString) == 0 {
			return errors.New("Invalid SQLite connection string.")
		}
	} else {
		return errors.New("Invalid database type.")
	}

	// Check blob storage settings
	if s.Storage == WeedFS {
		if _, err := url.Parse(s.WeedFS.MasterURL); err != nil {
			return err
		}
	} else if s.Storage == Filesystem {
		if _, err := filepath.Abs(s.Filesystem.Path); err != nil {
			return err
		}
	} else {
		return errors.New("Invalid blob storage type.")
	}

	// Check queue settings
	if s.Queue == NSQ {
		for _, addr := range s.NSQ.ServerAddresses {
			if _, err := net.ResolveTCPAddr("tcp", addr); err != nil {
				return err
			}
		}

		for _, addr := range s.NSQ.LookupdAddresses {
			if _, err := net.ResolveTCPAddr("tcp", addr); err != nil {
				return err
			}
		}
	} else if s.Queue == Memory {
		// no settings yet
	} else {
		return errors.New("Invalid queue type.")
	}

	return nil
}
