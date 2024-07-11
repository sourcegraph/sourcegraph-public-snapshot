package codyaccessservice

import (
	"context"

	"github.com/sourcegraph/conc/pool"

	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
	"github.com/sourcegraph/sourcegraph/internal/codygateway/codygatewayactor"
	"github.com/sourcegraph/sourcegraph/internal/codygateway/codygatewayevents"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	codyaccessv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/codyaccess/v1"
	subscriptionsv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptions/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
	// GetCodyGatewayUsage retrieves recent Cody Gateway usage data.
	// The subscriptionID should not be prefixed.
	GetCodyGatewayUsage(ctx context.Context, subscriptionID string) (*codyaccessv1.CodyGatewayUsage, error)
}

type storeV1 struct {
	SAMSClient        *sams.ClientV1
	CodyGatewayEvents *codygatewayevents.Service
}

type StoreV1Options struct {
	SAMSClient *sams.ClientV1
	// Optional.
	CodyGatewayEvents *codygatewayevents.Service
}

var errStoreUnimplemented = errors.New("unimplemented")

// NewStoreV1 returns a new StoreV1 using the given resource handles.
func NewStoreV1(opts StoreV1Options) StoreV1 {
	return &storeV1{
		SAMSClient:        opts.SAMSClient,
		CodyGatewayEvents: opts.CodyGatewayEvents,
	}
}

func (s *storeV1) IntrospectSAMSToken(ctx context.Context, token string) (*sams.IntrospectTokenResponse, error) {
	return s.SAMSClient.Tokens().IntrospectToken(ctx, token)
}

func (s *storeV1) GetCodyGatewayUsage(ctx context.Context, subscriptionID string) (*codyaccessv1.CodyGatewayUsage, error) {
	if s.CodyGatewayEvents == nil {
		return nil, errStoreUnimplemented
	}

	// Collect results concurrently, cancelling on the first error result.
	p := pool.NewWithResults[func(*codyaccessv1.CodyGatewayUsage)]().
		WithContext(ctx).
		WithCancelOnError()

	// Get code completions usage.
	p.Go(func(ctx context.Context) (func(*codyaccessv1.CodyGatewayUsage), error) {
		codeCompletions, err := s.CodyGatewayEvents.CompletionsUsageForActor(
			ctx,
			types.CompletionsFeatureCode,
			codygatewayactor.ActorSourceEnterpriseSubscription,
			subscriptionID,
		)
		if err != nil {
			return nil, errors.Wrap(err, "get code completions usage")
		}
		return func(u *codyaccessv1.CodyGatewayUsage) {
			u.CodeCompletionsUsage = convertCodyGatewayUsageDatapoints(codeCompletions)
		}, nil
	})

	// Get chat completions usage.
	p.Go(func(ctx context.Context) (func(*codyaccessv1.CodyGatewayUsage), error) {
		chatCompletions, err := s.CodyGatewayEvents.CompletionsUsageForActor(
			ctx,
			types.CompletionsFeatureChat,
			codygatewayactor.ActorSourceEnterpriseSubscription,
			subscriptionID,
		)
		if err != nil {
			return nil, errors.Wrap(err, "get chat completions usage")
		}
		return func(u *codyaccessv1.CodyGatewayUsage) {
			u.ChatCompletionsUsage = convertCodyGatewayUsageDatapoints(chatCompletions)
		}, nil
	})

	// Get embeddings usage
	p.Go(func(ctx context.Context) (func(*codyaccessv1.CodyGatewayUsage), error) {
		embeddings, err := s.CodyGatewayEvents.EmbeddingsUsageForActor(
			ctx,
			codygatewayactor.ActorSourceEnterpriseSubscription,
			subscriptionID,
		)
		if err != nil {
			return nil, errors.Wrap(err, "get embeddings usage")
		}
		return func(u *codyaccessv1.CodyGatewayUsage) {
			u.EmbeddingsUsage = convertCodyGatewayUsageDatapoints(embeddings)
		}, nil
	})

	// Collect all results
	results, err := p.Wait()
	if err != nil {
		return nil, err
	}

	// Apply all results synchronously
	usage := &codyaccessv1.CodyGatewayUsage{
		// Use public format ID
		SubscriptionId: subscriptionsv1.EnterpriseSubscriptionIDPrefix + subscriptionID,
	}
	for _, applyResult := range results {
		applyResult(usage)
	}
	return usage, nil
}
