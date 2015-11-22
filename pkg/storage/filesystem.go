package storage

import (
	"errors"
	"io"
	"os"
	"os/user"
	"path/filepath"

	"github.com/armed/mkdirp"
	"github.com/dchest/uniuri"

	"code.pgp.st/pgpst/pkg/config"
)

func NewFilesystem(cfg config.FilesystemConfig) (*Filesystem, error) {
	// Resolve the path
	if cfg.Path[0] == '~' {
		user, err := user.Current()
		if err != nil {
			return nil, err
		}
		cfg.Path = filepath.Join(user.HomeDir, cfg.Path[1:])
	}

	// Check if it exists, create it if it's missing
	fi, err := os.Stat(cfg.Path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := mkdirp.Mk(cfg.Path, 0755); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		if !fi.IsDir() {
			return nil, errors.New("Filesystem storage's base path is not a directory.")
		}
	}

	// Return a struct
	return &Filesystem{
		Base: cfg.Path,
	}, nil
}

type Filesystem struct {
	Base string
}

func (f *Filesystem) Create(data io.Reader) (string, error) {
	id := uniuri.NewLen(uniuri.UUIDLen)

	if err := f.Update(id, data); err != nil {
		return "", err
	}

	return id, nil
}

func (f *Filesystem) Fetch(id string) (io.ReadCloser, error) {
	file, err := os.Open(filepath.Join(f.Base, id))
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (f *Filesystem) Update(id string, data io.Reader) error {
	file, err := os.OpenFile(filepath.Join(f.Base, id), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	_, err = io.Copy(file, data)
	if err != nil {
		return err
	}

	return nil
}

func (f *Filesystem) Delete(id string) error {
	return os.Remove(filepath.Join(f.Base, id))
}
