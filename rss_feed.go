package andrew

import (
	"bytes"
	"fmt"
	"io/fs"
)

func GenerateRssFeed(f fs.FS, baseUrl string, title string, description string) []byte {
	buff := new(bytes.Buffer)

	const (
		header = `<?xml version="1.0"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
<channel>
`
		footer = `</channel>
`
	)

	fmt.Fprint(buff, header)

	fmt.Fprintf(buff, "\t<title>%s</title>\n\t<link>%s</link>\n\t<description>%s</description>\n", title, baseUrl, description)

	fmt.Fprint(buff, footer)

	return buff.Bytes()
}

// err := fs.WalkDir(a.SiteFiles, localContentRoot, func(path string, d fs.DirEntry, err error) error {
// 	if err != nil {
// 		return err
// 	}

// 	// We don't list index files in our collection of siblings and children, because I don't
// 	// want a link back to a page that contains only links.
// 	if strings.Contains(path, "index.html") {
// 		return nil
// 	}

// 	// If the file we're considering isn't an html file, let's move on with our day.
// 	if !strings.Contains(path, "html") {
// 		return nil
// 	}

// 	pageContent, err := fs.ReadFile(a.SiteFiles, path)
// 	if err != nil {
// 		return err
// 	}

// 	title, err := getTitle(path, pageContent)
// 	if err != nil {
// 		return err
// 	}

// 	publishTime, err := getPublishTime(a.SiteFiles, path, pageContent)
// 	if err != nil {
// 		return err
// 	}

// 	// links require a URL relative to the page we're discovering siblings from, not from
// 	// the root of the file system
// 	s_page := Page{
// 		Title:       title,
// 		UrlPath:     strings.TrimPrefix(path, localContentRoot+"/"),
// 		Content:     string(pageContent),
// 		PublishTime: publishTime,
// 	}

// 	pages = append(pages, s_page)

// 	return nil
// })
