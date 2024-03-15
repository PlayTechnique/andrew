package andrew

import (
	"net/http"
)

func ListenAndServe(address string, contentRoot string) error {

	andrewMuxer, err := NewAndrewMuxer(contentRoot)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", andrewMuxer.Serve)

	server := http.Server{
		Handler: mux,
		Addr:    address,
	}

	err = server.ListenAndServe()

	return err
}
