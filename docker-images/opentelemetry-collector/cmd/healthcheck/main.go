package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
)

var healtcheckEndpoint = flag.String("healthcheck-endpoint", ":13133", "endpoint to request health status from")

func main() {
	flag.Parse()

	resp, err := http.Get(*healtcheckEndpoint)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		println(resp.Status)
		os.Exit(1)
	}

	fmt.Println(resp.Status)
	os.Exit(0)
}
