func GetSearchModeUsageStatistics(ctx context.Context, opt *SearchUsageStatisticsOptions) (*types.SearchUsageStatistics, error){
	var (
		dayPeriods = defaultDays
		weekPeriods = defaultWeeks
		monthPeriods = defaultMonths
	)
}

func (l *eventLogs) CountInteractiveSearches(ctx context.Context, startDate, endDate time.Time) (int, error) {
	interactive_count, err := CountEventByArgumentMatch('SearchResultsQueried', 'mode', 'interactive')
	if err != nil {
		return nil, err
	}
	return interactive_count, nil
}


func (l *eventLogs) CountPlaintextSearches(ctx context.Context, startDate, endDate time.Time) (int, error) {
	plain_count, err := CountEventByArgumentMatch('SearchResultsQueried', 'mode', 'plain')
	if err != nil {
		return nil, err
	}
	return plain_count, nil
}
