package gitcli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/internal/common"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestMarkRepoMaybeCorrupt(t *testing.T) {
	dir := t.TempDir()
	gitDir := common.GitDir(dir)

	t.Run("creates file if not exists", func(t *testing.T) {
		require.NoError(t, markRepoMaybeCorrupt(gitDir))

		_, err := os.Stat(filepath.Join(dir, RepoMaybeCorruptFlagFilepath))
		require.NoError(t, err)
	})

	t.Run("updates mtime if file exists", func(t *testing.T) {
		filePath := filepath.Join(dir, RepoMaybeCorruptFlagFilepath)
		f, err := os.Create(filePath)
		require.NoError(t, err)
		require.NoError(t, f.Close())

		oldMtime, err := getFileMtime(filePath)
		require.NoError(t, err)

		err = markRepoMaybeCorrupt(gitDir)
		require.NoError(t, err)

		newMtime, err := getFileMtime(filePath)
		require.NoError(t, err)

		if !newMtime.After(oldMtime) {
			t.Errorf("file mtime not updated")
		}
	})
}

func getFileMtime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

func TestIsContextErr(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "context canceled",
			err:      context.Canceled,
			expected: true,
		},
		{
			name:     "context deadline exceeded",
			err:      context.DeadlineExceeded,
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "multi error",
			err:      errors.Append(errors.New("foo bar"), context.Canceled),
			expected: true,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := isContextErr(test.err)
			require.Equal(t, test.expected, actual)
		})
	}
}
