package search

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestDisableExhaustiveSearchJob(t *testing.T) {
	os.Setenv("DISABLE_SEARCH_JOBS", "true")
	defer os.Unsetenv("DISABLE_SEARCH_JOBS")

	job := NewSearchJob()
	routines, err := job.Routines(context.Background(), &observation.TestContext)
	require.NoError(t, err)
	require.Nil(t, routines)
}
