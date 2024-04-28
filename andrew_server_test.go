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

	contentRoot := t.TempDir()
	os.MkdirAll(contentRoot+"/pages", 0o755)

	// fstest.MapFS does not create directory-like objects, so we need a real file system in this test.
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

func TestAndrewIndexBodyIsGeneratedCorrectlyInContentrootDirectory(t *testing.T) {
	t.Parallel()

	contentRoot := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte(`
<!doctype HTML>
<head> </head>
<body>
{{ .AndrewIndexBody }}
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
<a class="andrewindexbodylink" id="andrewindexbodylink0" href="pages/1-2-3.html">1-2-3 Page</a>
</body>
`

	if !slices.Equal(received, []byte(expectedIndex)) {
		t.Fatalf("Diff of Expected and Actual: %s", cmp.Diff(expectedIndex, received))
	}
}

func TestAndrewIndexBodyIsGeneratedCorrectlyInAChildDirectory(t *testing.T) {
	t.Parallel()

	contentRoot := fstest.MapFS{
		"parentDir/index.html": &fstest.MapFile{Data: []byte(`
<!doctype HTML>
<head> </head>
<body>
{{ .AndrewIndexBody }}
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
<a class="andrewindexbodylink" id="andrewindexbodylink0" href="childDir/1-2-3.html">1-2-3 Page</a>
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

	t.Logf("Running server on %s\n", addr)

	return server
}
