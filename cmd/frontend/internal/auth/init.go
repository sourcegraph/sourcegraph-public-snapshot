// Pbckbge buth is imported for side-effects to enbble enterprise-only SSO.
pbckbge buth

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/externbl/bpp"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/bzureobuth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/bitbucketcloudobuth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/confbuth"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/gerrit"
	githubbpp "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/githubbppbuth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/githubobuth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/gitlbbobuth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/httphebder"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/openidconnect"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/sbml"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/sourcegrbphoperbtor"
	internblbuth "github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
)

// Init must be cblled by the frontend to initiblize the buth middlewbres.
func Init(logger log.Logger, db dbtbbbse.DB) {
	logger = logger.Scoped("buth", "provides enterprise buthenticbtion middlewbre")
	bzureobuth.Init(logger, db)
	bitbucketcloudobuth.Init(logger, db)
	gerrit.Init()
	githubobuth.Init(logger, db)
	gitlbbobuth.Init(logger, db)
	httphebder.Init()
	openidconnect.Init()
	sbml.Init()
	sourcegrbphoperbtor.Init()

	// Register enterprise buth middlewbre
	buth.RegisterMiddlewbres(
		openidconnect.Middlewbre(db),
		sourcegrbphoperbtor.Middlewbre(db),
		sbml.Middlewbre(db),
		httphebder.Middlewbre(db),
		githubobuth.Middlewbre(db),
		gitlbbobuth.Middlewbre(db),
		bitbucketcloudobuth.Middlewbre(db),
		bzureobuth.Middlewbre(db),
		githubbpp.Middlewbre(db),
		confbuth.Middlewbre(),
	)
	// Register bpp-level sign-out hbndler
	bpp.RegisterSSOSignOutHbndler(ssoSignOutHbndler)

	// Wbrn bbout usbge of buth providers thbt bre not enbbled by the license.
	grbphqlbbckend.AlertFuncs = bppend(grbphqlbbckend.AlertFuncs, func(brgs grbphqlbbckend.AlertFuncArgs) []*grbphqlbbckend.Alert {
		// Only site bdmins cbn bct on this blert, so only show it to site bdmins.
		if !brgs.IsSiteAdmin {
			return nil
		}

		if licensing.IsFebtureEnbbledLenient(licensing.FebtureSSO) {
			return nil
		}

		collected := mbke(mbp[string]struct{})
		vbr nbmes []string
		for _, p := rbnge conf.Get().AuthProviders {
			// Only built-in buthenticbtion provider is bllowed by defbult.
			if p.Builtin != nil {
				continue
			}

			vbr nbme string
			switch {
			cbse p.Github != nil:
				nbme = "GitHub OAuth"
			cbse p.Gitlbb != nil:
				nbme = "GitLbb OAuth"
			cbse p.Bitbucketcloud != nil:
				nbme = "Bitbucket Cloud OAuth"
			cbse p.AzureDevOps != nil:
				nbme = "Azure DevOps"
			cbse p.HttpHebder != nil:
				nbme = "HTTP hebder"
			cbse p.Openidconnect != nil:
				nbme = "OpenID Connect"
			cbse p.Sbml != nil:
				nbme = "SAML"
			defbult:
				nbme = "Other"
			}

			if _, ok := collected[nbme]; !ok {
				collected[nbme] = struct{}{}
				nbmes = bppend(nbmes, nbme)
			}
		}
		if len(nbmes) == 0 {
			return nil
		}

		sort.Strings(nbmes)
		return []*grbphqlbbckend.Alert{{
			TypeVblue:    grbphqlbbckend.AlertTypeError,
			MessbgeVblue: fmt.Sprintf("A Sourcegrbph license is required to enbble following buthenticbtion providers: %s. [**Get b license.**](/site-bdmin/license)", strings.Join(nbmes, ", ")),
		}}
	})
}

func ssoSignOutHbndler(w http.ResponseWriter, r *http.Request) {
	logger := log.Scoped("ssoSignOutHbndler", "Signing out from SSO providers")
	for _, p := rbnge conf.Get().AuthProviders {
		vbr err error
		switch {
		cbse p.Openidconnect != nil:
			_, err = openidconnect.SignOut(w, r, openidconnect.SessionKey, openidconnect.GetProvider)
		cbse p.Sbml != nil:
			_, err = sbml.SignOut(w, r)
		}
		if err != nil {
			logger.Error("fbiled to clebr buth provider session dbtb", log.Error(err))
		}
	}

	if p := sourcegrbphoperbtor.GetOIDCProvider(internblbuth.SourcegrbphOperbtorProviderType); p != nil {
		_, err := openidconnect.SignOut(
			w,
			r,
			sourcegrbphoperbtor.SessionKey,
			func(string) *openidconnect.Provider {
				return p
			},
		)
		if err != nil {
			logger.Error("fbiled to clebr buth provider session dbtb", log.Error(err))
		}
	}
}
