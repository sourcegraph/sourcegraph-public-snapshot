package app

import (
	"fmt"
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/errorutil"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/ui"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"
	httpapiauth "sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/httpapi/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/trace"
)

// NewHandler returns a new app handler that uses the app router.
func NewHandler() http.Handler {
	session.SetSessionStore(session.NewRedisStore(globals.AppURL.Scheme == "https"))

	r := router.Router()

	m := http.NewServeMux()

	m.Handle("/", r)

	m.Handle("/__version", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, env.Version)
	}))
	m.Handle("/healthz", http.HandlerFunc(handleHealthCheck)) // healthz is a conventional name for "health check"

	r.Get(router.RobotsTxt).Handler(trace.TraceRoute(http.HandlerFunc(robotsTxt)))
	r.Get(router.Favicon).Handler(trace.TraceRoute(http.HandlerFunc(favicon)))
	r.Get(router.OpenSearch).Handler(trace.TraceRoute(http.HandlerFunc(openSearch)))

	r.Get(router.RepoBadge).Handler(trace.TraceRoute(errorutil.Handler(serveRepoBadge)))

	// Redirects
	r.Get(router.OldToolsRedirect).Handler(trace.TraceRoute(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/beta", 301)
	})))

	r.Get(router.GopherconLiveBlog).Handler(trace.TraceRoute(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://about.sourcegraph.com/go", 302)
	})))

	r.Get(router.GoSymbolURL).Handler(trace.TraceRoute(errorutil.Handler(serveGoSymbolURL)))

	r.Get(router.UI).Handler(ui.Router())

	r.Get(router.SignUp).Handler(trace.TraceRoute(http.HandlerFunc(serveSignUp)))
	r.Get(router.SiteInit).Handler(trace.TraceRoute(http.HandlerFunc(serveSiteInit)))
	r.Get(router.SignIn).Handler(trace.TraceRoute(http.HandlerFunc(serveSignIn)))
	r.Get(router.SignOut).Handler(trace.TraceRoute(http.HandlerFunc(serveSignOut)))
	r.Get(router.VerifyEmail).Handler(trace.TraceRoute(http.HandlerFunc(serveVerifyEmail)))
	r.Get(router.ResetPasswordInit).Handler(trace.TraceRoute(http.HandlerFunc(serveResetPasswordInit)))
	r.Get(router.ResetPassword).Handler(trace.TraceRoute(http.HandlerFunc(serveResetPassword)))

	r.Get(router.GDDORefs).Handler(trace.TraceRoute(errorutil.Handler(serveGDDORefs)))
	r.Get(router.Editor).Handler(trace.TraceRoute(errorutil.Handler(serveEditor)))

	r.Get(router.DebugHeaders).Handler(trace.TraceRoute(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del("Cookie")
		r.Header.Write(w)
	})))
	addDebugHandlers(r.Get(router.Debug).Subrouter())

	var h http.Handler = m
	h = session.CookieMiddleware(h)
	h = httpapiauth.AuthorizationMiddleware(h)

	return h
}

func serveSignOut(w http.ResponseWriter, r *http.Request) {
	session.DeleteSession(w, r)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
