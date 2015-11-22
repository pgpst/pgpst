package storage

import (
	"io"
)

type Storage interface {
	Create(io.Reader) (string, error)
	Fetch(string) (io.ReadCloser, error)
	Update(string, io.Reader) error
	Delete(string) error
}
