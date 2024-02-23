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
	req := c.Client.R()
	if res != nil {
		req = req.SetResult(res)
	}
	return req.Get(url)
}

func (c *HttpClient) Put(url string, body interface{}, res interface{}) (*resty.Response, error) {
	slog.Debug("PUT", "url", url, "body", body)
	req := c.Client.R().SetBody(body)
	if res != nil {
		req = req.SetResult(res)
	}
	return req.Put(url)
}

func (c *HttpClient) Post(url string, body interface{}, res interface{}) (*resty.Response, error) {
	slog.Debug("POST", "url", url, "body", "[REDACTED]")
	req := c.Client.R().SetBody(body)
	if res != nil {
		req = req.SetResult(res)
	}
	return req.Post(url)
}
