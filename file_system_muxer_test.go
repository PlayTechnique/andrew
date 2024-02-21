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

	server := andrew.FileSystemMuxer{ContentRoot: "."}

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
