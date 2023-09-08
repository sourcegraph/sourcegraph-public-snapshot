package productsubscription

import (
	"context"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/slack"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

type fakeSlackClient struct {
	payloads []*slack.Payload
}

func (c *fakeSlackClient) Post(ctx context.Context, payload *slack.Payload) error {
	c.payloads = append(c.payloads, payload)
	return nil
}

func TestCheckForUpcomingLicenseExpirations(t *testing.T) {
	clock := glock.NewMockClock()

	cfg := conf.Get()
	conf.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			Dotcom: &schema.Dotcom{
				SlackLicenseExpirationWebhook: "https://slack.com/webhook",
			},
		},
	})
	mocks.subscriptions.List = func(ctx context.Context, opt dbSubscriptionsListOptions) ([]*dbSubscription, error) {
		return []*dbSubscription{
			{ID: "e9450fb2-87c7-47ae-a713-a376c4618faa"},
			{ID: "26136564-b319-4be4-98ff-7b8710abf4af"},
		}, nil
	}
	mocks.licenses.List = func(ctx context.Context, opt dbLicensesListOptions) ([]*dbLicense, error) {
		return []*dbLicense{{LicenseKey: opt.ProductSubscriptionID}}, nil
	}
	licensing.MockParseProductLicenseKeyWithBuiltinOrGenerationKey = func(licenseKey string) (*licensing.Info, string, error) {
		infos := map[string]*licensing.Info{
			"e9450fb2-87c7-47ae-a713-a376c4618faa": {
				Info: license.Info{
					ExpiresAt: clock.Now().Add((24 + 1) * time.Hour), // day away
				},
			},
			"26136564-b319-4be4-98ff-7b8710abf4af": {
				Info: license.Info{
					ExpiresAt: clock.Now().Add((7*24 + 1) * time.Hour), // week away
				},
			},
		}
		return infos[licenseKey], "", nil
	}

	users := dbmocks.NewStrictMockUserStore()
	users.GetByIDFunc.SetDefaultReturn(&types.User{Username: "alice"}, nil)

	db := dbmocks.NewStrictMockDB()
	db.UsersFunc.SetDefaultReturn(users)

	t.Cleanup(func() {
		conf.Mock(cfg)
		mocks.subscriptions = mockSubscriptions{}
		mocks.licenses = mockLicenses{}
		licensing.MockParseProductLicenseKeyWithBuiltinOrGenerationKey = nil
	})

	client := &fakeSlackClient{}
	checkForUpcomingLicenseExpirations(logtest.Scoped(t), db, clock, client)

	wantPayloads := []*slack.Payload{
		{Text: "The license for user `alice` <https://sourcegraph.com/site-admin/dotcom/product/subscriptions/e9450fb2-87c7-47ae-a713-a376c4618faa|will expire *in the next 24 hours*> :rotating_light:"},
		{Text: "The license for user `alice` <https://sourcegraph.com/site-admin/dotcom/product/subscriptions/26136564-b319-4be4-98ff-7b8710abf4af|will expire *in 7 days*>"},
	}
	if diff := cmp.Diff(wantPayloads, client.payloads); diff != "" {
		t.Fatalf("Payloads mismatch (-want +got):\n%s", diff)
	}
}
