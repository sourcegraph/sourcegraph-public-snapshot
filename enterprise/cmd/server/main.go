package main

import (
	"log"
	"os"
	"strconv"

	"sourcegraph.com/cmd/server/shared"
)

func main() {
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	if debug {
		log.Println("enterprise edition")
	}
	shared.Main()
}
