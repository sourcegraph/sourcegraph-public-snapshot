package auth

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc"
)

type (
	Provider         = auth.Provider
	ProviderConfigID = auth.ProviderConfigID
	Middleware       = auth.Middleware
	ProviderInfo     = auth.ProviderInfo
)

var (
	GetProviderByConfigID  = auth.GetProviderByConfigID
	Providers              = auth.Providers
	UpdateProviders        = auth.UpdateProviders
	SafeRedirectURL        = auth.SafeRedirectURL
	SetExternalAccountData = auth.SetExternalAccountData
	NormalizeUsername      = auth.NormalizeUsername
	CreateOrUpdateUser     = auth.CreateOrUpdateUser
	RegisterMiddlewares    = auth.RegisterMiddlewares
)

const (
	AuthURLPrefix = auth.AuthURLPrefix
)

func SetMockProviders(mockProviders []auth.Provider) {
	auth.MockProviders = mockProviders
}

func SetMockCreateOrUpdateUser(f func(db.NewUser, extsvc.ExternalAccountSpec) (int32, error)) {
	auth.MockCreateOrUpdateUser = f
}
