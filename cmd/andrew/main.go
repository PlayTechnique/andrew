package main

import (
	"os"

	"github.com/playtechnique/andrew"
)

func main() {
	os.Exit(andrew.Main(os.Args[1:], os.Stdout))
}
