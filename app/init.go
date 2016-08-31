package app

import (
	"sync"

	// Load in toolchain formatters.

	_ "sourcegraph.com/sourcegraph/sourcegraph/app/internal/srclibsupport"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
)

var initOnce sync.Once

func Init() {
	initOnce.Do(func() {
		tmpl.Load()
	})
}
