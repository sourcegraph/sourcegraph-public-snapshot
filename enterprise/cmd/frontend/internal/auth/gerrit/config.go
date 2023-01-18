package gerrit

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func Init() {
	const pkgName = "gerrit"
	go conf.Watch(func() {
		newProv := &Provider{}
		providers.Update(pkgName, []providers.Provider{newProv})
	})
}

type Provider struct{}

func (p *Provider) ConfigID() providers.ConfigID {
	return providers.ConfigID{
		Type: "gerrit",
		ID:   "http://localhost:8080",
	}
}

func (p *Provider) Config() schema.AuthProviders {
	return schema.AuthProviders{
		Gerrit: &schema.GerritAuthProvider{
			Type: "gerrit",
			Url:  "http://localhost:8080",
		},
	}
}

func (p *Provider) CachedInfo() *providers.Info {
	return &providers.Info{
		ServiceID:         "",
		ClientID:          "",
		DisplayName:       "Gerrit",
		AuthenticationURL: "http://localhost:8080",
	}
}

func (p *Provider) Refresh(ctx context.Context) error {
	return nil
}
