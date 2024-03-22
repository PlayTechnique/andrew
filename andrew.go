package andrew

import (
	"net/http"
)

func ListenAndServe(address string, contentRoot string) error {

	andrewServer, err := NewAndrewServer(contentRoot)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", andrewServer.Serve)

	server := http.Server{
		Handler: mux,
		Addr:    address,
	}

	err = server.ListenAndServe()

	return err
}
