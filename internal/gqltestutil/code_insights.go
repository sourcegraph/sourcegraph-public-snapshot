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
