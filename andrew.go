package andrew

import (
	"io/fs"
	"net/http"
)

// ListenAndServe creates a server in the contentRoot, listening at the address, with links on autogenerated
// pages to the baseUrl.
// contentRoot - an fs.FS at some location, whether that's a virtual fs.FS such as an fs.Testfs or an
//
//	fs.FS at a location on your file system such as os.DirFS.
//
// address - some ip:port combination. The AndrewServer
func ListenAndServe(contentRoot fs.FS, address string, baseUrl string) error {

	andrewServer := NewAndrewServer(contentRoot, address, baseUrl)

	mux := http.NewServeMux()
	mux.HandleFunc("/", andrewServer.Serve)
	mux.HandleFunc("/sitemap.xml", andrewServer.ServeSiteMap)

	server := http.Server{
		Handler: mux,
		Addr:    andrewServer.Address,
	}

	err := server.ListenAndServe()

	return err
}
