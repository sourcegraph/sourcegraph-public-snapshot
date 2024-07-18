package own

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDisableOwnJobs(t *testing.T) {
	os.Setenv("DISABLE_OWN", "true")
	defer os.Unsetenv("DISABLE_OWN")

	job := NewOwnRepoIndexingQueue()
	routines, err := job.Routines(nil, nil)
	require.NoError(t, err)
	require.Nil(t, routines)
}
