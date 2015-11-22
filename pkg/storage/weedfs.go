package storage

import (
	"io"
	"net/http"
	"strings"

	"github.com/ginuerzh/weedo"

	"code.pgp.st/pgpst/pkg/config"
)

func NewWeedFS(cfg config.WeedFSConfig) (*WeedFS, error) {
	// Create a new client
	client := weedo.NewClient(cfg.MasterURL)

	// Check if the system is working
	if err := client.Master().Status(); err != nil {
		return nil, err
	}

	// Return a new struct
	return &WeedFS{
		Client: client,
	}, nil
}

type WeedFS struct {
	Client *weedo.Client
}

func (w *WeedFS) Create(data io.Reader) (string, error) {
	fid, _, err := w.Client.AssignUpload("", "text/plain", data)
	if err != nil {
		return "", err
	}

	return fid, nil
}

func (w *WeedFS) Fetch(id string) (io.ReadCloser, error) {
	url, _, err := w.Client.GetUrl(id)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (w *WeedFS) Update(id string, data io.Reader) error {
	vid := id[:strings.Index(id, ",")+1]

	vol, err := w.Client.Volume(vid, "")
	if err != nil {
		return err
	}

	_, err = vol.Upload(id, 0, "", "text/plain", data)
	if err != nil {
		return err
	}

	return nil
}

func (w *WeedFS) Delete(id string) error {
	return w.Client.Delete(id, 1)
}
