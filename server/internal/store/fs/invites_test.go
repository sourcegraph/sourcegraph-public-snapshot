package fs

import (
	"testing"

	"src.sourcegraph.com/sourcegraph/store/testsuite"
)

func TestInvites(t *testing.T) {
	ctx, done := testContext()
	defer done()

	testsuite.Invites_test(ctx, t, &Invites{})
}
