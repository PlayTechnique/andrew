package andrew_test

import (
	"github.com/playtechnique/andrew"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"testing"
)

// TestServeIndexHtmlIfItExists will check an incoming request. If it is for a directory,
// and that directory contains an index.html, that index.html will be served.
func TestServeIndexHtmlIfItExists(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	contentRoot := "testdata/onlyIndexHtml"
	expected, err := os.ReadFile(contentRoot + "/index.html")

	if err != nil {
		t.Fatalf("Could not find test data: %v", err)
	}

	server := andrew.Server{ContentRoot: contentRoot, HttpServer: http.FileServer(http.Dir(contentRoot))}
	server.ServeUp(rec, req)

	res := rec.Result()
	defer res.Body.Close()

	indexContents, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}

	actual := indexContents

	if !slices.Equal(expected, actual) {
		t.Errorf("\nexpected:\n %s \nactual:\n %s", string(expected), string(actual))
	}
}

// TestServeIndexHtmlIfItExists will check an incoming request for a corresponding file
// on the file system and serve it.
// func TestServeHtmlIfItExists(t *testing.T) {
// 	rec := httptest.NewRecorder()
// 	req := httptest.NewRequest("GET", "/", nil)
// 	andrew.ServeUp(rec, req)
// 	res := rec.Result()
// 	defer res.Body.Close()
//
// 	data, err := io.ReadAll(res.Body)
// 	if err != nil {
// 		t.Errorf("expected error to be nil got %v", err)
// 	}
// 	if string(data) != "ABC" {
// 		t.Errorf("expected ABC got %v", string(data))
// 	}
// }

// func TestWhenADirectoryContainsNoFilesItsParentGetsNothing(t *testing.T) {
// 	rec := httptest.NewRecorder()
// 	req := httptest.NewRequest("GET", "/", nil)
// 	andrew.ServeRoot(rec, req)
// 	res := rec.Result()
// 	defer res.Body.Close()
//
// 	data, err := io.ReadAll(res.Body)
// 	if err != nil {
// 		t.Errorf("expected error to be nil got %v", err)
// 	}
// 	if string(data) != "ABC" {
// 		t.Errorf("expected ABC got %v", string(data))
// 	}
//
// }
//
// func TestWhenADirectoryContainsOneFileItsParentGetsAnIndexHtmlContainingIt(t *testing.T) {
//
// }
//
// func TestWhenADirectoryContainsManyFilesItsParentGetsAnIndexHtmlContainingThem(t *testing.T) {
//
// }
