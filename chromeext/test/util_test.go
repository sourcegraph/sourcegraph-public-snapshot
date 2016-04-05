package test

import (
	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/httpapi"
	"sourcegraph.com/sourcegraph/sourcegraph/httpapi/router"
	"sourcegraph.com/sourcegraph/sourcegraph/services/notif"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httptestutil"
)

func init() {
	notif.MustBeDisabled()
}

func newTest() (*httptestutil.Client, *httptestutil.MockClients) {
	mux := httpapi.NewHandler(router.New(mux.NewRouter()))
	return httptestutil.NewTest(mux)
}
