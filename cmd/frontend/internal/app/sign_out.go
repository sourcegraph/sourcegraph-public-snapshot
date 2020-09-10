package app

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
)

type SignOutURL struct {
	ProviderDisplayName string
	ProviderServiceType string
	URL                 string
}

var ssoSignOutHandler func(w http.ResponseWriter, r *http.Request) []SignOutURL

// RegisterSSOSignOutHandler registers a SSO sign-out handler that takes care of cleaning up SSO
// session state, both on Sourcegraph and on the SSO provider. This function should only be called
// once from an init function.
func RegisterSSOSignOutHandler(f func(w http.ResponseWriter, r *http.Request) []SignOutURL) {
	if ssoSignOutHandler != nil {
		panic("RegisterSSOSignOutHandler already called")
	}
	ssoSignOutHandler = f
}

func serveSignOut(w http.ResponseWriter, r *http.Request) {
	// Invalidate all user sessions first
	// This way, any other signout failures should not leave a valid session
	if err := session.InvalidateSessionCurrentUser(r); err != nil {
		log15.Error("Error in signout.", "err", err)
	}
	if err := session.SetActor(w, r, nil, 0); err != nil {
		log15.Error("Error in signout.", "err", err)
	}
	var signoutURLs []SignOutURL
	if ssoSignOutHandler != nil {
		signoutURLs = ssoSignOutHandler(w, r)
	}
	if len(signoutURLs) > 0 {
		renderSignoutPageTemplate(w, r, signoutURLs)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func renderSignoutPageTemplate(w http.ResponseWriter, r *http.Request, signoutURLs []SignOutURL) {
	data := struct {
		SignoutURLs []SignOutURL
	}{
		SignoutURLs: signoutURLs,
	}

	var buf bytes.Buffer
	if err := signoutPageTemplate.Execute(&buf, data); err != nil {
		log15.Error("Error rendering signout page template.", "err", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = buf.WriteTo(w)
}

var signoutPageTemplate = template.Must(template.New("").Parse(`
<style>
.auth-box,
.center {
    position: relative;
}
.auth-box,
a,
body {
    display: block;
}
a,
body,
h1 {
    text-align: center;
}
html {
    -ms-text-size-adjust: 100%;
    -webkit-text-size-adjust: 100%;
}
body {
	/* Body BG color from gray-23 browser/src/global-styles/colors.scss*/
    background-color: #0e121b;
}
* {
    font-family: system,-apple-system,San Francisco,\.SFNSDisplay-Regular,Segoe UI,Segoe,Segoe WP,Helvetica Neue,helvetica,Lucida Grande,arial,sans-serif;
    font-size: 14px;

	/* Text color from gray-01 browser/src/global-styles/colors.scss */
    color: #f2f4f8;
}
.auth-box {
    box-sizing: border-box;
    width: 100%;
    max-width: 420px;
    margin: 10vh auto auto;
    overflow: hidden;

	/* colors from browser/src/global-styles/colors.scss */
    border: 1px solid #2B3750;
    background-color: #1D2535;
    box-shadow: 0 0 8px 0 rgba(0,0,0,.4);
    padding: 14px;
    border-radius: 6px;
}
@media screen and (max-width:678px) {
    .auth-box {
        border: none;
        margin-top: 64px;
		/* colors from browser/src/global-styles/colors.scss */
        background-color: #0e121b;
    }
}
.logo {
    margin-top: 28px;
    width: atuo;
}
.center {
    padding-top: 48px;
}
h1 {
    font-size: 22px;
    font-weight: 500;
    margin-top: 40px;
}
a {
    font-size: 14px;
    line-height: 20px;
    border: none;
    border-radius: 2px;
	/* colors from browser/src/global-styles/colors.scss */
    background-color: #2b3750;
    margin-bottom: 20px;
    text-decoration: none;
    word-wrap: normal;
    transition: opacity 250ms ease;
    padding: 6px 12px;
}
a:last-of-type {
    margin-bottom: 0;
}
a:hover {
    opacity: 0.85;
}
a:active,
a:target {
    opacity: 0.7;
}
.primary {
    margin-top: 14px;
    color: #fff;
    background-color: #1663a9;
    border-color: #155d9e;
    margin-bottom: 48px;
}
</style>
<div class="auth-box">
<img class="logo" src="/.assets/img/sourcegraph-head-logo.svg">
<div class="center">
<h1>Signed out of Sourcegraph</h1>
<a class="primary" href="/">Return to Sourcegraph</a>
{{range .SignoutURLs}}
<a href="{{.URL}}">Sign out of {{if .ProviderDisplayName}}{{.ProviderDisplayName}}{{else}}{{.ProviderServiceType}} authentication provider{{end}}</a>
{{end}}
</div>
</div>
`))
