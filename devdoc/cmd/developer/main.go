package main

import (
	"flag"
	"log"
	"net/http"

	"src.sourcegraph.com/sourcegraph/devdoc"
)

var port = flag.String("http", ":9999", "port on which to serve")

func main() {
	flag.Parse()
	log.Println("http://localhost" + *port)
	log.Fatal(http.ListenAndServe(*port, devdoc.New(nil)))
}
