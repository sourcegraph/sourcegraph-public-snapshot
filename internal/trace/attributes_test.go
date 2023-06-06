package trace

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/stretchr/testify/require"
)

func Test_truncateError(t *testing.T) {
	cases := []struct {
		input  error
		limit  int
		output string
	}{{
		input:  errors.New("short error"),
		limit:  100,
		output: "short error",
	}, {
		input:  errors.New("super very very long error"),
		limit:  10,
		output: "super ...truncated... error",
	}}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			err := truncateError(tc.input, tc.limit)
			require.Equal(t, tc.output, err.Error())
		})
	}

	t.Run("nil error", func(t *testing.T) {
		err := truncateError(nil, 100)
		require.Nil(t, err)
	})
}
