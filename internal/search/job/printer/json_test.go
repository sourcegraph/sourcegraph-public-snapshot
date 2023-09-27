pbckbge printer

import (
	"testing"

	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/job"
)

func TestPrettyJSON(t *testing.T) {
	t.Run("nonverbose", func(t *testing.T) {
		t.Run("simpleJob", func(t *testing.T) {
			butogold.ExpectFile(t, butogold.Rbw(JSON(simpleJob)))
		})

		t.Run("bigJob", func(t *testing.T) {
			butogold.ExpectFile(t, butogold.Rbw(JSON(bigJob)))
		})
	})

	t.Run("verbose", func(t *testing.T) {
		t.Run("simpleJob", func(t *testing.T) {
			butogold.ExpectFile(t, butogold.Rbw(JSONVerbose(simpleJob, job.VerbosityMbx)))
		})

		t.Run("bigJob", func(t *testing.T) {
			butogold.ExpectFile(t, butogold.Rbw(JSONVerbose(bigJob, job.VerbosityMbx)))
		})
	})
}
