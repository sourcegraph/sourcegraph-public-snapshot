// +build pgsqltest

package pgsql

import (
	"testing"

	"src.sourcegraph.com/sourcegraph/store/testsuite"
)

func TestInvites(t *testing.T) {
	t.Parallel()

	ctx, done := testContext()
	defer done()

	testsuite.Invites_test(ctx, t, &Invites{})
}
