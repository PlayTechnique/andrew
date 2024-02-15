package andrew

import (
	"fmt"
	"golang.org/x/net/html"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// GetLinks is a function that walks a directory starting at contentRoot and
// gets a list of the html files inside that are not index.html. These
// represent the articles (files) or the next organisational unit (directories).
func GetLinks(contentRoot string) ([]string, error) {
	html := []string{}
	err := filepath.WalkDir(contentRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if strings.Contains(path, "index.html") {
			if path == contentRoot+"/index.html" {
				return nil
			}

			return nil
		}

		htmlSuffix := ".html"
		if filepath.Ext(path) == htmlSuffix {
			// foo/bar/bam.html becomes [foo, bar, bam.html]
			filenamePortions := strings.Split(path, "/")
			// path is contentroot/path/to/file.html. It needs to become
			// path/to/file.html
			link := strings.Join(filenamePortions[1:], "/")

			title, err := getTitle(path, filenamePortions, htmlSuffix)
			if err != nil {
				return err
			}

			// TODO: extract the formatting into its own function.
			path = fmt.Sprintf("<a href=%s>%s</a>", link, title)

			html = append(html, path)
		}

		return nil
	})

	return html, err
}

func getTitle(path string, filenamePortions []string, htmlSuffix string) (string, error) {
	title, err := titleFromHTMLTitleElement(path)

	if err != nil {
		if err.Error() != "no title element found" {
			return "", err
		}
		// filename is bam.html
		filename := filenamePortions[len(filenamePortions)-1]
		// title is bam
		title = filename[:len(filename)-len(htmlSuffix)]
	}
	return title, nil
}

// titleFromHTMLTitleElement returns the content of the "title" tag or an empty string.
// The error value "no title element found" is returned if title is not discovered
// or is set to an empty string.
func titleFromHTMLTitleElement(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	doc, err := html.Parse(f)
	if err != nil {
		return "", err
	}

	title := getAttribute("title", doc)
	if title == "" {
		return "", fmt.Errorf("no title element found")
	}
	return title, nil
}

// getAttribute recursively descends an html node tree, searching for
// the attribute provided. Once the attribute is discovered, it returns.
func getAttribute(attribute string, n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == attribute {
		if n.FirstChild != nil {
			return n.FirstChild.Data
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result := getAttribute(attribute, c)
		if result != "" {
			return result
		}
	}
	return ""
}
