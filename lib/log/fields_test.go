package log

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/privacy"
)

func TestShouldRedact(t *testing.T) {
	require.True(t, shouldRedact(privacy.Private))
	require.False(t, shouldRedact(privacy.Unknown))
	require.False(t, shouldRedact(privacy.Public))
}
