package andrew

import (
	"bytes"
	"fmt"
	"text/template"
	"time"
)

// RenderTemplates receives the path to a file, currently normally an index file.
// It traverses the file system starting at the directory containing
// that file, finds all html files that are _not_ index.html files and returns them
// as a list of html links to those pages.
func RenderTemplates(siblings []Page, startingPage Page) ([]byte, error) {

	var links bytes.Buffer

	links.Write([]byte("<ul>"))
	for i, sibling := range siblings {
		links.Write(buildAndrewTableOfContentsLink(sibling.UrlPath, sibling.Title, sibling.PublishTime.Format(time.DateOnly), i))
	}
	links.Write([]byte("</ul>"))

	templateBuffer := bytes.Buffer{}
	// execute template here, write it to something and then return it as the pageContent
	t, err := template.New(startingPage.UrlPath).Parse(startingPage.Content)

	if err != nil {
		// TODO: swap this for proper error handling
		panic(err)
	}

	err = t.Execute(&templateBuffer, map[string]string{"AndrewTableOfContents": links.String()})
	if err != nil {
		return templateBuffer.Bytes(), err
	}

	return templateBuffer.Bytes(), nil
}

// buildAndrewTableOfContentsLink encapsulates the format of the link
func buildAndrewTableOfContentsLink(urlPath string, title string, publishDate string, cssIdNumber int) []byte {
	link := fmt.Sprintf("<li><a class=\"andrewtableofcontentslink\" id=\"andrewtableofcontentslink%s\" href=\"%s\">%s</a> - <span class=\"publish-date\">%s</span></li>", fmt.Sprint(cssIdNumber), urlPath, title, publishDate)
	b := []byte(link)
	return b
}
