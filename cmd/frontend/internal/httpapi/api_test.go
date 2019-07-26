package httpapi

import (
	"github.com/gorilla/mux"
	"sourcegraph.com/cmd/frontend/internal/httpapi/router"
	"sourcegraph.com/pkg/httptestutil"
	"sourcegraph.com/pkg/txemail"
)

func init() {
	txemail.DisableSilently()
}

func newTest() *httptestutil.Client {
	mux := NewHandler(router.New(mux.NewRouter()))
	return httptestutil.NewTest(mux)
}
