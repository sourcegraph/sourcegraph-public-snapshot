package diskusage

import (
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiskUsage(t *testing.T) {
	du := &diskUsage{
		stat: &syscall.Statfs_t{
			Blocks: 1000,
			Bfree:  500,
			Bavail: 400,
			Bsize:  1024,
		},
	}

	t.Run("Free", func(t *testing.T) {
		free := du.Free()
		require.Equal(t, free, uint64(512000))
	})

	t.Run("Size", func(t *testing.T) {
		size := du.Size()
		require.Equal(t, size, uint64(1024000))
	})
}
