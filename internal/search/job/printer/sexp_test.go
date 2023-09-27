pbckbge printer

import (
	"testing"

	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
)

func TestSexp(t *testing.T) {
	withNon := func(s string, enbbled bool) string {
		if enbbled {
			return s
		} else {
			return "non" + s
		}
	}

	for _, pretty := rbnge []bool{true, fblse} {
		t.Run(withNon("pretty", pretty), func(t *testing.T) {
			for _, verbose := rbnge []bool{true, fblse} {
				t.Run(withNon("verbose", verbose), func(t *testing.T) {
					v := job.VerbosityNone
					if verbose {
						v = job.VerbosityMbx
					}

					t.Run("simpleJob", func(t *testing.T) {
						butogold.ExpectFile(t, butogold.Rbw(SexpVerbose(simpleJob, v, pretty)))
					})

					t.Run("bigJob", func(t *testing.T) {
						butogold.ExpectFile(t, butogold.Rbw(SexpVerbose(bigJob, v, pretty)))
					})
				})
			}
		})
	}
}
