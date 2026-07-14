package andrew_test

import (
	"bytes"
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

	feed := andrew.GenerateRssFeed(testFs, ".", baseUrl, rssInfo)

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

	rssDirectories := []string{"foo", "/foo"}

	for _, rootRssDir := range rssDirectories {

		rssInfo := andrew.RssInfo{Title: "PlayTechnique", Dir: rootRssDir, Description: "Learning to play better."}

		baseUrl := "http://localhost:8080"

		feed := andrew.GenerateRssFeed(testFs, ".", baseUrl, rssInfo)

		if !bytes.Equal(feed, expected) {
			t.Error(cmp.Diff(expected, feed))
		}
	}
}
