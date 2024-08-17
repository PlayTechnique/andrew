package andrew_test

import (
	"fmt"
	"io"
	"net/http"
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
	expected, err := regexp.Compile("(?s).*b_newest.html.*c_newer.html.*a_older.html.*")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	newer := now.Add(24 * time.Hour)
	newest := now.Add(48 * time.Hour)

	contentRoot := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte(`
{{ .AndrewTableOfContents }}
`)},
		"a_older.html":  &fstest.MapFile{ModTime: now},
		"b_newest.html": &fstest.MapFile{ModTime: newest, Data: []byte(fmt.Sprintf(`<meta name="andrew-publish-time" content="%s">`, newest.Format("2006-01-02")))},
		"c_newer.html":  &fstest.MapFile{ModTime: newer},
	}

	server := andrew.Server{SiteFiles: contentRoot}

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
func TestOneArticleAppearsUnderParentDirectoryForAndrewTableOfContentsWithDirectories(t *testing.T) {
	expected := `<div class="AndrewTableOfContentsWithDirectories">
<ul>
<li><a class="andrewtableofcontentslink" id="andrewtableofcontentslink0" href="otherPage.html">otherPage.html</a> - <span class="andrew-page-publish-date">0001-01-01</span></li>
</ul>
</div>
`
	contentRoot := fstest.MapFS{
		"groupedContents.html": &fstest.MapFile{Data: []byte(`{{.AndrewTableOfContentsWithDirectories}}`)},
		"otherPage.html":       &fstest.MapFile{},
	}

	s := newTestAndrewServer(t, contentRoot)
	resp, err := http.Get(s.BaseUrl + "/groupedContents.html")
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %q", resp.Status)
	}
	received, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if expected != string(received) {

		t.Errorf("Expected:\n" + expected + "\n Received:\n" + string(received))

	}
}

func TestArticlesFromChildDirectoriesAreShownForAndrewTableOfContentsWithDirectories(t *testing.T) {
	expected := `<div class="AndrewTableOfContentsWithDirectories">
<ul>
<h5>parentDir/</h5>
<li><a class="andrewtableofcontentslink" id="andrewtableofcontentslink0" href="parentDir/displayme.html">displayme.html</a> - <span class="andrew-page-publish-date">0001-01-01</span></li>
</ul>
<ul>
<h5><span class="AndrewParentDir">parentDir/</span>childDir/</h5>
<li><a class="andrewtableofcontentslink" id="andrewtableofcontentslink1" href="parentDir/childDir/1-2-3.html">1-2-3.html</a> - <span class="andrew-page-publish-date">0001-01-01</span></li>
</ul>
</div>
`
	contentRoot := fstest.MapFS{
		"groupedContents.html":          &fstest.MapFile{Data: []byte(`{{.AndrewTableOfContentsWithDirectories}}`)},
		"parentDir/index.html":          &fstest.MapFile{},
		"parentDir/styles.css":          &fstest.MapFile{},
		"parentDir/displayme.html":      &fstest.MapFile{},
		"parentDir/childDir/1-2-3.html": &fstest.MapFile{},
	}

	s := newTestAndrewServer(t, contentRoot)
	resp, err := http.Get(s.BaseUrl + "/groupedContents.html")
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %q", resp.Status)
	}
	received, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if expected != string(received) {
		t.Errorf("Expected:\n" + expected + "\n Received:\n" + string(received))
	}
}
