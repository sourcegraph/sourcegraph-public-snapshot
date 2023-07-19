package types

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/collections"
	"github.com/stretchr/testify/require"
)

func TestRepoURICache(t *testing.T) {
	index := collections.NewSet[string]("abc", "def")
	cache := NewRepoURICache(index)

	require.True(t, cache.Contains("abc"))
	require.False(t, cache.Contains("ghi"))
	require.Equal(t, index, cache.index)

	index2 := collections.NewSet[string]("xyz", "pqr")
	cache.Overwrite(index2)
	require.Equal(t, index2, cache.index)
}
