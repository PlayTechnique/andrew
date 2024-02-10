package andrew

import (
	"net/http"
	"os"
)

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
