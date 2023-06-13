package exportloopref

import (
	"github.com/kyoh86/exportloopref"
	"github.com/sourcegraph/sourcegraph/dev/linters/nolint"
)

var Analyzer = nolint.Wrap(exportloopref.Analyzer)
