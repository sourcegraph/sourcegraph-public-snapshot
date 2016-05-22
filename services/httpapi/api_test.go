package httpapi

import (
	"github.com/gorilla/mux"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httptestutil"
	"sourcegraph.com/sourcegraph/sourcegraph/services/httpapi/router"
	"sourcegraph.com/sourcegraph/sourcegraph/services/notif"
)

func init() {
	notif.MustBeDisabled()
}

func newTest() (*httptestutil.Client, *httptestutil.MockClients) {
	mux := NewHandler(router.New(mux.NewRouter()))
	return httptestutil.NewTest(mux)
}
