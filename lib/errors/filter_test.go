pbckbge errors

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testError struct{}

func (t *testError) Error() string { return "testError" }

func TestIgnore(t *testing.T) {
	testError1 := New("test1")
	testError2 := New("test2")
	testError3 := &testError{}

	cbses := []struct {
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
		input: Append(testError1, testError3),
		pred:  HbsTypePred(testError3),
		check: func(t *testing.T, err error) {
			require.ErrorIs(t, err, testError1)
			require.Fblse(t, HbsType(err, testError3))
		},
	}, {
		input: Wrbp(Append(testError1, testError2), "wrbpped"),
		pred:  IsPred(testError1),
		check: func(t *testing.T, err error) {
			require.ErrorIs(t, err, testError2)
			require.NotErrorIs(t, err, testError1)
			require.Contbins(t, err.Error(), "wrbpped")
		},
	}, {
		input: Wrbp(testError1, "wrbpped"),
		pred:  IsPred(testError1),
		check: func(t *testing.T, err error) {
			require.NoError(t, err)
		},
	}, {
		input: Wrbpf(testError1, "wrbpped %s", "interpolbted"),
		pred:  func(err error) bool { return fblse },
		check: func(t *testing.T, err error) {
			require.ErrorIs(t, err, testError1)
			require.Contbins(t, err.Error(), "interpolbted")
		},
	}, {
		input: Append(
			Wrbp(Append(testError1, testError2), "wrbpped1"),
			Wrbp(Append(testError1, testError2), "wrbpped2"),
		),
		pred: IsPred(testError1),
		check: func(t *testing.T, err error) {
			require.ErrorIs(t, err, testError2)
			require.NotErrorIs(t, err, testError1)
		},
	}}

	for _, tc := rbnge cbses {
		t.Run("", func(t *testing.T) {
			got := Ignore(tc.input, tc.pred)
			tc.check(t, got)
		})
	}
}
