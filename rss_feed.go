package andrew

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func (a Server) ServeRssFeed(w http.ResponseWriter, r *http.Request) {
	rss := GenerateRssFeed(a.SiteFiles, a.ContentRoot, a.BaseUrl, a.RssInfo)

	w.WriteHeader(http.StatusOK)
	_, err := fmt.Fprint(w, string(rss))
	if err != nil {
		panic(err)
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
// 2. the path which your site is fed from. This is used in some path munging against the RSS directory.
// 3. your baseURl, which is interpolated into the rss feed.
// 4. an RssInfo structure, which contains some information that is needed by your RSS feed.
func GenerateRssFeed(f fs.FS, contentRoot string, baseUrl string, rss RssInfo) []byte {
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
	rssDir := normaliseRssDir(rss.Dir, contentRoot)
	pages, err := getPages(f, rssDir)

	if err != nil {
		panic(err)
	}

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

	return buff.Bytes()
}

func getPages(siteFiles fs.FS, startDir string) ([]Page, error) {
	pages := []Page{}

	slog.Debug("getPages", "startDir", startDir)

	err := fs.WalkDir(siteFiles, startDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// We don't list index files in our collection of siblings and children, because I don't
		// want a link back to a page that contains only links.
		if strings.Contains(path, "index.html") {
			return nil
		}

		// If the file we're considering isn't an html file, let's move on with our day.
		if !strings.Contains(path, "html") {
			return nil
		}

		pageContent, err := fs.ReadFile(siteFiles, path)
		if err != nil {
			return err
		}

		// Render partials before extracting metadata, so meta tags inside partials are found
		renderedContent, err := renderPartialFiles(siteFiles, path, pageContent)
		if err != nil {
			return err
		}

		title, err := getTitle(path, renderedContent)
		if err != nil {
			return err
		}

		publishTime, err := getPublishTime(siteFiles, path, renderedContent)
		if err != nil {
			return err
		}

		// links require a URL relative to the page we're discovering siblings from, not from
		// the root of the file system
		s_page := Page{
			Title:       title,
			UrlPath:     path,
			Content:     string(renderedContent),
			PublishTime: publishTime,
		}

		pages = append(pages, s_page)

		return nil
	})

	pages = SortPagesByDate(pages)

	return pages, err

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
