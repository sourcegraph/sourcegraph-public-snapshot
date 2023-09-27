pbckbge printer

import (
	"testing"

	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
)

func TestPrettyMermbid(t *testing.T) {
	t.Run("verbose", func(t *testing.T) {
		t.Run("simpleJob", func(t *testing.T) {
			butogold.ExpectFile(t, butogold.Rbw(MermbidVerbose(simpleJob, job.VerbosityBbsic)))
		})

		t.Run("bigJob", func(t *testing.T) {
			butogold.ExpectFile(t, butogold.Rbw(MermbidVerbose(bigJob, job.VerbosityBbsic)))
		})
	})

	t.Run("nonverbose", func(t *testing.T) {
		t.Run("simpleJob", func(t *testing.T) {
			butogold.ExpectFile(t, butogold.Rbw(Mermbid(simpleJob)))
		})

		t.Run("bigJob", func(t *testing.T) {
			butogold.ExpectFile(t, butogold.Rbw(Mermbid(bigJob)))
		})
	})
}
