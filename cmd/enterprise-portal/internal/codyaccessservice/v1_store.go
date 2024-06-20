package codyaccessservice

import (
	"context"

	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
)

// StoreV1 is the data layer carrier for Cody access service v1. This interface
// is meant to abstract away and limit the exposure of the underlying data layer
// to the handler through a thin-wrapper.
type StoreV1 interface {
	// IntrospectSAMSToken takes a SAMS access token and returns relevant metadata.
	//
	// 🚨SECURITY: SAMS will return a successful result if the token is valid, but
	// is no longer active. It is critical that the caller not honor tokens where
	// `.Active == false`.
	IntrospectSAMSToken(ctx context.Context, token string) (*sams.IntrospectTokenResponse, error)
}

type storeV1 struct {
	SAMSClient *sams.ClientV1
}

type StoreV1Options struct {
	SAMSClient *sams.ClientV1
}

// NewStoreV1 returns a new StoreV1 using the given resource handles.
func NewStoreV1(opts StoreV1Options) StoreV1 {
	return &storeV1{
		SAMSClient: opts.SAMSClient,
	}
}

func (s *storeV1) IntrospectSAMSToken(ctx context.Context, token string) (*sams.IntrospectTokenResponse, error) {
	return s.SAMSClient.Tokens().IntrospectToken(ctx, token)
}
