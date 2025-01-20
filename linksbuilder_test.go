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

func TestAndrewTableOfContentsIsGeneratedCorrectlyInContentrootDirectory(t *testing.T) {
	t.Parallel()

	contentRoot := fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte(`
<!doctype HTML>
<head> </head>
<body>
{{ .AndrewTableOfContents }}
</body>
`)},
		"pages/1-2-3.html": &fstest.MapFile{Data: []byte(`
<!doctype HTML>
<head>
<title>1-2-3 Page</title>
</head>
`)},
	}

	s := newTestAndrewServer(t, contentRoot)

	resp, err := http.Get(s.BaseUrl + "/index.html")
	if err != nil {
		t.Fatal(err)
	}

	received, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	expected, err := regexp.Compile(".*<a class=\"andrewtableofcontentslink\" id=\"andrewtableofcontentslink0\" href=\"pages/1-2-3.html\">1-2-3 Page</a>.*")
	if err != nil {
		t.Fatal(err)
	}

	if expected.FindString(string(received)) == "" {
		t.Fatalf("Diff of Expected and Actual: %s", cmp.Diff(expected, received))
	}
}

func TestAndrewTableOfContentsIsGeneratedCorrectlyInAChildDirectory(t *testing.T) {
	t.Parallel()

	contentRoot := fstest.MapFS{
		"parentDir/index.html": &fstest.MapFile{Data: []byte(`
<!doctype HTML>
<head> </head>
<body>
{{ .AndrewTableOfContents }}
</body>
`)},
		"parentDir/childDir/1-2-3.html": &fstest.MapFile{Data: []byte(`
<!doctype HTML>
<head>
<title>1-2-3 Page</title>
</head>
`)},
	}

	s := newTestAndrewServer(t, contentRoot)

	resp, err := http.Get(s.BaseUrl + "/parentDir/index.html")
	if err != nil {
		t.Fatal(err)
	}

	received, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	expected, err := regexp.Compile(".*<a class=\"andrewtableofcontentslink\" id=\"andrewtableofcontentslink0\" href=\"childDir/1-2-3.html\">1-2-3 Page</a>")
	if err != nil {
		t.Fatal(err)
	}
	if expected.FindString(string(received)) == "" {
		t.Fatalf("Diff of Expected and Actual: %s", cmp.Diff(expected, received))
	}
}

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
		t.Error(cmp.Diff(expected, received))
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
		t.Errorf("Expected:\n%s\nReceived:\n%s", expected, string(received))
	}
}

func TestFullHTMLReturnedByAndrewTableOfContents(t *testing.T) {
	expected := `<div class="AndrewTableOfContents">
<ul>
<li><a class="andrewtableofcontentslink" id="andrewtableofcontentslink0" href="groupedContents.html">groupedContents.html</a> - <span class="andrew-page-publish-date">0001-01-01</span></li>
<li><a class="andrewtableofcontentslink" id="andrewtableofcontentslink1" href="parentDir/childDir/1-2-3.html">1-2-3.html</a> - <span class="andrew-page-publish-date">0001-01-01</span></li>
<li><a class="andrewtableofcontentslink" id="andrewtableofcontentslink2" href="parentDir/displayme.html">displayme.html</a> - <span class="andrew-page-publish-date">0001-01-01</span></li>
</ul>
</div>
`
	contentRoot := fstest.MapFS{
		"groupedContents.html":          &fstest.MapFile{Data: []byte(`{{.AndrewTableOfContents}}`)},
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
		t.Errorf("Expected:\n%s\nReceived:\n%s", expected, string(received))
	}
}

func TestAndrewTableOfContentsWithDirectoriesSortsDirectoriesByMostRecentContent(t *testing.T) {
	t.Parallel()

	now := time.Now()
	older := now.Add(-24 * time.Hour)
	newest := now.Add(24 * time.Hour)

	tests := []struct {
		name     string
		files    map[string]*fstest.MapFile
		expected string
	}{
		//In these tests, I picked the directory names so that lexicographic sorting
		//does not match the sort order by time.
		{
			name: "directories are sorted by most recent page",
			files: map[string]*fstest.MapFile{
				"index.html": {Data: []byte(`{{.AndrewTableOfContentsWithDirectories}}`)},
				"a-dir/content.html": {
					Data:    []byte(`<meta name="andrew-publish-time" content="` + older.Format("2006-01-02") + `">`),
					ModTime: older,
				},
				"y-dir/content.html": {
					Data:    []byte(`<meta name="andrew-publish-time" content="` + newest.Format("2006-01-02") + `">`),
					ModTime: newest,
				},
			},
			expected: "(?s).*<h5>y-dir/</h5>.*<h5>a-dir/</h5>.*",
		},
		{
			name: "nested directories maintain parent-child relationship and sort by newest content",
			files: map[string]*fstest.MapFile{
				"index.html": {Data: []byte(`{{.AndrewTableOfContentsWithDirectories}}`)},
				"music/rock.html": {
					Data:    []byte(`<meta name="andrew-publish-time" content="` + older.Format("2006-01-02") + `">`),
					ModTime: older,
				},
				"music/jazz/bebop.html": {
					Data:    []byte(`<meta name="andrew-publish-time" content="` + newest.Format("2006-01-02") + `">`),
					ModTime: newest,
				},
			},
			expected: "(?s).*<h5><span class=\"AndrewTableOfContentsWithDirectories\">music/</span>jazz/</h5>.*<h5>music/</h5>.*",
		},
		{
			name: "empty directories are ignored",
			files: map[string]*fstest.MapFile{
				"index.html": {Data: []byte(`{{.AndrewTableOfContentsWithDirectories}}`)},
				"y-dir/content.html": {
					Data:    []byte(`<meta name="andrew-publish-time" content="` + now.Format("2006-01-02") + `">`),
					ModTime: now,
				},
				"a-dir/": {},
			},
			expected: "(?s).*<h5>y-dir/</h5>.*",
		},
		{
			name: "directories with same newest content are sorted alphabetically",
			files: map[string]*fstest.MapFile{
				"index.html": {Data: []byte(`{{.AndrewTableOfContentsWithDirectories}}`)},
				"y-dir/content.html": {
					Data:    []byte(`<meta name="andrew-publish-time" content="` + now.Format("2006-01-02") + `">`),
					ModTime: now,
				},
				"a-dir/content.html": {
					Data:    []byte(`<meta name="andrew-publish-time" content="` + now.Format("2006-01-02") + `">`),
					ModTime: now,
				},
			},
			expected: "(?s).*<h5>a-dir/</h5>.*<h5>y-dir/</h5>.*",
		},
		{
			name: "root directory content is included in sorting",
			files: map[string]*fstest.MapFile{
				"index.html": {Data: []byte(`{{.AndrewTableOfContentsWithDirectories}}`)},
				"a-content.html": {
					Data:    []byte(`<meta name="andrew-publish-time" content="` + older.Format("2006-01-02") + `">`),
					ModTime: older,
				},
				"y-dir/content.html": {
					Data:    []byte(`<meta name="andrew-publish-time" content="` + newest.Format("2006-01-02") + `">`),
					ModTime: newest,
				},
			},
			expected: "(?s).*<h5>y-dir/</h5>.*href=\"a-content.html\".*",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := newTestAndrewServer(t, fstest.MapFS(tt.files))
			resp, err := http.Get(s.BaseUrl + "/index.html")
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

			expected, err := regexp.Compile(tt.expected)
			if err != nil {
				t.Fatal(err)
			}

			if expected.FindString(string(received)) == "" {
				t.Errorf("Expected pattern not found in response.\nExpected pattern: %s\nReceived: %s",
					tt.expected, string(received))
			}
		})
	}
}

func TestAndrewTableOfContentsUsesCorrectClasses(t *testing.T) {
	t.Parallel()

	files := map[string]*fstest.MapFile{
		"index.html": {Data: []byte(`{{.AndrewTableOfContents}}`)},
		"page.html": {
			Data:    []byte(`<meta name="andrew-publish-time" content="2024-01-01">`),
			ModTime: time.Now(),
		},
	}

	s := newTestAndrewServer(t, fstest.MapFS(files))
	resp, err := http.Get(s.BaseUrl + "/index.html")
	if err != nil {
		t.Fatal(err)
	}

	received, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	expected := "(?s)<div class=\"AndrewTableOfContents\">"
	matched, err := regexp.MatchString(expected, string(received))
	if err != nil {
		t.Fatal(err)
	}
	if !matched {
		t.Errorf("Class mismatch in div.\nExpected class: AndrewTableOfContents\nReceived output: %s", string(received))
	}
}

func TestAndrewTableOfContentsWithDirectoriesUsesCorrectClasses(t *testing.T) {
	t.Parallel()

	files := map[string]*fstest.MapFile{
		"index.html": {Data: []byte(`{{.AndrewTableOfContentsWithDirectories}}`)},
		"parent/child/page.html": {
			Data:    []byte(`<meta name="andrew-publish-time" content="2024-01-01">`),
			ModTime: time.Now(),
		},
	}

	s := newTestAndrewServer(t, fstest.MapFS(files))
	resp, err := http.Get(s.BaseUrl + "/index.html")
	if err != nil {
		t.Fatal(err)
	}

	received, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	expectedPatterns := map[string]string{
		"div":  "(?s)<div class=\"AndrewTableOfContentsWithDirectories\">",
		"span": "(?s)<h5><span class=\"AndrewTableOfContentsWithDirectories\">parent/</span>child/</h5>",
	}

	for element, pattern := range expectedPatterns {
		matched, err := regexp.MatchString(pattern, string(received))
		if err != nil {
			t.Fatal(err)
		}
		if !matched {
			t.Errorf("Class mismatch in %s.\nExpected class: AndrewTableOfContentsWithDirectories\nReceived output: %s",
				element, string(received))
		}
	}
}
