package resolvers

type usagePattern struct {
	query string
}

func newQueryUsagePattern(query string) usagePattern {
	return usagePattern{
		query: `repo:^github\.com/sourcegraph/sourcegraph$ ` + query,
	}
}
