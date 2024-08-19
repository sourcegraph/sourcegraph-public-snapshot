// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package buildkite

import "fmt"

// TeamService handles communication with the teams related
// methods of the buildkite API.
//
// buildkite API docs: https://buildkite.com/docs/api
type TeamsService struct {
	client *Client
}

// Team represents a buildkite team.
type Team struct {
	ID          *string    `json:"id,omitempty" yaml:"id,omitempty"`
	Name        *string    `json:"name,omitempty" yaml:"name,omitempty"`
	Slug        *string    `json:"slug,omitempty" yaml:"slug,omitempty"`
	Description *string    `json:"description,omitempty" yaml:"description,omitempty"`
	Privacy     *string    `json:"privacy,omitempty" yaml:"privacy,omitempty"`
	Default     *bool      `json:"default,omitempty" yaml:"default,omitempty"`
	CreatedAt   *Timestamp `json:"created_at,omitempty" yaml:"created_at,omitempty"`
	CreatedBy   *User      `json:"created_by,omitempty" yaml:"created_by,omitempty"`
}

// TeamsListOptions specifies the optional parameters to the
// TeamsService.List method.
type TeamsListOptions struct {
	ListOptions
	UserID string `url:"user_id,omitempty"`
}

// Get the teams for an org.
//
// buildkite API docs: https://buildkite.com/docs/api
func (ts *TeamsService) List(org string, opt *TeamsListOptions) ([]Team, *Response, error) {
	var u string

	u = fmt.Sprintf("v2/organizations/%s/teams", org)

	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := ts.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	teams := new([]Team)
	resp, err := ts.client.Do(req, teams)
	if err != nil {
		return nil, resp, err
	}

	return *teams, resp, err
}
