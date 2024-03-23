package andrew

import (
	"net/http"
)

func ListenAndServe(contentRoot string, address string, baseUrl string) error {

	andrewServer, err := NewAndrewServer(contentRoot, address, baseUrl)
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
