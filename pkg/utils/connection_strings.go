package utils

import (
	"errors"
	"net/url"
	"strconv"
	"time"

	r "github.com/pgpst/pgpst/internal/github.com/dancannon/gorethink"
)

type ConnectionData struct {
	Protocol string
	Address  string
	Item     string
	Options  map[string]string
}

func ParseRethinkDBString(input string) (r.ConnectOpts, error) {
	addr, err := url.Parse(input)
	if err != nil {
		return r.ConnectOpts{}, err
	}

	if addr.Scheme != "rethinkdb" {
		return r.ConnectOpts{}, errors.New("Invalid scheme. Expected rethinkdb, got " + addr.Scheme + ".")
	}

	opts := r.ConnectOpts{}

	opts.Address = addr.Host
	if addr.User != nil {
		opts.AuthKey = addr.User.String()
	}
	if len(addr.Path) > 1 {
		opts.Database = addr.Path[1:]
	}

	query := addr.Query()
	if query.Get("discover_hosts") == "true" {
		opts.DiscoverHosts = true
	}
	if s := query.Get("refresh_interval"); s != "" {
		if i, err := strconv.Atoi(s); err != nil {
			return r.ConnectOpts{}, err
		} else {
			opts.NodeRefreshInterval = time.Second * time.Duration(i)
		}
	}
	if s := query.Get("max_open"); s != "" {
		if i, err := strconv.Atoi(s); err != nil {
			return r.ConnectOpts{}, err
		} else {
			opts.MaxOpen = i
		}
	}
	if s := query.Get("max_idle"); s != "" {
		if i, err := strconv.Atoi(s); err != nil {
			return r.ConnectOpts{}, err
		} else {
			opts.MaxIdle = i
		}
	}

	return opts, nil
}
