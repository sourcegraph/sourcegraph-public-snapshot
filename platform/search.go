package platform

import (
	"fmt"
	"net/http"
)

// SearchFrame is a platform plugin point that lets an app add its own
// results to Sourcegraph search results.
type SearchFrame struct {
	// ID is a unique identifier for a registered search frame.
	ID string

	// Title is a human-readable that will be displayed on the
	// tab of results for this search frame.
	Title string

	// Icon specifies which octicon should serve as the applications
	// search icon.
	Icon string

	// Handler is the HTTP handler that is responsible for
	// returning search results specific to the SearchFrame.
	Handler http.Handler
}

type SearchOptions struct {
	Query   string
	PerPage int32
	Page    int32
}

var appSearchFrames = map[string]SearchFrame{}

func RegisterSearchFrame(appSearch SearchFrame) {
	if _, exists := appSearchFrames[appSearch.ID]; exists {
		panic(fmt.Sprintf("SearchFrame with ID %q already exists", appSearch.ID))
	}
	appSearchFrames[appSearch.ID] = appSearch
}

func SearchFrames() map[string]SearchFrame {
	return appSearchFrames
}
