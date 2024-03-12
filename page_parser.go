package andrew

import (
	"fmt"
	"os"

	"golang.org/x/net/html"
)

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
