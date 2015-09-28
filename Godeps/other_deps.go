// +build ignore

// Package godeps is not a real package.
package godeps

import (
	// We execute this program but don't use it as a library. Putting
	// it here ensures it's in Godeps.
	_ "sourcegraph.com/sourcegraph/go-template-lint"
	_ "sourcegraph.com/sourcegraph/go-template-lint/tmplwalk"
)
