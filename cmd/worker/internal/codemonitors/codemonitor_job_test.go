package codemonitors

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestDisableCodeMonitorsJob(t *testing.T) {
	os.Setenv("DISABLE_CODE_MONITORS", "true")
	defer os.Unsetenv("DISABLE_CODE_MONITORS")

	job := NewCodeMonitorJob()
	routines, err := job.Routines(context.Background(), &observation.TestContext)
	require.NoError(t, err)
	require.Nil(t, routines)
}
