package andrew_test

import (
	"github.com/playtechnique/andrew"
	"io"
	"net/http/httptest"
	"testing"
)

func TestWhenADirectoryContainsNoFilesItsParentGetsNothing(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	andrew.ServeRoot(rec, req)
	res := rec.Result()
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}
	if string(data) != "ABC" {
		t.Errorf("expected ABC got %v", string(data))
	}

}

func TestWhenADirectoryContainsOneFileItsParentGetsAnIndexHtmlContainingIt(t *testing.T) {

}

func TestWhenADirectoryContainsManyFilesItsParentGetsAnIndexHtmlContainingThem(t *testing.T) {

}
