package andrew_test

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"slices"
	"testing"

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

	contentRoot := t.TempDir()

	err := os.WriteFile(contentRoot+"/index.html", expected, 0o755)
	if err != nil {
		t.Fatal(err)
	}

	testPort, testUrl := getTestPortAndUrl(t)

	go func() {
		err := andrew.ListenAndServe(testPort, contentRoot)
		if err != nil {
			panic(err)
		}
	}()

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

	contentRoot := t.TempDir()

	err := os.WriteFile(contentRoot+"/index.html", expected, 0o755)
	if err != nil {
		t.Fatal(err)
	}

	testPort, testUrl := getTestPortAndUrl(t)

	go func() {
		err := andrew.ListenAndServe(testPort, contentRoot)
		if err != nil {
			panic(err)
		}
	}()

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

	contentRoot := t.TempDir()

	err := os.WriteFile(contentRoot+"/page.html", []byte("some text"), 0o755)
	if err != nil {
		t.Fatal(err)
	}

	testPort, testUrl := getTestPortAndUrl(t)
	go func() {
		err := andrew.ListenAndServe(testPort, contentRoot)
		if err != nil {
			panic(err)
		}
	}()

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

func TestAnIndexBodyIsBuilt(t *testing.T) {
	t.Parallel()

	contentRoot := t.TempDir()
	err := os.MkdirAll(contentRoot+"/pages", 0700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(contentRoot+"/index.html", []byte(`
<!doctype HTML>
<head> </head>
<body> 
{{ .AndrewIndexBody }}
</body>
`), 0o755)

	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(contentRoot+"/pages/1-2-3.html", []byte(`
<!doctype HTML>
<head>
<title>1-2-3 Page</title>
</head>
`), 0o700)
	if err != nil {
		t.Fatal(err)
	}

	testPort, testUrl := getTestPortAndUrl(t)

	go func() {
		err := andrew.ListenAndServe(testPort, contentRoot)
		if err != nil {
			panic(err)
		}
	}()

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
<a href="pages/1-2-3.html">1-2-3 Page</a>
</body>
`

	if !slices.Equal(received, []byte(expectedIndex)) {
		t.Fatalf("Diff of Expected and Actual: %s", cmp.Diff(expectedIndex, received))
	}
}

// func startAndrewServer(contentRoot string) {

// 	go func() {
// 		//how can I get a random free port here for the server to start on, and return it for the tests
// 		//add a server object to track this datum and for convenience methods like "shut down the server".
// 		err := andrew.ListenAndServe(":8084", contentRoot)
// 		if err != nil {
// 			panic(err)
// 		}
// 	}()
// }

func getTestPortAndUrl(t *testing.T) (string, string) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}

	listener.Close()
	url := fmt.Sprintf("http://localhost:%d/", listener.Addr().(*net.TCPAddr).Port)
	return listener.Addr().String(), url
}
