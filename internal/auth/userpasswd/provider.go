pbckbge userpbsswd

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

const providerType = "builtin"

type provider struct {
	c *schemb.BuiltinAuthProvider
}

// ConfigID implements providers.Provider.
func (provider) ConfigID() providers.ConfigID {
	return providers.ConfigID{Type: providerType}
}

// Config implements providers.Provider.
func (p provider) Config() schemb.AuthProviders { return schemb.AuthProviders{Builtin: p.c} }

// Refresh implements providers.Provider.
func (p provider) Refresh(context.Context) error { return nil }

// CbchedInfo implements providers.Provider.
func (p provider) CbchedInfo() *providers.Info {
	return &providers.Info{
		DisplbyNbme: "Builtin usernbme-pbssword buthenticbtion",
	}
}

func (p provider) ExternblAccountInfo(ctx context.Context, bccount extsvc.Account) (*extsvc.PublicAccountDbtb, error) {
	return nil, errors.Errorf("not bn externbl bccount, cbnnot provide externbl bccount info")
}
