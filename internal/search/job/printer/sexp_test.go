package printer

import (
	"testing"

	"github.com/hexops/autogold/v2"

	"github.com/sourcegraph/sourcegraph/internal/search/job"
)

func TestSexp(t *testing.T) {
	withNon := func(s string, enabled bool) string {
		if enabled {
			return s
		} else {
			return "non" + s
		}
	}

	for _, pretty := range []bool{true, false} {
		t.Run(withNon("pretty", pretty), func(t *testing.T) {
			for _, verbose := range []bool{true, false} {
				t.Run(withNon("verbose", verbose), func(t *testing.T) {
					v := job.VerbosityNone
					if verbose {
						v = job.VerbosityMax
					}

					t.Run("simpleJob", func(t *testing.T) {
						autogold.ExpectFile(t, autogold.Raw(SexpVerbose(simpleJob, v, pretty)))
					})

					t.Run("bigJob", func(t *testing.T) {
						autogold.ExpectFile(t, autogold.Raw(SexpVerbose(bigJob, v, pretty)))
					})
				})
			}
		})
	}
}
