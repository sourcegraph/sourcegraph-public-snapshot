package gitserver

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"

	"github.com/hexops/autogold/v2"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func Test_IsEmptyRepoError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		err  error
		want autogold.Value
	}{
		{
			err:  errors.Newf("Another message: %w", errors.New("Not an empty repo")),
			want: autogold.Expect(false),
		},
		{
			err:  errors.New("A different error"),
			want: autogold.Expect(false),
		},
		{
			err:  &gitdomain.RevisionNotFoundError{Repo: "foo", Spec: "HEAD"},
			want: autogold.Expect(true),
		},
		{
			err:  errors.Wrapf(&gitdomain.RevisionNotFoundError{Repo: "foo", Spec: "HEAD"}, "nice!"),
			want: autogold.Expect(true),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.err.Error(), func(t *testing.T) {
			got := isFirstCommitEmptyRepoError(tc.err)
			tc.want.Equal(t, got)
		})
	}
}
