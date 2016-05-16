// Package srclib_support imports def formatters.
//
// TODO(sqs): We should find a way to not have a Go source dependency on the
// toolchains.
package srclib_support

import (
	_ "github.com/sourcegraph/srclib-css/css_def"
	_ "sourcegraph.com/sourcegraph/srclib-docker/dockerfiledef"
	_ "sourcegraph.com/sourcegraph/srclib-go/golang_def"
	_ "sourcegraph.com/sourcegraph/srclib-haskell/haskell"
	_ "sourcegraph.com/sourcegraph/srclib-java/java_def"
	_ "sourcegraph.com/sourcegraph/srclib-python/python"
	_ "sourcegraph.com/sourcegraph/srclib-ruby/ruby_def"

	// Used by tests only, but tests spawn a separate sgx process that
	// expects this toolchain to be loaded, so we can't use a build
	// tag to only import it for the tests.
	_ "sourcegraph.com/sourcegraph/srclib-sample/sample"
)
