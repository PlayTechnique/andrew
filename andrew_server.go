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
	SiteFiles               fs.FS
	BaseUrl                 string
	Address                 string
	andrewindexbodytemplate string
}

func NewAndrewServer(contentRoot fs.FS, address string, baseUrl string) (AndrewServer, error) {
	return AndrewServer{SiteFiles: contentRoot, andrewindexbodytemplate: "AndrewIndexBody", Address: address, BaseUrl: baseUrl}, nil
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

	// /index.html becomes index.html
	// /articles/page.html becomes articles/page.html
	// without this the paths aren't found properly inside the fs.
	pagePath = strings.TrimPrefix(pagePath, "/")
	pageContent, err := fs.ReadFile(a.SiteFiles, pagePath)

	if err != nil {
		checkPageErrors(w, r, err)
	}

	t, err := template.New(pagePath).Parse(string(pageContent))
	if err != nil {
		panic(err)
	}

	indexBody, err := a.buildAndrewIndexBody(pagePath)

	if err != nil {
		panic(err)
	}

	body := strings.Join(indexBody, "\n")

	//write the executed template directly to the http writer
	err = t.Execute(w, map[string]string{a.andrewindexbodytemplate: body})

	if err != nil {
		panic(err)
	}
}

// buildAndrewIndexBody receives the path to a file. It traverses the file system starting at the directory containing
// that file, finds all html files that are _not_ index.html files and returns them
// as a list of html links to those pages.
func (a AndrewServer) buildAndrewIndexBody(indexPagePath string) ([]string, error) {

	html := []string{}

	//Given a dir structure <site root>/parentDir/index.html, localContentRoot is parentDir/
	localContentRoot := path.Dir(indexPagePath)
	linkNumber := 0

	err := fs.WalkDir(a.SiteFiles, localContentRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || strings.Contains(path, "index.html") {
			return nil
		}

		htmlSuffix := ".html"
		if filepath.Ext(path) == htmlSuffix {
			htmlContent, err := fs.ReadFile(a.SiteFiles, path)

			if err != nil {
				return err
			}

			title, err := getTitle(path, htmlContent)

			if err != nil {
				return err
			}

			if !strings.HasSuffix(localContentRoot, string(filepath.Separator)) {
				localContentRoot += string(filepath.Separator)
			}

			linkPath := strings.TrimPrefix(path, localContentRoot)

			// TODO: extract the formatting into its own function.
			// <a href=path/to/foo.html>what's the title?</a>
			link := fmt.Sprintf("<a class=\"andrewindexbodylink\" id=\"andrewindexbodylink%s\" href=\"%s\">%s</a>", fmt.Sprint(linkNumber), linkPath, title)
			linkNumber = linkNumber + 1

			html = append(html, link)
		}

		return nil
	})

	return html, err

}

// serveOther writes to the ResponseWriter any arbitrary html file, or css, javascript, images etc.
func (a AndrewServer) serveOther(w http.ResponseWriter, r *http.Request, pagePath string) {
	pagePath = strings.TrimPrefix(pagePath, "/")
	pageContent, err := fs.ReadFile(a.SiteFiles, pagePath)

	if err != nil {
		checkPageErrors(w, r, err)
		return
	}
	// Determine the content type based on the file extension
	switch filepath.Ext(pagePath) {
	case ".css":
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case ".html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
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
