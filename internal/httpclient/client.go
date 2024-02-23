package httpclient

import (
	"log/slog"

	"github.com/go-resty/resty/v2"
)

// HttpClient is a wrapper around the resty.Client
type HttpClient struct {
	Client *resty.Client
}

func New() *HttpClient {
	return &HttpClient{Client: resty.New()}
}

func NewWithClient(client *resty.Client) *HttpClient {
	return &HttpClient{Client: client}
}

func (c *HttpClient) Get(url string, res interface{}) (*resty.Response, error) {
	slog.Debug("GET", "url", url)
	if res != nil {
		return c.Client.R().SetResult(res).Get(url)
	}
	return c.Client.R().Get(url)
}

func (c *HttpClient) Put(url string, body interface{}, res interface{}) (*resty.Response, error) {
	slog.Debug("PUT", "url", url, "body", body)
	if res != nil {
		return c.Client.R().SetBody(body).SetResult(res).Put(url)
	}
	return c.Client.R().SetBody(body).Put(url)
}
