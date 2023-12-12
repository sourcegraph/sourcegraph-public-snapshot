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

const USE_LOG_INSTEAD = `use "github.com/sourcegraph/log" instead`

// Deny is a map which contains all the deprecated logging packages
// The key of the map is the package name that is not allowed - globs can be used as keys.
// The value of the key is the reason to give as to why the logger is not allowed.
var Deny = map[string]string{
	"log$":                              USE_LOG_INSTEAD,
	"github.com/inconshreveable/log15$": USE_LOG_INSTEAD,
	"go.uber.org/zap":                   USE_LOG_INSTEAD,
}

func createAnalyzer() *analysis.Analyzer {
	settings := &depguard.LinterSettings{
		"deprecated loggers": &depguard.List{
			Deny: Deny,
			Files: []string{
				// Let everything in dev use whatever they want
				"!**/dev",
				// // We allow one usage of a direct zap import here
				"!**/internal/observation/fields.go",
				// // Inits old loggers
				"!**/internal/logging/main.go",
				// // Dependencies require direct usage of zap
				"!**/cmd/frontend/internal/app/otlpadapter",
				// // Legacy and special case handling of panics in background routines
				"!**/lib/background/goroutine.go",
			},
		},
	}
	analyzer, err := depguard.NewAnalyzer(settings)
	if err != nil {
		panic(fmt.Sprintf("failed to create deprecated logging analyzer: %v", err))
	}

	return nolint.Wrap(analyzer)
}
