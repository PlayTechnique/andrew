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

	title := "PlayTechnique"
	baseUrl := "http://localhost:8080"
	description := "Learning to play better."

	feed := andrew.GenerateRssFeed(testFs, baseUrl, title, description)

	if !bytes.Equal(feed, expected) {
		t.Error(cmp.Diff(expected, feed))
	}

	if !bytes.Contains(feed, []byte(description)) {
		t.Errorf("Expected feed to contain description %s but it does not", description)
	}

	if !bytes.Contains(feed, []byte(title)) {
		t.Errorf("Expected feed to contain description %s but it does not", title)
	}
}
