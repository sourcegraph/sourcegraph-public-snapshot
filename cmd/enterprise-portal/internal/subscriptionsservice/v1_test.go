package subscriptionsservice

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	mockrequire "github.com/derision-test/go-mockgen/v2/testutil/require"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/subscriptions"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/dotcomdb"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/samsm2m"
	subscriptionsv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptions/v1"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/iam"
)

type testHandlerV1 struct {
	*handlerV1
	mockStore *MockStoreV1
}

func newTestHandlerV1() *testHandlerV1 {
	mockStore := NewMockStoreV1()
	mockStore.IntrospectSAMSTokenFunc.SetDefaultReturn(
		&sams.IntrospectTokenResponse{
			Active: true,
			Scopes: scopes.Scopes{
				samsm2m.EnterprisePortalScope("subscription", scopes.ActionRead),
				samsm2m.EnterprisePortalScope("subscription", scopes.ActionWrite),
				samsm2m.EnterprisePortalScope("permission.subscription", scopes.ActionWrite),
			},
		},
		nil,
	)
	return &testHandlerV1{
		handlerV1: &handlerV1{
			logger: logtest.NoOp(nil),
			store:  mockStore,
		},
		mockStore: mockStore,
	}
}

func TestHandlerV1_ListEnterpriseSubscriptions(t *testing.T) {
	ctx := context.Background()
	t.Run("more than one permission filter", func(t *testing.T) {
		req := connect.NewRequest(&subscriptionsv1.ListEnterpriseSubscriptionsRequest{
			Filters: []*subscriptionsv1.ListEnterpriseSubscriptionsFilter{
				{Filter: &subscriptionsv1.ListEnterpriseSubscriptionsFilter_Permission{
					Permission: &subscriptionsv1.Permission{
						Type:          subscriptionsv1.PermissionType_PERMISSION_TYPE_SUBSCRIPTION_CODY_ANALYTICS,
						Relation:      subscriptionsv1.PermissionRelation_PERMISSION_RELATION_VIEW,
						SamsAccountId: "018d21f2-04a6-7aaf-9f6f-6fc58c4187b9",
					},
				}},
				{Filter: &subscriptionsv1.ListEnterpriseSubscriptionsFilter_Permission{
					Permission: &subscriptionsv1.Permission{
						Type:          subscriptionsv1.PermissionType_PERMISSION_TYPE_SUBSCRIPTION_CODY_ANALYTICS,
						Relation:      subscriptionsv1.PermissionRelation_PERMISSION_RELATION_VIEW,
						SamsAccountId: "018d21f2-0b8d-756a-84f0-13b942a2bae5",
					},
				}},
			},
		})
		req.Header().Add("Authorization", "Bearer foolmeifyoucan")

		h := newTestHandlerV1()
		_, err := h.ListEnterpriseSubscriptions(ctx, req)
		require.EqualError(t, err, `invalid_argument: invalid filter: "permission" can only be provided once`)
	})

	t.Run("only subscription ID filter", func(t *testing.T) {
		req := connect.NewRequest(&subscriptionsv1.ListEnterpriseSubscriptionsRequest{
			Filters: []*subscriptionsv1.ListEnterpriseSubscriptionsFilter{
				{Filter: &subscriptionsv1.ListEnterpriseSubscriptionsFilter_SubscriptionId{
					SubscriptionId: "80ca12e2-54b4-448c-a61a-390b1a9c1224",
				}},
			},
		})
		req.Header().Add("Authorization", "Bearer foolmeifyoucan")

		h := newTestHandlerV1()
		h.mockStore.ListEnterpriseSubscriptionsFunc.SetDefaultHook(func(_ context.Context, opts subscriptions.ListEnterpriseSubscriptionsOptions) ([]*subscriptions.Subscription, error) {
			require.Len(t, opts.IDs, 1)
			assert.Equal(t, "80ca12e2-54b4-448c-a61a-390b1a9c1224", opts.IDs[0])
			return []*subscriptions.Subscription{}, nil
		})
		_, err := h.ListEnterpriseSubscriptions(ctx, req)
		require.NoError(t, err)
		mockrequire.Called(t, h.mockStore.ListEnterpriseSubscriptionsFunc)
	})

	t.Run("only permission filter", func(t *testing.T) {
		req := connect.NewRequest(&subscriptionsv1.ListEnterpriseSubscriptionsRequest{
			Filters: []*subscriptionsv1.ListEnterpriseSubscriptionsFilter{
				{Filter: &subscriptionsv1.ListEnterpriseSubscriptionsFilter_Permission{
					Permission: &subscriptionsv1.Permission{
						Type:          subscriptionsv1.PermissionType_PERMISSION_TYPE_SUBSCRIPTION_CODY_ANALYTICS,
						Relation:      subscriptionsv1.PermissionRelation_PERMISSION_RELATION_VIEW,
						SamsAccountId: "018d21f2-04a6-7aaf-9f6f-6fc58c4187b9",
					},
				}},
			},
		})
		req.Header().Add("Authorization", "Bearer foolmeifyoucan")

		h := newTestHandlerV1()
		h.mockStore.IAMListObjectsFunc.SetDefaultReturn([]string{"subscription_cody_analytics:80ca12e2-54b4-448c-a61a-390b1a9c1224"}, nil)
		h.mockStore.ListEnterpriseSubscriptionsFunc.SetDefaultHook(func(_ context.Context, opts subscriptions.ListEnterpriseSubscriptionsOptions) ([]*subscriptions.Subscription, error) {
			require.Len(t, opts.IDs, 1)
			assert.Equal(t, "80ca12e2-54b4-448c-a61a-390b1a9c1224", opts.IDs[0])
			return []*subscriptions.Subscription{}, nil
		})
		_, err := h.ListEnterpriseSubscriptions(ctx, req)
		require.NoError(t, err)
		mockrequire.Called(t, h.mockStore.ListEnterpriseSubscriptionsFunc)
	})

	t.Run("both subscription ID and permission filter that DO NOT intersect", func(t *testing.T) {
		req := connect.NewRequest(&subscriptionsv1.ListEnterpriseSubscriptionsRequest{
			Filters: []*subscriptionsv1.ListEnterpriseSubscriptionsFilter{
				{Filter: &subscriptionsv1.ListEnterpriseSubscriptionsFilter_Permission{
					Permission: &subscriptionsv1.Permission{
						Type:          subscriptionsv1.PermissionType_PERMISSION_TYPE_SUBSCRIPTION_CODY_ANALYTICS,
						Relation:      subscriptionsv1.PermissionRelation_PERMISSION_RELATION_VIEW,
						SamsAccountId: "018d21f2-04a6-7aaf-9f6f-6fc58c4187b9",
					},
				}},
				{Filter: &subscriptionsv1.ListEnterpriseSubscriptionsFilter_SubscriptionId{
					SubscriptionId: "80ca12e2-54b4-448c-a61a-390b1a9c1224",
				}},
			},
		})
		req.Header().Add("Authorization", "Bearer foolmeifyoucan")

		h := newTestHandlerV1()
		h.mockStore.IAMListObjectsFunc.SetDefaultReturn([]string{"subscription_cody_analytics:a-different-subscription-id"}, nil)
		h.mockStore.ListEnterpriseSubscriptionsFunc.SetDefaultHook(func(_ context.Context, opts subscriptions.ListEnterpriseSubscriptionsOptions) ([]*subscriptions.Subscription, error) {
			require.Len(t, opts.IDs, 1)
			assert.Equal(t, "a-different-subscription-id", opts.IDs[0])
			return []*subscriptions.Subscription{}, nil
		})
		resp, err := h.ListEnterpriseSubscriptions(ctx, req)
		require.NoError(t, err)
		assert.Empty(t, resp.Msg.Subscriptions)
	})

	t.Run("both subscription ID and permission filter", func(t *testing.T) {
		subscriptionID := "80ca12e2-54b4-448c-a61a-390b1a9c1224"
		req := connect.NewRequest(&subscriptionsv1.ListEnterpriseSubscriptionsRequest{
			Filters: []*subscriptionsv1.ListEnterpriseSubscriptionsFilter{
				{Filter: &subscriptionsv1.ListEnterpriseSubscriptionsFilter_Permission{
					Permission: &subscriptionsv1.Permission{
						Type:          subscriptionsv1.PermissionType_PERMISSION_TYPE_SUBSCRIPTION_CODY_ANALYTICS,
						Relation:      subscriptionsv1.PermissionRelation_PERMISSION_RELATION_VIEW,
						SamsAccountId: "018d21f2-04a6-7aaf-9f6f-6fc58c4187b9",
					},
				}},
				{Filter: &subscriptionsv1.ListEnterpriseSubscriptionsFilter_SubscriptionId{
					SubscriptionId: subscriptionID,
				}},
			},
		})
		req.Header().Add("Authorization", "Bearer foolmeifyoucan")

		h := newTestHandlerV1()
		h.mockStore.IAMListObjectsFunc.SetDefaultReturn([]string{"subscription_cody_analytics:" + subscriptionID}, nil)
		h.mockStore.ListEnterpriseSubscriptionsFunc.SetDefaultHook(func(_ context.Context, opts subscriptions.ListEnterpriseSubscriptionsOptions) ([]*subscriptions.Subscription, error) {
			require.Len(t, opts.IDs, 1)
			assert.Equal(t, subscriptionID, opts.IDs[0])
			return []*subscriptions.Subscription{{ID: opts.IDs[0]}}, nil
		})
		resp, err := h.ListEnterpriseSubscriptions(ctx, req)
		require.NoError(t, err)
		mockrequire.Called(t, h.mockStore.ListEnterpriseSubscriptionsFunc)
		require.NotEmpty(t, resp.Msg.Subscriptions)
		assert.Len(t, resp.Msg.Subscriptions, 1)
		assert.Equal(t, subscriptionsv1.EnterpriseSubscriptionIDPrefix+subscriptionID, resp.Msg.Subscriptions[0].Id)
	})

	t.Run("permission filter with no results", func(t *testing.T) {
		req := connect.NewRequest(&subscriptionsv1.ListEnterpriseSubscriptionsRequest{
			Filters: []*subscriptionsv1.ListEnterpriseSubscriptionsFilter{
				{Filter: &subscriptionsv1.ListEnterpriseSubscriptionsFilter_Permission{
					Permission: &subscriptionsv1.Permission{
						Type:          subscriptionsv1.PermissionType_PERMISSION_TYPE_SUBSCRIPTION_CODY_ANALYTICS,
						Relation:      subscriptionsv1.PermissionRelation_PERMISSION_RELATION_VIEW,
						SamsAccountId: "018d21f2-04a6-7aaf-9f6f-6fc58c4187b9",
					},
				}},
				{Filter: &subscriptionsv1.ListEnterpriseSubscriptionsFilter_SubscriptionId{
					SubscriptionId: "80ca12e2-54b4-448c-a61a-390b1a9c1224",
				}},
			},
		})
		req.Header().Add("Authorization", "Bearer foolmeifyoucan")

		h := newTestHandlerV1()
		h.mockStore.IAMListObjectsFunc.SetDefaultReturn([]string{}, nil)
		resp, err := h.ListEnterpriseSubscriptions(ctx, req)
		require.NoError(t, err)
		assert.Nil(t, resp.Msg.Subscriptions)
	})

	t.Run("no filters", func(t *testing.T) {
		req := connect.NewRequest(&subscriptionsv1.ListEnterpriseSubscriptionsRequest{})
		req.Header().Add("Authorization", "Bearer foolmeifyoucan")

		h := newTestHandlerV1()
		h.mockStore.ListEnterpriseSubscriptionsFunc.SetDefaultHook(func(_ context.Context, opts subscriptions.ListEnterpriseSubscriptionsOptions) ([]*subscriptions.Subscription, error) {
			return []*subscriptions.Subscription{}, nil
		})
		h.mockStore.ListDotcomEnterpriseSubscriptionsFunc.SetDefaultHook(func(_ context.Context, opts dotcomdb.ListEnterpriseSubscriptionsOptions) ([]*dotcomdb.SubscriptionAttributes, error) {
			assert.Empty(t, opts.SubscriptionIDs)
			assert.False(t, opts.IsArchived)
			return []*dotcomdb.SubscriptionAttributes{{
				ID: "80ca12e2-54b4-448c-a61a-390b1a9c1224",
			}}, nil
		})
		_, err := h.ListEnterpriseSubscriptions(ctx, req)
		require.NoError(t, err)
		mockrequire.Called(t, h.mockStore.ListEnterpriseSubscriptionsFunc)
		mockrequire.Called(t, h.mockStore.ListDotcomEnterpriseSubscriptionsFunc)
		resp, err := h.ListEnterpriseSubscriptions(ctx, req)
		require.NoError(t, err)
		assert.NotNil(t, resp.Msg.Subscriptions)
	})
}

func TestHandlerV1_UpdateEnterpriseSubscription(t *testing.T) {
	ctx := context.Background()
	t.Run("no update_mask", func(t *testing.T) {
		req := connect.NewRequest(&subscriptionsv1.UpdateEnterpriseSubscriptionRequest{
			Subscription: &subscriptionsv1.EnterpriseSubscription{
				Id:             "80ca12e2-54b4-448c-a61a-390b1a9c1224",
				InstanceDomain: "s1.sourcegraph.com",
			},
			UpdateMask: nil,
		})
		req.Header().Add("Authorization", "Bearer foolmeifyoucan")

		h := newTestHandlerV1()
		h.mockStore.ListDotcomEnterpriseSubscriptionsFunc.SetDefaultReturn([]*dotcomdb.SubscriptionAttributes{{ID: "80ca12e2-54b4-448c-a61a-390b1a9c1224"}}, nil)
		h.mockStore.UpsertEnterpriseSubscriptionFunc.SetDefaultHook(func(_ context.Context, _ string, opts subscriptions.UpsertSubscriptionOptions) (*subscriptions.Subscription, error) {
			assert.NotEmpty(t, opts.InstanceDomain)
			assert.False(t, opts.ForceUpdate)
			return &subscriptions.Subscription{}, nil
		})
		_, err := h.UpdateEnterpriseSubscription(ctx, req)
		require.NoError(t, err)
		mockrequire.Called(t, h.mockStore.UpsertEnterpriseSubscriptionFunc)
	})

	t.Run("specified update_mask", func(t *testing.T) {
		req := connect.NewRequest(&subscriptionsv1.UpdateEnterpriseSubscriptionRequest{
			Subscription: &subscriptionsv1.EnterpriseSubscription{
				Id:             "80ca12e2-54b4-448c-a61a-390b1a9c1224",
				InstanceDomain: "s1.sourcegraph.com",
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"instance_domain"}},
		})
		req.Header().Add("Authorization", "Bearer foolmeifyoucan")

		h := newTestHandlerV1()
		h.mockStore.ListDotcomEnterpriseSubscriptionsFunc.SetDefaultReturn([]*dotcomdb.SubscriptionAttributes{{ID: "80ca12e2-54b4-448c-a61a-390b1a9c1224"}}, nil)
		h.mockStore.UpsertEnterpriseSubscriptionFunc.SetDefaultHook(func(_ context.Context, _ string, opts subscriptions.UpsertSubscriptionOptions) (*subscriptions.Subscription, error) {
			assert.NotEmpty(t, opts.InstanceDomain)
			assert.False(t, opts.ForceUpdate)
			return &subscriptions.Subscription{}, nil
		})
		_, err := h.UpdateEnterpriseSubscription(ctx, req)
		require.NoError(t, err)
		mockrequire.Called(t, h.mockStore.UpsertEnterpriseSubscriptionFunc)
	})

	t.Run("* update_mask", func(t *testing.T) {
		req := connect.NewRequest(&subscriptionsv1.UpdateEnterpriseSubscriptionRequest{
			Subscription: &subscriptionsv1.EnterpriseSubscription{
				Id:             "80ca12e2-54b4-448c-a61a-390b1a9c1224",
				InstanceDomain: "",
			},
			UpdateMask: &fieldmaskpb.FieldMask{Paths: []string{"*"}},
		})
		req.Header().Add("Authorization", "Bearer foolmeifyoucan")

		h := newTestHandlerV1()
		h.mockStore.ListDotcomEnterpriseSubscriptionsFunc.SetDefaultReturn([]*dotcomdb.SubscriptionAttributes{{ID: "80ca12e2-54b4-448c-a61a-390b1a9c1224"}}, nil)
		h.mockStore.UpsertEnterpriseSubscriptionFunc.SetDefaultHook(func(_ context.Context, _ string, opts subscriptions.UpsertSubscriptionOptions) (*subscriptions.Subscription, error) {
			assert.Empty(t, opts.InstanceDomain)
			assert.True(t, opts.ForceUpdate)
			return &subscriptions.Subscription{}, nil
		})
		_, err := h.UpdateEnterpriseSubscription(ctx, req)
		require.NoError(t, err)
		mockrequire.Called(t, h.mockStore.UpsertEnterpriseSubscriptionFunc)
	})

	t.Run("noop update", func(t *testing.T) {
		req := connect.NewRequest(&subscriptionsv1.UpdateEnterpriseSubscriptionRequest{
			Subscription: &subscriptionsv1.EnterpriseSubscription{
				Id:             "80ca12e2-54b4-448c-a61a-390b1a9c1224",
				InstanceDomain: "",
			},
		})
		req.Header().Add("Authorization", "Bearer foolmeifyoucan")

		h := newTestHandlerV1()
		h.mockStore.ListDotcomEnterpriseSubscriptionsFunc.SetDefaultReturn([]*dotcomdb.SubscriptionAttributes{{ID: "80ca12e2-54b4-448c-a61a-390b1a9c1224"}}, nil)
		h.mockStore.UpsertEnterpriseSubscriptionFunc.SetDefaultHook(func(_ context.Context, _ string, opts subscriptions.UpsertSubscriptionOptions) (*subscriptions.Subscription, error) {
			assert.Empty(t, opts.InstanceDomain)
			assert.False(t, opts.ForceUpdate)
			return &subscriptions.Subscription{}, nil
		})
		_, err := h.UpdateEnterpriseSubscription(ctx, req)
		require.NoError(t, err)
		mockrequire.Called(t, h.mockStore.UpsertEnterpriseSubscriptionFunc)
	})
}

func TestHandlerV1_UpdateEnterpriseSubscriptionMembership(t *testing.T) {
	ctx := context.Background()
	t.Run("via subscription ID", func(t *testing.T) {
		req := connect.NewRequest(&subscriptionsv1.UpdateEnterpriseSubscriptionMembershipRequest{
			Membership: &subscriptionsv1.EnterpriseSubscriptionMembership{
				SubscriptionId:      "80ca12e2-54b4-448c-a61a-390b1a9c1224",
				MemberSamsAccountId: "018d21f2-04a6-7aaf-9f6f-6fc58c4187b9",
				MemberRoles:         []subscriptionsv1.Role{subscriptionsv1.Role_ROLE_SUBSCRIPTION_CODY_ANALYTICS_CUSTOMER_ADMIN},
			},
		})
		req.Header().Add("Authorization", "Bearer foolmeifyoucan")

		h := newTestHandlerV1()
		h.mockStore.ListDotcomEnterpriseSubscriptionsFunc.SetDefaultReturn([]*dotcomdb.SubscriptionAttributes{{ID: "80ca12e2-54b4-448c-a61a-390b1a9c1224"}}, nil)
		h.mockStore.IAMWriteFunc.SetDefaultHook(func(_ context.Context, opts iam.WriteOptions) error {
			assert.Len(t, opts.Writes, 2)
			return nil
		})
		_, err := h.UpdateEnterpriseSubscriptionMembership(ctx, req)
		require.NoError(t, err)
		mockrequire.Called(t, h.mockStore.IAMWriteFunc)
	})

	t.Run("via instance domain", func(t *testing.T) {
		req := connect.NewRequest(&subscriptionsv1.UpdateEnterpriseSubscriptionMembershipRequest{
			Membership: &subscriptionsv1.EnterpriseSubscriptionMembership{
				InstanceDomain:      "s1.sourcegraph.com",
				MemberSamsAccountId: "018d21f2-04a6-7aaf-9f6f-6fc58c4187b9",
				MemberRoles:         []subscriptionsv1.Role{subscriptionsv1.Role_ROLE_SUBSCRIPTION_CODY_ANALYTICS_CUSTOMER_ADMIN},
			},
		})
		req.Header().Add("Authorization", "Bearer foolmeifyoucan")

		h := newTestHandlerV1()
		h.mockStore.ListEnterpriseSubscriptionsFunc.SetDefaultReturn([]*subscriptions.Subscription{{ID: "80ca12e2-54b4-448c-a61a-390b1a9c1224"}}, nil)
		h.mockStore.IAMWriteFunc.SetDefaultHook(func(_ context.Context, opts iam.WriteOptions) error {
			assert.Len(t, opts.Writes, 2)
			return nil
		})
		_, err := h.UpdateEnterpriseSubscriptionMembership(ctx, req)
		require.NoError(t, err)
		mockrequire.Called(t, h.mockStore.IAMWriteFunc)
	})
}
