package andrew

import "net/http"

type muxer interface {
	Serve(http.ResponseWriter, *http.Request)
	serveNonIndexPage(http.ResponseWriter, *http.Request)
	serveIndexPage(http.ResponseWriter, *http.Request)
}
