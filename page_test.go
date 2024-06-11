package andrew

import (
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestTitleDiscoveryReturnsErrorWhenNoTitleElementInPageContent(t *testing.T) {
	_, err := titleFromHTMLTitleElement([]byte("snibble"))

	if err.Error() != "no title element found" {
		t.Fatalf("Expected error, received %v", err)
	}
}

func TestTitleElementDiscoveredWhenPresentInPageContent(t *testing.T) {
	expected := "my title"
	received, err := titleFromHTMLTitleElement([]byte("<title>" + expected + "</title>"))

	if err != nil {
		t.Fatal(err)
	}

	if received != expected {
		t.Fatal(cmp.Diff(expected, received))
	}
}

func TestGetTitleReturnsPageFileNameWhenNoTitleInDocument(t *testing.T) {
	received, err := getTitle("page title", []byte{})

	if err != nil {
		t.Fatal(err)
	}

	if received != "page title" {
		t.Fatal(cmp.Diff("", received))
	}
}

func TestMetaPopulatesATag(t *testing.T) {
	expected := []string{"andrew-created-at 2025-03-01"}
	received, err := GetMetaElements([]byte("<meta name=andrew-created-at content=2025-03-01>"))

	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(expected, received) {
		t.Fatal(cmp.Diff(expected, received))
	}
}

// func TestMetaIsPopulatedWithExpectedElements(t *testing.T) {
// 	expected := map[string]string{"andrew-created-at": "2025-03-01"}
// 	received, err := GetMetaElements([]byte("<meta name=andrew-created-at content=2025-03-01>"))

// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if received != expected {
// 		t.Fatal(cmp.Diff(expected, received))
// 	}
// }
