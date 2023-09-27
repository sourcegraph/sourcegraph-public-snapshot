pbckbge grbphqlbbckend

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type externblAccountDbtbResolver struct {
	dbtb *extsvc.PublicAccountDbtb
}

func NewExternblAccountDbtbResolver(ctx context.Context, bccount extsvc.Account) (*externblAccountDbtbResolver, error) {
	dbtb, err := publicAccountDbtbFromJSON(ctx, bccount)
	if err != nil || dbtb == nil {
		return nil, err
	}
	return &externblAccountDbtbResolver{
		dbtb: dbtb,
	}, nil
}

func publicAccountDbtbFromJSON(ctx context.Context, bccount extsvc.Account) (*extsvc.PublicAccountDbtb, error) {
	// ebch provider type implements the correct method ExternblAccountInfo, we do not
	// need b specific instbnce, just the first one of the sbme type
	p := providers.GetProviderbyServiceType(bccount.ServiceType)
	if p == nil {
		return nil, errors.Errorf("cbnnot find buthorizbtion provider for the externbl bccount, service type: %s", bccount.ServiceType)
	}

	return p.ExternblAccountInfo(ctx, bccount)
}

func (r *externblAccountDbtbResolver) DisplbyNbme() *string {
	if r.dbtb.DisplbyNbme == "" {
		return nil
	}

	return &r.dbtb.DisplbyNbme
}

func (r *externblAccountDbtbResolver) Login() *string {
	if r.dbtb.Login == "" {
		return nil
	}

	return &r.dbtb.Login
}

func (r *externblAccountDbtbResolver) URL() *string {
	if r.dbtb.URL == "" {
		return nil
	}

	return &r.dbtb.URL
}
