package codyaccessservice

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"connectrpc.com/connect"
	mockrequire "github.com/derision-test/go-mockgen/v2/testutil/require"
	"github.com/hexops/autogold/v2"
	"github.com/hexops/valast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/sourcegraph/log/logtest"

	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/codyaccess"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/samsm2m"
	codyaccessv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/codyaccess/v1"
)

type testHandlerV1 struct {
	*HandlerV1
	mockStore *MockStoreV1
}

func newTestHandlerV1() *testHandlerV1 {
	mockStore := NewMockStoreV1()
	mockStore.IntrospectSAMSTokenFunc.SetDefaultReturn(
		&sams.IntrospectTokenResponse{
			Active: true,
			Scopes: scopes.Scopes{
				samsm2m.EnterprisePortalScope(
					scopes.PermissionEnterprisePortalCodyAccess, scopes.ActionRead),
				samsm2m.EnterprisePortalScope(
					scopes.PermissionEnterprisePortalCodyAccess, scopes.ActionWrite),
			},
		},
		nil,
	)
	return &testHandlerV1{
		HandlerV1: &HandlerV1{
			logger: logtest.NoOp(nil),
			store:  mockStore,
		},
		mockStore: mockStore,
	}
}

func TestHandlerV1UpdateCodyGatewayAccess(t *testing.T) {
	ctx := context.Background()
	const mockSubscriptionID = "es_80ca12e2-54b4-448c-a61a-390b1a9c1224"

	for _, tc := range []struct {
		name           string
		update         *codyaccessv1.UpdateCodyGatewayAccessRequest
		wantUpdateOpts autogold.Value
	}{
		{
			name: "no update mask",
			update: &codyaccessv1.UpdateCodyGatewayAccessRequest{
				Access: &codyaccessv1.CodyGatewayAccess{
					SubscriptionId: mockSubscriptionID,
					Enabled:        true,
					ChatCompletionsRateLimit: &codyaccessv1.CodyGatewayRateLimit{
						Limit: 100,
					},
					EmbeddingsRateLimit: &codyaccessv1.CodyGatewayRateLimit{
						IntervalDuration: durationpb.New(time.Minute),
					},
				},
			},
			// All non-zero values should be included in update
			wantUpdateOpts: autogold.Expect(codyaccess.UpsertCodyGatewayAccessOptions{
				Enabled: valast.Ptr(true),
				ChatCompletionsRateLimit: &sql.NullInt64{
					Int64: 100,
					Valid: true,
				},
				EmbeddingsRateLimitIntervalSeconds: &sql.NullInt32{
					Int32: 60,
					Valid: true,
				},
			}),
		},
		{
			name: "specified field mask",
			update: &codyaccessv1.UpdateCodyGatewayAccessRequest{
				Access: &codyaccessv1.CodyGatewayAccess{
					SubscriptionId: mockSubscriptionID,
					Enabled:        true,
					ChatCompletionsRateLimit: &codyaccessv1.CodyGatewayRateLimit{
						Limit: 100, // Should be the only thing included in update
					},
				},
				UpdateMask: &fieldmaskpb.FieldMask{
					Paths: []string{"chat_completions_rate_limit.limit"},
				},
			},
			wantUpdateOpts: autogold.Expect(codyaccess.UpsertCodyGatewayAccessOptions{ChatCompletionsRateLimit: &sql.NullInt64{
				Int64: 100,
				Valid: true,
			}}),
		},
		{
			name: "specified field masks to reset limits",
			update: &codyaccessv1.UpdateCodyGatewayAccessRequest{
				Access: &codyaccessv1.CodyGatewayAccess{
					SubscriptionId:           mockSubscriptionID,
					Enabled:                  true,
					ChatCompletionsRateLimit: nil,
					CodeCompletionsRateLimit: nil,
					EmbeddingsRateLimit:      nil,
				},
				UpdateMask: &fieldmaskpb.FieldMask{
					Paths: []string{
						"chat_completions_rate_limit.limit",
						"chat_completions_rate_limit.interval_duration",
						"code_completions_rate_limit.limit",
						"code_completions_rate_limit.interval_duration",
						"embeddings_rate_limit.limit",
						"embeddings_rate_limit.interval_duration",
					},
				},
			},
			wantUpdateOpts: autogold.Expect(codyaccess.UpsertCodyGatewayAccessOptions{
				ChatCompletionsRateLimit:                &sql.NullInt64{},
				ChatCompletionsRateLimitIntervalSeconds: &sql.NullInt32{},
				CodeCompletionsRateLimit:                &sql.NullInt64{},
				CodeCompletionsRateLimitIntervalSeconds: &sql.NullInt32{},
				EmbeddingsRateLimit:                     &sql.NullInt64{},
				EmbeddingsRateLimitIntervalSeconds:      &sql.NullInt32{},
			}),
		},
		{
			name: "* update_mask",
			update: &codyaccessv1.UpdateCodyGatewayAccessRequest{
				Access: &codyaccessv1.CodyGatewayAccess{
					SubscriptionId: mockSubscriptionID,
				},
				// All update-able values are zero
			},

			// All update-able values should be set to their defaults explicitly
			wantUpdateOpts: autogold.Expect(codyaccess.UpsertCodyGatewayAccessOptions{}),
		},
		{
			name: "no-op update",
			update: &codyaccessv1.UpdateCodyGatewayAccessRequest{
				Access: &codyaccessv1.CodyGatewayAccess{
					SubscriptionId: mockSubscriptionID,
				},
				UpdateMask: nil,
			},

			// Update should be empty
			wantUpdateOpts: autogold.Expect(codyaccess.UpsertCodyGatewayAccessOptions{}),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := connect.NewRequest(tc.update)
			req.Header().Add("Authorization", "Bearer foolmeifyoucan")

			h := newTestHandlerV1()
			h.mockStore.UpsertCodyGatewayAccessFunc.SetDefaultHook(func(_ context.Context, subscriptionID string, opts codyaccess.UpsertCodyGatewayAccessOptions) (*codyaccess.CodyGatewayAccessWithSubscriptionDetails, error) {
				tc.wantUpdateOpts.Equal(t, opts)
				assert.Equal(t, mockSubscriptionID, subscriptionID)
				return &codyaccess.CodyGatewayAccessWithSubscriptionDetails{
					CodyGatewayAccess: codyaccess.CodyGatewayAccess{
						SubscriptionID: subscriptionID,
					},
				}, nil
			})

			_, err := h.UpdateCodyGatewayAccess(ctx, req)
			require.NoError(t, err)
			if tc.wantUpdateOpts != nil {
				mockrequire.CalledOnce(t, h.mockStore.UpsertCodyGatewayAccessFunc)
			} else {
				mockrequire.NotCalled(t, h.mockStore.UpsertCodyGatewayAccessFunc)
			}
		})
	}
}
