package graphqlbackend

import (
	"context"
	"errors"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/usagestats"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

type codeIntelUsageStatisticsResolver struct {
	codeIntelUsageStatistics *types.CodeIntelUsageStatistics
}

func (r *siteResolver) CodeIntelUsageStatistics(ctx context.Context, args *struct {
	Days   *int32
	Weeks  *int32
	Months *int32
}) (*codeIntelUsageStatisticsResolver, error) {
	if envvar.SourcegraphDotComMode() {
		return nil, errors.New("code intel usage statistics are not available on sourcegraph.com")
	}
	opt := &usagestats.CodeIntelUsageStatisticsOptions{}
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
	activity, err := usagestats.GetCodeIntelUsageStatistics(ctx, opt)
	if err != nil {
		return nil, err
	}
	return &codeIntelUsageStatisticsResolver{activity}, nil
}

func (s *codeIntelUsageStatisticsResolver) DailyActivities() []*codeIntelUsagePeriodResolver {
	return s.activities(s.codeIntelUsageStatistics.DailyActivities)
}

func (s *codeIntelUsageStatisticsResolver) WeeklyActivities() []*codeIntelUsagePeriodResolver {
	return s.activities(s.codeIntelUsageStatistics.WeeklyActivities)
}

func (s *codeIntelUsageStatisticsResolver) MonthlyActivities() []*codeIntelUsagePeriodResolver {
	return s.activities(s.codeIntelUsageStatistics.MonthlyActivities)
}

func (s *codeIntelUsageStatisticsResolver) activities(periods []*types.CodeIntelUsagePeriod) []*codeIntelUsagePeriodResolver {
	resolvers := make([]*codeIntelUsagePeriodResolver, 0, len(periods))
	for _, p := range periods {
		resolvers = append(resolvers, &codeIntelUsagePeriodResolver{codeIntelUsagePeriod: p})
	}
	return resolvers
}

type codeIntelUsagePeriodResolver struct {
	codeIntelUsagePeriod *types.CodeIntelUsagePeriod
}

func (s *codeIntelUsagePeriodResolver) StartTime() string {
	return s.codeIntelUsagePeriod.StartTime.Format(time.RFC3339)
}

func (s *codeIntelUsagePeriodResolver) PreciseHoverStatistics() *codeIntelEventStatisticsResolver {
	return &codeIntelEventStatisticsResolver{codeIntelEventStatistics: s.codeIntelUsagePeriod.PreciseHoverStatistics}
}

func (s *codeIntelUsagePeriodResolver) FuzzyHoverStatistics() *codeIntelEventStatisticsResolver {
	return &codeIntelEventStatisticsResolver{codeIntelEventStatistics: s.codeIntelUsagePeriod.FuzzyHoverStatistics}
}

func (s *codeIntelUsagePeriodResolver) PreciseDefinitionsStatistics() *codeIntelEventStatisticsResolver {
	return &codeIntelEventStatisticsResolver{codeIntelEventStatistics: s.codeIntelUsagePeriod.PreciseDefinitionsStatistics}
}

func (s *codeIntelUsagePeriodResolver) FuzzyDefinitionsStatistics() *codeIntelEventStatisticsResolver {
	return &codeIntelEventStatisticsResolver{codeIntelEventStatistics: s.codeIntelUsagePeriod.FuzzyDefinitionsStatistics}
}

func (s *codeIntelUsagePeriodResolver) PreciseReferencesStatistics() *codeIntelEventStatisticsResolver {
	return &codeIntelEventStatisticsResolver{codeIntelEventStatistics: s.codeIntelUsagePeriod.PreciseReferencesStatistics}
}

func (s *codeIntelUsagePeriodResolver) FuzzyReferencesStatistics() *codeIntelEventStatisticsResolver {
	return &codeIntelEventStatisticsResolver{codeIntelEventStatistics: s.codeIntelUsagePeriod.FuzzyReferencesStatistics}
}

type codeIntelEventStatisticsResolver struct {
	codeIntelEventStatistics *types.CodeIntelEventStatistics
}

func (s *codeIntelEventStatisticsResolver) UsersCount() int32 {
	return s.codeIntelEventStatistics.UsersCount
}

func (s *codeIntelEventStatisticsResolver) EventsCount() int32 {
	return s.codeIntelEventStatistics.EventsCount
}

func (s *codeIntelEventStatisticsResolver) EventLatencies() *codeIntelEventLatenciesResolver {
	return &codeIntelEventLatenciesResolver{codeIntelEventLatencies: s.codeIntelEventStatistics.EventLatencies}
}

type codeIntelEventLatenciesResolver struct {
	codeIntelEventLatencies *types.CodeIntelEventLatencies
}

func (s *codeIntelEventLatenciesResolver) P50() float64 {
	return s.codeIntelEventLatencies.P50
}

func (s *codeIntelEventLatenciesResolver) P90() float64 {
	return s.codeIntelEventLatencies.P90
}

func (s *codeIntelEventLatenciesResolver) P99() float64 {
	return s.codeIntelEventLatencies.P99
}
