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

type AndrewPage struct {
	Title string
	// According to https://datatracker.ietf.org/doc/html/rfc1738#section-3.1, the subsection of a
	// URL after the procol://hostname is the UrlPath.
	UrlPath string
	//
	Content     string
	PublishTime time.Time
}

// NewPage creates a Page from a URL by reading the corresponding file from the
// file system.
// If a page cannot be read, it just hands the error back up the stack. This is
// because NewPage is being called in a web server's context, and errors are handled
// by printing both an http.StatusCode and a well understood warning to a tcp socket,
// not by panicking or something.
// The Page constructor does not have access to the tcp socket, so it cannot actually
// handle the error correctly.
func NewPage(server AndrewServer, pageUrl string) (AndrewPage, error) {
	// /index.html becomes index.html
	// /articles/page.html becomes articles/page.html
	// without this the paths aren't found properly inside the fs.
	pageContent, err := fs.ReadFile(server.SiteFiles, pageUrl)
	if err != nil {
		return AndrewPage{}, err
	}

	pagePath := strings.TrimPrefix(pageUrl, "/")
	pageTitle, err := getTitle(pagePath, pageContent)

	if err != nil {
		return AndrewPage{}, err
	}

	//TODO: This constructor does not seem like the right place to hide the knowledge
	//TODO: that index.html isn't treated the same as everything else, but it's good
	//TODO: for making the functionality work.
	if strings.Contains(pageUrl, "index.html") {
		pageContent, err = buildAndrewIndexBody(server, pageUrl, pageContent)

		if err != nil {
			return AndrewPage{}, err
		}
	}

	return AndrewPage{Content: string(pageContent), UrlPath: pageUrl, Title: pageTitle}, nil
}

func (a AndrewPage) SetUrlPath(urlPath string) AndrewPage {
	return AndrewPage{Title: a.Title, Content: a.Content, PublishTime: a.PublishTime, UrlPath: urlPath}
}

func buildAndrewIndexBody(server AndrewServer, startingPageUrl string, pageContent []byte) ([]byte, error) {
	// The index body should not contain other index.html files.
	// It also should not contain any files that are not html files.
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
	//execute template here, write it to something and then return it as the pageContent
	t, err := template.New(startingPageUrl).Parse(string(pageContent))

	if err != nil {
		//TODO: swap this for proper error handling
		panic(err)
	}

	err = t.Execute(&templateBuffer, map[string]string{server.andrewindexbodytemplate: links.String()})

	if err != nil {
		//TODO: swap this for proper error handling
		panic(err)
	}
	return templateBuffer.Bytes(), nil
}

// buildAndrewIndexBody receives the path to a file, currently normally an index file.
// It traverses the file system starting at the directory containing
// that file, finds all html files that are _not_ index.html files and returns them
// as a list of html links to those pages.
func buildAndrewIndexLink(page AndrewPage, cssIdNumber int) []byte {
	link := fmt.Sprintf("<a class=\"andrewindexbodylink\" id=\"andrewindexbodylink%s\" href=\"%s\">%s</a>", fmt.Sprint(cssIdNumber), page.UrlPath, page.Title)
	b := []byte(link)
	return b
}

// titleFromHTMLTitleElement returns the content of the "title" tag or an empty string.
// The error value "no title element found" is returned if title is not discovered
// or is set to an empty string.
func titleFromHTMLTitleElement(fileContent []byte) (string, error) {

	doc, err := html.Parse(bytes.NewReader(fileContent))
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
