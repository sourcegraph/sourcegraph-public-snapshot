package platform

import (
	"fmt"
	"html/template"
	"net/http"
)

// SearchFrame is a platform plugin point that lets an app add its own
// results to Sourcegraph search results.
type SearchFrame struct {
	// ID is a unique identifier for a registered search frame.
	ID string

	// Title is a human-readable that will be displayed on the
	// tab of results for this search frame.
	Name string

	// Icon specifies which octicon should serve as the applications
	// search icon.
	Icon string

	// Handler is the HTTP handler that is responsible for
	// returning search results specific to the SearchFrame.
	Handler http.Handler `json:"-"`

	// PerPage indicates how many results the app wishes
	// to return per page.
	PerPage int
}

// ResponseJSON defines the expected format of the json response
// body expected to be returned from any app SearchFrame.
//
// SearchFrameResponse should be imported by SearchFrame
// applications and used to serialize responses to search requests
// against the SearchFrame.
type SearchFrameResponse struct {
	// HTML is raw html to be rendered as search results
	// on the front end. This contract allows for simplicty
	// of implementation and flexibility in the rendered format.
	HTML template.HTML `json:"HTML"`

	// Total is the total number of results for the given
	// query.
	Total uint64 `json:"Total"`
}

// SearchFrameErrorResponse should be marshalled as a response on any error
// case during handling of an HTTP request by the search frame.
// Note that a status code should also be set by the handler
// to reflect the error case.
type SearchFrameErrorResponse struct {
	// Error contains a human-readable error message
	Error string `json:"Error"`
}

type SearchOptions struct {
	Query   string `json:"q"`
	PerPage int
	Page    int
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
