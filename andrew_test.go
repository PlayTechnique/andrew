package andrew_test

import (
	"github.com/playtechnique/andrew"
	"io"
	"net/http"
	"os"
	"slices"
	"testing"
)

func TestRetrieveEmptySetWhenOnlyIndexHtml(t *testing.T) {
	testDir := t.TempDir()

	contentRoot := testDir + "/onlyIndexHtml"
	os.Mkdir(contentRoot, os.ModePerm)
	os.WriteFile(contentRoot+"/index.html", []byte{}, os.ModePerm)

	expected := []string{}
	actual, err := andrew.GetLinks(contentRoot)

	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(expected, actual) {
		t.Fatalf("expected to generate %s actually received %s", expected, actual)
	}

}

func TestGenerateLinkToOneFile(t *testing.T) {
	testDir := t.TempDir()

	contentRoot := "website"
	absPath := testDir + "/" + contentRoot
	err := os.Mkdir(absPath, os.ModePerm)

	if err != nil {
		t.Fatal(err)
	}

	os.WriteFile(absPath+"/index.html", []byte{}, os.ModePerm)
	os.WriteFile(absPath+"/somearticle.html", []byte{}, os.ModePerm)
	os.WriteFile(absPath+"/main.css", []byte{}, os.ModePerm)
	os.WriteFile(absPath+"/main.js", []byte{}, os.ModePerm)

	expected := []string{"<a href=somearticle.html>somearticle</a>"}

	err = os.Chdir(testDir)
	if err != nil {
		t.Fatal(err)
	}

	actual, err := andrew.GetLinks(contentRoot)

	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(expected, actual) {
		t.Fatalf("expected to generate %s actually received %s", expected, actual)
	}

}

// TODO: reduce duplication in these tests
func TestListenAndServeReturnsHelloWorld(t *testing.T) {
	t.Parallel()

	go func() {
		err := andrew.ListenAndServe(":8080")
		if err != nil {
			t.Fatal(err)
		}
	}()

	resp, err := http.Get("http://localhost:8080/")

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

	if !slices.Equal(received, []byte("hello, world!")) {
		t.Fatalf("Expected 'hello, world!', received %q", received)
	}
}

func TestListenAndServeServesMultiplePages(t *testing.T) {
	t.Parallel()

	go func() {
		err := andrew.ListenAndServe(":8081")
		if err != nil {
			t.Fatal(err)
		}
	}()

	resp, err := http.Get("http://localhost:8081/pageone")

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

	if !slices.Equal(received, []byte("this is page one")) {
		t.Fatalf("Expected 'this is page one', received %q", received)
	}

	resp, err = http.Get("http://localhost:8081/pagetwo")

	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status code 200, got %s", resp.Status)
	}

	received, err = io.ReadAll(resp.Body)

	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(received, []byte("this is page two")) {
		t.Fatalf("Expected 'this is page two', received %q", received)
	}
}
