package logging

import (
	"fmt"

	"github.com/OpenPeeDeeP/depguard/v2"
	"golang.org/x/tools/go/analysis"

	"github.com/sourcegraph/sourcegraph/dev/linters/nolint"
)

// This analyzer is modeled after the one in
// https://sourcegraph.sourcegraph.com/github.com/sourcegraph/sourcegraph@f6ae87add606c65876b87d378929fcb80c3bb493/-/blob/dev/linters/depguard/depguard.go
// These could potentially be combined into one analyzer.
var Analyzer *analysis.Analyzer = createAnalyzer()

const useLogInsteadMessage = `use "github.com/sourcegraph/log" instead`

// Deny is a map which contains all the deprecated logging packages
// The key of the map is the package name that is not allowed - globs can be used as keys.
// The value of the key is the reason to give as to why the logger is not allowed.
var Deny = map[string]string{
	"log$":                              useLogInsteadMessage,
	"github.com/inconshreveable/log15$": useLogInsteadMessage,
	"go.uber.org/zap":                   useLogInsteadMessage,
}

func createAnalyzer() *analysis.Analyzer {
	settings := &depguard.LinterSettings{
		"deprecated loggers": &depguard.List{
			Deny: Deny,
			Files: []string{
				// Let everything in dev use whatever they want
				"!**/dev/**/*.go",
			},
		},
	}
	analyzer, err := depguard.NewAnalyzer(settings)
	if err != nil {
		panic(fmt.Sprintf("failed to create deprecated logging analyzer: %v", err))
	}
	analyzer.Name = "logging"

	return nolint.Wrap(analyzer)
}
