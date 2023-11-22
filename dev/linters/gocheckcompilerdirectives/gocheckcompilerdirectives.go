package gocheckcompilerdirectives

import (
	"4d63.com/gocheckcompilerdirectives/checkcompilerdirectives"
	"github.com/sourcegraph/sourcegraph/dev/linters/nolint"
)

var Analyzer = nolint.Wrap(checkcompilerdirectives.Analyzer())
