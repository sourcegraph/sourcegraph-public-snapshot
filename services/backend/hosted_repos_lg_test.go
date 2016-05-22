// +build exectest

package backend_test

import (
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/testserver"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httptestutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/testutil"
)

var httpClient = &httptestutil.Client{}

func TestHostedRepo_CreateCloneAndView(t *testing.T) {
	a, ctx := testserver.NewUnstartedServer()
	if err := a.Start(); err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	_, _, done, err := testutil.CreateAndPushRepo(t, ctx, "r/r")
	if err != nil {
		t.Fatal(err)
	}
	defer done()
}
