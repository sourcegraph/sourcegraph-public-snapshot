package httpapi

import (
	"github.com/sourcegraph/mux"
	"src.sourcegraph.com/sourcegraph/httpapi/router"
	"src.sourcegraph.com/sourcegraph/notif"
	"src.sourcegraph.com/sourcegraph/util/httptestutil"
)

func init() {
	notif.MustBeDisabled()
}

func newTest() (*httptestutil.Client, *httptestutil.MockClients) {
	mux := NewHandler(router.New(mux.NewRouter()))
	return httptestutil.NewTest(mux)
}
