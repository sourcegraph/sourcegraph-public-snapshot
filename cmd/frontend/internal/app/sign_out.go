package app

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/openidconnect"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/saml"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/session"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/sourcegraphoperator"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	internalauth "github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/cookie"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/telemetry"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/telemetryrecorder"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/telemetrystore/teestore"
)

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
			recorder.Record(ctx, "signOut", telemetry.ActionFailed, nil)
			logger.Error("serveSignOutHandler", log.Error(err))
		}

		if err = session.SetActor(w, r, nil, 0, time.Time{}); err != nil {
			recordSecurityEvent(r, db, database.SecurityEventNameSignOutFailed, err)
			recorder.Record(ctx, "signOut", telemetry.ActionFailed, nil)
			logger.Error("serveSignOutHandler", log.Error(err))
		}

		session.SetSignOutCookie(w)

		ssoSignOutHandler(w, r)

		if err == nil {
			recordSecurityEvent(r, db, database.SecurityEventNameSignOutSucceeded, nil)

			recorder.Record(ctx, "signOut", telemetry.ActionSucceeded, nil)
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

	// Safe to ignore this error
	anonymousUserID, _ := cookie.AnonymousUID(r)

	if errsec := db.SecurityEventLogs().LogSecurityEvent(ctx, name, r.URL.Path, uint32(a.UID), anonymousUserID, "BACKEND", arg); errsec != nil {
		log.Error(errsec)
	}

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
	//lint:ignore SA1019 existing usage of deprecated functionality. Use EventRecorder from internal/telemetryrecorder instead.
	_ = db.EventLogs().Insert(ctx, logEvent)
}

func ssoSignOutHandler(w http.ResponseWriter, r *http.Request) {
	logger := log.Scoped("ssoSignOutHandler")
	for _, p := range conf.Get().AuthProviders {
		var err error
		switch {
		case p.Openidconnect != nil:
			_, err = openidconnect.SignOut(w, r, openidconnect.SessionKey, openidconnect.GetProvider)
		case p.Saml != nil:
			_, err = saml.SignOut(w, r)
		}
		if err != nil {
			logger.Error("failed to clear auth provider session data", log.Error(err))
		}
	}

	if p := sourcegraphoperator.GetOIDCProvider(internalauth.SourcegraphOperatorProviderType); p != nil {
		_, err := openidconnect.SignOut(
			w,
			r,
			sourcegraphoperator.SessionKey,
			func(string) *openidconnect.Provider {
				return p
			},
		)
		if err != nil {
			logger.Error("failed to clear auth provider session data", log.Error(err))
		}
	}
}
