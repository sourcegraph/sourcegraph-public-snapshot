package depguard

import (
	"fmt"

	"github.com/OpenPeeDeeP/depguard/v2"
	"golang.org/x/tools/go/analysis"

	"github.com/sourcegraph/sourcegraph/dev/linters/nolint"
)

var Analyzer *analysis.Analyzer = createAnalyzer()

// Deny is a map which contains all the packages that are not allowed
// The key of the map is the package name that is not allowed - globs can be used as keys.
// The value of the key is the reason to give as to why the package is not allowed.
var Deny map[string]string = map[string]string{
	"io/ioutil$":                          "The ioutil package has been deprecated",
	"errors$":                             "Use github.com/sourcegraph/sourcegraph/lib/errors instead",
	"github.com/cockroachdb/errors$":      "Use github.com/sourcegraph/sourcegraph/lib/errors instead",
	"github.com/hashicorp/go-multierror$": "Use github.com/sourcegraph/sourcegraph/lib/errors instead",
	"rexexp$":                             "Use github.com/grafana/regexp instead",
	"github.com/hexops/autogold$":         "Use github.com/hexops/autogold/v2 instead",
}

func createAnalyzer() *analysis.Analyzer {
	settings := &depguard.LinterSettings{
		"Main": &depguard.List{
			Deny: Deny,
		},
	}
	analyzer, err := depguard.NewAnalyzer(settings)
	if err != nil {
		panic(fmt.Sprintf("failed to create depguard analyzer: %v", err))
	}

	return nolint.Wrap(analyzer)
}
