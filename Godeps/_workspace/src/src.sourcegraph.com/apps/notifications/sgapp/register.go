package sgapp

import (
	"net/http"

	"golang.org/x/net/context"
	notificationsapp "src.sourcegraph.com/apps/notifications"
	"src.sourcegraph.com/apps/notifications/common"
	kvnotifications "src.sourcegraph.com/apps/notifications/notifications/kv"
	"src.sourcegraph.com/sourcegraph/conf/feature"
	"src.sourcegraph.com/sourcegraph/events"
	"src.sourcegraph.com/sourcegraph/platform"
	"src.sourcegraph.com/sourcegraph/platform/notifications"
	"src.sourcegraph.com/sourcegraph/platform/pctx"
	"src.sourcegraph.com/sourcegraph/platform/putil"
)

func init() {
	if !feature.Features.NotificationCenter {
		return
	}

	// Arrange sgapp so we can get a background app-level context during Start,
	// create a service with it and register the app frame.
	events.RegisterListener(sgapp{})
}

// sgapp implements events.EventListener.
type sgapp struct{}

func (sgapp) Scopes() []string {
	return []string{"app:notifications"}
}

// Start creates a service using ctx and registers the app frame.
func (sgapp) Start(ctx context.Context) {
	service := kvnotifications.NewService(ctx, "notifications")

	// TODO: Try this out initially, see if this can/should be made better.
	//
	// Make notifications external API available.
	notifications.Service = service

	opt := notificationsapp.Options{
		Context: func(req *http.Request) context.Context {
			return putil.Context(req)
		},
		BaseURI: func(req *http.Request) string {
			ctx := putil.Context(req)
			return pctx.BaseURI(ctx)
		},
		CSRFToken: func(req *http.Request) string {
			ctx := putil.Context(req)
			return pctx.CSRFToken(ctx)
		},
		Verbatim: func(w http.ResponseWriter) {
			w.Header().Set("X-Sourcegraph-Verbatim", "true")
		},
		BaseState: func(req *http.Request) notificationsapp.BaseState {
			ctx := putil.Context(req)
			reqPath := req.URL.Path
			if reqPath == "/" {
				reqPath = ""
			}
			return notificationsapp.BaseState{
				State: common.State{
					BaseURI:   pctx.BaseURI(ctx),
					ReqPath:   reqPath,
					CSRFToken: pctx.CSRFToken(ctx),
				},
			}
		},
		HeadPre: `<style type="text/css">
	#main {
		margin: 20px auto 0 auto;
		line-height: initial;
	}
</style>`,
	}
	handler := notificationsapp.New(service, opt)

	platform.RegisterGlobalApp(platform.GlobalApp{
		ID:      "notifications",
		Title:   "Notification Center",
		Icon:    "TODO",
		Handler: handler,
	})
}
