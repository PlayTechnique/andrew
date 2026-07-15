package andrew_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/google/go-cmp/cmp"
	"github.com/playtechnique/andrew"
)

func TestGenerateRssFeedIncludesRequiredElements(t *testing.T) {
	expected := []byte(`<?xml version="1.0"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
<channel>
	<title>PlayTechnique</title>
	<link>http://localhost:8080</link>
	<description>Learning to play better.</description>
	<generator>Andrew</generator>
	<item>
		<title>page.html</title>
		<link>http://localhost:8080/page.html</link>
		<pubDate>Mon, 01 Jan 0001 00:00:00 +0000</pubDate>
		<source url="http://localhost:8080/rss.xml">PlayTechnique</source>
	</item>
</channel>
</rss>
`)

	testFs := fstest.MapFS{
		"index.html": {},
		"page.html":  {},
	}

	rssInfo := andrew.RssInfo{Title: "PlayTechnique", Dir: ".", Description: "Learning to play better."}

	baseUrl := "http://localhost:8080"

	feed, err := andrew.GenerateRssFeed(testFs, baseUrl, rssInfo)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(feed, expected) {
		t.Error(cmp.Diff(expected, feed))
	}

	if !bytes.Contains(feed, []byte(rssInfo.Description)) {
		t.Errorf("Expected feed to contain description %s but it does not", rssInfo.Description)
	}

	if !bytes.Contains(feed, []byte(rssInfo.Title)) {
		t.Errorf("Expected feed to contain description %s but it does not", rssInfo.Title)
	}
}

func TestGenerateRssFeedLinksToPagesInTheRssDirCorrectly(t *testing.T) {
	expected := []byte(`<?xml version="1.0"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
<channel>
	<title>PlayTechnique</title>
	<link>http://localhost:8080</link>
	<description>Learning to play better.</description>
	<generator>Andrew</generator>
	<item>
		<title>barpage.html</title>
		<link>http://localhost:8080/foo/barpage.html</link>
		<pubDate>Mon, 01 Jan 0001 00:00:00 +0000</pubDate>
		<source url="http://localhost:8080/rss.xml">PlayTechnique</source>
	</item>
	<item>
		<title>foopage.html</title>
		<link>http://localhost:8080/foo/foopage.html</link>
		<pubDate>Mon, 01 Jan 0001 00:00:00 +0000</pubDate>
		<source url="http://localhost:8080/rss.xml">PlayTechnique</source>
	</item>
</channel>
</rss>
`)

	testFs := fstest.MapFS{
		"index.html":       {},
		"page.html":        {},
		"foo/foopage.html": {},
		"foo/barpage.html": {},
	}

	// GenerateRssFeed takes a Dir that is already normalised; turning "/foo" into "foo" is
	// option parsing's job.
	rssInfo := andrew.RssInfo{Title: "PlayTechnique", Dir: "foo", Description: "Learning to play better."}

	baseUrl := "http://localhost:8080"

	feed, err := andrew.GenerateRssFeed(testFs, baseUrl, rssInfo)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(feed, expected) {
		t.Error(cmp.Diff(expected, feed))
	}
}

// TestGenerateRssFeedReturnsAnErrorRatherThanPanicking uses an rss dir that isn't in the
// site. Main resolves the rss dir at startup so a running andrew won't reach this, but
// GenerateRssFeed is exported and a panic here used to escape into an http handler, where it
// cost the client its connection instead of an http status.
func TestGenerateRssFeedReturnsAnErrorRatherThanPanicking(t *testing.T) {
	t.Parallel()

	rssInfo := andrew.RssInfo{Title: "PlayTechnique", Dir: "does-not-exist", Description: "Learning to play better."}

	_, err := andrew.GenerateRssFeed(fstest.MapFS{"index.html": {}}, "http://localhost:8080", rssInfo)
	if err == nil {
		t.Fatal("expected an error for an rss dir that is not in the site, got nil")
	}
}

// TestServeRssFeedReturnsAnHttpErrorWhenTheFeedCannotBeGenerated is the handler-level
// contract: a feed that cannot be built has to reach the client as a status code. Before
// GenerateRssFeed returned errors, this panicked before WriteHeader, so net/http recovered
// it and closed the connection, and the client saw an empty reply rather than an error.
func TestServeRssFeedReturnsAnHttpErrorWhenTheFeedCannotBeGenerated(t *testing.T) {
	t.Parallel()

	rssInfo := andrew.RssInfo{Title: "PlayTechnique", Dir: "does-not-exist", Description: "Learning to play better."}
	s := andrew.NewServer(fstest.MapFS{"index.html": {}}, ":0", "http://localhost:8080", rssInfo)

	w := httptest.NewRecorder()
	s.ServeRssFeed(w, httptest.NewRequest(http.MethodGet, "/rss.xml", nil))

	// CheckPageErrors maps the fs.ErrNotExist from the walk onto a 404.
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}
