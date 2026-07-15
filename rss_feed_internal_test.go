package andrew

import (
	"testing"
)

// TestNormaliseRssDir validates that different classes of input for normaliseRssDir are validated.
// This leaks ever so slightly the interface to fs.FS, because "." is a magic string in that context
// which means "the root of the fs.FS", but that's not my major concern today.
func TestNormaliseRssDir(t *testing.T) {
	tests := []struct {
		name        string
		rssdir      string
		contentRoot string
		want        string
	}{
		{name: "relative, default root", rssdir: "foo", contentRoot: ".", want: "foo"},
		{name: "leading slash stripped", rssdir: "/foo", contentRoot: ".", want: "foo"},
		{name: "absolute root + absolute dir", rssdir: "/site/foo", contentRoot: "/site", want: "foo"},
		{name: "rss dir is the content root", rssdir: "/site", contentRoot: "/site", want: "."},
		{name: "both default", rssdir: ".", contentRoot: ".", want: "."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normaliseRssDir(tt.rssdir, tt.contentRoot); got != tt.want {
				t.Errorf("normaliseRssDir(%q, %q) = %q, want %q", tt.rssdir, tt.contentRoot, got, tt.want)
			}
		})
	}
}

func TestResolveRssDirRejectsAnRssDirThatIsNotADirectoryInTheSite(t *testing.T) {
	tests := []struct {
		name   string
		rssDir string
	}{
		{name: "a directory that is not in the site at all", rssDir: "does-not-exist"},
		{name: "an rss dir that is a file rather than a directory", rssDir: "blog/newest.html"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := resolveRssDir(testSiteFiles(), tt.rssDir, ".")
			if err == nil {
				t.Fatalf("expected an error for rss dir %q, got nil", tt.rssDir)
			}
		})
	}
}
