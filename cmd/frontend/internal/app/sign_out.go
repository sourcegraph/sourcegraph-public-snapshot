package app

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	log15 "gopkg.in/inconshreveable/log15.v2"
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
	if err := session.SetActor(w, r, nil, 0); err != nil {
		log15.Error("Error in signout.", "err", err)
	}

	// TODO(sqs): Show the auth provider name corresponding to each signout URL (helpful when there
	// are multiple).
	var signoutURLs []SignOutURL
	if ssoSignOutHandler != nil {
		signoutURLs = ssoSignOutHandler(w, r)
	}
	if conf.MultipleAuthProvidersEnabled() {
		if len(signoutURLs) > 0 {
			renderSignoutPageTemplate(w, r, signoutURLs)
			return
		}
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
	buf.WriteTo(w)
}

var (
	signoutPageTemplate = template.Must(template.New("").Parse(`
<pre>
<strong>Signed out of Sourcegraph</strong>
<br>
<a href="/">Return to Sourcegraph</a>
<br>
{{range .SignoutURLs}}
<a href="{{.URL}}">Sign out of {{if .ProviderDisplayName}}{{.ProviderDisplayName}}{{else}}{{.ProviderServiceType}} authentication provider{{end}}</a><br>
{{end}}
</pre>
`))
)
