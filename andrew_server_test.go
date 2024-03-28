package andrew_test

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"path/filepath"
	"slices"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"
	"github.com/playtechnique/andrew"
)

func TestGetPages(t *testing.T) {
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
		"index.html": &fstest.MapFile{Data: expected, Mode: 0o755},
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

func TestGetPagesDefaultsToIndexHtml(t *testing.T) {
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
		"index.html": &fstest.MapFile{Data: expected, Mode: 0o755},
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
		"page.html": &fstest.MapFile{Data: []byte("some text"), Mode: 0o755},
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

func TestIndexBodyFromTopLevelIndexHtmlPage(t *testing.T) {
	t.Parallel()

	contentRoot := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte(`
<!doctype HTML>
<head> </head>
<body>
{{ .AndrewIndexBody }}
</body>
`), Mode: 0o755},
		"pages/1-2-3.html": &fstest.MapFile{Data: []byte(`
<!doctype HTML>
<head>
<title>1-2-3 Page</title>
</head>
`), Mode: 0o755},
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

func TestIndexBodyFromADirectoryTwoLevelsDown(t *testing.T) {
	t.Parallel()

	contentRoot := fstest.MapFS{
		"parentDir/index.html": &fstest.MapFile{Data: []byte(`
<!doctype HTML>
<head> </head>
<body>
{{ .AndrewIndexBody }}
</body>
`), Mode: 0o755},
		"parentDir/childDir/1-2-3.html": &fstest.MapFile{Data: []byte(`
<!doctype HTML>
<head>
<title>1-2-3 Page</title>
</head>
`), Mode: 0o755},
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

func TestMineTypesAreSetCorrectly(t *testing.T) {
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

	for page, _ := range contentRoot {
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
