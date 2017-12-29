package app

import (
	"fmt"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/errorutil"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/redirects"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/ui2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"
	httpapiauth "sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/traceutil"
)

// NewHandler returns a new app handler that uses the provided app
// router.
func NewHandler(r *router.Router) http.Handler {
	session.SetSessionStore(session.MustNewRedisStore(globals.AppURL.Scheme == "https"))

	m := http.NewServeMux()

	m.Handle("/", r)

	m.Handle("/__version", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, env.Version)
	}))

	r.Get(router.RobotsTxt).Handler(traceutil.TraceRoute(http.HandlerFunc(robotsTxt)))
	r.Get(router.Favicon).Handler(traceutil.TraceRoute(http.HandlerFunc(favicon)))
	r.Get(router.OpenSearch).Handler(traceutil.TraceRoute(http.HandlerFunc(openSearch)))

	r.Get(router.RepoBadge).Handler(traceutil.TraceRoute(errorutil.Handler(serveRepoBadge)))

	// Redirects
	r.Get(router.OldToolsRedirect).Handler(traceutil.TraceRoute(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/beta", 301)
	})))

	r.Get(router.GopherconLiveBlog).Handler(traceutil.TraceRoute(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://about.sourcegraph.com/go", 302)
	})))

	r.Get(router.GoSymbolURL).Handler(traceutil.TraceRoute(errorutil.Handler(serveGoSymbolURL)))

	r.Get(router.UI).Handler(ui2.Router())

	r.Get(router.SignUp).Handler(traceutil.TraceRoute(http.HandlerFunc(serveSignUp)))
	r.Get(router.SignIn).Handler(traceutil.TraceRoute(http.HandlerFunc(serveSignIn)))
	r.Get(router.SignOut).Handler(traceutil.TraceRoute(http.HandlerFunc(serveSignOut)))
	r.Get(router.VerifyEmail).Handler(traceutil.TraceRoute(http.HandlerFunc(serveVerifyEmail)))
	r.Get(router.ResetPasswordInit).Handler(traceutil.TraceRoute(http.HandlerFunc(serveResetPasswordInit)))
	r.Get(router.ResetPassword).Handler(traceutil.TraceRoute(http.HandlerFunc(serveResetPassword)))

	r.Get(router.GDDORefs).Handler(traceutil.TraceRoute(errorutil.Handler(serveGDDORefs)))
	r.Get(router.Editor).Handler(traceutil.TraceRoute(errorutil.Handler(serveEditor)))

	r.Get(router.DebugHeaders).Handler(traceutil.TraceRoute(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del("Cookie")
		r.Header.Write(w)
	})))

	var h http.Handler = m
	h = redirects.RedirectsMiddleware(h)
	h = session.CookieMiddleware(h)
	h = httpapiauth.AuthorizationMiddleware(h)

	return h
}

func serveSignOut(w http.ResponseWriter, r *http.Request) {
	session.DeleteSession(w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
