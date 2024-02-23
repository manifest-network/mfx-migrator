package store

import (
	"net/url"

	"github.com/liftedinit/mfx-migrator/internal/httpclient"
)

type Store struct {
	client  *httpclient.HttpClient
	rootUrl *url.URL
}

func New(url *url.URL) *Store {
	return &Store{client: httpclient.New(), rootUrl: url}
}

func NewWithClient(url *url.URL, client *httpclient.HttpClient) *Store {
	return &Store{client: client, rootUrl: url}
}
