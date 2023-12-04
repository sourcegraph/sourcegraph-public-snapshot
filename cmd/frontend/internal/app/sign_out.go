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
	"github.com/sourcegraph/sourcegraph/internal/telemetry/teestore"
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
	logger = logger.Scoped("signOut")
	recorder := telemetryrecorder.NewBestEffort(logger, db)

	return func(w http.ResponseWriter, r *http.Request) {
		// In this code, we still use legacy events (usagestats.LogBackendEvent),
		// so do not tee events automatically.
		// TODO: We should remove this in 5.3 entirely
		ctx := teestore.WithoutV1(r.Context())

		recordSecurityEvent(r, db, database.SecurityEventNameSignOutAttempted, nil)

		// Invalidate all user sessions first
		// This way, any other signout failures should not leave a valid session
		var err error
		if err = session.InvalidateSessionCurrentUser(w, r, db); err != nil {
			recordSecurityEvent(r, db, database.SecurityEventNameSignOutFailed, err)
			recorder.Record(ctx, telemetry.FeatureSignOut, telemetry.ActionFailed, nil)
			logger.Error("serveSignOutHandler", log.Error(err))
		}

		if err = session.SetActor(w, r, nil, 0, time.Time{}); err != nil {
			recordSecurityEvent(r, db, database.SecurityEventNameSignOutFailed, err)
			recorder.Record(ctx, telemetry.FeatureSignOut, telemetry.ActionFailed, nil)
			logger.Error("serveSignOutHandler", log.Error(err))
		}

		auth.SetSignOutCookie(w)

		if ssoSignOutHandler != nil {
			ssoSignOutHandler(w, r)
		}

		if err == nil {
			recordSecurityEvent(r, db, database.SecurityEventNameSignOutSucceeded, nil)
			recorder.Record(ctx, telemetry.FeatureSignOut, telemetry.ActionSucceeded, nil)
		}

		http.Redirect(w, r, "/search", http.StatusSeeOther)
	}
}

// recordSecurityEvent records an event into the security event log.
func recordSecurityEvent(r *http.Request, db database.DB, name database.SecurityEventName, err error) {
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

	// Legacy event - TODO: Remove in 5.3, alongside the teestore.WithoutV1
	// context.
	logEvent := &database.Event{
		Name:            string(name),
		URL:             r.URL.Host,
		UserID:          uint32(a.UID),
		AnonymousUserID: "backend",
		Argument:        marshalled,
		Source:          "BACKEND",
		Timestamp:       time.Now(),
	}
	//lint:ignore SA1019 existing usage of deprecated functionality.
	// Use EventRecorder from internal/telemetryrecorder instead.
	_ = db.EventLogs().Insert(ctx, logEvent)
}
