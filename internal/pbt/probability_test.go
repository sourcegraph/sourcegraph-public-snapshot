package pbt

import (
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestBool(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		bools := rapid.SliceOfN(Bool(0.5), 256, 256).Draw(t, "bools")
		hasTrue := false
		hasFalse := false
		for _, b := range bools {
			hasTrue = hasTrue || b
			hasFalse = hasFalse || !b
			if hasTrue && hasFalse {
				break
			}
		}
		require.True(t, hasTrue, "failure probability should be 1/2^256")
		require.True(t, hasFalse, "failure probability should be 1/2^256")
	})
}
