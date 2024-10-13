package andrew

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server holds a reference to the paths in the fs.FS that correspond to
// each page that should be served.
// When a URL is requested, Server creates an Page for the file referenced
// in that URL and then serves the Page.
type Server struct {
	SiteFiles                     fs.FS  // The files being served
	BaseUrl                       string // The URL used in any links generated for this website that should contain the hostname.
	Address                       string // IpAddress:Port combo to be served on.
	Andrewtableofcontentstemplate string // The string we're searching for inside a Page that should be replaced with a template. Mightn't belong in the Server.
	RssTitle                      string // The title of your RSS feed.
	RssDescription                string // The description of your RSS feed. Go wild.
	HTTPServer                    *http.Server
}

// allRequestsByPathCounter creates a new prometheus counter for use in the Serve function, tracking all requests made, segregated by path.
// Note there can be many requests for a single page, as css etc is served.
var allRequestsByPathCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "andrew_server_serve_allrequestsbypath",
	Help: "The total number of all requests received by the andrew server, segregated by path",
}, []string{"allrequests"})

// allRequestsCounter creates a new prometheus counter for use in the Serve function, tracking all requests made.
// Note there can be many requests for a single page, as css etc is served.
var allRequestsCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "andrew_server_serve_allrequests",
	Help: "The total number of all requests received by the andrew server",
})

// allRequestsErrorsByPathCounter creates a new prometheus counter for use in the Serve function, tracking all of the error codes generated,
// organised by the path that generates the error.
var allRequestsErrorsByPathCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "andrew_server_serve_allrequests_errorsbypath",
	Help: "The total number of all requests received by the andrew server, segregated by path",
}, []string{"path", "status"})

// NewServer builds your web server.
// contentRoot: an fs.FS of the files that you're serving.
// address: The ip address to bind this web server to.
// baseUrl: https://example.com or http://www.example.com
// rssTitle: The title of the RSS feed that shares your site.
// rssDescription: The description for your RSS feed. Jazz it up.
// Returns an [Server].
func NewServer(contentRoot fs.FS, address, baseUrl string, rssInfo RssInfo) *Server {
	s := &Server{
		SiteFiles:                     contentRoot,
		Andrewtableofcontentstemplate: "AndrewTableOfContents",
		Address:                       address,
		BaseUrl:                       baseUrl,
		RssTitle:                      rssInfo.Title,
		RssDescription:                rssInfo.Description,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.Serve)
	mux.HandleFunc("/sitemap.xml", s.ServeSiteMap)
	mux.HandleFunc("/rss.xml", s.ServeRssFeed)
	mux.Handle("/metrics", promhttp.Handler())

	s.HTTPServer = &http.Server{
		Handler: mux,
		Addr:    address,
	}

	return s
}

// Serve handles requests for any URL. It checks whether the request is for
// an index.html page or for anything else (another page, css, javascript etc).
// If a directory is requested, Serve defaults to finding the index.html page
// within that directory. Detecting this case for
func (a Server) Serve(w http.ResponseWriter, r *http.Request) {

	pagePath := path.Clean(r.RequestURI)
	allRequestsByPathCounter.WithLabelValues(pagePath).Inc()
	allRequestsCounter.Inc()

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
		allRequestsErrorsByPathCounter.WithLabelValues(pagePath, strconv.Itoa(status)).Inc()
		w.WriteHeader(status)
		fmt.Fprint(w, message)
		return
	}

	a.serve(w, page)
}

func (a *Server) ListenAndServe() error {
	return a.HTTPServer.ListenAndServe()
}

func (a *Server) ListenAndServeTLS(certPath string, privateKeyPath string) error {
	return a.HTTPServer.ListenAndServeTLS(certPath, privateKeyPath)
}

func (a *Server) Close() error {
	return a.HTTPServer.Close()
}

// serve writes to the ResponseWriter any arbitrary html file, or css, javascript, images etc.
func (a Server) serve(w http.ResponseWriter, page Page) {
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
		promauto.NewCounter(prometheus.CounterOpts{
			Name: "andrew_404_total",
			Help: "The total number of 404s",
		})
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
func (a Server) GetSiblingsAndChildren(pagePath string) ([]Page, error) {

	pages := []Page{}
	localContentRoot := path.Dir(pagePath)

	err := fs.WalkDir(a.SiteFiles, localContentRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// We don't list index files in our collection of siblings and children, because I don't
		// want a link back to a page that contains only links.
		if strings.Contains(path, "index.html") {
			return nil
		}

		// If the file we're considering isn't an html file, let's move on with our day.
		if !strings.Contains(path, "html") {
			return nil
		}

		pageContent, err := fs.ReadFile(a.SiteFiles, path)
		if err != nil {
			return err
		}

		title, err := getTitle(path, pageContent)
		if err != nil {
			return err
		}

		publishTime, err := getPublishTime(a.SiteFiles, path, pageContent)
		if err != nil {
			return err
		}

		// links require a URL relative to the page we're discovering siblings from, not from
		// the root of the file system
		s_page := Page{
			Title:       title,
			UrlPath:     strings.TrimPrefix(path, localContentRoot+"/"),
			Content:     string(pageContent),
			PublishTime: publishTime,
		}

		pages = append(pages, s_page)

		return nil
	})

	return pages, err
}
