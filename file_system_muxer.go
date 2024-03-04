package andrew

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type FileSystemMuxer struct {
	ContentRoot string
}

func NewFileSystemMuxer(contentRoot string) (FileSystemMuxer, error) {
	cr, err := filepath.Abs(contentRoot)
	if err != nil {
		return FileSystemMuxer{}, err
	}
	return FileSystemMuxer{ContentRoot: cr}, nil
}

func (f FileSystemMuxer) Serve(w http.ResponseWriter, r *http.Request) {
	pagePath := f.ContentRoot + r.RequestURI

	if strings.HasSuffix(pagePath, "/") {
		pagePath = pagePath + "index.html"
	}

	if isIndexPage(pagePath) {
		f.serveIndexPage(w, r, pagePath)
		return
	}

	f.serveNonIndexPage(w, r, pagePath)
}

// websiteStorage
// WebsiteFromFileSystem is a function that walks a directory starting at contentRoot and
// gets a list of the html files inside that are not index.html. These
// represent the articles (files) or the next organisational unit (directories).
func (f FileSystemMuxer) serveIndexPage(w http.ResponseWriter, r *http.Request, pagePath string) {

	pageContent, err := os.ReadFile(pagePath)

	if err != nil {
		checkPageErrors(w, r, err)
	}

	// TODO: This check doesnt work because the page has not been read
	w.WriteHeader(http.StatusOK)
	t, err := template.New(pagePath).Parse(string(pageContent))
	if err != nil {
		panic(err)
	}

	indexBody, err := buildIndexBody(pagePath)

	if err != nil {
		panic(err)
	}

	body := strings.Join(indexBody, "\n")

	err = t.Execute(w, map[string]string{"AndrewIndexBody": body})

	if err != nil {
		panic(err)
	}
}

func buildIndexBody(indexPagePath string) ([]string, error) {

	html := []string{}

	//when at root - list of articles of all pages
	// when in a subdirectory - list of pages from here down
	//Given a path to the index page of ./foo/bar/index.html, I want the contentRoot
	//to be the containing directory i.e. ./foo/bar/
	pathSegments := strings.Split(indexPagePath, "/")
	contentRoot := strings.Join(pathSegments[:len(pathSegments)-1], "/")

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
			// foo/bar/bam.html becomes [foo, bar, bam.html]
			filenamePortions := strings.Split(path, "/")
			// path is contentroot/path/to/file.html. It needs to become
			// path/to/file.html
			link := strings.Join(filenamePortions[1:], "/")
			title, err := getTitle(path, filenamePortions, htmlSuffix)

			if err != nil {
				return err
			}

			// TODO: extract the formatting into its own function.
			// <a href=path/to/foo.html>what's the title?</a>
			path = fmt.Sprintf("<a href=%s>%s</a>", link, title)

			html = append(html, path)
		}

		return nil
	})

	return html, err

}

func (f FileSystemMuxer) serveNonIndexPage(w http.ResponseWriter, r *http.Request, page string) {
	pageContent, err := os.ReadFile(page)

	if err != nil {
		checkPageErrors(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = fmt.Fprint(w, string(pageContent))

	if err != nil {
		panic(err)
	}

}

func checkPageErrors(w http.ResponseWriter, r *http.Request, err error) {
	// if a file doesn't exist
	// http 404
	if os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, " 404 not found %s", r.RequestURI)
		return
	}

	// if the file does exist but is unreadable
	// http 403
	if os.IsPermission(err) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "403 Forbidden")
		return
	}

	// other errors; not sure what they are, but catchall
	// http 500
	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "500 something went wrong")
}

func isIndexPage(uri string) bool {
	isIndex := strings.HasSuffix(uri, "index.html")
	return isIndex
}

func getTitle(path string, filenamePortions []string, htmlSuffix string) (string, error) {
	title, err := titleFromHTMLTitleElement(path)

	if err != nil {
		if err.Error() != "no title element found" {
			return "", err
		}
		// filename is bam.html
		filename := filenamePortions[len(filenamePortions)-1]
		// title is bam
		title = filename[:len(filename)-len(htmlSuffix)]
	}
	return title, nil
}
