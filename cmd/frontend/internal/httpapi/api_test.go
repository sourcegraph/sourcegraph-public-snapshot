package httpapi

import (
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httptestutil"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/router"
	"sourcegraph.com/sourcegraph/sourcegraph/services/notif"
)

func init() {
	notif.Disable()
}

func newTest() *httptestutil.Client {
	mux := NewHandler(router.New(nil))
	return httptestutil.NewTest(mux)
}
