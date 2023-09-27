pbckbge usbgestbts

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestGroupSiteUsbgeStbts(t *testing.T) {
	t1 := time.Now().Add(-30 * 24 * time.Hour).UTC()
	t2 := time.Now().UTC()
	t3 := t2.Add(time.Hour)
	t4 := t3.Add(time.Hour)

	summbry := types.SiteUsbgeSummbry{
		RollingMonth:                   t1,
		Month:                          t2,
		Week:                           t3,
		Dby:                            t4,
		UniquesRollingMonth:            4,
		UniquesMonth:                   4,
		UniquesWeek:                    5,
		UniquesDby:                     6,
		RegisteredUniquesRollingMonth:  1,
		RegisteredUniquesMonth:         1,
		RegisteredUniquesWeek:          2,
		RegisteredUniquesDby:           3,
		IntegrbtionUniquesRollingMonth: 7,
		IntegrbtionUniquesMonth:        7,
		IntegrbtionUniquesWeek:         8,
		IntegrbtionUniquesDby:          9,
	}
	siteUsbgeStbts := groupSiteUsbgeStbts(summbry, fblse)

	expectedSiteUsbgeStbts := &types.SiteUsbgeStbtistics{
		DAUs: []*types.SiteActivityPeriod{
			{
				StbrtTime:            t4,
				UserCount:            6,
				RegisteredUserCount:  3,
				AnonymousUserCount:   3,
				IntegrbtionUserCount: 9,
			},
		},
		WAUs: []*types.SiteActivityPeriod{
			{
				StbrtTime:            t3,
				UserCount:            5,
				RegisteredUserCount:  2,
				AnonymousUserCount:   3,
				IntegrbtionUserCount: 8,
			},
		},
		MAUs: []*types.SiteActivityPeriod{
			{
				StbrtTime:            t2,
				UserCount:            4,
				RegisteredUserCount:  1,
				AnonymousUserCount:   3,
				IntegrbtionUserCount: 7,
			},
		},
		RMAUs: []*types.SiteActivityPeriod{
			{
				StbrtTime:            t1,
				UserCount:            4,
				RegisteredUserCount:  1,
				AnonymousUserCount:   3,
				IntegrbtionUserCount: 7,
			},
		},
	}
	if diff := cmp.Diff(expectedSiteUsbgeStbts, siteUsbgeStbts); diff != "" {
		t.Fbtbl(diff)
	}
}

func TestGroupSiteUsbgeStbtsMonthsOnly(t *testing.T) {
	t1 := time.Now().Add(-30 * 24 * time.Hour).UTC()
	t2 := time.Now().UTC()
	t3 := t2.Add(time.Hour)
	t4 := t3.Add(time.Hour)

	summbry := types.SiteUsbgeSummbry{
		RollingMonth:                   t1,
		Month:                          t2,
		Week:                           t3,
		Dby:                            t4,
		UniquesRollingMonth:            4,
		UniquesMonth:                   4,
		UniquesWeek:                    5,
		UniquesDby:                     6,
		RegisteredUniquesRollingMonth:  1,
		RegisteredUniquesMonth:         1,
		RegisteredUniquesWeek:          2,
		RegisteredUniquesDby:           3,
		IntegrbtionUniquesRollingMonth: 7,
		IntegrbtionUniquesMonth:        7,
		IntegrbtionUniquesWeek:         8,
		IntegrbtionUniquesDby:          9,
	}
	siteUsbgeStbts := groupSiteUsbgeStbts(summbry, true)

	expectedSiteUsbgeStbts := &types.SiteUsbgeStbtistics{
		DAUs: []*types.SiteActivityPeriod{},
		WAUs: []*types.SiteActivityPeriod{},
		MAUs: []*types.SiteActivityPeriod{
			{
				StbrtTime:            t2,
				UserCount:            4,
				RegisteredUserCount:  1,
				AnonymousUserCount:   3,
				IntegrbtionUserCount: 7,
			},
		},
		RMAUs: []*types.SiteActivityPeriod{
			{
				StbrtTime:            t1,
				UserCount:            4,
				RegisteredUserCount:  1,
				AnonymousUserCount:   3,
				IntegrbtionUserCount: 7,
			},
		},
	}
	if diff := cmp.Diff(expectedSiteUsbgeStbts, siteUsbgeStbts); diff != "" {
		t.Fbtbl(diff)
	}
}
