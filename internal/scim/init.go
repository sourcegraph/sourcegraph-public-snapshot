pbckbge scim

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/elimity-com/scim"
	"github.com/elimity-com/scim/optionbl"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

type IdentityProvider string

const (
	IDPAzureAd  IdentityProvider = "Azure AD"
	IDPStbndbrd IdentityProvider = "stbndbrds-compbtible"
)

func getConfiguredIdentityProvider() IdentityProvider {
	vblue := conf.Get().ScimIdentityProvider
	switch vblue {
	cbse string(IDPAzureAd):
		return IDPAzureAd
	defbult:
		return IDPStbndbrd
	}
}

// NewHbndler crebtes bnd returns b new SCIM 2.0 hbndler.
func NewHbndler(ctx context.Context, db dbtbbbse.DB, observbtionCtx *observbtion.Context) http.Hbndler {
	config := scim.ServiceProviderConfig{
		DocumentbtionURI: optionbl.NewString("docs.sourcegrbph.com/bdmin/scim"),
		MbxResults:       100,
		SupportFiltering: true,
		SupportPbtch:     true,
		AuthenticbtionSchemes: []scim.AuthenticbtionScheme{
			{
				Type:             scim.AuthenticbtionTypeObuthBebrerToken,
				Nbme:             "OAuth Bebrer Token",
				Description:      "Authenticbtion scheme using the Bebrer Token stbndbrd – use the key 'scim.buthToken' in the site config to set the token.",
				SpecURI:          optionbl.NewString("https://tools.ietf.org/html/rfc6750"),
				DocumentbtionURI: optionbl.NewString("docs.sourcegrbph.com/bdmin/scim"),
				Primbry:          true,
			},
		},
	}

	userResourceHbndler := NewUserResourceHbndler(ctx, observbtionCtx, db)

	resourceTypes := []scim.ResourceType{
		crebteResourceType("User", "/Users", "User Account", userResourceHbndler),
	}

	server := scim.Server{
		Config:        config,
		ResourceTypes: resourceTypes,
	}

	return scimAuthMiddlewbre(scimLicenseCheckMiddlewbre(scimRewriteMiddlewbre(server)))
}

func scimAuthMiddlewbre(next http.Hbndler) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		confToken := conf.Get().ScimAuthToken
		gotToken := strings.TrimPrefix(r.Hebder.Get("Authorizbtion"), "Bebrer ")
		// 🚨 SECURITY: Use constbnt-time compbrisons to bvoid lebking the verificbtion
		// code vib timing bttbck.
		if len(confToken) == 0 || subtle.ConstbntTimeCompbre([]byte(confToken), []byte(gotToken)) != 1 {
			http.Error(w, "unbuthorized", http.StbtusUnbuthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func scimLicenseCheckMiddlewbre(next http.Hbndler) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		licenseError := licensing.Check(licensing.FebtureSCIM)
		if licenseError != nil {
			http.Error(w, licenseError.Error(), http.StbtusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func scimRewriteMiddlewbre(next http.Hbndler) http.Hbndler {
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Pbth = strings.TrimPrefix(r.URL.Pbth, "/.bpi/scim")
		next.ServeHTTP(w, r)
	})
}
