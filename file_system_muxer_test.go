package andrew_test

import (
	"github.com/playtechnique/andrew"
	"io"
	"net/http"
	"os"
	"slices"
	"testing"
)

func TestGetPages(t *testing.T) {

	expected := []byte(`
<!DOCTYPE html>
<head>
  <title>index title</title>
</head>
<body>
</body>
`)

	fixtureDir := t.TempDir()
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

	server, err := andrew.NewFileSystemMuxer(".")
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		err := andrew.ListenAndServe(":8082", server)
		if err != nil {
			t.Fatal(err)
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
	expected := []byte(`
<!DOCTYPE html>
<head>
  <title>index title</title>
</head>
<body>
</body>
`)

	fixtureDir := t.TempDir()

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

	server, err := andrew.NewFileSystemMuxer(".")
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		err := andrew.ListenAndServe(":8083", server)
		if err != nil {
			t.Fatal(err)
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
	indexPage := []byte(`
this is an index
`)
	expected := []byte(`
<!DOCTYPE html>
<head>
  <title>expected title</title>
</head>
<body>
</body>
`)

	fixtureDir := t.TempDir()
	err := os.Chdir(fixtureDir)
	if err != nil {
		t.Fatal(err)
	}

	i, err := os.Create(fixtureDir + "/index.html")
	if err != nil {
		t.Fatal(err)
	}

	_, err = i.Write(indexPage)
	if err != nil {
		t.Fatal(err)
	}

	e, err := os.Create(fixtureDir + "/expected.html")
	if err != nil {
		t.Fatal(err)
	}

	_, err = e.Write(expected)
	if err != nil {
		t.Fatal(err)
	}

	err = e.Sync()
	if err != nil {
		t.Fatal(err)
	}

	server, err := andrew.NewFileSystemMuxer(".")
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		err := andrew.ListenAndServe(":8084", server)
		if err != nil {
			t.Fatal(err)
		}
	}()

	resp, err := http.Get("http://localhost:8084/expected.html")

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
