package app

import (
	"net/http"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
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

func serveSignOutHandler(db dbutil.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logSignOutEvent(r, db, "SignOutAttempted")

		// Invalidate all user sessions first
		// This way, any other signout failures should not leave a valid session
		var err error
		if err = session.InvalidateSessionCurrentUser(w, r); err != nil {
			logSignOutEvent(r, db, "SignOutFailed")
			log15.Error("serveSignOutHandler", "err", err)
		}
		if err = session.SetActor(w, r, nil, 0, time.Time{}); err != nil {
			logSignOutEvent(r, db, "SignOutFailed")
			log15.Error("serveSignOutHandler", "err", err)
		}
		if ssoSignOutHandler != nil {
			ssoSignOutHandler(w, r)
		}
		if err == nil {
			logSignOutEvent(r, db, "SignOutSucceeded")
		}

		http.Redirect(w, r, "/search", http.StatusSeeOther)
	}
}

// logSignOutEvent records an event into the security event log.
func logSignOutEvent(r *http.Request, db dbutil.DB, name string) {
	// We don't want to begin logging events in on-premises installations yet.
	if !envvar.SourcegraphDotComMode() {
		return
	}

	ctx := r.Context()
	actor := actor.FromContext(ctx)

	if err := database.SecurityEventLogs(db).Insert(ctx, &database.SecurityEvent{
		Name:            name,
		URL:             r.URL.Path,
		UserID:          uint32(actor.UID),
		AnonymousUserID: "",
		Argument:        nil,
		Source:          "BACKEND",
		Timestamp:       time.Now(),
	}); err != nil {
		log15.Error("logSignOutEvent", "err", err)
	}
}
