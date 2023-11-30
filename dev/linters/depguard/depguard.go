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
	"regexp$":                             "Use github.com/grafana/regexp instead",
	"github.com/hexops/autogold$":         "Use github.com/hexops/autogold/v2 instead",
	"github.com/google/go-github/github$": "Use github.com/google/go-github/v55/github instead. To convert between v48 and v55, use the internal/extsvc/github/githubconvert package",
}

func createAnalyzer() *analysis.Analyzer {
	for i := 1; i < 55; i++ {
		Deny[fmt.Sprintf("github.com/google/go-github/v%d/github$", i)] = "Use github.com/google/go-github/v55/github instead. To convert between v48 and v55, use the internal/extsvc/github/githubconvert package"
	}

	// We don't provide anything for the Files attribute, which means the "Main" list will apply
	// to all files. If we wanted to restrict our Deny list to a subset of files, we would add
	// Files: []string{"dev/**"}, which would mean it will only deny the import of some packages
	// in code under dev/**, thus ignore the rest of the code base.
	//
	// You can also create other lists, that apply different deny/allow lists. Ie:
	// "Test": &depguard.List{
	//	Files: []string{"*.test"}, // can also just use $test to match all test files
	//	Allow: []string{"$gostd", "github.com/strechr/testify"}
	// }
	// The above settings will make it that only imports from the Go standard lib and testify is allowed.
	// The rest will be denied
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
