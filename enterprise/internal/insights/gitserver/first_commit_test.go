package gitserver

import (
	"testing"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func Test_IsEmptyRepoError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		err  error
		want autogold.Value
	}{
		{
			err:  errors.New(emptyRepoErrMessage),
			want: autogold.Want("EmptyRepo", true),
		},
		{
			err:  errors.Newf("Another message: %w", errors.New(emptyRepoErrMessage)),
			want: autogold.Want("NestedEmptyRepoError", true),
		},
		{
			err:  errors.Newf("Another message: %w", errors.Newf("Deep nested: %w", errors.New(emptyRepoErrMessage))),
			want: autogold.Want("DeepNestedError", true),
		},
		{
			err:  errors.Newf("Another message: %w", errors.New("Not an empty repo")),
			want: autogold.Want("NestedNotEmptyRepoError", false),
		},
		{
			err:  errors.New("A different error"),
			want: autogold.Want("NotEmptyRepo", false),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			got := isFirstCommitEmptyRepoError(tc.err)
			tc.want.Equal(t, got)
		})
	}
}
