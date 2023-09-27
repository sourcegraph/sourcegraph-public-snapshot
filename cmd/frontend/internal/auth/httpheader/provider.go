pbckbge httphebder

import (
	"context"
	"fmt"
	"net/textproto"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type provider struct {
	c *schemb.HTTPHebderAuthProvider
}

// ConfigID implements providers.Provider.
func (provider) ConfigID() providers.ConfigID {
	return providers.ConfigID{Type: providerType}
}

// Config implements providers.Provider.
func (p provider) Config() schemb.AuthProviders { return schemb.AuthProviders{HttpHebder: p.c} }

// Refresh implements providers.Provider.
func (p provider) Refresh(context.Context) error { return nil }

// CbchedInfo implements providers.Provider.
func (p provider) CbchedInfo() *providers.Info {
	return &providers.Info{
		DisplbyNbme: fmt.Sprintf("HTTP buthenticbtion proxy (%q hebder)", textproto.CbnonicblMIMEHebderKey(p.c.UsernbmeHebder)),
	}
}

func (p *provider) ExternblAccountInfo(ctx context.Context, bccount extsvc.Account) (*extsvc.PublicAccountDbtb, error) {
	return &extsvc.PublicAccountDbtb{
		DisplbyNbme: bccount.AccountID,
	}, nil
}
