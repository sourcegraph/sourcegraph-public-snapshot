package graphqlbackend

import (
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

type siteActivityResolver struct {
	siteActivity *types.SiteActivity
}

func (s *siteActivityResolver) DAUs() []*siteActivityPeriodResolver {
	daus := make([]*siteActivityPeriodResolver, 0, len(s.siteActivity.DAUs))
	for _, d := range s.siteActivity.DAUs {
		daus = append(daus, &siteActivityPeriodResolver{
			siteActivityPeriod: d,
		})
	}
	return daus
}

func (s *siteActivityResolver) WAUs() []*siteActivityPeriodResolver {
	waus := make([]*siteActivityPeriodResolver, 0, len(s.siteActivity.WAUs))
	for _, w := range s.siteActivity.WAUs {
		waus = append(waus, &siteActivityPeriodResolver{
			siteActivityPeriod: w,
		})
	}
	return waus
}

func (s *siteActivityResolver) MAUs() []*siteActivityPeriodResolver {
	maus := make([]*siteActivityPeriodResolver, 0, len(s.siteActivity.MAUs))
	for _, m := range s.siteActivity.MAUs {
		maus = append(maus, &siteActivityPeriodResolver{
			siteActivityPeriod: m,
		})
	}
	return maus
}

type siteActivityPeriodResolver struct {
	siteActivityPeriod *types.SiteActivityPeriod
}

func (s *siteActivityPeriodResolver) StartTime() string {
	return s.siteActivityPeriod.StartTime.Format(time.RFC3339)
}

func (s *siteActivityPeriodResolver) UserCount() int32 {
	return s.siteActivityPeriod.UserCount
}

func (s *siteActivityPeriodResolver) RegisteredUserCount() int32 {
	return s.siteActivityPeriod.RegisteredUserCount
}

func (s *siteActivityPeriodResolver) AnonymousUserCount() int32 {
	return s.siteActivityPeriod.AnonymousUserCount
}

func (s *siteActivityPeriodResolver) IntegrationUserCount() int32 {
	return s.siteActivityPeriod.IntegrationUserCount
}
