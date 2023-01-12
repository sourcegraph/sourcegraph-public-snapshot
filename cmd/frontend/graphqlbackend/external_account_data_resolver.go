package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

type externalAccountDataResolver struct {
	data *extsvc.PublicAccountData
}

func NewExternalAccountDataResolver(ctx context.Context, account extsvc.Account) (*externalAccountDataResolver, error) {
	data, err := publicAccountDataFromJSON(ctx, account)
	if err != nil {
		return nil, err
	}
	return &externalAccountDataResolver{
		data: data,
	}, nil
}

func publicAccountDataFromJSON(ctx context.Context, account extsvc.Account) (*extsvc.PublicAccountData, error) {
	p := providers.GetProviderByConfigID(providers.ConfigID{Type: account.ServiceType, ID: account.ServiceID})
	if p == nil {
		return nil, nil
	}

	return p.ExternalAccountInfo(ctx, account)
}

func (r *externalAccountDataResolver) Username() *string {
	return r.data.DisplayName
}

func (r *externalAccountDataResolver) Login() *string {
	return r.data.Login
}

func (r *externalAccountDataResolver) URL() *string {
	return r.data.URL
}
