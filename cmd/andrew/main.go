package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/playtechnique/andrew"
)

func main() {

	for _, arg := range os.Args {
		if arg == "-h" || arg == "--help" {
			printHelp()
			return
		}
	}

	contentRoot := "."
	address := ":8080"
	baseUrl := "http://localhost:8080"

	if len(os.Args) >= 2 {
		contentRoot = os.Args[1]
	}

	if len(os.Args) >= 3 {
		address = os.Args[2]
	}

	if len(os.Args) >= 4 {
		baseUrl = os.Args[3]
	}

	fmt.Printf("Listening on %s, serving on %s", address, baseUrl)

	cr, err := filepath.Abs(contentRoot)
	if err != nil {
		panic(err)
	}

	err = andrew.ListenAndServe(os.DirFS(cr), address, baseUrl)

	if err != nil {
		panic(err)
	}
}

func printHelp() {
	fmt.Println("Usage: andrew [contentRoot] [address] [baseUrl]")
	fmt.Println(" - contentRoot: The root directory of your content. Defaults to '.' if not specified.")
	fmt.Println(" - address: The address to bind to. Defaults to 'localhost:8080' if not specified. If in doubt, you probably want 0.0.0.0:<something>")
	fmt.Println(" - base URL: The protocol://hostname for your server. Defaults to 'http://localhost:8080' if not specified. Used to generate sitemap/rss feed accurately.")
	fmt.Println(" -h, --help: Display this help message.")
}
