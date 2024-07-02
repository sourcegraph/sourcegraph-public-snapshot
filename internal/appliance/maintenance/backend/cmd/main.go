package main

import (
	"log"

	"github.com/sourcegraph/sourcegraph/internal/appliance/maintenance/backend/api"
)

func main() {
	server := api.New()
	log.Println("Starting mock API server")
	server.Run()
}
