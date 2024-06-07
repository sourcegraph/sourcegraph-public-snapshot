package codyaccessservice

import (
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/dotcomdb"
	codyaccessv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/codyaccess/v1"
	subscriptionsv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptions/v1"
)

func convertAccessAttrsToProto(attrs *dotcomdb.CodyGatewayAccessAttributes) *codyaccessv1.CodyGatewayAccess {
	// Provide ID in prefixed format.
	subscriptionID := subscriptionsv1.EnterpriseSubscriptionIDPrefix + attrs.SubscriptionID

	// Always try to return the full response, since even when disabled, some
	// features may be allowed via Cody Gateway (notably attributions). This
	// also allows Cody Gateway to cache the state of actors that are disabled.
	limits := attrs.EvaluateRateLimits()
	return &codyaccessv1.CodyGatewayAccess{
		SubscriptionId:          subscriptionID,
		SubscriptionDisplayName: attrs.GetSubscriptionDisplayName(),
		Enabled:                 attrs.CodyGatewayEnabled,
		// Rate limits return nil if not enabled, per API spec
		ChatCompletionsRateLimit: nilIfNotEnabled(attrs.CodyGatewayEnabled, &codyaccessv1.CodyGatewayRateLimit{
			Source:           limits.ChatSource,
			Limit:            limits.Chat.Limit,
			IntervalDuration: durationpb.New(limits.Chat.IntervalDuration()),
		}),
		CodeCompletionsRateLimit: nilIfNotEnabled(attrs.CodyGatewayEnabled, &codyaccessv1.CodyGatewayRateLimit{
			Source:           limits.CodeSource,
			Limit:            limits.Code.Limit,
			IntervalDuration: durationpb.New(limits.Code.IntervalDuration()),
		}),
		EmbeddingsRateLimit: nilIfNotEnabled(attrs.CodyGatewayEnabled, &codyaccessv1.CodyGatewayRateLimit{
			Source:           limits.EmbeddingsSource,
			Limit:            limits.Embeddings.Limit,
			IntervalDuration: durationpb.New(limits.Embeddings.IntervalDuration()),
		}),
		// This is always provided, even if access is disabled
		AccessTokens: func() []*codyaccessv1.CodyGatewayAccessToken {
			accessTokens := attrs.GenerateAccessTokens()
			if len(accessTokens) == 0 {
				return []*codyaccessv1.CodyGatewayAccessToken{}
			}

			results := make([]*codyaccessv1.CodyGatewayAccessToken, len(accessTokens))
			for i, token := range accessTokens {
				results[i] = &codyaccessv1.CodyGatewayAccessToken{
					Token: token,
				}
			}
			return results
		}(),
	}
}

func nilIfNotEnabled[T any](enabled bool, value *T) *T {
	if !enabled {
		return nil
	}
	return value
}
