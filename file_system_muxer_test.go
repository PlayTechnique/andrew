package andrew_test

import (
	"bytes"
	"io"
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

	fixtureDir := t.TempDir()
	defer os.RemoveAll(fixtureDir)
	err := os.Chdir(fixtureDir)
	if err != nil {
		t.Fatal(err)
	}

	i, err := os.Create(fixtureDir + "/index.html")
	if err != nil {
		t.Fatal(err)
	}

	_, err = i.Write(expected)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		err := andrew.ListenAndServe(":8082", fixtureDir)
		if err != nil {
			panic(err)
		}
	}()

	resp, err := http.Get("http://localhost:8082/index.html")

	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code 200, got %s", resp.Status)
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

	fixtureDir := t.TempDir()
	defer os.RemoveAll(fixtureDir)

	err := os.Chdir(fixtureDir)
	if err != nil {
		t.Fatal(err)
	}

	i, err := os.Create(fixtureDir + "/index.html")
	if err != nil {
		t.Fatal(err)
	}

	_, err = i.Write(expected)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		err := andrew.ListenAndServe(":8083", ".")
		if err != nil {
			panic(err)
		}
	}()

	resp, err := http.Get("http://localhost:8083/")

	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code 200, got %s", resp.Status)
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

	fixtureDir := newFixtureDir(t)

	startAndrewServer(fixtureDir)

	resp, err := http.Get("http://localhost:8084/page.html")

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

func startAndrewServer(fixtureDir string) {
	go func() {
		//how can I get a random free port here for the server to start on, and return it for the tests
		//add a server object to track this datum and for convenience methods like "shut down the server".
		err := andrew.ListenAndServe(":8084", fixtureDir)
		if err != nil {
			panic(err)
		}
	}()
}

func newFixtureDir(t *testing.T) string {
	fixtureDir := t.TempDir()

	_, err := os.Create(fixtureDir + "/index.html")
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(fixtureDir+"/page.html", []byte("some text"), 0o755)
	if err != nil {
		t.Fatal(err)
	}
	return fixtureDir
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
<title>1-2-3 Page</a>
</head>
	`), 0o700)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		err := andrew.ListenAndServe(":9091", contentRoot)
		if err != nil {
			panic(err)
		}
	}()

	resp, err := http.Get("http://localhost:9091/index.html")

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
		t.Error(cmp.Diff(expectedIndex, received))
	}
}
