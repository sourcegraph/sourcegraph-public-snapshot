package sgapp

import (
	"golang.org/x/net/context"
	"src.sourcegraph.com/apps/updater"
	"src.sourcegraph.com/sourcegraph/conf/feature"
	"src.sourcegraph.com/sourcegraph/events"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/platform"
)

func init() {
	// Arrange sgapp so we can get a background app-level context during Start,
	// create a service with it and register the app frame.
	events.RegisterListener(sgapp{})
}

// sgapp implements events.EventListener.
type sgapp struct{}

func (sgapp) Scopes() []string {
	return []string{"app:updater"}
}

// Start creates a service using ctx and registers the app frame.
func (sgapp) Start(ctx context.Context) {
	if !feature.Features.RepoUpdater {
		return
	}

	handler := updater.New(ctx)

	platform.RegisterFrame(platform.RepoFrame{
		ID:      "updater",
		Title:   "Updater",
		Icon:    "cloud-download",
		Handler: handler,
		Enable:  func(*sourcegraph.Repo) bool { return true },
	})
}
