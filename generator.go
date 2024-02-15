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
			title, err := getTitleFromHTMLFile(path)

			if err != nil {
				if err.Error() == "no title element found" {
					filenames := strings.Split(path, "/")
					filename := filenames[len(filenames)-1]
					indexFileName := filename[:len(filename)-len(htmlSuffix)]
					title = indexFileName
				} else {
					return err
				}
			}

			// TODO: extract the formatting into its own function.
			path = fmt.Sprintf("<a href=%s>%s</a>", path, title)

			html = append(html, path)
		}

		return nil
	})

	return html, err
}

// getTitleFromHTMLFile returns the content of the "title" tag or an empty string.
// The error value "no title element found" is returned if title is not discovered
// or is set to an empty string.
func getTitleFromHTMLFile(filePath string) (string, error) {
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
