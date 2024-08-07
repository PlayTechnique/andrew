package andrew_test

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"testing"
	"testing/fstest"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/playtechnique/andrew"
)

// TestArticlesOrderInAndrewTableOfContentsIsOverridable is verifying that
// when a page contains an andrew-publish-time meta element then the list of links andrew
// generates for the {{.AndrewTableOfContents}} are
// sorted by the meta element, then the mtime, not using the ascii sorting order.
// This test requires having several files which are in one order when sorted
// by modtime and in another order by andrew-publish-time time, so that we can tell
// what file attribute andrew is actually sorting on.
func TestArticlesOrderInAndrewTableOfContentsIsOverridable(t *testing.T) {
	expected, err := regexp.Compile(".*b_newest.html.*c_newer.html.*a_older.html.*")
	if err != nil {
		t.Fatal(err)
	}

	contentRoot := t.TempDir()

	err = os.WriteFile(contentRoot+"/index.html", []byte("{{.AndrewTableOfContents}}"), 0o700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(contentRoot+"/a_older.html", []byte{}, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(contentRoot+"/c_newer.html", []byte{}, 0o700)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()

	newest := now.Add(24 * time.Hour)
	content := fmt.Sprintf(`<meta name="andrew-publish-time" content="%s">`, newest.Format("2006-01-02"))

	err = os.WriteFile(contentRoot+"/b_newest.html", []byte(content), 0o700)
	if err != nil {
		t.Fatal(err)
	}

	older := now.Add(-24 * time.Hour)

	os.Chtimes(contentRoot+"/a_older.html", older, older)
	os.Chtimes(contentRoot+"/b_newest.html", older, older)
	os.Chtimes(contentRoot+"/c_newer.html", now, now)

	server := andrew.Server{SiteFiles: os.DirFS(contentRoot)}

	page, err := andrew.NewPage(server, "index.html")

	if err != nil {
		t.Fatal(err)
	}

	received := page.Content

	if expected.FindString(received) == "" {
		t.Errorf(cmp.Diff(expected, received))
	}
}

// TestInvalidMetaContentDoesNotCrashTheWebServer checks that if there's
// garbage data inside a meta element named andrew-publish-at that we do
// something sensible rather than crashing the web server and emitting a 502.
func TestInvalidAndrewPublishTimeContentDoesNotCrashTheWebServer(t *testing.T) {
	t.Parallel()

	contentRoot := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte(`
<!doctype HTML>
<head> </head>
<body>
{{ .AndrewTableOfContents }}
</body>
`)},
		"a.html": &fstest.MapFile{Data: []byte(`
<!doctype HTML>
<head>
<meta name="andrew-publish-time" content="<no value>"
</head>
`)},
	}

	s := newTestAndrewServer(t, contentRoot)

	resp, err := http.Get(s.BaseUrl)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected http 200, received %d", resp.StatusCode)
	}
}

// TestArticlesOrderInAndrewTableOfContentsIsOverridable is verifying that
// when a page contains an andrew-publish-time meta element then the list of links andrew
// generates for the {{.AndrewTableOfContents}} are
// sorted by the meta element, then the mtime, not using the ascii sorting order.
// This test requires having several files which are in one order when sorted
// by modtime and in another order by andrew-publish-time time, so that we can tell
// what file attribute andrew is actually sorting on.
// func TestArticlesAppearUnderParentDirectoryForAndrewTableOfContentsGrouped(t *testing.T) {
// 	expected := `<a class="andrewtableofcontentslink" id="andrewtableofcontentslink1" href="parentDir/displayMe.html">parentDir/displayMe.html</a>` +
// 		`<a class="andrewtableofcontentslink" id="andrewtableofcontentslink2" href="parentDir/childDir/1-2-3.html">parentDir/childDir/1-2-3.html</a>`

// 	contentRoot := fstest.MapFS{
// 		"groupedContents.html": &fstest.MapFile{Data: []byte(`
// {{ .AndrewTableOfContentsGrouped}}`)},

// 		"parentDir/index.html": &fstest.MapFile{Data: []byte(`
// 	<!doctype HTML>
// 	<head> </head>
// 	<body>
// 	</body>
// 	`)},
// 		"parentDir/displayme.html": &fstest.MapFile{Data: []byte(`
// 	<!doctype HTML>
// 	<head> </head>
// 	<body>
// 	</body>
// 	`)},
// 		"parentDir/childDir/1-2-3.html": &fstest.MapFile{Data: []byte(`
// 	<!doctype HTML>
// 	<head>
// 	<title>1-2-3 Page</title>
// 	</head>
// 	`)},
// 	}

// 	s := newTestAndrewServer(t, contentRoot)
// 	resp, err := http.Get(s.BaseUrl + "/groupedContents.html")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if resp.StatusCode != http.StatusOK {
// 		t.Fatalf("unexpected status %q", resp.Status)
// 	}
// 	received, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if expected != string(received) {
// 		t.Errorf(cmp.Diff(expected, received))
// 	}
// }
