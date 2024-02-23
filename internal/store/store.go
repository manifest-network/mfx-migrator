package store

import (
	"log/slog"
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

func (s *Store) SetClient(client *httpclient.HttpClient) {
	s.client = client
}

func (s *Store) Login(username, password string) (string, error) {
	slog.Debug("logging in", "username", username, "password", "[REDACTED]")

	fullUrl, err := url.JoinPath(s.rootUrl.String(), GetAuthEndpoint())
	if err != nil {
		return "", err
	}

	response, err := s.client.Post(fullUrl, &Credentials{
		Username: username,
		Password: password,
	}, &Token{})
	if err != nil {
		return "", err
	}
	token := response.Result().(*Token)

	return token.AccessToken, nil
}
