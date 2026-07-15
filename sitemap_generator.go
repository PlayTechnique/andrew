package andrew

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
)

// SiteMap
func (a Server) ServeSiteMap(w http.ResponseWriter, r *http.Request) {
	sitemap, err := GenerateSiteMap(a.SiteFiles, a.BaseUrl)
	if err != nil {
		message, status := CheckPageErrors(err)
		w.WriteHeader(status)
		fmt.Fprint(w, message)
		return
	}

	w.WriteHeader(http.StatusOK)

	// The response is already on the wire, so there is no status left to set. A client that
	// hangs up mid-write is routine rather than exceptional, so log it and move on.
	if _, err := fmt.Fprint(w, string(sitemap)); err != nil {
		slog.Info("could not finish writing the sitemap", "error", err)
	}
}

// Generates and returns a sitemap.xml.
// An error from the walk is returned rather than swallowed, so that a partial walk surfaces
// as an http error instead of a sitemap that looks complete but silently omits pages.
func GenerateSiteMap(f fs.FS, baseUrl string) ([]byte, error) {
	buff := new(bytes.Buffer)

	const (
		header = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
`
		footer = `</urlset>
`
	)

	fmt.Fprint(buff, header)

	err := fs.WalkDir(f, ".", func(path string, dir fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".html" {
			// index.html
			// foo/bar/index.html
			path = strings.TrimSuffix(path, "index.html")

			fmt.Fprintf(buff, "\t<url>\n\t\t<loc>%s/%s</loc>\n\t</url>\n", baseUrl, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	fmt.Fprint(buff, footer)

	return buff.Bytes(), nil
}
