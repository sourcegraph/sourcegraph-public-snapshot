package app

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/cookie"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/session"
	"github.com/sourcegraph/sourcegraph/internal/telemetry"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/telemetryrecorder"
)

type SignOutURL struct {
	ProviderDisplayName string
	ProviderServiceType string
	URL                 string
}

var ssoSignOutHandler func(w http.ResponseWriter, r *http.Request)

// RegisterSSOSignOutHandler registers a SSO sign-out handler that takes care of cleaning up SSO
// session state, both on Sourcegraph and on the SSO provider. This function should only be called
// once from an init function.
func RegisterSSOSignOutHandler(f func(w http.ResponseWriter, r *http.Request)) {
	if ssoSignOutHandler != nil {
		panic("RegisterSSOSignOutHandler already called")
	}
	ssoSignOutHandler = f
}

func serveSignOutHandler(logger log.Logger, db database.DB) http.HandlerFunc {
	logger = logger.Scoped("signout", "sign-out request handler")
	events := telemetryrecorder.NewBestEffort(logger, db)

	return func(w http.ResponseWriter, r *http.Request) {
		logSignOutEvent(r, db, database.SecurityEventNameSignOutAttempted, nil)

		// Invalidate all user sessions first
		// This way, any other signout failures should not leave a valid session
		var err error
		if err = session.InvalidateSessionCurrentUser(w, r, db); err != nil {
			logSignOutEvent(r, db, database.SecurityEventNameSignOutFailed, err)
			events.Record(r.Context(), // context has actor
				telemetry.FeatureSignOut, telemetry.ActionFailed,
				telemetry.EventParameters{
					PrivateMetadata: map[string]any{"error": err.Error()},
				})
			logger.Error("serveSignOutHandler", log.Error(err))
		}

		if err = session.SetActor(w, r, nil, 0, time.Time{}); err != nil {
			logSignOutEvent(r, db, database.SecurityEventNameSignOutFailed, err)
			events.Record(r.Context(), // context has actor
				telemetry.FeatureSignOut, telemetry.ActionFailed,
				telemetry.EventParameters{
					PrivateMetadata: map[string]any{"error": err.Error()},
				})
			logger.Error("serveSignOutHandler", log.Error(err))
		}

		auth.SetSignOutCookie(w)

		if ssoSignOutHandler != nil {
			ssoSignOutHandler(w, r)
		}

		if err == nil {
			logSignOutEvent(r, db, database.SecurityEventNameSignOutSucceeded, nil)
			events.Record(r.Context(), // context has actor
				telemetry.FeatureSignOut, telemetry.ActionSucceeded,
				telemetry.EventParameters{})
		}

		http.Redirect(w, r, "/search", http.StatusSeeOther)
	}
}

// logSignOutEvent records an event into the security event log.
func logSignOutEvent(r *http.Request, db database.DB, name database.SecurityEventName, err error) {
	ctx := r.Context()
	a := actor.FromContext(ctx)

	arg := struct {
		Error string `json:"error"`
	}{}
	if err != nil {
		arg.Error = err.Error()
	}

	marshalled, _ := json.Marshal(arg)

	event := &database.SecurityEvent{
		Name:      name,
		URL:       r.URL.Path,
		UserID:    uint32(a.UID),
		Argument:  marshalled,
		Source:    "BACKEND",
		Timestamp: time.Now(),
	}

	// Safe to ignore this error
	event.AnonymousUserID, _ = cookie.AnonymousUID(r)

	db.SecurityEventLogs().LogEvent(ctx, event)
}
