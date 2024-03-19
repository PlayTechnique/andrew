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
)

type AndrewServer struct {
	SiteFiles fs.FS
}

func NewAndrewServer(contentRoot string) (AndrewServer, error) {
	cr, err := filepath.Abs(contentRoot)
	if err != nil {
		return AndrewServer{}, err
	}

	return AndrewServer{SiteFiles: os.DirFS(cr)}, nil
}

// The Serve function handles requests for any URL. It checks whether the request is for
// an index.html page or for anything else. The special behaviour for the index page is documented
// below.
func (a AndrewServer) Serve(w http.ResponseWriter, r *http.Request) {
	pagePath := path.Clean(r.RequestURI)

	if strings.HasSuffix(pagePath, "/") {
		pagePath = pagePath + "index.html"
	}

	if isIndexPage(pagePath) {

		a.serveIndexPage(w, r, pagePath)
		return
	}

	a.serveOther(w, r, pagePath)
}

// serveIndexPage treats the index page as a template with a single data element: AndrewIndexBody.
// If the page does not contain this element, it is written to the http.ResponseWriter as it is.
// If the page does contain an AndrewIndexBody element, serveIndexPage calls out to buildIndexBody to create
// the correct body of the page and then renders it into the AndrewIndexBody.
func (a AndrewServer) serveIndexPage(w http.ResponseWriter, r *http.Request, pagePath string) {

	pageContent, err := fs.ReadFile(a.SiteFiles, pagePath)

	if err != nil {
		checkPageErrors(w, r, err)
	}

	t, err := template.New(pagePath).Parse(string(pageContent))
	if err != nil {
		panic(err)
	}

	indexBody, err := a.buildIndexBody(pagePath)

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

// buildIndexBody receives the path to a file. It traverses the file system starting at the directory containing
// that file, finds all html files that are _not_ index.html files and returns them
// as a list of html links to those pages.
func (a AndrewServer) buildIndexBody(indexPagePath string) ([]string, error) {

	html := []string{}

	//Given a path to the index page of ./foo/bar/index.html, I want the contentRoot
	//to be the containing directory i.e. ./foo/bar/
	pathSegments := strings.Split(indexPagePath, "/")
	localContentRoot := strings.Join(pathSegments[:len(pathSegments)-1], "/")
	linkNumber := 0

	err := fs.WalkDir(a.SiteFiles, localContentRoot, func(path string, d fs.DirEntry, err error) error {
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
			localPath := strings.Replace(path, localContentRoot+"/", "", 1)
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

// serveOther writes to the ResponseWriter any arbitrary html file, or css, javascript, images etc.
func (a AndrewServer) serveOther(w http.ResponseWriter, r *http.Request, pagePath string) {
	pageContent, err := os.ReadFile(pagePath)

	if err != nil {
		checkPageErrors(w, r, err)
		return
	}
	// Determine the content type based on the file extension
	switch filepath.Ext(pagePath) {
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".html":
		w.Header().Set("Content-Type", "text/html")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".jpg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case ".webp":
		w.Header().Set("Content-Type", "image/webp")
	case ".ico":
		w.Header().Set("Content-Type", "image/x-icon")
	}

	w.WriteHeader(http.StatusOK)
	_, err = fmt.Fprint(w, string(pageContent))

	if err != nil {
		panic(err)
	}

}

// checkPageErrors is a helper function that will convert an error handed into it
// into the appropriate http error code.
// If no specific error is found, a 500 is the default value written to the
// http.ResponseWriter
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

// isIndexPage is a helper function to check if a file being requested
// is an index.html file.
func isIndexPage(uri string) bool {
	isIndex := strings.HasSuffix(uri, "index.html")
	return isIndex
}
