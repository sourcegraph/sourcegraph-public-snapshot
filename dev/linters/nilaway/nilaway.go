package nilaway

import (
	"go.uber.org/nilaway"
	"golang.org/x/tools/go/analysis"

	"github.com/sourcegraph/sourcegraph/dev/linters/nolint"
)

var Analyzer *analysis.Analyzer = nolint.Wrap(nilaway.Analyzer)
