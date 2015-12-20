package platform

import (
	"fmt"
	"net/http"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
)

// A RepoFrame is platform plugin point that allows an application to
// install a frame into the repository page. A menu item will appear in
// the repository subnav, to the right of "Code".
type RepoFrame struct {
	// ID is a unique identifier for the application. It is used in
	// the URL to identify URLs that belong to this app.
	ID string

	// Title is the user-visible name for the application.
	Title string

	// Icon specifies which octicon should serve as the application's
	// icon.
	Icon string

	// Handler is the HTTP handler that should return the HTML that
	// should be injected into the main repository page area.
	Handler http.Handler

	// Enable, if set, is called to determine whether the frame should
	// be enabled for a given repo. A frame is enabled for a repo if
	// EITHER its Enable func returns true OR if the repo config
	// specifies the app as being enabled.
	//
	// The Enable func, if provided, should not rely on external
	// resources or heavy computation, as it is called on every repo
	// page load for every registered frame.
	Enable func(*sourcegraph.Repo) bool
}

// repoFrames map key is app ID.
var repoFrames = map[string]RepoFrame{}

// CLI exposes the command line interface to platform applications. Applications
// may add their own CLI commands during initialization.
var CLI = cli.CLI

// RegisterFrame is the function that applications should call to
// add a frame to the repository page menu. This will panic if another
// frame has already been registered with the same ID.
//
// Requests to app's root page should always have "/" as the request URL.Path value,
// but the canonical app root page does not have a trailing slash.
func RegisterFrame(frame RepoFrame) {
	if _, exists := repoFrames[frame.ID]; exists {
		panic(fmt.Sprintf("RepoFrame with ID %s already exists", frame.ID))
	}
	repoFrames[frame.ID] = frame
}

// Frames returns all the frames registered in this instance of Sourcegraph.
// The map key is the app ID.
func Frames() map[string]RepoFrame {
	return repoFrames
}
