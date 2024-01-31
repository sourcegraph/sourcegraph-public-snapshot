package clouddeploy

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArchive(t *testing.T) {
	_, err := NewCloudRunCustomTargetSkaffoldAssetsArchive()
	require.NoError(t, err)
}
