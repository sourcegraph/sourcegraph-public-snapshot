package graphqlbackend

import (
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

type siteUsageStatisticsResolver struct {
	siteUsageStatistics *types.SiteUsageStatistics
}

func (s *siteUsageStatisticsResolver) DAUs() []*siteUsagePeriodResolver {
	daus := make([]*siteUsagePeriodResolver, 0, len(s.siteUsageStatistics.DAUs))
	for _, d := range s.siteUsageStatistics.DAUs {
		daus = append(daus, &siteUsagePeriodResolver{
			siteUsagePeriod: d,
		})
	}
	return daus
}

func (s *siteUsageStatisticsResolver) WAUs() []*siteUsagePeriodResolver {
	waus := make([]*siteUsagePeriodResolver, 0, len(s.siteUsageStatistics.WAUs))
	for _, w := range s.siteUsageStatistics.WAUs {
		waus = append(waus, &siteUsagePeriodResolver{
			siteUsagePeriod: w,
		})
	}
	return waus
}

func (s *siteUsageStatisticsResolver) MAUs() []*siteUsagePeriodResolver {
	maus := make([]*siteUsagePeriodResolver, 0, len(s.siteUsageStatistics.MAUs))
	for _, m := range s.siteUsageStatistics.MAUs {
		maus = append(maus, &siteUsagePeriodResolver{
			siteUsagePeriod: m,
		})
	}
	return maus
}

type siteUsagePeriodResolver struct {
	siteUsagePeriod *types.SiteActivityPeriod
}

func (s *siteUsagePeriodResolver) StartTime() string {
	return s.siteUsagePeriod.StartTime.Format(time.RFC3339)
}

func (s *siteUsagePeriodResolver) UserCount() int32 {
	return s.siteUsagePeriod.UserCount
}

func (s *siteUsagePeriodResolver) RegisteredUserCount() int32 {
	return s.siteUsagePeriod.RegisteredUserCount
}

func (s *siteUsagePeriodResolver) AnonymousUserCount() int32 {
	return s.siteUsagePeriod.AnonymousUserCount
}

func (s *siteUsagePeriodResolver) IntegrationUserCount() int32 {
	return s.siteUsagePeriod.IntegrationUserCount
}
