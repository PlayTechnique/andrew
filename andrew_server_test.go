package andrew_test

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"testing"
	"testing/fstest"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/playtechnique/andrew"
)

func TestServerRespondsStatusOKForExistingPage(t *testing.T) {
	t.Parallel()
	expected := []byte(`
<!DOCTYPE html>
<head>
  <title>index title</title>
</head>
<body>
</body>
`)

	s := newTestAndrewServer(t, fstest.MapFS{
		"index.html": &fstest.MapFile{
			Data: expected,
		},
	})

	resp, err := http.Get(s.BaseUrl + "/index.html")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %q", resp.Status)
	}
	received, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(received, expected) {
		t.Fatalf("Expected %q, received %q", expected, received)
	}
}

func TestGetForNonExistentPageGeneratesStatusNotFound(t *testing.T) {
	t.Parallel()

	s := newTestAndrewServer(t, fstest.MapFS{})

	resp, err := http.Get(s.BaseUrl + "/page.html")
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("Expected a 404 response for a non-existent page, received %d", resp.StatusCode)
	}
}

func TestGetForUnreadablePageGeneratesStatusForbidden(t *testing.T) {
	t.Parallel()

	contentRoot := t.TempDir()

	// fstest.MapFS does not enforce file permissions, so we need a real file system in this test.
	err := os.WriteFile(contentRoot+"/index.html", []byte{}, 0o222)

	if err != nil {
		t.Fatal(err)
	}
	s := newTestAndrewServer(t, os.DirFS(contentRoot))

	resp, err := http.Get(s.BaseUrl + "/index.html")
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("Expected a 403 response for a page without permission, received %d", resp.StatusCode)
	}
}

func Test500ErrorForUnforeseenErrorCase(t *testing.T) {
	t.Parallel()

	_, status := andrew.CheckPageErrors(errors.New("novel error"))
	if status != http.StatusInternalServerError {
		t.Errorf("Expected status 500 for unknown error, received %q", status)
	}
}

func TestGetSitemapReturnsTheSitemap(t *testing.T) {
	t.Parallel()

	s := newTestAndrewServer(t, fstest.MapFS{})

	resp, err := http.Get(s.BaseUrl + "/sitemap.xml")
	if err != nil {
		t.Fatal(err)
	}

	expected := []byte("http://www.sitemaps.org/schemas/sitemap/0.9")
	received, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Contains(received, expected) {
		t.Fatalf("Expected %q, received %q", expected, received)
	}
}

func TestGettingADirectoryDefaultsToIndexHtml(t *testing.T) {
	t.Parallel()

	expected := []byte(`
<!DOCTYPE html>
<head>
<title>index title</title>
</head>
<body>
</body>
	`)

	// fstest.MapFS does not create directory-like objects, so we need a real file system in this test.
	contentRoot := t.TempDir()
	os.MkdirAll(contentRoot+"/pages", 0o755)

	err := os.WriteFile(contentRoot+"/pages/index.html", expected, 0o755)
	if err != nil {
		t.Fatal(err)
	}
	s := newTestAndrewServer(t, os.DirFS(contentRoot))

	resp, err := http.Get(s.BaseUrl + "/pages/")
	if err != nil {
		t.Fatal(err)
	}

	received, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(received, expected) {
		t.Fatalf("Expected %q, received %q", expected, received)
	}
}

func TestServerServesRequestedPage(t *testing.T) {
	t.Parallel()

	contentRoot := fstest.MapFS{
		"page.html": &fstest.MapFile{Data: []byte("some text")},
	}

	s := newTestAndrewServer(t, contentRoot)
	t.Logf("Server running on %s\n", s.BaseUrl)

	resp, err := http.Get(s.BaseUrl + "/page.html")
	if err != nil {
		t.Fatal(err)
	}

	received, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(received, []byte("some text")) {
		t.Fatalf("Expected %q, received %q", []byte("some text"), received)
	}
}

func TestServerServesIndexPageByDefault(t *testing.T) {
	t.Parallel()

	expected := []byte(`
<!DOCTYPE html>
<head>
<title>index title</title>
</head>
<body>
</body>
	`)

	contentRoot := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: expected},
	}

	s := newTestAndrewServer(t, contentRoot)

	resp, err := http.Get(s.BaseUrl)
	if err != nil {
		t.Fatal(err)
	}

	received, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(received, expected) {
		t.Fatalf("Expected %q, received %q", expected, received)
	}
}

func TestCorrectMimeTypeIsSetForKnownFileTypes(t *testing.T) {
	t.Parallel()

	expectedMimeTypes := map[string]string{
		".css":  "text/css; charset=utf-8",
		".html": "text/html; charset=utf-8",
		".js":   "application/javascript; charset=utf-8",
		".jpg":  "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".ico":  "image/x-icon",
	}

	contentRoot := fstest.MapFS{
		"page.css":  {},
		"page.html": {},
		"page.js":   {},
		"page.jpg":  {},
		"page.png":  {},
		"page.gif":  {},
		"page.webp": {},
		"page.ico":  {},
	}

	s := newTestAndrewServer(t, contentRoot)

	for page := range contentRoot {
		resp, err := http.Get(s.BaseUrl + "/" + page)
		if err != nil {
			t.Fatal(err)
		}
		// Read the body to prevent resource leaks
		_, err = io.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}

		expectedMimeType := expectedMimeTypes[filepath.Ext(page)]
		contentType := resp.Header.Get("Content-Type")

		if contentType != expectedMimeType {
			t.Errorf("Incorrect MIME type for %s: got %s, want %s", page, contentType, expectedMimeType)
		}
	}
}

// TestArticlesInAndrewTableOfContentsAreDefaultSortedByModTime is verifying that
// when the list of links andrew generates for the {{.AndrewTableOfContents}} are
// sorted by mtime, not using the ascii sorting order.
// This test requires having two files which are in one order when sorted
// ascii-betically and in another order by date time, so that we can tell
// what file attribute andrew is actually sorting on.
func TestArticlesInAndrewTableOfContentsAreDefaultSortedByModTime(t *testing.T) {

	expectedOrder, err := regexp.Compile(`(?s).*b_newer.*a_older.*`)

	if err != nil {
		t.Fatal(err)
	}

	contentRoot := t.TempDir()

	// fstest.MapFS does not enforce file permissions, so we need a real file system in this test.
	// above might be wrong
	err = os.WriteFile(contentRoot+"/index.html", []byte("{{.AndrewTableOfContents}}"), 0o700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(contentRoot+"/a_older.html", []byte{}, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(contentRoot+"/b_newer.html", []byte{}, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	older := now.Add(-10 * time.Minute)

	os.Chtimes(contentRoot+"/b_newer.html", now, now)
	os.Chtimes(contentRoot+"/a_older.html", older, older)

	server := andrew.Server{SiteFiles: os.DirFS(contentRoot)}

	page, err := andrew.NewPage(server, "index.html")

	if err != nil {
		t.Fatal(err)
	}

	received := page.Content

	if expectedOrder.FindString(received) == "" {
		t.Fatal(cmp.Diff(expectedOrder, received))
	}

}

// newTestAndrewServer starts an andrew and returns the localhost url that you can run http gets against
// to retrieve data from that server
func newTestAndrewServer(t *testing.T, contentRoot fs.FS) *andrew.Server {
	t.Helper()

	// Listen on IPv4 localhost on any available port
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}

	addr := listener.Addr().String()
	listener.Close()

	server := andrew.NewServer(contentRoot, addr, "http://"+addr)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Log("Server stopped with error:", err)
		}
	}()

	// Ensure server is ready by attempting to dial the server repeatedly
	ready := make(chan bool)
	go func() {
		defer close(ready)
		for i := 0; i < 10; i++ {
			conn, err := net.Dial("tcp", addr)
			if err == nil {
				conn.Close()
				ready <- true
				return
			}
			time.Sleep(50 * time.Millisecond) // Brief sleep to wait for server to be ready
		}
		t.Log("Failed to connect to server after retries:", addr)
	}()

	// Wait for server to be confirmed ready
	<-ready

	return server
}
