package andrew_test

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"
	"github.com/playtechnique/andrew"
)

func TestGetForExistingPageRetrievesThePage(t *testing.T) {
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

	testUrl := startAndrewServer(t, contentRoot)

	resp, err := http.Get(testUrl + "/index.html")

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

func TestGetForNonExistentPageGenerates404(t *testing.T) {
	t.Parallel()

	contentRoot := fstest.MapFS{}
	testUrl := startAndrewServer(t, contentRoot)

	resp, err := http.Get(testUrl + "/index.html")

	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 404 {
		t.Fatalf("Expected a 404 response for a non-existent page, received %d", resp.StatusCode)
	}
}

func TestGetForUnreadablePageGenerates403(t *testing.T) {
	t.Parallel()

	contentRoot := t.TempDir()

	// fstest.MapFS does not enforce file permissions, so we need a real file system in this test.
	err := os.WriteFile(contentRoot+"/index.html", []byte(""), 0o222)
	if err != nil {
		t.Fatal(err)
	}
	testUrl := startAndrewServer(t, os.DirFS(contentRoot))

	resp, err := http.Get(testUrl + "/index.html")

	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 403 {
		t.Fatalf("Expected a 403 response for a page without permission, received %d", resp.StatusCode)
	}
}

func TestGetPagesWithoutSpecifyingPageDefaultsToIndexHtml(t *testing.T) {
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

	testUrl := startAndrewServer(t, contentRoot)

	resp, err := http.Get(testUrl)

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

func TestGetPagesCanRetrieveOtherPages(t *testing.T) {
	t.Parallel()

	contentRoot := fstest.MapFS{
		"page.html": &fstest.MapFile{Data: []byte("some text")},
	}

	testUrl := startAndrewServer(t, contentRoot)

	resp, err := http.Get(testUrl + "/page.html")

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

func TestAndrewIndexBodyIsGeneratedCorrectlyInTopLevelIndexHtmlPage(t *testing.T) {
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

	testUrl := startAndrewServer(t, contentRoot)
	resp, err := http.Get(testUrl + "/index.html")

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

	testUrl := startAndrewServer(t, contentRoot)

	resp, err := http.Get(testUrl + "/parentDir/index.html")

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

func TestCorrectMimeTypeIsSetForCommonFileTypes(t *testing.T) {
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

	testUrl := startAndrewServer(t, contentRoot)

	for page := range contentRoot {
		resp, err := http.Get(testUrl + "/" + page)

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

// startAndrewServer starts an andrew and returns the localhost url that you can run http gets against
// to retrieve data from that server
func startAndrewServer(t *testing.T, contentRoot fs.FS) string {
	t.Helper()

	testPort, testUrl := getTestPortAndUrl(t)
	go func() {
		//how can I get a random free port here for the server to start on, and return it for the tests
		//add a server object to track this datum and for convenience methods like "shut down the server".
		err := andrew.ListenAndServe(contentRoot, testPort, testUrl)
		if err != nil {
			panic(err)
		}
	}()

	return testUrl
}

// getTestPortAndUrl creates a net.Listen to retrieve a random, currently open port on the system.
// It then closes the net.Listen, as andrew will want to bind to the discovered port, but returns
// a preformatted localhost url with the new port as a test convenience.
func getTestPortAndUrl(t *testing.T) (string, string) {
	t.Helper()

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}

	listener.Close()
	url := fmt.Sprintf("http://localhost:%d/", listener.Addr().(*net.TCPAddr).Port)
	return listener.Addr().String(), url
}
