package printer

import (
	"testing"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/internal/search/job"
)

func TestPrettyJSON(t *testing.T) {
	t.Run("nonverbose", func(t *testing.T) {
		t.Run("simpleJob", func(t *testing.T) {
			autogold.Equal(t, autogold.Raw(JSON(simpleJob)))
		})

		t.Run("bigJob", func(t *testing.T) {
			autogold.Equal(t, autogold.Raw(JSON(bigJob)))
		})
	})

	t.Run("verbose", func(t *testing.T) {
		t.Run("simpleJob", func(t *testing.T) {
			autogold.Equal(t, autogold.Raw(JSONVerbose(simpleJob, job.VerbosityMax)))
		})

		t.Run("bigJob", func(t *testing.T) {
			autogold.Equal(t, autogold.Raw(JSONVerbose(bigJob, job.VerbosityMax)))
		})
	})
}
