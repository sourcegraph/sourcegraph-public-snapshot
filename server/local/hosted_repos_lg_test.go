// +build exectest

package local_test

import (
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/auth/authutil"
	"sourcegraph.com/sourcegraph/sourcegraph/server/testserver"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httptestutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/testutil"
)

var httpClient = &httptestutil.Client{}

func TestHostedRepo_CreateCloneAndView(t *testing.T) {
	if testserver.Store == "pgsql" {
		t.Skip()
	}

	a, ctx := testserver.NewUnstartedServer()
	a.Config.ServeFlags = append(a.Config.ServeFlags,
		&authutil.Flags{Source: "none", AllowAnonymousReaders: true},
	)
	if err := a.Start(); err != nil {
		t.Fatal(err)
	}
	defer a.Close()

	_, done, err := testutil.CreateAndPushRepo(t, ctx, "r/r")
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
