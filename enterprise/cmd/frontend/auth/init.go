// Package auth is imported for side-effects to enable enterprise-only SSO.
package auth

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/external/app"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/auth/githuboauth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/auth/gitlaboauth"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/auth/httpheader"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/auth/openidconnect"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/auth/saml"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func init() {
	// Register enterprise auth middleware
	auth.RegisterMiddlewares(
		openidconnect.Middleware,
		saml.Middleware,
		httpheader.Middleware,
		githuboauth.Middleware,
		gitlaboauth.Middleware,
	)
	// Register app-level sign-out handler
	app.RegisterSSOSignOutHandler(ssoSignOutHandler)
}

func ssoSignOutHandler(w http.ResponseWriter, r *http.Request) (signOutURLs []app.SignOutURL) {
	for _, p := range conf.Get().Critical.AuthProviders {
		var e app.SignOutURL
		var err error
		switch {
		case p.Openidconnect != nil:
			e.ProviderDisplayName = p.Openidconnect.DisplayName
			e.ProviderServiceType = p.Openidconnect.Type
			e.URL, err = openidconnect.SignOut(w, r)
		case p.Saml != nil:
			e.ProviderDisplayName = p.Saml.DisplayName
			e.ProviderServiceType = p.Saml.Type
			e.URL, err = saml.SignOut(w, r)
		case p.Github != nil:
			e.ProviderDisplayName = p.Github.DisplayName
			e.ProviderServiceType = p.Github.Type
			e.URL, err = githuboauth.SignOutURL(p.Github.Url)
		case p.Gitlab != nil:
			e.ProviderDisplayName = p.Gitlab.DisplayName
			e.ProviderServiceType = p.Gitlab.Type
			e.URL, err = gitlaboauth.SignOutURL(p.Gitlab.Url)
		}
		if e.URL != "" {
			signOutURLs = append(signOutURLs, e)
		}
		if err != nil {
			log15.Error("Error clearing auth provider session data.", "err", err)
		}
	}

	return signOutURLs
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_589(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
