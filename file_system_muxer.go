package andrew

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sirupsen/logrus"
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
	logrus.Info("Serving ", pagePath)

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

	t, err := template.New(pagePath).Parse(string(pageContent))
	if err != nil {
		panic(err)
	}

	indexBody, err := buildIndexBody(pagePath)

	if err != nil {
		panic(err)
	}

	body := strings.Join(indexBody, "\n")

	//write the executed template directly to the http writer
	err = t.Execute(w, map[string]string{"AndrewIndexBody": body})

	if err != nil {
		panic(err)
	}
}

// buildIndexBody will traverse the file system starting at the directory containing the index.html
// whose body we're building. It finds all html files (excepting index.html files) and returns them
// as a list of html links to those pages.
func buildIndexBody(indexPagePath string) ([]string, error) {

	html := []string{}

	//Given a path to the index page of ./foo/bar/index.html, I want the contentRoot
	//to be the containing directory i.e. ./foo/bar/
	pathSegments := strings.Split(indexPagePath, "/")
	contentRoot := strings.Join(pathSegments[:len(pathSegments)-1], "/")
	linkNumber := 0

	err := filepath.WalkDir(contentRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if strings.Contains(path, "index.html") {
			return nil
		}

		htmlSuffix := ".html"
		if filepath.Ext(path) == htmlSuffix {
			// path is contentroot/path/to/file.html. It needs to become
			// path/to/file.html for generating the link to the path.
			localPath := strings.Replace(path, contentRoot+"/", "", 1)
			title, err := getTitle(path)

			if err != nil {
				return err
			}

			// TODO: extract the formatting into its own function.
			// <a href=path/to/foo.html>what's the title?</a>
			link := fmt.Sprintf("<a class=\"andrewindexbodylink\" id=\"andrewindexbodylink%s\" href=\"%s\">%s</a>", fmt.Sprint(linkNumber), localPath, title)
			linkNumber = linkNumber + 1

			html = append(html, link)
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

func getTitle(filePath string) (string, error) {
	title, err := titleFromHTMLTitleElement(filePath)

	if err != nil {
		if err.Error() != "no title element found" {
			return "", err
		}
		// filename is bam.html
		title = path.Base(filePath)
	}
	return title, nil
}
