package main

type rawResult struct {
	Data   result        `json:"data,omitempty"`
	Errors []interface{} `json:"errors,omitempty"`
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
	Results                    []map[string]interface{}
	LimitHit                   bool
	Cloning, Missing, Timedout []map[string]interface{}
	ResultCount                int
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
	took  int64
	trace string
}
