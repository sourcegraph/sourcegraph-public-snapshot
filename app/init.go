package app

import (
	"log"
	"os"

	"sync"

	// Load in toolchain formatters.

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	_ "sourcegraph.com/sourcegraph/sourcegraph/srclib_support"
)

var initOnce sync.Once

func Init() {
	initOnce.Do(func() {
		tmpl.Load()
	})
}

var hostname string

func init() {
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
}
