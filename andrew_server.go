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
)

// AndrewServer holds a reference to the paths in the fs.FS that correspond to
// each page that should be served.
// When a URL is requested, AndrewServer creates an AndrewPage for the file referenced
// in that URL and then serves the AndrewPage.
type AndrewServer struct {
	SiteFiles               fs.FS  //The files being served
	BaseUrl                 string //The URL used in any links generated for this website that should contain the hostname.
	Address                 string //IpAddress:Port combo to be served on.
	Andrewindexbodytemplate string //The string we're searching for inside a Page that should be replaced with a template. Mightn't belong in the Server.
}

const (
	AndrewIndexBodyTemplate = "AndrewIndexBody"
	DefaultContentRoot      = "."
	DefaultAddress          = ":8080"
	DefaultBaseUrl          = "http://localhost:8080"
)

func NewAndrewServer(contentRoot fs.FS, address string, baseUrl string) (AndrewServer, error) {
	return AndrewServer{SiteFiles: contentRoot, Andrewindexbodytemplate: "AndrewIndexBody", Address: address, BaseUrl: baseUrl}, nil
}

func Main(args []string, printDest io.Writer) int {
	help := `Usage: andrew [contentRoot] [address] [baseUrl]
	- contentRoot: The root directory of your content. Defaults to '.' if not specified.
	- address: The address to bind to. Defaults to 'localhost:8080' if not specified. If in doubt, you probably want '0.0.0.0:<some free port>'
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

// Serve handles requests for any URL. It checks whether the request is for
// an index.html page or for anything else (another page, css, javascript etc).
// If a directory is requested, Serve defaults to finding the index.html page
// within that directory. Detecting this case for
func (a AndrewServer) Serve(w http.ResponseWriter, r *http.Request) {
	pagePath := path.Clean(r.RequestURI)

	// Ensure the pagePath is relative to the root of a.SiteFiles.
	// This involves trimming a leading slash if present.
	pagePath = strings.TrimPrefix(pagePath, "/")

	maybeDir, _ := fs.Stat(a.SiteFiles, pagePath)

	// In most cases, pagePath does not need to be manipulated.
	// There are three cases where we need to append "index.html" to the pagePath, though:
	// 1. If we receive a request for a directory within the file system, the default file to serve is index.html
	// 2. If we receive a request for www.example.com/, pagePath will be /. This means "please serve the index.html
	//    in whatever directory the web server is started from."
	// 3. If we receive a request for www.example.com, pagePath will be an empty string. We should serve index.html.
	switch {
	case maybeDir != nil && maybeDir.IsDir():
		pagePath = pagePath + "/index.html"
	case strings.HasSuffix(pagePath, "/"):
		pagePath = "index.html"
	case pagePath == "":
		pagePath = "index.html"
	}

	page, err := NewPage(a, pagePath)

	if err != nil {
		message, status := CheckPageErrors(err)
		w.WriteHeader(status)
		fmt.Fprint(w, message)
		return
	}

	a.serve(w, page)
}

// serve writes to the ResponseWriter any arbitrary html file, or css, javascript, images etc.
func (a AndrewServer) serve(w http.ResponseWriter, page AndrewPage) {

	// Determine the content type based on the file extension
	switch filepath.Ext(page.UrlPath) {
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
	fmt.Fprint(w, page.Content)

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

// GetSiblingsAndChildren accepts a path to a file and a filter function.
// It infers the directory that the file resides within, and then recurses the Server's fs.FS
// to return all of the files both in the same directory and further down in the directory structure.
// To filter these down to only files that you care about, pass in a filter function.
// The filter is called in the context of fs.WalkDir. It is handed fs.WalkDir's path and directory entry,
// in that order, and is expected to return a boolean false.
// If that error is nil then the current file being evaluated is skipped for consideration.
func (a AndrewServer) GetSiblingsAndChildren(pagePath string, filter func(string, fs.DirEntry) bool) ([]AndrewPage, error) {
	pages := []AndrewPage{}
	localContentRoot := path.Dir(pagePath)

	err := fs.WalkDir(a.SiteFiles, localContentRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if filter(path, d) {
			// links require a URL relative to the page we're discovering siblings from, not from
			// the root of the file system
			page, err := NewPage(a, path)
			page = page.SetUrlPath(strings.TrimPrefix(path, localContentRoot+"/"))

			if err != nil {
				return err
			}

			pages = append(pages, page)
		}

		return nil
	})

	return pages, err
}
