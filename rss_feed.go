package andrew

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func (a Server) ServeRssFeed(w http.ResponseWriter, r *http.Request) {
	rss, err := GenerateRssFeed(a.SiteFiles, a.BaseUrl, a.RssInfo)
	if err != nil {
		message, status := CheckPageErrors(err)
		w.WriteHeader(status)
		fmt.Fprint(w, message)
		return
	}

	w.WriteHeader(http.StatusOK)

	// The response is already on the wire, so there is no status left to set. A client that
	// hangs up mid-write is routine rather than exceptional, so log it and move on.
	if _, err := fmt.Fprint(w, string(rss)); err != nil {
		slog.Info("could not finish writing the rss feed", "error", err)
	}
}

// The RSS format's pretty simple.
// First we add a constant header identifying the version of the RSS feed.
// Then we add the "channel" information. A "channel" is this RSS document.
// Inside the "channel", we add all of the "items".
// For Andrew, an "item" is synonymous with a page that is not an index.html page.
// Finally, we close the channel.
// It's sort of an anachronistic site to visit, but https://www.rssboard.org/rss-specification is the reference for
// what I'm including in these items and the channel.
// Args:
// 1. an fs.FS which contains your full site
// 2. your baseURl, which is interpolated into the rss feed.
// 3. an RssInfo structure, which contains some information that is needed by your RSS feed.
func GenerateRssFeed(f fs.FS, baseUrl string, rss RssInfo) ([]byte, error) {
	buff := new(bytes.Buffer)
	rssUrl := baseUrl + "/rss.xml"

	const (
		header = `<?xml version="1.0"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
<channel>
`
		footer = `</channel>
</rss>
`
	)
	pages, err := pagesInDir(f, rss.Dir)
	if err != nil {
		return nil, err
	}

	pages = SortPagesByDate(pages)

	fmt.Fprint(buff, header)

	fmt.Fprintf(buff, "\t<title>%s</title>\n"+
		"\t<link>%s</link>\n"+
		"\t<description>%s</description>\n"+
		"\t<generator>Andrew</generator>\n", rss.Title, baseUrl, rss.Description)

	for _, page := range pages {
		fmt.Fprintf(buff, "\t<item>\n"+
			"\t\t<title>%s</title>\n"+
			"\t\t<link>%s</link>\n"+
			"\t\t<pubDate>%s</pubDate>\n"+
			"\t\t<source url=\"%s\">%s</source>\n"+
			"\t</item>\n", page.Title, baseUrl+"/"+page.UrlPath, page.PublishTime.Format(time.RFC1123Z), rssUrl, rss.Title)
	}

	fmt.Fprint(buff, footer)

	return buff.Bytes(), nil
}

// resolveRssDir turns the rss directory as the end user typed it on the command line into a
// path inside siteFiles, and confirms that directory is really there.
// Resolving at startup means a typo'd --rssdir fails immediately, rather than serving a
// broken feed later on.
func resolveRssDir(siteFiles fs.FS, rssDir string, contentRoot string) (string, error) {
	rssDir = normaliseRssDir(rssDir, contentRoot)

	if err := checkRssDirExists(siteFiles, rssDir, contentRoot); err != nil {
		return "", err
	}

	return rssDir, nil
}

// The rss directory is always a directory inside SiteFiles. But the end-user might supply it as an absolute path.
// If the end user does submit it as an absolute path, we need to make it relative to the sitefiles root, which can be either
// an absolute path or a relative path.
// And while we're at it, the difference between the directory being "foo" and "foo/" trips up the fs.FS.
// As does the difference betwee /foo and foo; and it is worth noting that trimming the contentRoot from the rss directory will
// often leave a leading / behind.
// After all of that, what if both contentRoot and rssDir were the pwd i.e. "."? That would mean that the end result of our machinations
// would be an empty string, which also would not do.
func normaliseRssDir(rd string, contentRoot string) string {
	//	if baseOfSiteFiles contains a leading / then it's fine; otherwise, turn it into a path relative to the base of sitefiles
	rd = strings.TrimSuffix(rd, "/")
	rd = strings.TrimPrefix(rd, contentRoot)
	rd = strings.TrimPrefix(rd, "/")
	if rd == "" {
		rd = "."
	}
	return rd
}

// checkRssDirExists confirms that rssDir, which must already have been through
// normaliseRssDir, is a directory that siteFiles actually contains. contentRoot is only
// here so that a missing directory can say where we went looking for it.
func checkRssDirExists(siteFiles fs.FS, rssDir string, contentRoot string) error {
	rssDirInfo, err := fs.Stat(siteFiles, rssDir)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("rss directory %q must be a directory inside the content root %s: %w", rssDir, contentRoot, err)
		}
		return fmt.Errorf("rss directory %q: %w", rssDir, err)
	}

	if !rssDirInfo.IsDir() {
		return fmt.Errorf("rss directory %q is not a directory", rssDir)
	}

	return nil
}
