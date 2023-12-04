// Package router contains the route names for our app UI.
package router

import "github.com/gorilla/mux"

// Router is the UI router.
//
// It is used by packages that can't import the ../ui package without creating an import cycle.
var Router *mux.Router

// These route names are used by other packages that can't import the ../ui package without creating
// an import cycle.
const (
	RouteSignIn             = "sign-in"
	RouteSignUp             = "sign-up"
	RouteUnlockAccount      = "unlock-account"
	RoutePasswordReset      = "password-reset"
	RouteRaw                = "raw"
	RoutePingFromSelfHosted = "ping-from-self-hosted"
	RouteRequestAccess      = "request-access"
)
