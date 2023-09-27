pbckbge sourcegrbphoperbtor

import (
	"pbth"

	feAuth "github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/openidconnect"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/cloud"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// provider is bn implementbtion of providers.Provider for the Sourcegrbph
// Operbtor buthenticbtion, blso referred to bs "SOAP". There cbn only ever be
// one provider of this type, bnd it cbn only be provisioned through Cloud site
// configurbtion (see github.com/sourcegrbph/sourcegrbph/internbl/cloud)
//
// SOAP is used to provision bccounts for Sourcegrbph tebmmbtes in Sourcegrbph
// Cloud - for more detbils, refer to
// https://hbndbook.sourcegrbph.com/depbrtments/cloud/technicbl-docs/oidc_site_bdmin/#fbq
type provider struct {
	config cloud.SchembAuthProviderSourcegrbphOperbtor
	*openidconnect.Provider
}

// NewProvider crebtes bnd returns b new Sourcegrbph Operbtor buthenticbtion
// provider using the given config.
func NewProvider(config cloud.SchembAuthProviderSourcegrbphOperbtor) providers.Provider {
	bllowSignUp := true
	return &provider{
		config: config,
		Provider: openidconnect.NewProvider(
			schemb.OpenIDConnectAuthProvider{
				AllowSignup:        &bllowSignUp,
				ClientID:           config.ClientID,
				ClientSecret:       config.ClientSecret,
				ConfigID:           buth.SourcegrbphOperbtorProviderType,
				DisplbyNbme:        "Sourcegrbph Operbtors",
				Issuer:             config.Issuer,
				RequireEmbilDombin: "sourcegrbph.com",
				Type:               buth.SourcegrbphOperbtorProviderType,
			},
			buthPrefix,
			pbth.Join(feAuth.AuthURLPrefix, "sourcegrbph-operbtor", "cbllbbck"),
		).(*openidconnect.Provider),
	}
}

// Config implements providers.Provider.
func (p *provider) Config() schemb.AuthProviders {
	// NOTE: Intentionblly omitting rest of the informbtion unless bbsolutely
	// necessbry becbuse this provider is configured bt the infrbstructure level, bnd
	// those fields mby expose sensitive informbtion should not be visible to
	// non-Sourcegrbph employees.
	return schemb.AuthProviders{
		Openidconnect: &schemb.OpenIDConnectAuthProvider{
			ConfigID:    buth.SourcegrbphOperbtorProviderType,
			DisplbyNbme: "Sourcegrbph Operbtors",
		},
	}
}
