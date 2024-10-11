package andrew

import (
	"bytes"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"
)

func (a Server) ServeRssFeed(w http.ResponseWriter, r *http.Request) {
	rss := GenerateRssFeed(a.SiteFiles, a.BaseUrl, a.RssTitle, a.RssDescription)

	w.WriteHeader(http.StatusOK)
	_, err := fmt.Fprint(w, string(rss))
	if err != nil {
		panic(err)
	}
}

// The RSS format's pretty simple.
// First we add a constant header identifying the vesion of the RSS feed.
// Then we add the "channel" information. A "channel" is this RSS document.
// Inside the "channel", we add all of the "items".
// For Andrew, an "item" is synonymous with a page that is not an index.html page.
// Finally, we close the channel.
// It's sort of an anachronistic site to visit, but https://www.rssboard.org/rss-specification is the reference for
// what I'm including in these items and the channel.
func GenerateRssFeed(f fs.FS, baseUrl string, rssChannelTitle string, rssChannelDescription string) []byte {
	buff := new(bytes.Buffer)
	rssUrl := baseUrl + "/rss.xml"

	const (
		header = `<?xml version="1.0"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
<channel>
`
		footer = `</channel>
`
	)

	pages, err := getPages(f)

	if err != nil {
		panic(err)
	}

	fmt.Fprint(buff, header)

	fmt.Fprintf(buff, "\t<title>%s</title>\n"+
		"\t<link>%s</link>\n"+
		"\t<description>%s</description>\n"+
		"\t<generator>Andrew</generator>\n", rssChannelTitle, baseUrl, rssChannelDescription)

	for _, page := range pages {
		fmt.Fprintf(buff, "\t<item>\n"+
			"\t\t<title>%s</title>\n"+
			"\t\t<link>%s</link>\n"+
			"\t\t<pubDate>%s</pubDate>\n"+
			"\t\t<source url=\"%s\">%s</source>\n"+
			"\t</item>\n", page.Title, baseUrl+"/"+page.UrlPath, page.PublishTime.Format(time.RFC1123Z), rssUrl, rssChannelTitle)
	}

	fmt.Fprint(buff, footer)

	return buff.Bytes()
}

func getPages(siteFiles fs.FS) ([]Page, error) {
	pages := []Page{}
	localContentRoot := path.Dir(".")

	err := fs.WalkDir(siteFiles, localContentRoot, func(path string, d fs.DirEntry, err error) error {
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

		title, err := getTitle(path, pageContent)
		if err != nil {
			return err
		}

		publishTime, err := getPublishTime(siteFiles, path, pageContent)
		if err != nil {
			return err
		}

		// links require a URL relative to the page we're discovering siblings from, not from
		// the root of the file system
		s_page := Page{
			Title:       title,
			UrlPath:     strings.TrimPrefix(path, localContentRoot+"/"),
			Content:     string(pageContent),
			PublishTime: publishTime,
		}

		pages = append(pages, s_page)

		return nil
	})

	pages = SortPagesByDate(pages)

	return pages, err

}
