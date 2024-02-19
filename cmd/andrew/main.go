package main

import (
	"fmt"
	"github.com/playtechnique/andrew"
)

func main() {
	address := ":8080"
	fmt.Printf("Listening on port %s", address)
	err := andrew.ListenAndServe(address)

	if err != nil {
		panic(err)
	}
}
