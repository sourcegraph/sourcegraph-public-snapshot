package gqltestutil

import (
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (c *Client) GetInsights() ([]string, error) {
	const query = `
		query InsightViews {
			insightViews {
				nodes { id }
			}
		}
	`
	var resp struct {
		Data struct {
			InsightViews struct {
				Nodes []struct {
					Id string `json:"id"`
				} `json:"nodes"`
			} `json:"insightviews"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, map[string]any{}, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "request GraphQL")
	}

	ids := []string{}
	for _, node := range resp.Data.InsightViews.Nodes {
		ids = append(ids, node.Id)
	}

	return ids, nil
}

type GetDashboardArgs struct {
	First *int
	After *string
	Id    *string
}

type DashboardInputArgs struct {
	Title       string
	UserGrant   string
	OrgGrant    string
	GlobalGrant bool
}

type DashboardResponse struct {
	Id     string         `json:"id"`
	Title  string         `json:"title"`
	Grants GrantsResponse `json:"grants"`
}

type GrantsResponse struct {
	Users         []string `json:"users"`
	Organizations []string `json:"organizations"`
	Global        bool     `json:"global"`
}

func (c *Client) CreateDashboard(args DashboardInputArgs) (DashboardResponse, error) {
	const query = `
		mutation CreateInsightsDashboard($input: CreateInsightsDashboardInput!) {
			createInsightsDashboard(input: $input) {
				dashboard {
					id, title, grants { users, organizations, global }
				}
			}
		}
	`

	variables := map[string]any{
		"input": map[string]any{
			"title":  args.Title,
			"grants": buildGrants(args.UserGrant, args.OrgGrant, args.GlobalGrant),
		},
	}

	var resp struct {
		Data struct {
			CreateInsightsDashboard struct {
				Dashboard DashboardResponse `json:"dashboard"`
			} `json:"createInsightsDashboard"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return DashboardResponse{}, errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.CreateInsightsDashboard.Dashboard, nil
}

func (c *Client) GetDashboards(args GetDashboardArgs) ([]DashboardResponse, error) {
	const query = `
		query CreateInsightsDashboard($first: Int, $after: String, $id: ID) {
			insightsDashboards(first: $first, after: $after, id: $id) {
				nodes {
					id, title
				}
			}
		}
	`

	variables := map[string]any{}
	if args.First != nil {
		variables["first"] = args.First
	}
	if args.After != nil {
		variables["after"] = args.After
	}
	if args.Id != nil {
		variables["id"] = args.Id
	}

	var resp struct {
		Data struct {
			InsightsDashboards struct {
				Nodes []DashboardResponse
			} `json:"insightsDashboards"`
		} `json:"data"`
	}

	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return []DashboardResponse{}, errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.InsightsDashboards.Nodes, nil
}

func (c *Client) UpdateDashboard(id string, args DashboardInputArgs) (DashboardResponse, error) {
	const query = `
		mutation UpdateInsightsDashboard($id: ID!, $input: UpdateInsightsDashboardInput!) {
			updateInsightsDashboard(id: $id, input: $input) {
				dashboard {
					id, title, grants { users, organizations, global }
				}
			}
		}
	`

	variables := map[string]any{
		"id": id,
		"input": map[string]any{
			"title":  args.Title,
			"grants": buildGrants(args.UserGrant, args.OrgGrant, args.GlobalGrant),
		},
	}

	var resp struct {
		Data struct {
			UpdateInsightsDashboard struct {
				Dashboard DashboardResponse `json:"dashboard"`
			} `json:"updateInsightsDashboard"`
		} `json:"data"`
	}

	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return DashboardResponse{}, errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.UpdateInsightsDashboard.Dashboard, nil
}

func (c *Client) DeleteDashboard(id string) error {
	const query = `
		mutation DeleteInsightsDashboard($id: ID!) {
			deleteInsightsDashboard(id: $id) {
				alwaysNil
			}
		}
	`
	variables := map[string]any{
		"id": id,
	}
	var resp struct {
		Data struct {
			DeleteInsightsDashboard struct {
				AlwaysNil string `json:"alwaysNil"`
			} `json:"deleteInsightsDashboard"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return errors.Wrap(err, "request GraphQL")
	}
	return nil
}

func buildGrants(userGrant string, orgGrant string, globalGrant bool) map[string]any {
	userGrants := []string{}
	orgGrants := []string{}

	if userGrant != "" {
		userGrants = []string{userGrant}
	}
	if orgGrant != "" {
		orgGrants = []string{orgGrant}
	}

	return map[string]any{
		"users":         userGrants,
		"organizations": orgGrants,
		"global":        globalGrant,
	}
}

func (c *Client) CreateSearchInsight(title string, series map[string]any, viewRepoScope map[string]any, viewTimeScope map[string]any) (SearchInsight, error) {
	const query = `
	mutation CreateSearchBasedInsight($input: LineChartSearchInsightInput!) {
		createLineChartSearchInsight(input: $input) {
		  view {
			id
			presentation {
			  ... on LineChartInsightViewPresentation {
				seriesPresentation {
				  seriesId
				  label
				  color
				}
			  }
			}
			repositoryDefinition {
			  ... on InsightRepositoryScope {
				repositories
			  }
			}
			timeScope {
			  ... on InsightIntervalTimeScope {
				unit
				value
			  }
			}
		  }
		}
	  }
	`

	variables := map[string]any{
		"input": map[string]any{
			"options": map[string]any{
				"title": title,
			},
			"dataSeries": []any{
				series,
			},
			"repositoryScope": viewRepoScope,
			"timeScope":       viewTimeScope,
		},
	}
	var resp struct {
		Data struct {
			CreateLineChartSearchInsight struct {
				View struct {
					Id                   string `json:"id"`
					RepositoryDefinition struct {
						Repositories []string `json:"Repositories"`
					} `json:"repositoryDefinition"`
					TimeScope struct {
						Unit  string `json:"Unit"`
						Value int32  `json:"Value"`
					} `json:"timeScope"`
					Presentation struct {
						SeriesPresentation []struct {
							SeriesId string `json:"SeriesId"`
							Label    string `json:"Label"`
							Color    string `json:"Color"`
						} `json:"SeriesPresentation"`
					} `json:"Presentation"`
				} `json:"View"`
			} `json:"createLineChartSearchInsight"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return SearchInsight{}, errors.Wrap(err, "request GraphQL")
	}
	singleSeries := resp.Data.CreateLineChartSearchInsight.View.Presentation.SeriesPresentation
	if len(singleSeries) != 1 {
		return SearchInsight{}, errors.Newf("Received %d insight series when expecting 1: resp [%v]", len(singleSeries), resp)
	}
	return SearchInsight{
		InsightViewId: resp.Data.CreateLineChartSearchInsight.View.Id,
		SeriesId:      singleSeries[0].SeriesId,
		Label:         singleSeries[0].Label,
		Color:         singleSeries[0].Color,
		Repos:         resp.Data.CreateLineChartSearchInsight.View.RepositoryDefinition.Repositories,
		IntervalUnit:  resp.Data.CreateLineChartSearchInsight.View.TimeScope.Unit,
		IntervalValue: resp.Data.CreateLineChartSearchInsight.View.TimeScope.Value,
	}, nil
}

type SearchInsight struct {
	InsightViewId string
	SeriesId      string
	Label         string
	Color         string
	Repos         []string
	IntervalUnit  string
	IntervalValue int32
	NumSamples    int32
}

func (c *Client) UpdateSearchInsight(insightViewID string, input map[string]any) (SearchInsight, error) {
	const query = `
		mutation UpdateLineChartSearchInsight($input: UpdateLineChartSearchInsightInput!, $id: ID!) {
			updateLineChartSearchInsight(input: $input, id: $id) {
				view {
                  id
				  presentation {
					... on LineChartInsightViewPresentation {
					  seriesPresentation {
						seriesId
						label
						color
					  }
					}
				  }
				  repositoryDefinition {
					... on InsightRepositoryScope {
					  repositories
					}
				  }
				  timeScope {
					... on InsightIntervalTimeScope {
					  unit
					  value
					}
				  }
				  defaultSeriesDisplayOptions {
					numSamples
				  }
				}
			}
		}
	`

	variables := map[string]any{
		"id":    insightViewID,
		"input": input,
	}
	var resp struct {
		Data struct {
			UpdateLineChartSearchInsight struct {
				View struct {
					Id                   string `json:"id"`
					RepositoryDefinition struct {
						InsightRepositoryScope struct {
							Repositories []string `json:"Repositories"`
						} `json:"InsightRepositoryScope"`
					} `json:"RepositoryDefinition"`
					TimeScope struct {
						InsightIntervalTimeScope struct {
							Unit  string `json:"unit"`
							Value int32  `json:"value"`
						} `json:"InsightIntervalTimeScope"`
					} `json:"TimeScope"`
					Presentation struct {
						SeriesPresentation []struct {
							SeriesId string `json:"SeriesId"`
							Label    string `json:"Label"`
							Color    string `json:"Color"`
						} `json:"SeriesPresentation"`
					} `json:"Presentation"`
					DefaultSeriesDisplayOptions struct {
						NumSamples int32 `json:"NumSamples"`
					} `json:"DefaultSeriesDisplayOptions"`
				} `json:"View"`
			} `json:"updateLineChartSearchInsight"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return SearchInsight{}, errors.Wrap(err, "request GraphQL")
	}
	singleSeries := resp.Data.UpdateLineChartSearchInsight.View.Presentation.SeriesPresentation
	if len(singleSeries) != 1 {
		return SearchInsight{}, errors.Newf("Received %d insight series when expecting 1: resp [%v]", len(singleSeries), resp)
	}
	return SearchInsight{
		InsightViewId: resp.Data.UpdateLineChartSearchInsight.View.Id,
		SeriesId:      singleSeries[0].SeriesId,
		Label:         singleSeries[0].Label,
		Color:         singleSeries[0].Color,
		NumSamples:    resp.Data.UpdateLineChartSearchInsight.View.DefaultSeriesDisplayOptions.NumSamples,
	}, nil
}

func (c *Client) SaveInsightAsNewView(input map[string]any) ([]SearchInsight, error) {
	const query = `
		mutation saveAsNewView($new: SaveInsightAsNewViewInput!) {
		  saveInsightAsNewView(input:$new) {
			view {
			  id
			  dataSeries {
				seriesId
			  }
			}
		  }
		}
	`

	variables := map[string]any{
		"new": input,
	}
	var resp struct {
		Data struct {
			SaveInsightAsNewView struct {
				View struct {
					Id         string `json:"id"`
					DataSeries []struct {
						SeriesId string `json:"seriesId"`
					} `json:"dataSeries"`
				} `json:"view"`
			} `json:"saveInsightAsNewView"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "request GraphQL")
	}

	insights := []SearchInsight{}
	for _, series := range resp.Data.SaveInsightAsNewView.View.DataSeries {
		insights = append(insights, SearchInsight{
			InsightViewId: resp.Data.SaveInsightAsNewView.View.Id,
			SeriesId:      series.SeriesId,
		})
	}
	return insights, nil
}

func (c *Client) DeleteInsightView(insightViewId string) error {
	const query = `
		mutation DeleteInsightView ($id: ID!) {
		  deleteInsightView(id: $id){
			alwaysNil
		  }
		}
	`
	variables := map[string]any{
		"id": insightViewId,
	}
	var resp struct {
		Data struct {
			DeleteInsightView struct {
				AlwaysNil string `json:"alwaysNil"`
			} `json:"deleteInsightView"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return errors.Wrap(err, "request GraphQL")
	}
	return nil
}
