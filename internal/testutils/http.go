package testutils

import (
	"net/http"
	"net/http/httptest"
)

func CreateHTTPTestServer(routes map[string]http.HandlerFunc) *httptest.Server {
	mux := http.NewServeMux()

	for route, handler := range routes {
		mux.HandleFunc(route, handler)
	}

	server := httptest.NewServer(mux)

	return server
}
