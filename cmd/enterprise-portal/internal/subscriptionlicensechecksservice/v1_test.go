package subscriptionlicensechecksservice

import (
	"bytes"
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	mockrequire "github.com/derision-test/go-mockgen/v2/testutil/require"
	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	subscriptions "github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/subscriptions"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/utctime"
	"github.com/sourcegraph/sourcegraph/internal/hashutil"
	"github.com/sourcegraph/sourcegraph/internal/license"
	slack "github.com/sourcegraph/sourcegraph/internal/slack"
	subscriptionlicensechecksv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptionlicensechecks/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestNewMultipleInstancesUsageNotification(t *testing.T) {
	autogold.ExpectFile(t, autogold.Raw(
		newMultipleInstancesUsageNotification(multipleInstancesUsageNotificationOpts{
			subscriptionID:          "subscription-id",
			subscriptionDisplayName: "subscription-display-name",
			licenseID:               "license-id",
			instanceIDs: []string{
				"instance-id-1",
				"instance-id-2",
			},
		}).Text,
	))
}

func TestCheckLicenseKey(t *testing.T) {
	mockTime := time.Date(2024, 7, 8, 16, 39, 16, 4277000, time.Local)

	const (
		knownInstanceID                   = "known-instance-id"
		licenseKeyWithDetectedInstance    = "key-with-detected-instance"
		licenseKeyWithoutDetectedInstance = "key-without-detected-instance"
		licenseKeyExpired                 = "key-expired"
		licenseKeyRevoked                 = "key-revoked"
	)

	store := NewMockStoreV1()
	store.NowFunc.SetDefaultReturn(mockTime)
	store.GetByLicenseKeyFunc.SetDefaultHook(func(ctx context.Context, id string) (*subscriptions.SubscriptionLicense, error) {
		switch id {
		case licenseKeyWithDetectedInstance:
			return &subscriptions.SubscriptionLicense{
				ID:                 "license-id-with-detected-instance",
				SubscriptionID:     "subscription-id",
				DetectedInstanceID: pointers.Ptr(knownInstanceID),
				ExpireAt:           utctime.FromTime(mockTime.Add(time.Hour)),
			}, nil
		case licenseKeyWithoutDetectedInstance:
			return &subscriptions.SubscriptionLicense{
				ID:                 "license-id-without-detected-instance",
				SubscriptionID:     "subscription-id",
				DetectedInstanceID: nil,
				ExpireAt:           utctime.FromTime(mockTime.Add(time.Hour)),
			}, nil
		case licenseKeyExpired:
			return &subscriptions.SubscriptionLicense{
				ID:       "license-id-expired",
				ExpireAt: utctime.FromTime(mockTime.Add(-time.Hour)),
			}, nil
		case licenseKeyRevoked:
			return &subscriptions.SubscriptionLicense{
				ID:        "license-id-revoked",
				RevokedAt: pointers.Ptr(utctime.FromTime(mockTime.Add(-time.Hour))),
				ExpireAt:  utctime.FromTime(mockTime.Add(time.Hour)),
			}, nil
		default:
			return nil, errors.New("license not found")
		}
	})
	store.GetByLicenseKeyHashFunc.SetDefaultHook(func(ctx context.Context, hash string) (*subscriptions.SubscriptionLicense, error) {
		// Only test for once case, just to check the hash extraction path
		if bytes.Equal(hashutil.ToSHA256Bytes([]byte(licenseKeyExpired)), []byte(hash)) {
			return store.GetByLicenseKey(ctx, licenseKeyExpired)
		}
		return nil, errors.New("license not found")
	})
	store.GetSubscriptionFunc.SetDefaultHook(func(ctx context.Context, s string) (*subscriptions.Subscription, error) {
		if s == "subscription-id" {
			return &subscriptions.Subscription{
				ID:          "subscription-id",
				DisplayName: pointers.Ptr("subscription-display-name"),
			}, nil
		}
		return nil, errors.New("subscription not found")
	})

	for _, tc := range []struct {
		name   string
		req    *subscriptionlicensechecksv1.CheckLicenseKeyRequest
		bypass bool

		wantResult autogold.Value
		wantErr    autogold.Value

		wantSetDetectedInstance autogold.Value
		wantPostToSlack         autogold.Value
	}{{
		name: "bypass enabled",
		req: &subscriptionlicensechecksv1.CheckLicenseKeyRequest{
			InstanceId: "instance-id",
			LicenseKey: "license-key",
		},
		bypass:     true,
		wantResult: autogold.Expect(map[string]interface{}{"reason": "", "valid": true}),
	}, {
		name: "instance_id required",
		req: &subscriptionlicensechecksv1.CheckLicenseKeyRequest{
			InstanceId: "",
			LicenseKey: "license-key",
		},
		wantErr: autogold.Expect("invalid_argument: instance_id is required"),
	}, {
		name: "license_key required",
		req: &subscriptionlicensechecksv1.CheckLicenseKeyRequest{
			InstanceId: "instance-id",
			LicenseKey: "",
		},
		wantErr: autogold.Expect("invalid_argument: license_key is required"),
	}, {
		name: "legacy hash support",
		req: &subscriptionlicensechecksv1.CheckLicenseKeyRequest{
			InstanceId: "instance-id",
			LicenseKey: license.GenerateLicenseKeyBasedAccessToken(licenseKeyExpired),
		},
		// expect expired license
		wantResult: autogold.Expect(map[string]interface{}{"reason": "license has expired", "valid": false}),
	}, {
		name: "expired license",
		req: &subscriptionlicensechecksv1.CheckLicenseKeyRequest{
			InstanceId: "instance-id",
			LicenseKey: licenseKeyExpired,
		},
		wantResult: autogold.Expect(map[string]interface{}{"reason": "license has expired", "valid": false}),
	}, {
		name: "revoked license",
		req: &subscriptionlicensechecksv1.CheckLicenseKeyRequest{
			InstanceId: "instance-id",
			LicenseKey: licenseKeyRevoked,
		},
		wantResult: autogold.Expect(map[string]interface{}{"reason": "license has been revoked", "valid": false}),
	}, {
		name: "license already in use by another instance",
		req: &subscriptionlicensechecksv1.CheckLicenseKeyRequest{
			InstanceId: "unknown-instance-id",
			LicenseKey: licenseKeyWithDetectedInstance,
		},
		wantResult: autogold.Expect(map[string]interface{}{
			"reason": "license has already been used by another instance",
			"valid":  false,
		}),
		wantPostToSlack: autogold.Expect(&slack.Payload{Text: "Subscription \"subscription-display-name\"'s license <https://sourcegraph.com/site-admin/dotcom/product/subscriptions/subscription-id#license-id-with-detected-instance|license-id-with-detected-instance> failed a license check, as it seems to be used by multiple Sourcegraph instance IDs:\n\n- `known-instance-id`\n- `unknown-instance-id`\n\nThis could mean that the license key is attempting to be used on multiple Sourcegraph instances.\n\nTo fix it, <https://docs.google.com/document/d/1xzlkJd3HXGLzB67N7o-9T1s1YXhc1LeGDdJyKDyqfbI/edit#heading=h.mr6npkexi05j|follow the guide to update the siteID and license key for all customer instances>."}),
	}, {
		name: "instance usage assigned",
		req: &subscriptionlicensechecksv1.CheckLicenseKeyRequest{
			InstanceId: "new-instance-id",
			LicenseKey: licenseKeyWithoutDetectedInstance,
		},
		wantResult:              autogold.Expect(map[string]interface{}{"reason": "", "valid": true}),
		wantSetDetectedInstance: autogold.Expect(map[string]string{"instance_id": "new-instance-id", "license_id": "license-id-without-detected-instance"}),
	}, {
		name: "license already used by same instance",
		req: &subscriptionlicensechecksv1.CheckLicenseKeyRequest{
			InstanceId: knownInstanceID,
			LicenseKey: licenseKeyWithDetectedInstance,
		},
		wantResult: autogold.Expect(map[string]interface{}{"reason": "", "valid": true}),
	}} {
		t.Run(tc.name, func(t *testing.T) {
			// Clone the underlying mock store, to avoid polluting other test
			// cases
			store := NewMockStoreV1From(store)
			store.BypassAllLicenseChecksFunc.SetDefaultReturn(tc.bypass)

			h := &handlerV1{
				logger: logtest.Scoped(t),
				store:  store,
			}
			resp, err := h.CheckLicenseKey(context.Background(), connect.NewRequest(tc.req))
			if tc.wantErr != nil {
				require.Error(t, err)
				tc.wantErr.Equal(t, err.Error())
			} else {
				assert.NoError(t, err)
			}
			if tc.wantResult != nil {
				tc.wantResult.Equal(t, map[string]any{
					"valid":  resp.Msg.Valid,
					"reason": resp.Msg.Reason,
				})
			} else {
				assert.Nil(t, resp)
			}

			if tc.wantSetDetectedInstance != nil {
				mockrequire.CalledOnce(t, store.SetDetectedInstanceFunc)
				call := store.SetDetectedInstanceFunc.History()[0]
				tc.wantSetDetectedInstance.Equal(t, map[string]string{
					"license_id":  call.Arg1,
					"instance_id": call.Arg2,
				})
			} else {
				mockrequire.NotCalled(t, store.SetDetectedInstanceFunc)
			}
			if tc.wantPostToSlack != nil {
				mockrequire.CalledOnce(t, store.PostToSlackFunc)
				tc.wantPostToSlack.Equal(t, store.PostToSlackFunc.History()[0].Arg1)
			} else {
				mockrequire.NotCalled(t, store.PostToSlackFunc)
			}
		})
	}

}
