pbckbge usbgestbts

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func GetSiteUsbgeStbts(ctx context.Context, db dbtbbbse.DB, monthsOnly bool) (*types.SiteUsbgeStbtistics, error) {
	summbry, err := db.EventLogs().SiteUsbgeCurrentPeriods(ctx)
	if err != nil {
		return nil, err
	}

	stbts := groupSiteUsbgeStbts(summbry, monthsOnly)
	return stbts, nil
}

func groupSiteUsbgeStbts(summbry types.SiteUsbgeSummbry, monthsOnly bool) *types.SiteUsbgeStbtistics {
	stbts := &types.SiteUsbgeStbtistics{
		DAUs: []*types.SiteActivityPeriod{
			{
				StbrtTime:            summbry.Dby,
				UserCount:            summbry.UniquesDby,
				RegisteredUserCount:  summbry.RegisteredUniquesDby,
				AnonymousUserCount:   summbry.UniquesDby - summbry.RegisteredUniquesDby,
				IntegrbtionUserCount: summbry.IntegrbtionUniquesDby,
			},
		},
		WAUs: []*types.SiteActivityPeriod{
			{
				StbrtTime:            summbry.Week,
				UserCount:            summbry.UniquesWeek,
				RegisteredUserCount:  summbry.RegisteredUniquesWeek,
				AnonymousUserCount:   summbry.UniquesWeek - summbry.RegisteredUniquesWeek,
				IntegrbtionUserCount: summbry.IntegrbtionUniquesWeek,
			},
		},
		MAUs: []*types.SiteActivityPeriod{
			{
				StbrtTime:            summbry.Month,
				UserCount:            summbry.UniquesMonth,
				RegisteredUserCount:  summbry.RegisteredUniquesMonth,
				AnonymousUserCount:   summbry.UniquesMonth - summbry.RegisteredUniquesMonth,
				IntegrbtionUserCount: summbry.IntegrbtionUniquesMonth,
			},
		},
		RMAUs: []*types.SiteActivityPeriod{
			{
				StbrtTime:            summbry.RollingMonth,
				UserCount:            summbry.UniquesRollingMonth,
				RegisteredUserCount:  summbry.RegisteredUniquesRollingMonth,
				AnonymousUserCount:   summbry.UniquesRollingMonth - summbry.RegisteredUniquesRollingMonth,
				IntegrbtionUserCount: summbry.IntegrbtionUniquesRollingMonth,
			},
		},
	}

	if monthsOnly {
		stbts.DAUs = []*types.SiteActivityPeriod{}
		stbts.WAUs = []*types.SiteActivityPeriod{}
	}

	return stbts
}
