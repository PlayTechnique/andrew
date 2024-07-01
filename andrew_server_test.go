package andrew_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
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

func TestAndrewTableOfContentsIsGeneratedCorrectlyInContentrootDirectory(t *testing.T) {
	t.Parallel()

	contentRoot := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte(`
<!doctype HTML>
<head> </head>
<body>
{{ .AndrewTableOfContents }}
</body>
`)},
		"pages/1-2-3.html": &fstest.MapFile{Data: []byte(`
<!doctype HTML>
<head>
<title>1-2-3 Page</title>
</head>
`)},
	}

	s := newTestAndrewServer(t, contentRoot)

	resp, err := http.Get(s.BaseUrl + "/index.html")
	if err != nil {
		t.Fatal(err)
	}

	received, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	expectedIndex := `
<!doctype HTML>
<head> </head>
<body>
<a class="andrewtableofcontentslink" id="andrewtableofcontentslink0" href="pages/1-2-3.html">1-2-3 Page</a>
</body>
`

	if !slices.Equal(received, []byte(expectedIndex)) {
		t.Fatalf("Diff of Expected and Actual: %s", cmp.Diff(expectedIndex, received))
	}
}

func TestAndrewTableOfContentsIsGeneratedCorrectlyInAChildDirectory(t *testing.T) {
	t.Parallel()

	contentRoot := fstest.MapFS{
		"parentDir/index.html": &fstest.MapFile{Data: []byte(`
<!doctype HTML>
<head> </head>
<body>
{{ .AndrewTableOfContents }}
</body>
`)},
		"parentDir/childDir/1-2-3.html": &fstest.MapFile{Data: []byte(`
<!doctype HTML>
<head>
<title>1-2-3 Page</title>
</head>
`)},
	}

	s := newTestAndrewServer(t, contentRoot)

	resp, err := http.Get(s.BaseUrl + "/parentDir/index.html")
	if err != nil {
		t.Fatal(err)
	}

	received, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	expectedIndex := `
<!doctype HTML>
<head> </head>
<body>
<a class="andrewtableofcontentslink" id="andrewtableofcontentslink0" href="childDir/1-2-3.html">1-2-3 Page</a>
</body>
`

	if !slices.Equal(received, []byte(expectedIndex)) {
		t.Fatalf("Diff of Expected and Actual: %s", cmp.Diff(expectedIndex, received))
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

func TestMainCalledWithHelpOptionDisplaysHelp(t *testing.T) {
	t.Parallel()

	args := []string{"--help"}
	received := new(bytes.Buffer)

	exit := andrew.Main(args, received)

	if exit != 0 {
		t.Error("Expected exit value 0, received %i", exit)
	}

	if !strings.Contains(received.String(), "Usage") {
		t.Errorf("Expected help message containing 'Usage', received %s", received)
	}
}

func TestMainCalledWithNoArgsUsesDefaults(t *testing.T) {
	t.Parallel()

	contentRoot, address, baseUrl := andrew.ParseArgs([]string{})

	if contentRoot != andrew.DefaultContentRoot {
		t.Errorf("contentroot should be %s, received %q", andrew.DefaultContentRoot, contentRoot)
	}

	if address != andrew.DefaultAddress {
		t.Errorf("address should be %s, received %q", andrew.DefaultAddress, address)
	}

	if baseUrl != andrew.DefaultBaseUrl {
		t.Errorf("baseUrl should be %s, received %q", andrew.DefaultBaseUrl, baseUrl)
	}
}

func TestMainCalledWithArgsOverridesDefaults(t *testing.T) {
	t.Parallel()

	contentRoot, address, baseUrl := andrew.ParseArgs([]string{"1", "2", "3"})

	if contentRoot != "1" {
		t.Errorf("contentroot should be %s, received %q", "1", contentRoot)
	}

	if address != "2" {
		t.Errorf("address should be %s, received %q", "2", address)
	}

	if baseUrl != "3" {
		t.Errorf("baseUrl should be %s, received %q", "3", baseUrl)
	}
}

func TestMainCalledWithInvalidAddressPanics(t *testing.T) {
	t.Parallel()
	args := []string{".", "notanipaddress"}
	nullLogger := new(bytes.Buffer)

	// No need to check whether `recover()` is nil. Just turn off the panic.
	defer func() {
		err := recover()
		if err == nil {
			t.Fatalf("Expected panic with invalid address, received %v", err)
		}
	}()

	andrew.Main(args, nullLogger)
}

// TestArticlesInAndrewTableOfContentsAreDefaultSortedByModTime is verifying that
// when the list of links andrew generates for the {{.AndrewTableOfContents}} are
// sorted by mtime, not using the ascii sorting order.
// This test requires having two files which are in one order when sorted
// ascii-betically and in another order by date time, so that we can tell
// what file attribute andrew is actually sorting on.
func TestArticlesInAndrewTableOfContentsAreDefaultSortedByModTime(t *testing.T) {
	expected := `<a class="andrewtableofcontentslink" id="andrewtableofcontentslink0" href="b_newer.html">b_newer.html</a>` +
		`<a class="andrewtableofcontentslink" id="andrewtableofcontentslink1" href="a_older.html">a_older.html</a>`

	contentRoot := t.TempDir()

	// fstest.MapFS does not enforce file permissions, so we need a real file system in this test.
	// above might be wrong
	err := os.WriteFile(contentRoot+"/index.html", []byte("{{.AndrewTableOfContents}}"), 0o700)
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

	if expected != string(received) {
		t.Errorf(cmp.Diff(expected, received))
	}

}

// TestArticlesOrderInAndrewTableOfContentsIsOverridable is verifying that
// when a page contains an andrew-publish-time meta element then the list of links andrew
// generates for the {{.AndrewTableOfContents}} are
// sorted by the meta element, then the mtime, not using the ascii sorting order.
// This test requires having several files which are in one order when sorted
// by modtime and in another order by andrew-publish-time time, so that we can tell
// what file attribute andrew is actually sorting on.
func TestArticlesOrderInAndrewTableOfContentsIsOverridable(t *testing.T) {
	expected := `<a class="andrewtableofcontentslink" id="andrewtableofcontentslink0" href="b_newest.html">b_newest.html</a>` +
		`<a class="andrewtableofcontentslink" id="andrewtableofcontentslink1" href="c_newer.html">c_newer.html</a>` +
		`<a class="andrewtableofcontentslink" id="andrewtableofcontentslink2" href="a_older.html">a_older.html</a>`

	contentRoot := t.TempDir()

	err := os.WriteFile(contentRoot+"/index.html", []byte("{{.AndrewTableOfContents}}"), 0o700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(contentRoot+"/a_older.html", []byte{}, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(contentRoot+"/c_newer.html", []byte{}, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()

	newest := now.Add(24 * time.Hour)
	content := fmt.Sprintf(`<meta name="andrew-publish-time" content="%s">`, newest.Format("2006-01-02"))

	err = os.WriteFile(contentRoot+"/b_newest.html", []byte(content), 0o700)
	if err != nil {
		t.Fatal(err)
	}

	older := now.Add(-24 * time.Hour)

	os.Chtimes(contentRoot+"/a_older.html", older, older)
	os.Chtimes(contentRoot+"/b_newest.html", older, older)
	os.Chtimes(contentRoot+"/c_newer.html", now, now)

	server := andrew.Server{SiteFiles: os.DirFS(contentRoot)}

	page, err := andrew.NewPage(server, "index.html")

	if err != nil {
		t.Fatal(err)
	}

	received := page.Content

	if expected != string(received) {
		t.Errorf(cmp.Diff(expected, received))
	}

}

// TestInvalidMetaContentDoesNotCrashTheWebServer checks that if there's
// garbage data inside a meta element named andrew-publish-at that we do
// something sensible rather than crashing the web server and emitting a 502.
func TestInvalidAndrewPublishTimeContentDoesNotCrashTheWebServer(t *testing.T) {
	t.Parallel()

	contentRoot := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte(`
<!doctype HTML>
<head> </head>
<body>
{{ .AndrewTableOfContents }}
</body>
`)},
		"a.html": &fstest.MapFile{Data: []byte(`
<!doctype HTML>
<head>
<meta name="andrew-publish-time" content="<no value>"
</head>
`)},
	}

	s := newTestAndrewServer(t, contentRoot)

	resp, err := http.Get(s.BaseUrl)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected http 200, received %d", resp.StatusCode)
	}

	received, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	expectedIndex := `
<!doctype HTML>
<head> </head>
<body>
<a class="andrewtableofcontentslink" id="andrewtableofcontentslink0" href="a.html">a.html</a>
</body>
`

	if !slices.Equal(received, []byte(expectedIndex)) {
		t.Fatalf("Diff of Expected and Actual: %s", cmp.Diff(expectedIndex, received))
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
