package platform

import (
	"fmt"
	"net/http"
)

type GlobalApp struct {
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
}

// GlobalApps map key is app ID.
var GlobalApps = map[string]GlobalApp{}

func RegisterGlobalApp(app GlobalApp) {
	if _, exists := GlobalApps[app.ID]; exists {
		panic(fmt.Sprintf("Global App with ID %s already exists", app.ID))
	}
	GlobalApps[app.ID] = app
}
