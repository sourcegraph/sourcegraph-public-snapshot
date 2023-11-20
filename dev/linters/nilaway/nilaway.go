package nilaway

import (
	"github.com/sourcegraph/sourcegraph/dev/linters/nolint"

	"go.uber.org/nilaway"
)

// Analyzer is the top-level instance of Analyzer - it coordinates the entire dataflow to report
// nil flow errors in this package. It is needed here for nogo to recognize the package.
var Analyzer = nolint.Wrap(nilaway.Analyzer)
