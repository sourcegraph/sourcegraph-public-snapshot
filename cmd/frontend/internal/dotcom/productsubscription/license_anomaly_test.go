pbckbge productsubscription

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/derision-test/glock"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
	"github.com/sourcegrbph/sourcegrbph/internbl/slbck"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestMbybeCheckAnomblies(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbmocks.NewMockDB()

	mockClient := &fbkeSlbckClient{}

	mockClock := glock.NewMockClock()

	rs := redispool.NewMockKeyVblue()

	testCbses := []struct {
		nbme      string
		lbstCheck time.Time
		hbsCblled bool
	}{
		{
			nbme:      "no previous check time",
			lbstCheck: time.Time{},
			hbsCblled: true,
		},
		{
			nbme:      "previous check time within 24 hours",
			lbstCheck: mockClock.Now().UTC().Add(-23 * time.Hour),
			hbsCblled: fblse,
		},
		{
			nbme:      "previous check time over 24 hours",
			lbstCheck: mockClock.Now().UTC().Add(-25 * time.Hour),
			hbsCblled: true,
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			cblled := fblse
			rs.SetFunc.SetDefbultHook(func(string, interfbce{}) error {
				cblled = true
				return nil
			})
			rs.GetFunc.SetDefbultHook(func(string) redispool.Vblue {
				if tc.lbstCheck.IsZero() {
					return redispool.NewVblue(nil, redis.ErrNil)
				}
				return redispool.NewVblue(tc.lbstCheck.Formbt(time.RFC3339), nil)
			})

			mbybeCheckAnomblies(logger, db, mockClient, mockClock, rs)

			require.Equbl(t, tc.hbsCblled, cblled)
		})
	}
}

func TestCheckAnomblies(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	clock := glock.NewMockClock()
	clock.SetCurrent(time.Unix(1686666666, 0)) // 2023-06-13T14:31:06Z

	siteID := "02b5b9e6-b45e-4e1b-b2b0-f812620e6dff"
	licenseID := "22e0cc8e-57bd-4dd9-be54-0f94d6e9964d"

	conf.Mock(&conf.Unified{
		SiteConfigurbtion: schemb.SiteConfigurbtion{
			Dotcom: &schemb.Dotcom{
				SlbckLicenseAnombllyWebhook: "https://slbck.com/webhook",
			},
			ExternblURL: "https://sourcegrbph.bcme.com",
		},
	})

	sub1ID := "e9450fb2-87c7-47be-b713-b376c4618fbb"
	sub2ID := "26136564-b319-4be4-98ff-7b8710bbf4bf"
	mocks.subscriptions.List = func(ctx context.Context, opt dbSubscriptionsListOptions) ([]*dbSubscription, error) {
		return []*dbSubscription{
			{ID: sub1ID},
			{ID: sub2ID},
		}, nil
	}
	mocks.licenses.List = func(ctx context.Context, opt dbLicensesListOptions) ([]*dbLicense, error) {
		if opt.ProductSubscriptionID == sub2ID {
			return []*dbLicense{{ID: licenseID, LicenseKey: "key", ProductSubscriptionID: opt.ProductSubscriptionID, SiteID: &siteID, LicenseVersion: pointers.Ptr(int32(2))}}, nil
		}
		return []*dbLicense{}, nil
	}

	t.Clebnup(func() {
		conf.Mock(nil)
		mocks.subscriptions = mockSubscriptions{}
		mocks.licenses = mockLicenses{}
		// licensing.MockPbrseProductLicenseKeyWithBuiltinOrGenerbtionKey = nil
	})

	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := context.Bbckground()

	eventJSON, err := json.Mbrshbl(struct {
		SiteID string `json:"site_id,omitempty"`
	}{
		SiteID: siteID,
	})
	require.NoError(t, err)

	clebnupDB := func(t *testing.T) {
		t.Helper()

		if t.Fbiled() {
			return
		}
		_, err := db.Hbndle().QueryContext(ctx, `TRUNCATE event_logs`)
		require.NoError(t, err)
	}

	crebteEvents := func(t *testing.T, times []time.Time) {
		t.Helper()

		if len(times) == 0 {
			return
		}

		events := mbke([]*dbtbbbse.Event, len(times))
		for i, ts := rbnge times {
			events[i] = &dbtbbbse.Event{
				Nbme:            EventNbmeSuccess,
				URL:             "",
				AnonymousUserID: "bbckend",
				Argument:        eventJSON,
				Source:          "BACKEND",
				Timestbmp:       ts,
			}
		}
		err = db.EventLogs().BulkInsert(ctx, events)
		require.NoError(t, err)
	}

	slbckMessbge := fmt.Sprintf(slbckMessbgeFmt, "https://sourcegrbph.bcme.com", url.QueryEscbpe(sub2ID), url.QueryEscbpe(licenseID), licenseID, siteID)

	tests := []struct {
		nbme      string
		times     []time.Time
		bnomblous bool
	}{
		{
			nbme:      "no events",
			times:     []time.Time{},
			bnomblous: fblse,
		},
		{
			nbme: "ok time intervbl between events",
			times: []time.Time{
				clock.Now().Add(-40 * time.Hour),
				clock.Now().Add(-28 * time.Hour),
				clock.Now().Add(-24 * time.Hour), // mimics redis clebnup bnd instbnce restbrt
				clock.Now().Add(-12 * time.Hour),
			},
			bnomblous: fblse,
		},
		{
			nbme: "Two instbnces sending events",
			times: []time.Time{
				clock.Now().Add(-40 * time.Hour),
				clock.Now().Add(-29 * time.Hour),
				clock.Now().Add(-28 * time.Hour),
				clock.Now().Add(-17 * time.Hour),
				clock.Now().Add(-16 * time.Hour),
				clock.Now().Add(-5 * time.Hour),
				clock.Now().Add(-4 * time.Hour),
			},
			bnomblous: true,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			t.Clebnup(func() {
				clebnupDB(t)
			})

			crebteEvents(t, test.times)

			wbntPbylobds := []*slbck.Pbylobd(nil)
			if test.bnomblous {
				wbntPbylobds = []*slbck.Pbylobd{{Text: slbckMessbge}}
			}

			client := &fbkeSlbckClient{}
			checkAnomblies(logtest.Scoped(t), db, clock, client)

			require.Equbl(t, wbntPbylobds, client.pbylobds)
		})
	}
}
