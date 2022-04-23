package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/sourcegraph/sourcegraph/logsvc"
)

func main() {
	fmt.Println("Listening on :8081")
	http.ListenAndServe(":8081", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req logsvc.Request
		json.NewDecoder(r.Body).Decode(&req)
		log.Printf(req.Fmt, req.Args...)
	}))
}
