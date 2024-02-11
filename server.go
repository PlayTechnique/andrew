package andrew

import (
	"fmt"
	"net/http"
	"os"
)

func Serve() {
	contentRoot := os.Args[0]

	if _, err := os.Stat(contentRoot); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Usage: andrew <content directory>\n")
		os.Exit(1)
	}

	server := Server{ContentRoot: contentRoot, HttpServer: http.FileServer(http.Dir(contentRoot))}

	// Setup route
	http.HandleFunc("/", server.ServeUp)

	// Start HTTP server
	fmt.Println("Serving content on http://0.0.0.0:8080")
	if err := http.ListenAndServe("0.0.0.0:8080", nil); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %s\n", err)
		os.Exit(1)
	}

}

type Server struct {
	ContentRoot string
	HttpServer  http.Handler
}

// serveUp will check to see if a static file exists  that matches the name of the request.
// If that file does exist, it gets served.
// In the special case that you are serving from a directory that contains other directories,
// and those child directories contain files but the cwd does not, an index.html will be
// automatically generated.
func (s *Server) ServeUp(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	trailingCharacter := path[len(path)-1]

	if trailingCharacter == '/' {
		indexPath := s.ContentRoot + path + "index.html"

		if _, err := os.Stat(indexPath); os.IsNotExist(err) {
			s.generateIndex(w, r)
			return
		}
	}

	s.HttpServer.ServeHTTP(w, r)
}

func (s *Server) generateIndex(w http.ResponseWriter, r *http.Request) {
	return
}
