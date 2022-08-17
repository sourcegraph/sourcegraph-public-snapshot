package streaming

type testAggregator struct {
	results map[string]int
}

func (r *testAggregator) AddResult(result *AggregationMatchResult, err error) {
	if err != nil {
		return
	}
	current, _ := r.results[result.Key.Group]
	r.results[result.Key.Group] = result.Count + current
}
