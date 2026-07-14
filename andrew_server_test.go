package andrew_test

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
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

	s := newTestAndrewServer(t, ".",fstest.MapFS{
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

	s := newTestAndrewServer(t, ".",fstest.MapFS{})

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
	s := newTestAndrewServer(t, ".",os.DirFS(contentRoot))

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

	s := newTestAndrewServer(t, ".",fstest.MapFS{})

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
	s := newTestAndrewServer(t, ".",os.DirFS(contentRoot))

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

	s := newTestAndrewServer(t, ".",contentRoot)
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

	s := newTestAndrewServer(t, ".",contentRoot)

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

	s := newTestAndrewServer(t, ".",contentRoot)

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

	page, err := server.NewPage("index.html")

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
func newTestAndrewServer(t *testing.T, contentRoot string, siteFiles fs.FS) *andrew.Server {
	t.Helper()

	// Listen on IPv4 localhost on any available port
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}

	rssInfo := andrew.RssInfo{Title: "exampleTitle", Description: "exampleDescription"}

	addr := listener.Addr().String()
	listener.Close()

	server := andrew.NewServer(contentRoot, siteFiles, addr, "http://"+addr, rssInfo)

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

// TestServeLogsRequestMetadata verifies that each request through Serve emits
// an access-log line containing the client's user-agent, referer, and path.
// This is the data needed to spot bots that impersonate Googlebot.
func TestServeLogsRequestMetadata(t *testing.T) {
	// Not parallel: this test swaps the process-global slog default.
	var logs bytes.Buffer
	original := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(original) })

	s := newTestAndrewServer(t, ".",fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<body></body>")},
	})

	req, err := http.NewRequest(http.MethodGet, s.BaseUrl+"/index.html", nil)
	if err != nil {
		t.Fatal(err)
	}
	const wantUA = "FakeGooglebot/2.1"
	const wantReferer = "http://evil.example.com/"
	req.Header.Set("User-Agent", wantUA)
	req.Header.Set("Referer", wantReferer)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	got := logs.String()
	for _, want := range []string{wantUA, wantReferer, "/index.html"} {
		if !strings.Contains(got, want) {
			t.Errorf("request log missing %q; got: %s", want, got)
		}
	}
}

// TestServeLogsXForwardedFor verifies that when Andrew runs behind a reverse
// proxy (e.g. Traefik), logRequest extracts the real client IP from the
// X-Forwarded-For header instead of logging the proxy's container IP.
func TestServeLogsXForwardedFor(t *testing.T) {
	// Not parallel: this test swaps the process-global slog default.
	var logs bytes.Buffer
	original := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(original) })

	s := newTestAndrewServer(t, ".",fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<body></body>")},
	})

	req, err := http.NewRequest(http.MethodGet, s.BaseUrl+"/index.html", nil)
	if err != nil {
		t.Fatal(err)
	}
	const wantIP = "203.0.113.42"
	req.Header.Set("X-Forwarded-For", wantIP+", 10.0.0.1") // client, then proxy

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	got := logs.String()
	if !strings.Contains(got, wantIP) {
		t.Errorf("request log missing client IP %q; got: %s", wantIP, got)
	}
	// Ensure the proxy IP (10.0.0.1) is NOT logged.
	if strings.Contains(got, "10.0.0.1") {
		t.Errorf("request log incorrectly included proxy IP; got: %s", got)
	}
}

// TestGetSiblingsAndChildrenHonorsPublishTimeFromPartial verifies
// that a sibling's <meta name="andrew-publish-time"> is respected even when
// it lives inside an .AndrewPartialFile partial.
// The bug: GetSiblingsAndChildren reads meta tags from the raw page bytes before partials are rendered, so a
// meta tag inside a partial is invisible and PublishTime silently falls back
// to mtime.
func TestGetSiblingsAndChildrenHonorsPublishTimeFromPartial(t *testing.T) {
	t.Parallel()

	contentRoot := t.TempDir()

	partial := []byte(`<meta name="andrew-publish-time" content="{{ .pubtime }}" />`)
	if err := os.WriteFile(contentRoot+"/.AndrewPartialFile", partial, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(contentRoot+"/index.html", []byte(""), 0o700); err != nil {
		t.Fatal(err)
	}

	page := []byte(`{{ .AndrewPartialFile pubtime="2024-01-28" }}`)
	if err := os.WriteFile(contentRoot+"/post.html", page, 0o700); err != nil {
		t.Fatal(err)
	}

	// Set mtime far away from the meta-tag date so an mtime fallback is obvious.
	mtime := time.Date(2030, 6, 1, 0, 0, 0, 0, time.UTC)
	if err := os.Chtimes(contentRoot+"/post.html", mtime, mtime); err != nil {
		t.Fatal(err)
	}

	server := andrew.Server{SiteFiles: os.DirFS(contentRoot)}

	siblings, err := server.GetSiblingsAndChildren("index.html")
	if err != nil {
		t.Fatal(err)
	}

	var post andrew.Page
	for _, s := range siblings {
		if s.UrlPath == "post.html" {
			post = s
			break
		}
	}

	want := time.Date(2024, 1, 28, 0, 0, 0, 0, time.UTC)
	if !post.PublishTime.Equal(want) {
		t.Fatalf("PublishTime = %v, want %v (meta tag is inside a partial; GetSiblingsAndChildren is falling back to mtime)", post.PublishTime, want)
	}
}
