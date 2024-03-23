package andrew

import (
	"bytes"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
)

// SiteMap
func (a AndrewServer) ServeSiteMap(w http.ResponseWriter, r *http.Request) {
	sitemap, err := GenerateSiteMap(a.SiteFiles, a.BaseUrl)

	if err != nil {
		checkPageErrors(w, r, err)
	}

	w.WriteHeader(http.StatusOK)
	_, err = fmt.Fprint(w, string(sitemap))

	if err != nil {
		panic(err)
	}

}

// Generates and returns a sitemap.xml.
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

	fs.WalkDir(f, ".", func(path string, dir fs.DirEntry, err error) error {
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

	fmt.Fprint(buff, footer)

	return buff.Bytes(), nil
}
