package andrew

import (
	"fmt"
	"io"
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

const (
	DefaultContentRoot = "."
	DefaultAddress     = ":8080"
	DefaultBaseUrl     = "http://localhost:8080"
)

func NewAndrewServer(contentRoot fs.FS, address string, baseUrl string) (AndrewServer, error) {
	return AndrewServer{SiteFiles: contentRoot, andrewindexbodytemplate: "AndrewIndexBody", Address: address, BaseUrl: baseUrl}, nil
}

func Main(args []string, printDest io.Writer) int {
	help := `Usage: andrew [contentRoot] [address] [baseUrl]
	- contentRoot: The root directory of your content. Defaults to '.' if not specified.
	- address: The address to bind to. Defaults to 'localhost:8080' if not specified. If in doubt, you probably want 0.0.0.0:<something>
	- base URL: The protocol://hostname for your server. Defaults to 'http://localhost:8080' if not specified. Used to generate sitemap/rss feed accurately.
	-h, --help: Display this help message.
`

	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			fmt.Fprint(printDest, help)
			return 0
		}
	}

	contentRoot, address, baseUrl := ParseArgs(args)

	fmt.Fprintf(printDest, "Serving from %s, listening on %s, serving on %s", contentRoot, address, baseUrl)

	err := ListenAndServe(os.DirFS(contentRoot), address, baseUrl)

	if err != nil {
		panic(err)
	}

	return 0
}

func ParseArgs(args []string) (string, string, string) {
	contentRoot := DefaultContentRoot
	address := DefaultAddress
	baseUrl := DefaultBaseUrl

	if len(args) >= 1 {
		contentRoot = args[0]
	}

	if len(args) >= 2 {
		address = args[1]
	}

	if len(args) >= 3 {
		baseUrl = args[2]
	}

	return contentRoot, address, baseUrl
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

	a.serveOther(w, pagePath)
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
		message, status := CheckPageErrors(err)
		w.WriteHeader(status)
		fmt.Fprint(w, message)
		return
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
func (a AndrewServer) serveOther(w http.ResponseWriter, pagePath string) {
	pagePath = strings.TrimPrefix(pagePath, "/")
	pageContent, err := fs.ReadFile(a.SiteFiles, pagePath)

	if err != nil {
		message, status := CheckPageErrors(err)
		w.WriteHeader(status)
		fmt.Fprint(w, message)
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
	fmt.Fprint(w, string(pageContent))

}

// CheckPageErrors is a helper function that will convert an error handed into it
// into the appropriate http error code and a message.
// If no specific error is found, a 500 is the default value returned.
func CheckPageErrors(err error) (string, int) {
	// if a file doesn't exist
	// http 404
	if os.IsNotExist(err) {
		return "404 not found", http.StatusNotFound
	}

	// if the file does exist but is unreadable
	// http 403
	if os.IsPermission(err) {
		return "403 Forbidden", http.StatusForbidden
	}

	// other errors; not sure what they are, but catchall
	// http 500
	return "500 something went wrong", http.StatusInternalServerError
}

// isIndexPage is a helper function to check if a file being requested
// is an index.html file.
func isIndexPage(uri string) bool {
	isIndex := strings.HasSuffix(uri, "index.html")
	return isIndex
}
