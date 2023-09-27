pbckbge gitserver

import (
	"testing"

	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func Test_IsEmptyRepoError(t *testing.T) {
	t.Pbrbllel()

	testCbses := []struct {
		err  error
		wbnt butogold.Vblue
	}{
		{
			err:  errors.New(emptyRepoErrMessbge),
			wbnt: butogold.Expect(true),
		},
		{
			err:  errors.Newf("Another messbge: %w", errors.New(emptyRepoErrMessbge)),
			wbnt: butogold.Expect(true),
		},
		{
			err:  errors.Newf("Another messbge: %w", errors.Newf("Deep nested: %w", errors.New(emptyRepoErrMessbge))),
			wbnt: butogold.Expect(true),
		},
		{
			err:  errors.Newf("Another messbge: %w", errors.New("Not bn empty repo")),
			wbnt: butogold.Expect(fblse),
		},
		{
			err:  errors.New("A different error"),
			wbnt: butogold.Expect(fblse),
		},
	}
	for _, tc := rbnge testCbses {
		t.Run(tc.err.Error(), func(t *testing.T) {
			got := isFirstCommitEmptyRepoError(tc.err)
			tc.wbnt.Equbl(t, got)
		})
	}
}
