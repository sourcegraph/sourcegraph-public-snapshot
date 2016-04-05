// +build exectest

package local_test

import (
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/server/testserver"
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

	// TODO(sqs): also test when there are no commits that we show a
	// "no commits yet" page, and same for when there are no files.

	if _, err := httpClient.GetOK(a.AbsURL("/r/r")); err != nil {
		t.Fatal(err)
	}
}
