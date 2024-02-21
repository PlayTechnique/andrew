package main

import (
	"fmt"
	"github.com/playtechnique/andrew"
)

func main() {
	address := ":8080"
	fmt.Printf("Listening on port %s", address)
	server := andrew.FileSystemMuxer{ContentRoot: "."}
	err := andrew.ListenAndServe(address, server)

	if err != nil {
		panic(err)
	}
}
