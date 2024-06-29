package andrew

import (
	"bytes"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
	"time"

	"golang.org/x/net/html"
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
	Content     string
	PublishTime time.Time
}

type TagInfo struct {
	Data       string
	Attributes map[string]string
}

// NewPage creates a Page from a URL by reading the corresponding file from the
// AndrewServer's SiteFiles.
// NewPage does this by reading the page content from disk, then parsing out various
// metadata that are convenient to have quick access to, such as the page title or the
// publish time.
func NewPage(server Server, pageUrl string) (Page, error) {
	pageContent, err := fs.ReadFile(server.SiteFiles, pageUrl)
	if err != nil {
		return Page{}, err
	}

	pageInfo, err := fs.Stat(server.SiteFiles, pageUrl)
	if err != nil {
		return Page{}, err
	}

	// The fs.FS documentation notes that paths should not start with a leading slash.
	pagePath := strings.TrimPrefix(pageUrl, "/")

	pageTitle, err := getTitle(pagePath, pageContent)
	if err != nil {
		return Page{}, err
	}

	page := Page{Content: string(pageContent), UrlPath: pageUrl, Title: pageTitle, PublishTime: pageInfo.ModTime()}

	meta, err := GetMetaElements(pageContent)
	if err != nil {
		return Page{}, err
	}

	publishTime, ok := meta["andrew-publish-time"]

	if ok {
		andrewCreatedAt, err := time.Parse(time.DateOnly, publishTime)

		if err != nil {
			return Page{}, err
			// log.Logger("could not parse meta tag andrew-publish-time using time.Parse. Defaulting to mod time")
		} else {
			page.PublishTime = andrewCreatedAt
		}
	}

	if strings.Contains(pageUrl, "index.html") {
		siblings, err := server.GetSiblingsAndChildren(page.UrlPath)

		if err != nil {
			return page, err
		}

		orderedSiblings := sortPages(siblings)

		pageContent, err = BuildAndrewTOCLinks(orderedSiblings, page)
		if err != nil {
			return Page{}, err
		}

		page.Content = string(pageContent)
	}

	return page, nil
}

// SetUrlPath updates the UrlPath on a pre-existing Page.
func (a Page) SetUrlPath(urlPath string) Page {
	return Page{Title: a.Title, Content: a.Content, UrlPath: urlPath, PublishTime: a.PublishTime}
}

// getTagInfo recursively descends an html node tree for the requested tag,
// searching both data and attributes to find information about the node that's requested.
// getTagInfo recursively descends an html node tree for the requested tag,
// searching both data and attributes to find information about the node that's requested.
func getTagInfo(tag string, n *html.Node) TagInfo {
	var tagDataAndAttributes TagInfo = TagInfo{
		Data:       "",
		Attributes: make(map[string]string),
	}

	// getTag recursively descends an html node tree, searching for
	// the attribute provided. Once the attribute is discovered, it first checks
	// for any Attributes available on the html node. If there are no Attributes,
	// the key won't exist in the tagDataAndAttributes map.
	// If there is data, it will append to attributes.
	var getTag func(n *html.Node)

	getTag = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == tag {
			attrName := ""
			attrVal := ""

			if n.Attr != nil {
				for _, attr := range n.Attr {
					switch attr.Key {
					case "content":
						attrVal = attr.Val
					case "name":
						attrName = attr.Val
					}
					tagDataAndAttributes.Attributes[attrName] = attrVal
				}
			}

			if n.FirstChild != nil {
				tagDataAndAttributes.Data = n.FirstChild.Data
			}

			return
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			getTag(c)
		}
	}

	// Start the recursion from the root node
	getTag(n)

	return tagDataAndAttributes
}

func GetMetaElements(htmlContent []byte) (map[string]string, error) {
	element := "meta"

	doc, err := html.Parse(bytes.NewReader(htmlContent))
	if err != nil {
		return map[string]string{}, err
	}

	tagInfo := getTagInfo(element, doc)

	return tagInfo.Attributes, nil
}

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

	tagInfo := getTagInfo("title", doc)
	if len(tagInfo.Data) == 0 {
		return "", fmt.Errorf("no title element found")
	}
	return tagInfo.Data, nil
}

func sortPages(pages []Page) []Page {

	sort.Slice(pages, func(i, j int) bool {
		return pages[i].PublishTime.After(pages[j].PublishTime)
	})

	return pages
}
