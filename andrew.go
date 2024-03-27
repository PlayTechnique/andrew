package andrew

import (
	"net/http"
	"os"
	"path/filepath"
)

func ListenAndServe(contentRoot string, address string, baseUrl string) error {

	cr, err := filepath.Abs(contentRoot)
	if err != nil {
		return err
	}

	andrewServer, err := NewAndrewServer(os.DirFS(cr), address, baseUrl)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", andrewServer.Serve)
	mux.HandleFunc("/sitemap.xml", andrewServer.ServeSiteMap)

	server := http.Server{
		Handler: mux,
		Addr:    address,
	}

	err = server.ListenAndServe()

	return err
}
