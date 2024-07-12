package andrew

import (
	"bytes"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"time"
)

// RenderTemplates receives the path to a file, currently normally an index file.
// It traverses the file system starting at the directory containing
// that file, finds all html files that are _not_ index.html files and returns them
// as a list of html links to those pages.
func RenderTemplates(siblings []Page, startingPage Page) ([]byte, error) {

	tableOfContents, err := regexp.Compile(`.*{{\s*\.AndrewTableOfContents\s*}}.*`)
	if err != nil {
		return nil, err
	}

	if tableOfContents.FindString(startingPage.Content) != "" {
		return renderAndrewTableOfContents(siblings, startingPage)
	}

	tableOfContentsWithDirs, err := regexp.Compile(`.*{{\s*\.AndrewTableOfContentsWithDirectories\s*}}.*`)
	if err != nil {
		return nil, err
	}

	if tableOfContentsWithDirs.FindString(startingPage.Content) != "" {
		return renderAndrewTableOfContentsWithDirectories(siblings, startingPage)
	}

	return []byte(startingPage.Content), nil
}

func countSlashes(s string) int {
	return strings.Count(s, "/")
}

func renderAndrewTableOfContentsWithDirectories(siblings []Page, startingPage Page) ([]byte, error) {
	var links bytes.Buffer
	var templateBuffer bytes.Buffer
	directoriesAndContents := make(map[string][]Page)

	for _, sibling := range siblings {
		path, _ := path.Split(sibling.UrlPath)
		directoriesAndContents[path] = append(directoriesAndContents[path], sibling)
	}

	keys := make([]string, 0, len(directoriesAndContents))
	for k := range directoriesAndContents {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		slashesI := countSlashes(keys[i])
		slashesJ := countSlashes(keys[j])
		if slashesI == slashesJ {
			return keys[i] < keys[j] // Lexicographical order if the number of slashes is the same
		}
		return slashesI < slashesJ // Primary order by number of slashes
	})

	linkCount := 0
	dirCount := 1

	links.Write([]byte("<ul style=\"padding-left: 10px;\">\n"))

	for parentDir, siblingPages := range directoriesAndContents {
		if linkCount != 0 {
			links.Write([]byte("<ul>\n"))
		}
		if parentDir != "" {
			links.Write([]byte("<h5 style=\"display: inline;\">" + parentDir + "</h5>\n"))
		}
		for _, sibling := range siblingPages {
			links.Write(buildAndrewTableOfContentsLink(sibling.UrlPath, sibling.Title, sibling.PublishTime.Format(time.DateOnly), linkCount))
			linkCount = linkCount + 1
		}
		dirCount = dirCount + 1
	}

	for range dirCount - 1 {
		links.Write([]byte("</ul>\n"))
	}

	t, err := template.New(startingPage.UrlPath).Parse(startingPage.Content)

	if err != nil {
		panic(err)
	}

	err = t.Execute(&templateBuffer, map[string]string{"AndrewTableOfContentsWithDirectories": links.String()})
	if err != nil {
		return templateBuffer.Bytes(), err
	}

	return templateBuffer.Bytes(), nil
}

func renderAndrewTableOfContents(siblings []Page, startingPage Page) ([]byte, error) {
	var links bytes.Buffer

	links.Write([]byte("<ul>\n"))
	for i, sibling := range siblings {
		links.Write(buildAndrewTableOfContentsLink(sibling.UrlPath, sibling.Title, sibling.PublishTime.Format(time.DateOnly), i))
	}
	links.Write([]byte("</ul>\n"))

	templateBuffer := bytes.Buffer{}

	t, err := template.New(startingPage.UrlPath).Parse(startingPage.Content)

	if err != nil {

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
	link := fmt.Sprintf("<li><a class=\"andrewtableofcontentslink\" id=\"andrewtableofcontentslink%s\" href=\"%s\">%s</a> - <span class=\"publish-date\">%s</span></li>\n", fmt.Sprint(cssIdNumber), urlPath, title, publishDate)
	b := []byte(link)
	return b
}
