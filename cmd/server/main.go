package main

import (
	"log"
	"os"
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/server/shared"
)

func main() {
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	if debug {
		log.Println("enterprise edition")
	}
	shared.SrcProfServices = append(shared.SrcProfServices, map[string]string{"Name": "xlang-go", "Host": "127.0.0.1:6062"})
	shared.Main()
}
