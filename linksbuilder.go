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
	var html bytes.Buffer
	var templateBuffer bytes.Buffer
	directoriesAndContents := mapFromPagePaths(siblings)

	directoriesInDepthOrder := keysOrderedByNumberOfSlashes(directoriesAndContents)
	linkCount := 0

	html.Write([]byte("<div class=\"AndrewTableOfContentsWithDirectories\">\n"))

	for _, parentDir := range directoriesInDepthOrder {
		// Skip the root directory if it only contains the starting page
		if parentDir == "" && len(directoriesAndContents[parentDir]) == 1 && directoriesAndContents[parentDir][0] == startingPage {
			continue
		}

		// Start the list for the directory
		html.Write([]byte("<ul>\n"))

		// Add the directory heading inside the <ul>
		if parentDir != "" {
			if countSlashes(parentDir) == 1 {
				html.Write([]byte("<h5>" + parentDir + "</h5>\n"))
			} else {
				dirs := strings.Split(parentDir, "/")
				html.Write([]byte("<h5><span class=\"AndrewParentDir\">" + dirs[0] + "/</span>" + strings.Join(dirs[1:], "/") + "</h5>\n"))
			}
		}

		// Add the links to the list
		for _, sibling := range directoriesAndContents[parentDir] {
			// Skip the starting page
			if sibling == startingPage {
				continue
			}
			html.Write(buildAndrewTableOfContentsLink(sibling.UrlPath, sibling.Title, sibling.PublishTime.Format(time.DateOnly), linkCount))
			linkCount++
		}

		html.Write([]byte("</ul>\n"))
	}

	html.Write([]byte("</div>\n"))

	t, err := template.New(startingPage.UrlPath).Parse(startingPage.Content)
	if err != nil {
		panic(err)
	}

	err = t.Execute(&templateBuffer, map[string]string{"AndrewTableOfContentsWithDirectories": html.String()})
	if err != nil {
		return templateBuffer.Bytes(), err
	}

	return templateBuffer.Bytes(), nil
}

// mapFromPagePaths takes an array of pages and returns a map of those pages in which the keys
// are the directories containing a specific page and the value is the path inside the directory
// to that page.
// So pages at page.html, parent/page1.html, parent/page2.html and parent/child/page.html
// become {"": "page.html", "parent": ["page1.html","page2.html"], "parent/child": ["page.html"]}
// The indexes are directory names as strings; the values are arrays of Pages.
func mapFromPagePaths(siblings []Page) map[string][]Page {
	directoriesAndContents := make(map[string][]Page)

	for _, sibling := range siblings {
		path, _ := path.Split(sibling.UrlPath)
		directoriesAndContents[path] = append(directoriesAndContents[path], sibling)
	}
	return directoriesAndContents
}

func keysOrderedByNumberOfSlashes(directoriesAndContents map[string][]Page) []string {
	keysOrderedByLength := make([]string, 0, len(directoriesAndContents))
	for k := range directoriesAndContents {
		keysOrderedByLength = append(keysOrderedByLength, k)
	}

	sort.Slice(keysOrderedByLength, func(i, j int) bool {
		slashesI := countSlashes(keysOrderedByLength[i])
		slashesJ := countSlashes(keysOrderedByLength[j])
		if slashesI == slashesJ {
			return keysOrderedByLength[i] < keysOrderedByLength[j]
		}
		return slashesI < slashesJ
	})
	return keysOrderedByLength
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
