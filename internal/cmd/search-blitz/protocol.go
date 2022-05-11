package main

import "time"

type rawResult struct {
	Data   result `json:"data,omitempty"`
	Errors []any  `json:"errors,omitempty"`
}

type result struct {
	Site struct {
		BuildVersion string
	}
	Search struct {
		Results searchResults
	}
}

// searchResults represents the data we get back from the GraphQL search request.
type searchResults struct {
	Results                    []map[string]any
	LimitHit                   bool
	Cloning, Missing, Timedout []map[string]any
	MatchCount                 int
	ElapsedMilliseconds        int
	Alert                      searchResultsAlert
}

// searchResultsAlert is a type that can be used to unmarshal values returned by
// the searchResultsAlertFragment GraphQL fragment below.
type searchResultsAlert struct {
	Title           string
	Description     string
	ProposedQueries []struct {
		Description string
		Query       string
	}
}

type metrics struct {
	took        time.Duration
	firstResult time.Duration
	matchCount  int
	trace       string
}
