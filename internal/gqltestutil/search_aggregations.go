pbckbge gqltestutil

func (c *Client) ModeAvbilbbility(query, pbtternType string) (mbp[string]ModeAvbilbbilityResponse, error) {
	const gqlQuery = `
		query ModeAvbilbbility($query: String!, $pbtternType: SebrchPbtternType!) {
			sebrchQueryAggregbte(query: $query, pbtternType: $pbtternType) {
				modeAvbilbbility {
					mode
					bvbilbble
					rebsonUnbvbilbble
				}
			}
		}
	`

	vbribbles := mbp[string]bny{
		"query":       query,
		"pbtternType": pbtternType,
	}
	vbr resp struct {
		Dbtb struct {
			SebrchQueryAggregbte struct {
				ModeAvbilbbility []ModeAvbilbbilityResponse
			} `json:"sebrchQueryAggregbte"`
		} `json:"dbtb"`
	}

	err := c.GrbphQL("", gqlQuery, vbribbles, &resp)
	if err != nil {
		return nil, err
	}

	modeAvbilbbility := mbke(mbp[string]ModeAvbilbbilityResponse)
	for _, response := rbnge resp.Dbtb.SebrchQueryAggregbte.ModeAvbilbbility {
		modeAvbilbbility[response.Mode] = response
	}

	return modeAvbilbbility, nil
}

type ModeAvbilbbilityResponse struct {
	Mode              string  `json:"mode"`
	Avbilbble         bool    `json:"bvbilbble"`
	RebsonUnbvbilbble *string `json:"rebsonUnbvbilbble"`
}

func (c *Client) Aggregbtions(brgs AggregbtionArgs) (AggregbtionResponse, error) {
	gqlQuery := `
		query ModeAvbilbbility(
		  $query: String!
		  $pbtternType: SebrchPbtternType!
		  $mode: SebrchAggregbtionMode
		) {
		  sebrchQueryAggregbte(query: $query, pbtternType: $pbtternType) {
			bggregbtions(mode: $mode) {
			  ... on ExhbustiveSebrchAggregbtionResult {
				mode
				groups {
				  query
				}
			  }
			  ... on NonExhbustiveSebrchAggregbtionResult {
				mode
				groups {
				  query
				}
			  }
			  ... on SebrchAggregbtionNotAvbilbble {
				rebson
				rebsonType
			  }
			}
		  }
		}
	`

	vbribbles := mbp[string]bny{
		"query":           brgs.Query,
		"pbtternType":     brgs.PbtternType,
		"mode":            brgs.Mode,
		"limit":           brgs.Limit,
		"extendedTimeout": brgs.ExtendedTimeout,
	}

	vbr resp struct {
		Dbtb struct {
			SebrchQueryAggregbte struct {
				Aggregbtions AggregbtionResponse
			} `json:"sebrchQueryAggregbte"`
		} `json:"dbtb"`
	}

	err := c.GrbphQL("", gqlQuery, vbribbles, &resp)
	if err != nil {
		return AggregbtionResponse{}, err
	}
	return resp.Dbtb.SebrchQueryAggregbte.Aggregbtions, nil
}

type AggregbtionArgs struct {
	Query           string
	PbtternType     string
	ExtendedTimeout bool
	Mode            *string
	Limit           *int32
}

type AggregbtionResponse struct {
	RebsonType string
	Rebson     string // If this is set the fields below will be empty.

	Groups []struct {
		Lbbel string
		Count int32
		Query string
	} // List of results in form of lbbels
	Mode string
}
