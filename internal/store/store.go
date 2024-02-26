package store

import (
	"log/slog"

	"github.com/go-resty/resty/v2"
)

type Store struct {
	client *resty.Client
}

func NewWithClient(client *resty.Client) *Store {
	return &Store{client: client}
}

func (s *Store) Login(username, password string) (string, error) {
	slog.Debug("logging in", "username", username, "password", "[REDACTED]")

	req := s.client.R().SetBody(&Credentials{
		Username: username,
		Password: password,
	}).SetResult(&Token{})
	response, err := req.Post("/auth/login")
	if err != nil {
		return "", err
	}
	token := response.Result().(*Token)

	return token.AccessToken, nil
}
