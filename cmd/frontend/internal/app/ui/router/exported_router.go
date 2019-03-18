package router

import "github.com/gorilla/mux"

// Router is the UI router.
//
// It is used by packages that can't import the ../ui package without creating an import cycle.
var Router *mux.Router

// These route names are used by other packages that can't import the ../ui package without creating
// an import cycle.
const (
	RouteSignIn        = "sign-in"
	RouteSignUp        = "sign-up"
	RoutePasswordReset = "password-reset"
)
