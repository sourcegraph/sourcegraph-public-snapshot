package app

import (
	"fmt"
	"html"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/oauth2client"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/gcstracker"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/hubspot/hubspotutil"
)

// serveShowAuth prints the user's auth cookie. Local tools that need
// to access Sourcegraph as a certain user can direct the user to
// http://sourcegraph.com/-/show-auth to let the user obtain their own
// auth cookie.
func serveShowAuth(w http.ResponseWriter, r *http.Request) error {
	sessionCookie := session.SessionCookie(r)

	// If the user isn't logged in, redirect them to log in and come
	// back here.
	if sessionCookie == "" {
		redirectURL := conf.AppURL.ResolveReference(router.Rel.URLTo(router.GitHubOAuth2Receive))
		returnTo := router.Rel.URLTo(router.ShowAuth)
		return oauth2client.GitHubOAuth2Initiate(w, r, nil, redirectURL.String(), returnTo.String(), returnTo.String())
	}

	trackZapAuth(actor.FromContext(r.Context()))

	// sessionCookie is the value of the user's cookie. If an attacker
	// can set the user's cookie to an arbitrary value, then they
	// could inject content into this HTTP response. We mitigate this
	// risk by (1) setting the Content-Type to text/plain to reduce
	// the chance it's interpreted as a script, and (2) HTML-escaping
	// the string. In practice, if an attacker could SET a user's
	// cookie, they could probably do much more damage than anything
	// this endpoint would allow.
	w.Header().Set("content-type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "To authenticate with Sourcegraph enter your auth token in the 'zap auth' prompt:\n\n sg-session=%s", html.EscapeString(sessionCookie))
	return nil
}

// Zap authorization metrics.
func trackZapAuth(actor *actor.Actor) error {
	if actor.Email != "" {
		hubspotclient, err := hubspotutil.Client()
		if err != nil {
			return err
		}
		hubspotclient.LogEvent(actor.Email, hubspotutil.EventNameToHubSpotID["ZapAuthCompleted"], map[string]string{})

		gcsclient, err := gcstracker.NewFromUserInfo(&gcstracker.UserInfo{
			Email:          actor.Email,
			BusinessUserID: actor.Login,
		})
		if err != nil {
			return err
		}
		if gcsclient == nil {
			return nil
		}
		tos := gcsclient.NewTrackedObjects("ZapAuthCompleted")
		tos.AddUserDetailsObject(&gcstracker.UserDetailsContext{
			ZapAuthCompleted: true,
		})
		return gcsclient.Write(tos)
	}
	return nil
}
