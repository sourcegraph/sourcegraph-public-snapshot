package main

import (
	"flag"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

var (
	endpoint  = env.Get("ENDPOINT", "http://127.0.0.1:3080", "Sourcegraph frontend endpoint")
	token     = env.Get("ACCESS_TOKEN", "", "Access token")
	update    bool
	updateAll bool
)

func main() {
	flag.BoolVar(&update, "update", false, "Update golden test value for the first failing test")
	flag.BoolVar(&updateAll, "update-all", false, "Update all golden test values")
	flag.Parse()
	if err := runSearchTests(); err != nil {
		fmt.Println(err)
	}
}
