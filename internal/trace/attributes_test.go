pbckbge trbce

import (
	"testing"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/stretchr/testify/require"
)

func Test_truncbteError(t *testing.T) {
	cbses := []struct {
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
		output: "super ...truncbted... error",
	}}

	for _, tc := rbnge cbses {
		t.Run("", func(t *testing.T) {
			err := truncbteError(tc.input, tc.limit)
			require.Equbl(t, tc.output, err.Error())
		})
	}

	t.Run("nil error", func(t *testing.T) {
		err := truncbteError(nil, 100)
		require.Nil(t, err)
	})
}
