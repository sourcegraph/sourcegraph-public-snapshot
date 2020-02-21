package graphqlbackend

import (
	"context"

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
	opt := &usagestats.CodeIntelUsageStatisticsOptions{
		IncludeEventCounts:    true,
		IncludeEventLatencies: true,
	}
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

func (s *codeIntelUsageStatisticsResolver) Daily() []*codeIntelUsagePeriodResolver {
	return s.activities(s.codeIntelUsageStatistics.Daily)
}

func (s *codeIntelUsageStatisticsResolver) Weekly() []*codeIntelUsagePeriodResolver {
	return s.activities(s.codeIntelUsageStatistics.Weekly)
}

func (s *codeIntelUsageStatisticsResolver) Monthly() []*codeIntelUsagePeriodResolver {
	return s.activities(s.codeIntelUsageStatistics.Monthly)
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

func (s *codeIntelUsagePeriodResolver) StartTime() DateTime {
	return DateTime{s.codeIntelUsagePeriod.StartTime}
}

func (s *codeIntelUsagePeriodResolver) Hover() *codeIntelEventCategoryStatisticsResolver {
	return &codeIntelEventCategoryStatisticsResolver{CodeIntelEventCategoryStatistics: s.codeIntelUsagePeriod.Hover}
}

func (s *codeIntelUsagePeriodResolver) Definitions() *codeIntelEventCategoryStatisticsResolver {
	return &codeIntelEventCategoryStatisticsResolver{CodeIntelEventCategoryStatistics: s.codeIntelUsagePeriod.Definitions}
}

func (s *codeIntelUsagePeriodResolver) References() *codeIntelEventCategoryStatisticsResolver {
	return &codeIntelEventCategoryStatisticsResolver{CodeIntelEventCategoryStatistics: s.codeIntelUsagePeriod.References}
}

type codeIntelEventCategoryStatisticsResolver struct {
	CodeIntelEventCategoryStatistics *types.CodeIntelEventCategoryStatistics
}

func (s *codeIntelEventCategoryStatisticsResolver) LSIF() *codeIntelEventStatisticsResolver {
	return &codeIntelEventStatisticsResolver{codeIntelEventStatistics: s.CodeIntelEventCategoryStatistics.LSIF}
}

func (s *codeIntelEventCategoryStatisticsResolver) LSP() *codeIntelEventStatisticsResolver {
	return &codeIntelEventStatisticsResolver{codeIntelEventStatistics: s.CodeIntelEventCategoryStatistics.LSP}
}

func (s *codeIntelEventCategoryStatisticsResolver) Search() *codeIntelEventStatisticsResolver {
	return &codeIntelEventStatisticsResolver{codeIntelEventStatistics: s.CodeIntelEventCategoryStatistics.Search}
}

type codeIntelEventStatisticsResolver struct {
	codeIntelEventStatistics *types.CodeIntelEventStatistics
}

func (s *codeIntelEventStatisticsResolver) UsersCount() int32 {
	return s.codeIntelEventStatistics.UsersCount
}

func (s *codeIntelEventStatisticsResolver) EventsCount() int32 {
	// Dereference is safe: we queried with IncludeEventCounts=true so this field is guaranteed to be populated.
	return *s.codeIntelEventStatistics.EventsCount
}

func (s *codeIntelEventStatisticsResolver) EventLatencies() *codeIntelEventLatenciesResolver {
	// Dereference is safe: we queried with IncludeEventLatencies=true so this field is guaranteed to be populated.
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
