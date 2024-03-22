package main

import (
	"fmt"
	"os"

	"github.com/playtechnique/andrew"
)

func main() {
	contentRoot := os.Args[1]

	if contentRoot == "" {
		contentRoot = "."
	}

	address := ":8080"
	fmt.Printf("Listening on port %s", address)

	err := andrew.ListenAndServe(address, contentRoot)

	if err != nil {
		panic(err)
	}
}
