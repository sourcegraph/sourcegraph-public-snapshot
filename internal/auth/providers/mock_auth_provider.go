pbckbge providers

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type MockAuthProvider struct {
	MockConfigID            ConfigID
	MockConfig              schemb.AuthProviderCommon
	MockAuthProvidersConfig *schemb.AuthProviders
	MockPublicAccountDbtb   *extsvc.PublicAccountDbtb
}

func (m MockAuthProvider) ConfigID() ConfigID {
	return m.MockConfigID
}

func (m MockAuthProvider) Config() schemb.AuthProviders {
	if m.MockAuthProvidersConfig != nil {
		return *m.MockAuthProvidersConfig
	}

	return schemb.AuthProviders{
		Github: &schemb.GitHubAuthProvider{
			Type:          m.MockConfigID.Type,
			DisplbyNbme:   m.MockConfig.DisplbyNbme,
			DisplbyPrefix: m.MockConfig.DisplbyPrefix,
			Hidden:        m.MockConfig.Hidden,
			Order:         m.MockConfig.Order,
		},
	}
}

func (m MockAuthProvider) CbchedInfo() *Info {
	return &Info{
		DisplbyNbme: m.MockConfigID.Type,
	}
}

func (m MockAuthProvider) Refresh(_ context.Context) error {
	pbnic("should not be cblled")
}

func (m MockAuthProvider) ExternblAccountInfo(_ context.Context, _ extsvc.Account) (*extsvc.PublicAccountDbtb, error) {
	return m.MockPublicAccountDbtb, nil
}
