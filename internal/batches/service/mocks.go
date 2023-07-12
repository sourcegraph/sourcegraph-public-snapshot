package service

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
)

type ServiceMocks struct {
	ValidateAuthenticator func(ctx context.Context, externalServiceID, externalServiceType string, a auth.Authenticator) error
}

func (sm *ServiceMocks) Reset() {
	sm.ValidateAuthenticator = nil
}

var Mocks = ServiceMocks{}
