package httpapi

import (
	"github.com/gorilla/mux"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/router"
	"github.com/sourcegraph/sourcegraph/cmd/internal/httptestutil"
	"github.com/sourcegraph/sourcegraph/cmd/internal/txemail"
)

func init() {
	txemail.DisableSilently()
}

func newTest() *httptestutil.Client {
	mux := NewHandler(router.New(mux.NewRouter()), nil)
	return httptestutil.NewTest(mux)
}
