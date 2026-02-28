//go:build production

package embed

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var distFS embed.FS

// spaHandler serves embedded assets with SPA fallback to index.html.
type spaHandler struct {
	fs      http.FileSystem
	fileServer http.Handler
}

func (h *spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Open to check existence (http.FileSystem has no Stat); close immediately.
	if f, err := h.fs.Open(r.URL.Path); err != nil {
		// File not found: serve index.html for SPA client-side routing
		r.URL.Path = "/"
	} else {
		f.Close()
	}

	// Cache-Control: hashed assets get long-term cache, index.html gets no-cache
	path := r.URL.Path
	if path == "/" || path == "/index.html" {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	} else if strings.Contains(path, "/assets/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	}

	h.fileServer.ServeHTTP(w, r)
}

// Handler returns a file server for the embedded frontend assets.
// Falls back to index.html for paths that don't match a static file (SPA routing).
func Handler() (http.Handler, error) {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return nil, fmt.Errorf("embedded frontend: %w", err)
	}
	httpFS := http.FS(sub)
	return &spaHandler{
		fs:         httpFS,
		fileServer: http.FileServerFS(sub),
	}, nil
}

// IsDev returns false in production mode.
func IsDev() bool { return false }
