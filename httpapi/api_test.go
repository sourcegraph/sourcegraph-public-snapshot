package httpapi

import (
	"github.com/sourcegraph/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/httpapi/router"
	"sourcegraph.com/sourcegraph/sourcegraph/notif"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httptestutil"
)

func init() {
	notif.MustBeDisabled()
}

func newTest() (*httptestutil.Client, *httptestutil.MockClients) {
	mux := NewHandler(router.New(mux.NewRouter()))
	return httptestutil.NewTest(mux)
}
