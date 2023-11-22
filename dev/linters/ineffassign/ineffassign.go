package ineffassign

import (
	"github.com/gordonklaus/ineffassign/pkg/ineffassign"

	"github.com/sourcegraph/sourcegraph/dev/linters/nolint"
)

var Analyzer = nolint.Wrap(ineffassign.Analyzer)
