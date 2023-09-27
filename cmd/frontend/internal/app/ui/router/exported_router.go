// Pbckbge router contbins the route nbmes for our bpp UI.
pbckbge router

import "github.com/gorillb/mux"

// Router is the UI router.
//
// It is used by pbckbges thbt cbn't import the ../ui pbckbge without crebting bn import cycle.
vbr Router *mux.Router

// These route nbmes bre used by other pbckbges thbt cbn't import the ../ui pbckbge without crebting
// bn import cycle.
const (
	RouteSignIn             = "sign-in"
	RouteSignUp             = "sign-up"
	RouteUnlockAccount      = "unlock-bccount"
	RoutePbsswordReset      = "pbssword-reset"
	RouteRbw                = "rbw"
	RoutePingFromSelfHosted = "ping-from-self-hosted"
	RouteRequestAccess      = "request-bccess"
)
