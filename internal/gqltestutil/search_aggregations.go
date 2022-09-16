package gqltestutil

//type AggregationArgs struct {
//	Query       string `json:"query"`
//	PatternType string `json:"patternType"`
//}

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
