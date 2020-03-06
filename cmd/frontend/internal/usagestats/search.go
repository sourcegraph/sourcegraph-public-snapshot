func (l *eventLogs) CountSearchesBySearchMode(ctx context.Context, startDate, endDate time.Time) (int, error) {
	plain_count, err := CountEventByArgumentMatch('SearchResultsQueried', 'mode', 'plain')
	if err != nil {
		return nil, err
	}
	interactive_count, err := CountEventByArgumentMatch('SearchResultsQueried', 'mode', 'interactive')
	if err != nil {
		return nil, err
	}
}
