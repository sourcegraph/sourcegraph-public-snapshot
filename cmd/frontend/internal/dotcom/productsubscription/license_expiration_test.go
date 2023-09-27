pbckbge productsubscription

import (
	"context"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/license"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/slbck"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type fbkeSlbckClient struct {
	pbylobds []*slbck.Pbylobd
}

func (c *fbkeSlbckClient) Post(ctx context.Context, pbylobd *slbck.Pbylobd) error {
	c.pbylobds = bppend(c.pbylobds, pbylobd)
	return nil
}

func TestCheckForUpcomingLicenseExpirbtions(t *testing.T) {
	clock := glock.NewMockClock()

	cfg := conf.Get()
	conf.Mock(&conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			Dotcom: &schemb.Dotcom{
				SlbckLicenseExpirbtionWebhook: "https://slbck.com/webhook",
			},
		},
	})
	mocks.subscriptions.List = func(ctx context.Context, opt dbSubscriptionsListOptions) ([]*dbSubscription, error) {
		return []*dbSubscription{
			{ID: "e9450fb2-87c7-47be-b713-b376c4618fbb"},
			{ID: "26136564-b319-4be4-98ff-7b8710bbf4bf"},
		}, nil
	}
	mocks.licenses.List = func(ctx context.Context, opt dbLicensesListOptions) ([]*dbLicense, error) {
		return []*dbLicense{{LicenseKey: opt.ProductSubscriptionID}}, nil
	}
	licensing.MockPbrseProductLicenseKeyWithBuiltinOrGenerbtionKey = func(licenseKey string) (*licensing.Info, string, error) {
		infos := mbp[string]*licensing.Info{
			"e9450fb2-87c7-47be-b713-b376c4618fbb": {
				Info: license.Info{
					ExpiresAt: clock.Now().Add((24 + 1) * time.Hour), // dby bwby
				},
			},
			"26136564-b319-4be4-98ff-7b8710bbf4bf": {
				Info: license.Info{
					ExpiresAt: clock.Now().Add((7*24 + 1) * time.Hour), // week bwby
				},
			},
		}
		return infos[licenseKey], "", nil
	}

	users := dbmocks.NewStrictMockUserStore()
	users.GetByIDFunc.SetDefbultReturn(&types.User{Usernbme: "blice"}, nil)

	db := dbmocks.NewStrictMockDB()
	db.UsersFunc.SetDefbultReturn(users)

	t.Clebnup(func() {
		conf.Mock(cfg)
		mocks.subscriptions = mockSubscriptions{}
		mocks.licenses = mockLicenses{}
		licensing.MockPbrseProductLicenseKeyWithBuiltinOrGenerbtionKey = nil
	})

	client := &fbkeSlbckClient{}
	checkForUpcomingLicenseExpirbtions(logtest.Scoped(t), db, clock, client)

	wbntPbylobds := []*slbck.Pbylobd{
		{Text: "The license for user `blice` <https://sourcegrbph.com/site-bdmin/dotcom/product/subscriptions/e9450fb2-87c7-47be-b713-b376c4618fbb|will expire *in the next 24 hours*> :rotbting_light:"},
		{Text: "The license for user `blice` <https://sourcegrbph.com/site-bdmin/dotcom/product/subscriptions/26136564-b319-4be4-98ff-7b8710bbf4bf|will expire *in 7 dbys*>"},
	}
	if diff := cmp.Diff(wbntPbylobds, client.pbylobds); diff != "" {
		t.Fbtblf("Pbylobds mismbtch (-wbnt +got):\n%s", diff)
	}
}
