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

	// If not enabled, return a minimal response.
	if !attrs.CodyGatewayEnabled {
		return &codyaccessv1.CodyGatewayAccess{
			SubscriptionId: subscriptionID,
			Enabled:        false,
		}
	}

	// If enabled, return the full response.
	limits := attrs.EvaluateRateLimits()
	return &codyaccessv1.CodyGatewayAccess{
		SubscriptionId: subscriptionID,
		Enabled:        attrs.CodyGatewayEnabled,
		ChatCompletionsRateLimit: &codyaccessv1.CodyGatewayRateLimit{
			Source:           limits.ChatSource,
			Limit:            limits.Chat.Limit,
			IntervalDuration: durationpb.New(limits.Chat.IntervalDuration()),
		},
		CodeCompletionsRateLimit: &codyaccessv1.CodyGatewayRateLimit{
			Source:           limits.CodeSource,
			Limit:            limits.Code.Limit,
			IntervalDuration: durationpb.New(limits.Code.IntervalDuration()),
		},
		EmbeddingsRateLimit: &codyaccessv1.CodyGatewayRateLimit{
			Source:           limits.EmbeddingsSource,
			Limit:            limits.Embeddings.Limit,
			IntervalDuration: durationpb.New(limits.Embeddings.IntervalDuration()),
		},
		AccessTokens: func() []*codyaccessv1.CodyGatewayAccessToken {
			accessTokens := attrs.GenerateAccessTokens()
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
