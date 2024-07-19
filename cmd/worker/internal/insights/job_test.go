package insights

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestDisableInsightsJobs(t *testing.T) {
	os.Setenv("DISABLE_CODE_INSIGHTS", "true")
	defer os.Unsetenv("DISABLE_CODE_INSIGHTS")

	job := NewInsightsJob()
	routines, err := job.Routines(context.Background(), &observation.TestContext)
	require.NoError(t, err)
	require.Nil(t, routines)
}
