package graphqlbackend

import (
	"context"
	"errors"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/usagestats"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func (r *siteResolver) UsageStatistics(ctx context.Context, args *struct {
	Days   *int32
	Weeks  *int32
	Months *int32
}) (*siteUsageStatisticsResolver, error) {
	if envvar.SourcegraphDotComMode() {
		return nil, errors.New("site usage statistics are not available on sourcegraph.com")
	}
	opt := &usagestats.SiteUsageStatisticsOptions{}
	if args.Days != nil {
		d := int(*args.Days)
		opt.DayPeriods = &d
	}
	if args.Weeks != nil {
		w := int(*args.Weeks)
		opt.WeekPeriods = &w
	}
	if args.Months != nil {
		m := int(*args.Months)
		opt.MonthPeriods = &m
	}
	activity, err := usagestats.GetSiteUsageStatistics(ctx, opt)
	if err != nil {
		return nil, err
	}
	return &siteUsageStatisticsResolver{activity}, nil
}

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

func (s *siteUsagePeriodResolver) Stages() *siteUsageStagesResolver {
	if s.siteUsagePeriod.Stages == nil {
		return nil
	}
	return &siteUsageStagesResolver{
		stages: s.siteUsagePeriod.Stages,
	}
}

type siteUsageStagesResolver struct {
	stages *types.Stages
}

func (s *siteUsageStagesResolver) Manage() int32 {
	return s.stages.Manage
}

func (s *siteUsageStagesResolver) Plan() int32 {
	return s.stages.Plan
}

func (s *siteUsageStagesResolver) Code() int32 {
	return s.stages.Code
}

func (s *siteUsageStagesResolver) Review() int32 {
	return s.stages.Review
}

func (s *siteUsageStagesResolver) Verify() int32 {
	return s.stages.Verify
}

func (s *siteUsageStagesResolver) Package() int32 {
	return s.stages.Package
}

func (s *siteUsageStagesResolver) Deploy() int32 {
	return s.stages.Deploy
}

func (s *siteUsageStagesResolver) Configure() int32 {
	return s.stages.Configure
}

func (s *siteUsageStagesResolver) Monitor() int32 {
	return s.stages.Monitor
}

func (s *siteUsageStagesResolver) Secure() int32 {
	return s.stages.Secure
}

func (s *siteUsageStagesResolver) Automate() int32 {
	return s.stages.Automate
}
