package bodyclose

import (
	"github.com/timakin/bodyclose/passes/bodyclose"

	"github.com/sourcegraph/sourcegraph/dev/linters/nolint"
)

var Analyzer = nolint.Wrap(bodyclose.Analyzer)
