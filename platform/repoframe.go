package platform

import (
	"fmt"
	"net/http"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
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

	// EnabledRepos is a whitelist of repositories that have this
	// frame enabled. If empty, then the frame will be enabled for all
	// repositories.
	EnabledRepos map[sourcegraph.RepoSpec]struct{}
}

var repoFrames = map[string]RepoFrame{}

// RegisterFrame is the function that applications should call to
// add a frame to the repository page menu. This will panic if another
// frame has already been registered with the same ID.
func RegisterFrame(frame RepoFrame) {
	if _, exists := repoFrames[frame.ID]; exists {
		panic(fmt.Sprintf("RepoFrame with ID %s already exists", frame.ID))
	}
	repoFrames[frame.ID] = frame
}

// Frames returns the frames registered in this instance of Sourcegraph
func Frames(repo sourcegraph.RepoSpec) map[string]RepoFrame {
	frames := make(map[string]RepoFrame)
	for _, frame := range repoFrames {
		if frame.EnabledRepos != nil {
			if _, enabled := frame.EnabledRepos[repo]; enabled {
				frames[frame.ID] = frame
			}
		} else {
			frames[frame.ID] = frame
		}
	}
	return frames
}
