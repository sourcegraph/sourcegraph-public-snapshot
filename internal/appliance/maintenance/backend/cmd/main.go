package main

import (
	"log"

	"sourcegraph.com/operator/api/api"
)

func main() {
	server := api.New()
	log.Println("Starting mock API server")
	server.Run()
}
