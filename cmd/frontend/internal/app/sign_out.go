package app

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/openidconnect"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func serveSignOut(w http.ResponseWriter, r *http.Request) {
	if err := session.SetActor(w, r, nil, 0); err != nil {
		log15.Error("Error in signout.", "err", err)
	}

	var err error
	p := conf.AuthProvider()
	switch {
	case p.Openidconnect != nil:
		var endSessionEndpoint string
		endSessionEndpoint, err = openidconnect.SignOut(w, r)
		if endSessionEndpoint != "" {
			// Load the end-session endpoint *and* redirect.
			renderEndSessionTemplate(w, r, endSessionEndpoint)
			return
		}
	}
	if err != nil {
		log15.Error("Error clearing auth provider session data.", "err", err)
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func renderEndSessionTemplate(w http.ResponseWriter, r *http.Request, endSessionEndpoint string) {
	data := struct {
		EndSessionURL string
	}{
		EndSessionURL: endSessionEndpoint,
	}

	var buf bytes.Buffer
	if err := renderEndSessionPageTemplate.Execute(&buf, data); err != nil {
		log15.Error("Error rendering end-session template.", "err", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	buf.WriteTo(w)
}

var (
	renderEndSessionPageTemplate = template.Must(template.New("").Parse(`
<pre>
<strong>Logged out of Sourcegraph</strong>
<br>
<a href="/">Go to Sourcegraph</a>
<br>
<a href="{{.EndSessionURL}}">Log out of authentication provider</a>
</pre>
`))
)
