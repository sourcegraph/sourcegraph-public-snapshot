package app

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/openidconnect"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/saml"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func serveSignOut(w http.ResponseWriter, r *http.Request) {
	if err := session.SetActor(w, r, nil, 0); err != nil {
		log15.Error("Error in signout.", "err", err)
	}

	// TODO(sqs): Show the auth provider name corresponding to each signout URL (helpful when there
	// are multiple).
	var signoutURLs []string
	for _, p := range conf.AuthProviders() {
		var signoutURL string
		var err error
		switch {
		case p.Openidconnect != nil:
			signoutURL, err = openidconnect.SignOut(w, r)
		case p.Saml != nil && conf.EnhancedSAMLEnabled():
			signoutURL, err = saml.SignOut(w, r)
		}
		if signoutURL != "" {
			signoutURLs = append(signoutURLs, signoutURL)
		}
		if err != nil {
			log15.Error("Error clearing auth provider session data.", "err", err)
		}
	}
	if len(signoutURLs) > 0 {
		renderSignoutPageTemplate(w, r, signoutURLs)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func renderSignoutPageTemplate(w http.ResponseWriter, r *http.Request, signoutURLs []string) {
	data := struct {
		SignoutURLs []string
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
<a href="/">Go to Sourcegraph</a>
<br>
{{range .SignoutURLs}}
<a href="{{.}}">Sign out of authentication provider</a><br>
{{end}}
</pre>
`))
)
