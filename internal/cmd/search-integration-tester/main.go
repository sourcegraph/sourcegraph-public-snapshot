package main

import (
	"flag"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

var (
	endpoint = env.Get("ENDPOINT", "http://127.0.0.1:3080", "Sourcegraph frontend endpoint")
	token    = env.Get("ACCESS_TOKEN", "", "Access token")
	update   bool
)

func main() {
	flag.BoolVar(&update, "update", false, "Update golden test values in files")
	flag.Parse()
	if err := runSearchTests(); err != nil {
		fmt.Println(err)
	}
}
