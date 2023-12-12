// "OpenTracing interop with OpenTelemetry is set up, but the libraries are deprecated - use OpenTelemetry directly instead: https://go.opentelemetry.io/otel/trace"
package tracinglibraries

import (
	"fmt"

	"github.com/OpenPeeDeeP/depguard/v2"
	"golang.org/x/tools/go/analysis"

	"github.com/sourcegraph/sourcegraph/dev/linters/nolint"
)

const deniedMessage = `use "go.opentelemetry.io/otel/trace" instead`

var allowedFiles = []string{
	// Banned imports will match on the linter here
	"!**/dev/sg/linters",
	// Adapters here
	"!**/internal/tracer",
}

var deniedLibraries = map[string]string{
	// No OpenTracing
	"github.com/opentracing/opentracing-go": deniedMessage,
}

var Analyzer = createAnalyzer()

func createAnalyzer() *analysis.Analyzer {
	settings := &depguard.LinterSettings{
		"tracing libraries": &depguard.List{
			Files: allowedFiles,
			Deny:  deniedLibraries,
		},
	}

	analyzer, err := depguard.NewAnalyzer(settings)
	if err != nil {
		panic(fmt.Sprintf("failed to create deprecated tracing libraries analyzer: %v", err))
	}

	return nolint.Wrap(analyzer)
}
