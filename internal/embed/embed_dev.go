//go:build !production

package embed

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Handler returns a reverse proxy to the Vite dev server.
func Handler() (http.Handler, error) {
	target, err := url.Parse("http://localhost:5173")
	if err != nil {
		return nil, err
	}
	return httputil.NewSingleHostReverseProxy(target), nil
}

// IsDev returns true in development mode.
func IsDev() bool { return true }
