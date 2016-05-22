package app

import (
	"sync"

	// Load in toolchain formatters.

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
	_ "sourcegraph.com/sourcegraph/sourcegraph/pkg/srclib_support"
)

var initOnce sync.Once

func Init() {
	initOnce.Do(func() {
		tmpl.Load()
	})
}
