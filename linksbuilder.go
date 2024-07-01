package andrew

import (
	"bytes"
	"fmt"
	"text/template"
)

// RenderTemplate receives the path to a file, currently normally an index file.
// It traverses the file system starting at the directory containing
// that file, finds all html files that are _not_ index.html files and returns them
// as a list of html links to those pages.
func RenderTemplate(siblings []Page, startingPage Page) ([]byte, error) {

	var links bytes.Buffer

	for i, sibling := range siblings {
		links.Write(buildAndrewTableOfContentsLink(sibling.UrlPath, sibling.Title, i))
	}

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
func buildAndrewTableOfContentsLink(urlPath string, title string, cssIdNumber int) []byte {
	link := fmt.Sprintf("<a class=\"andrewtableofcontentslink\" id=\"andrewtableofcontentslink%s\" href=\"%s\">%s</a>", fmt.Sprint(cssIdNumber), urlPath, title)
	b := []byte(link)
	return b
}
