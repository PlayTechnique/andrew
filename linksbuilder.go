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
		return renderAndrewTableOfContentsWithDirectories(siblings, startingPage, DefaultPageSort)
	}

	return []byte(startingPage.Content), nil
}

func countSlashes(s string) int {
	return strings.Count(s, "/")
}

// PageSortFunc is a function type that defines how to sort Pages within a directory
type PageSortFunc func([]Page) []Page

// directorySortKey represents a directory and its most recent page time
type directorySortKey struct {
	dir      string
	lastTime time.Time
}

// getDirectoriesOrderedByMostRecent returns directories sorted first by their most recent page,
// then by depth (number of slashes), then alphabetically
func getDirectoriesOrderedByMostRecent(dirMap map[string][]Page) []string {
	// Create slice of directorySortKey
	dirs := make([]directorySortKey, 0, len(dirMap))

	// For each directory, find the most recent page
	for dir, pages := range dirMap {
		var mostRecent time.Time
		for _, page := range pages {
			if page.PublishTime.After(mostRecent) {
				mostRecent = page.PublishTime
			}
		}
		dirs = append(dirs, directorySortKey{dir, mostRecent})
	}

	// Sort directories by most recent first, then by depth, then alphabetically
	sort.Slice(dirs, func(i, j int) bool {
		// If times are different, sort by most recent first
		if !dirs[i].lastTime.Equal(dirs[j].lastTime) {
			return dirs[i].lastTime.After(dirs[j].lastTime)
		}

		// If times are equal, sort by depth (fewer slashes first)
		iSlashes := countSlashes(dirs[i].dir)
		jSlashes := countSlashes(dirs[j].dir)
		if iSlashes != jSlashes {
			return iSlashes < jSlashes
		}

		// If depths are equal, sort alphabetically
		return dirs[i].dir < dirs[j].dir
	})

	// Convert back to string slice
	result := make([]string, len(dirs))
	for i, dir := range dirs {
		result[i] = dir.dir
	}
	return result
}

func renderAndrewTableOfContentsWithDirectories(siblings []Page, startingPage Page, sortFn PageSortFunc) ([]byte, error) {
	var html bytes.Buffer
	var templateBuffer bytes.Buffer
	directoriesAndContents := mapFromPagePaths(siblings)

	directoriesInDepthOrder := getDirectoriesOrderedByMostRecent(directoriesAndContents)
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
				html.Write([]byte("<h5><span class=\"AndrewTableOfContentsWithDirectories\">" + dirs[0] + "/</span>" + strings.Join(dirs[1:], "/") + "</h5>\n"))
			}
		}

		// Sort the pages in this directory using the provided sort function
		pages := directoriesAndContents[parentDir]
		if sortFn != nil {
			pages = sortFn(pages)
		}

		// Add the links to the list
		for _, sibling := range pages {
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

func renderAndrewTableOfContents(siblings []Page, startingPage Page) ([]byte, error) {
	var html bytes.Buffer

	html.Write([]byte("<div class=\"AndrewTableOfContents\">\n"))
	html.Write([]byte("<ul>\n"))
	for i, sibling := range siblings {
		html.Write(buildAndrewTableOfContentsLink(sibling.UrlPath, sibling.Title, sibling.PublishTime.Format(time.DateOnly), i))
	}
	html.Write([]byte("</ul>\n"))
	html.Write([]byte("</div>\n"))

	templateBuffer := bytes.Buffer{}

	t, err := template.New(startingPage.UrlPath).Parse(startingPage.Content)

	if err != nil {

		panic(err)
	}

	err = t.Execute(&templateBuffer, map[string]string{"AndrewTableOfContents": html.String()})
	if err != nil {
		return templateBuffer.Bytes(), err
	}

	return templateBuffer.Bytes(), nil
}

// buildAndrewTableOfContentsLink creates an HTML list item containing a link to a page.
// It formats the link with a CSS class, unique ID, URL path, title, and publish date.
//
// Parameters:
//   - urlPath: The path to the linked page
//   - title: The display text for the link
//   - publishDate: The formatted date string to display
//   - cssIdNumber: A unique number used to generate the link's ID attribute
//
// Returns a byte slice containing the formatted HTML list item.
func buildAndrewTableOfContentsLink(urlPath string, title string, publishDate string, cssIdNumber int) []byte {
	link := fmt.Sprintf("<li><a class=\"andrewtableofcontentslink\" id=\"andrewtableofcontentslink%s\" href=\"%s\">%s</a> - <span class=\"andrew-page-publish-date\">%s</span></li>\n", fmt.Sprint(cssIdNumber), urlPath, title, publishDate)
	b := []byte(link)
	return b
}

// DefaultPageSort provides the default sorting behavior for pages
// (current implementation preserved as a separate function)
func DefaultPageSort(pages []Page) []Page {
	sorted := make([]Page, len(pages))
	copy(sorted, pages)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].PublishTime.After(sorted[j].PublishTime)
	})
	return sorted
}
