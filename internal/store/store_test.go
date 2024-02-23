package store_test

import (
	"net/url"
	"testing"

	"github.com/liftedinit/mfx-migrator/internal/httpclient"
	"github.com/liftedinit/mfx-migrator/internal/store"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	tests := []struct {
		desc   string
		url    string
		client *httpclient.HttpClient
	}{
		{"new", "http://localhost:3001", nil},
		{"new with client", "http://localhost:3001", httpclient.New()},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			u, err := url.Parse(tt.url)
			if err != nil {
				t.Error("Error")
			}
			if tt.client == nil {
				s := store.New(u)
				require.NotNil(t, s)
			} else {
				s := store.NewWithClient(u, tt.client)
				require.NotNil(t, s)
			}
		})
	}
}
