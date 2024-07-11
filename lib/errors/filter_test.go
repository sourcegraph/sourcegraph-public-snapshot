package errors

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testError struct{}

func (t *testError) Error() string { return "testError" }

func TestIgnore(t *testing.T) {
	testError1 := New("test1")
	testError2 := New("test2")

	cases := []struct {
		input error
		pred  func(error) bool
		check func(*testing.T, error)
	}{{
		input: testError1,
		pred:  IsPred(testError2),
		check: func(t *testing.T, err error) {
			require.ErrorIs(t, err, testError1)
		},
	}, {
		input: testError1,
		pred:  IsPred(testError1),
		check: func(t *testing.T, err error) {
			require.NoError(t, err)
		},
	}, {
		input: Append(testError1, testError2),
		pred:  IsPred(testError1),
		check: func(t *testing.T, err error) {
			require.ErrorIs(t, err, testError2)
			require.NotErrorIs(t, err, testError1)
		},
	}, {
		input: Append(testError1, testError2, testError1),
		pred:  IsPred(testError1),
		check: func(t *testing.T, err error) {
			require.ErrorIs(t, err, testError2)
			require.NotErrorIs(t, err, testError1)
		},
	}, {
		input: Append(testError1, testError1),
		pred:  IsPred(testError1),
		check: func(t *testing.T, err error) {
			require.NoError(t, err)
		},
	}, {
		input: Wrap(Append(testError1, testError2), "wrapped"),
		pred:  IsPred(testError1),
		check: func(t *testing.T, err error) {
			require.ErrorIs(t, err, testError2)
			require.NotErrorIs(t, err, testError1)
			require.Contains(t, err.Error(), "wrapped")
		},
	}, {
		input: Wrap(testError1, "wrapped"),
		pred:  IsPred(testError1),
		check: func(t *testing.T, err error) {
			require.NoError(t, err)
		},
	}, {
		input: Wrapf(testError1, "wrapped %s", "interpolated"),
		pred:  func(err error) bool { return false },
		check: func(t *testing.T, err error) {
			require.ErrorIs(t, err, testError1)
			require.Contains(t, err.Error(), "interpolated")
		},
	}, {
		input: Append(
			Wrap(Append(testError1, testError2), "wrapped1"),
			Wrap(Append(testError1, testError2), "wrapped2"),
		),
		pred: IsPred(testError1),
		check: func(t *testing.T, err error) {
			require.ErrorIs(t, err, testError2)
			require.NotErrorIs(t, err, testError1)
		},
	}}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			got := Ignore(tc.input, tc.pred)
			tc.check(t, got)
		})
	}
}
