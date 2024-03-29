package andrew_test

import (
	"bytes"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"
	"github.com/playtechnique/andrew"
)

func TestGenerateSitemapCreatesACorrectSiteMap(t *testing.T) {
	expected := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	<url>
		<loc>http://localhost:8080/</loc>
	</url>
	<url>
		<loc>http://localhost:8080/page.html</loc>
	</url>
</urlset>
`)

	testFs := fstest.MapFS{
		"index.html": {},
		"page.html":  {},
	}

	baseUrl := "http://localhost:8080"

	sitemap, err := andrew.GenerateSiteMap(testFs, baseUrl)

	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(sitemap, expected) {
		t.Error(cmp.Diff(expected, sitemap))
	}

}
