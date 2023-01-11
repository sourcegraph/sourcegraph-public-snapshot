package gqltestutil

func (c *Client) ModeAvailability(query, patternType string) (map[string]ModeAvailabilityResponse, error) {
	const gqlQuery = `
		query ModeAvailability($query: String!, $patternType: SearchPatternType!) {
			searchQueryAggregate(query: $query, patternType: $patternType) {
				modeAvailability {
					mode
					available
					reasonUnavailable
				}
			}
		}
	`

	variables := map[string]any{
		"query":       query,
		"patternType": patternType,
	}
	var resp struct {
		Data struct {
			SearchQueryAggregate struct {
				ModeAvailability []ModeAvailabilityResponse
			} `json:"searchQueryAggregate"`
		} `json:"data"`
	}

	err := c.GraphQL("", gqlQuery, variables, &resp)
	if err != nil {
		return nil, err
	}

	modeAvailability := make(map[string]ModeAvailabilityResponse)
	for _, response := range resp.Data.SearchQueryAggregate.ModeAvailability {
		modeAvailability[response.Mode] = response
	}

	return modeAvailability, nil
}

type ModeAvailabilityResponse struct {
	Mode              string  `json:"mode"`
	Available         bool    `json:"available"`
	ReasonUnavailable *string `json:"reasonUnavailable"`
}

func (c *Client) Aggregations(args AggregationArgs) (AggregationResponse, error) {
	gqlQuery := `
		query ModeAvailability(
		  $query: String!
		  $patternType: SearchPatternType!
		  $mode: SearchAggregationMode
		) {
		  searchQueryAggregate(query: $query, patternType: $patternType) {
			aggregations(mode: $mode) {
			  ... on ExhaustiveSearchAggregationResult {
				mode
				groups {
				  query
				}
			  }
			  ... on NonExhaustiveSearchAggregationResult {
				mode
				groups {
				  query
				}
			  }
			  ... on SearchAggregationNotAvailable {
				reason
				reasonType
			  }
			}
		  }
		}
	`

	variables := map[string]any{
		"query":           args.Query,
		"patternType":     args.PatternType,
		"mode":            args.Mode,
		"limit":           args.Limit,
		"extendedTimeout": args.ExtendedTimeout,
	}

	var resp struct {
		Data struct {
			SearchQueryAggregate struct {
				Aggregations AggregationResponse
			} `json:"searchQueryAggregate"`
		} `json:"data"`
	}

	err := c.GraphQL("", gqlQuery, variables, &resp)
	if err != nil {
		return AggregationResponse{}, err
	}
	return resp.Data.SearchQueryAggregate.Aggregations, nil
}

type AggregationArgs struct {
	Query           string
	PatternType     string
	ExtendedTimeout bool
	Mode            *string
	Limit           *int32
}

type AggregationResponse struct {
	ReasonType string
	Reason     string // If this is set the fields below will be empty.

	Groups []struct {
		Label string
		Count int32
		Query string
	} // List of results in form of labels
	Mode string
}
