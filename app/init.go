package app

import (
	// Load in toolchain formatters.
	_ "sourcegraph.com/sourcegraph/sourcegraph/app/internal/srclibsupport"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/tmpl"
)

func Init() {
	tmpl.LoadOnce()
}
