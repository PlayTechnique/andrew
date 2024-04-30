package andrew

import (
	"bytes"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"text/template"
	"time"

	"golang.org/x/net/html"
)

const (
	// The index.html has overhead associated with processing its internals, so it gets
	// processed separately from other pages.
	indexIdentifier = "index.html"
)

// Page tracks the content of a specific file and various pieces of metadata about it.
// The Page makes creating links and serving content convenient, as it lets me offload
// the parsing of any elements into a constructor, so that when I need to present those
// elements to an end-user they're easy for me to reason about.
type Page struct {
	// Page title
	Title string
	// According to https://datatracker.ietf.org/doc/html/rfc1738#section-3.1, the subsection of a
	// URL after the procol://hostname is the UrlPath.
	UrlPath     string
	Meta        []string
	Content     string
	PublishTime time.Time
}

// NewPage creates a Page from a URL by reading the corresponding file from the
// AndrewServer's SiteFiles.
func NewPage(server Server, pageUrl string) (Page, error) {
	pageContent, err := fs.ReadFile(server.SiteFiles, pageUrl)
	if err != nil {
		return Page{}, err
	}

	// The fs.FS documentation notes that paths should not start with a leading slash.
	pagePath := strings.TrimPrefix(pageUrl, "/")
	pageTitle, err := getTitle(pagePath, pageContent)
	if err != nil {
		return Page{}, err
	}

	if strings.Contains(pageUrl, indexIdentifier) {
		pageContent, err = buildAndrewIndexBody(server, pageUrl, pageContent)
		if err != nil {
			return Page{}, err
		}
	}

	// pageMeta := getMeta(pagePath, pageContent)
	return Page{Content: string(pageContent), UrlPath: pageUrl, Title: pageTitle}, nil
}

// SetUrlPath updates the UrlPath on a pre-existing Page.
func (a Page) SetUrlPath(urlPath string) Page {
	return Page{Title: a.Title, Content: a.Content, UrlPath: urlPath}
}

// buildAndrewIndexBody receives the path to a file, currently normally an index file.
// It traverses the file system starting at the directory containing
// that file, finds all html files that are _not_ index.html files and returns them
// as a list of html links to those pages.
func buildAndrewIndexBody(server Server, startingPageUrl string, pageContent []byte) ([]byte, error) {
	filterIndexFiles := func(path string, d fs.DirEntry) bool {
		if strings.Contains(path, "index.html") {
			return false
		}

		if !strings.Contains(path, "html") {
			return false
		}

		return true
	}

	siblings, err := server.GetSiblingsAndChildren(startingPageUrl, filterIndexFiles)
	if err != nil {
		return pageContent, err
	}

	var links bytes.Buffer
	cssIdNumber := 0

	for _, sibling := range siblings {
		links.Write(buildAndrewIndexLink(sibling, cssIdNumber))
		cssIdNumber = cssIdNumber + 1
	}

	templateBuffer := bytes.Buffer{}
	// execute template here, write it to something and then return it as the pageContent
	t, err := template.New(startingPageUrl).Parse(string(pageContent))
	if err != nil {
		// TODO: swap this for proper error handling
		panic(err)
	}

	err = t.Execute(&templateBuffer, map[string]string{server.Andrewindexbodytemplate: links.String()})
	if err != nil {
		// TODO: swap this for proper error handling
		panic(err)
	}
	return templateBuffer.Bytes(), nil
}

// buildAndrewIndexLink encapsulates the format of the link
func buildAndrewIndexLink(page Page, cssIdNumber int) []byte {
	link := fmt.Sprintf("<a class=\"andrewindexbodylink\" id=\"andrewindexbodylink%s\" href=\"%s\">%s</a>", fmt.Sprint(cssIdNumber), page.UrlPath, page.Title)
	b := []byte(link)
	return b
}

// getAttribute recursively descends an html node tree, searching for
// the attribute provided. Once the attribute is discovered, it returns.
func getAttributes(attribute string, n *html.Node) []string {
	var attributes []string

	//n.Type no longer matches html.ElementNode; n is now a document, not a node
	if n.Type == html.ElementNode {
		for _, a := range n.Attr {
			if a.Key == attribute {
				attributes = append(attributes, a.Val)
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		attributes = append(attributes, getAttributes(attribute, c)...)
	}

	return attributes
}

// func getMeta(htmlContent []byte) ([]string, error) {
// 	element := "meta"

// 	doc, err := html.Parse(bytes.NewReader(htmlContent))
// 	if err != nil {
// 		return []string{}, err
// 	}

// 	meta := getAttributes(element, doc)

// 	if len(meta) == 0 {
// 		return meta, fmt.Errorf("no %s element found", element)
// 	}
// 	return meta, nil
// }

func getTitle(htmlFilePath string, htmlContent []byte) (string, error) {
	title, err := titleFromHTMLTitleElement(htmlContent)
	if err != nil {
		if err.Error() != "no title element found" {
			return "", err
		}
		// filename is bam.html
		title = path.Base(htmlFilePath)
	}
	return title, nil
}

// titleFromHTMLTitleElement returns the content of the "title" tag or an empty string.
// The error value "no title element found" is returned if title is not discovered
// or is set to an empty string.
func titleFromHTMLTitleElement(fileContent []byte) (string, error) {

	doc, err := html.Parse(bytes.NewReader(fileContent))
	if err != nil {
		return "", err
	}

	title := getAttributes("title", doc)
	if len(title) == 0 {
		return "", fmt.Errorf("no title element found")
	}
	return title[0], nil
}