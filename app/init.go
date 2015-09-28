package app

import (
	"log"
	"os"

	"sync"

	// Load in toolchain formatters.

	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	_ "src.sourcegraph.com/sourcegraph/srclib_support"
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
