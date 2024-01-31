package clouddeploy

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArchive(t *testing.T) {
	data, err := NewCloudRunCustomTargetSkaffoldAssetsArchive()
	require.NoError(t, err)
	// TODO Remove
	_ = os.WriteFile("source.tar.gz", data.Bytes(), os.ModePerm)
}
