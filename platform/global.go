package platform

import (
	"fmt"
	"net/http"

	"golang.org/x/net/context"
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

	// IconBadge, if not nil, is called with a per-request context
	// to check if there should be a badge dispayed over app icon.
	//
	// This func is called on every request to render a Sourcegraph page,
	// so it should be fast. It should also be safe for concurrent access.
	IconBadge func(ctx context.Context) (bool, error)

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

// OrderedEnabledGlobalApps returns an ordered list of global apps.
func OrderedEnabledGlobalApps() []GlobalApp {
	var apps []GlobalApp
	// There's at most 1 global app in the map at this time. Will need to sort in some way once there's more.
	for _, app := range GlobalApps {
		apps = append(apps, app)
	}
	return apps
}
