package pbt

import (
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

func TestCommitID(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		commitID := CommitID().Draw(t, "")
		_, err := api.NewCommitID(string(commitID))
		require.NoError(t, err, "CommitID generator is buggy")
	})
}
