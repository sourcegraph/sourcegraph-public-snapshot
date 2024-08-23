package sdk

/*
   Copyright 2016 Alexander I.Grafov <grafov@gmail.com>
   Copyright 2016-2022 The Grafana SDK authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

	   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.

   ॐ तारे तुत्तारे तुरे स्व
*/

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// SearchTeams search teams with optional parameters.
// Reflects GET /api/teams/search API call.
func (r *Client) SearchTeams(ctx context.Context, params ...SearchTeamParams) (PageTeams, error) {
	var (
		raw       []byte
		pageTeams PageTeams
		code      int
		err       error
		requestParams = make(url.Values)
	)

	for _, p := range params {
		p(requestParams)
	}

	if raw, code, err = r.get(ctx, "api/teams/search", requestParams); err != nil {
		return pageTeams, err
	}
	if code != 200 {
		return pageTeams, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&pageTeams); err != nil {
		return pageTeams, fmt.Errorf("unmarshal teams: %s\n%s", err, raw)
	}
	return pageTeams, err
}

func (r *Client) GetTeamByName(ctx context.Context, name string) (Team, error) {
	var (
		team Team
		err  error
	)
	search, err := r.SearchTeams(ctx, WithTeam(name))
	if err != nil {
		return team, err
	}
	if len(search.Teams) == 0 {
		return Team{}, TeamNotFound
	}

	return search.Teams[0], nil
}

// GetTeam gets an team by ID.
// Reflects GET /api/teams/:id API call.
func (r *Client) GetTeam(ctx context.Context, id uint) (Team, error) {
	var (
		raw  []byte
		team Team
		code int
		err  error
	)
	if raw, code, err = r.get(ctx, fmt.Sprintf("api/teams/%d", id), nil); err != nil {
		return team, err
	}
	if code != 200 {
		return team, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&team); err != nil {
		return team, fmt.Errorf("unmarshal team: %s\n%s", err, raw)
	}
	return team, err
}

// CreateTeam creates a new team.
// Reflects POST /api/teams API call.
func (r *Client) CreateTeam(ctx context.Context, t Team) (StatusMessage, error) {
	var (
		raw  []byte
		resp StatusMessage
		err  error
	)
	if raw, err = json.Marshal(t); err != nil {
		return StatusMessage{}, err
	}
	if raw, _, err = r.post(ctx, "api/teams", nil, raw); err != nil {
		return StatusMessage{}, err
	}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return StatusMessage{}, err
	}
	return resp, nil
}

// UpdateTeam updates a team.
// Reflects PUT /api/teams/:id API call.
func (r *Client) UpdateTeam(ctx context.Context, id uint, t Team) (StatusMessage, error) {
	var (
		raw  []byte
		resp StatusMessage
		err  error
	)
	if raw, err = json.Marshal(t); err != nil {
		return StatusMessage{}, err
	}
	if raw, _, err = r.put(ctx, fmt.Sprintf("api/teams/%d", id), nil, raw); err != nil {
		return StatusMessage{}, err
	}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return StatusMessage{}, err
	}
	return resp, nil
}

// DeleteTeam deletes a team.
// Reflects DELETE /api/teams/:id API call.
func (r *Client) DeleteTeam(ctx context.Context, id uint) (StatusMessage, error) {
	var (
		raw  []byte
		resp StatusMessage
		err  error
	)
	if raw, _, err = r.delete(ctx, fmt.Sprintf("api/teams/%d", id)); err != nil {
		return StatusMessage{}, err
	}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return StatusMessage{}, err
	}
	return resp, nil
}

// GetTeamMembers gets the members of a team by id.
// Reflects GET /api/teams/:teamId/members API call.
func (r *Client) GetTeamMembers(ctx context.Context, teamId uint) ([]TeamMember, error) {
	var (
		raw         []byte
		teamMembers []TeamMember
		code        int
		err         error
	)
	if raw, code, err = r.get(ctx, fmt.Sprintf("api/teams/%d/members", teamId), nil); err != nil {
		return teamMembers, err
	}
	if code != 200 {
		return teamMembers, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&teamMembers); err != nil {
		return teamMembers, fmt.Errorf("unmarshal team: %s\n%s", err, raw)
	}
	return teamMembers, err
}

// AddTeamMember adds a member to a team.
// Reflects POST /api/teams/:teamId/members API call.
func (r *Client) AddTeamMember(ctx context.Context, teamId uint, userId uint) (StatusMessage, error) {
	var (
		raw  []byte
		resp StatusMessage
		err  error
	)
	if raw, err = json.Marshal(struct {
		UserId uint `json:"userId"`
	}{
		UserId: userId,
	}); err != nil {
		return StatusMessage{}, err
	}
	if raw, _, err = r.post(ctx, fmt.Sprintf("api/teams/%d/members", teamId), nil, raw); err != nil {
		return StatusMessage{}, err
	}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return StatusMessage{}, err
	}
	return resp, nil
}

// DeleteTeamMember removes a ream member from a team by id.
// Reflects DELETE /api/teams/:teamId/:userId API call.
func (r *Client) DeleteTeamMember(ctx context.Context, teamId uint, userId uint) (StatusMessage, error) {
	var (
		raw  []byte
		resp StatusMessage
		err  error
	)
	if raw, _, err = r.delete(ctx, fmt.Sprintf("api/teams/%d/members/%d", teamId, userId)); err != nil {
		return StatusMessage{}, err
	}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return StatusMessage{}, err
	}
	return resp, nil
}

// GetTeamPreferences gets the preferences for a team by id.
// Reflects GET /api/teams/:teamId/preferences API call.
func (r *Client) GetTeamPreferences(ctx context.Context, teamId uint) (TeamPreferences, error) {
	var (
		raw             []byte
		teamPreferences TeamPreferences
		code            int
		err             error
	)
	if raw, code, err = r.get(ctx, fmt.Sprintf("api/teams/%d/preferences", teamId), nil); err != nil {
		return teamPreferences, err
	}
	if code != 200 {
		return teamPreferences, fmt.Errorf("HTTP error %d: returns %s", code, raw)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&teamPreferences); err != nil {
		return teamPreferences, fmt.Errorf("unmarshal team: %s\n%s", err, raw)
	}
	return teamPreferences, err
}

// UpdateTeamPreferences updates the preferences for a team by id.
// Reflects PUT /api/teams/:teamId/preferences API call.
func (r *Client) UpdateTeamPreferences(ctx context.Context, teamId uint, tp TeamPreferences) (StatusMessage, error) {
	var (
		raw  []byte
		resp StatusMessage
		err  error
	)
	if raw, err = json.Marshal(tp); err != nil {
		return StatusMessage{}, err
	}
	if raw, _, err = r.put(ctx, fmt.Sprintf("api/teams/%d/preferences", teamId), nil, raw); err != nil {
		return StatusMessage{}, err
	}
	if err = json.Unmarshal(raw, &resp); err != nil {
		return StatusMessage{}, err
	}
	return resp, nil
}

// TeamNotFound is an error returned if the given team was not found.
var TeamNotFound = fmt.Errorf("team not found")

// SearchTeamParams is the type for all options implementing query parameters
// perpage optional. default 1000
// page optional. default 1
// http://docs.grafana.org/http_api/team/#search-teams
// http://docs.grafana.org/http_api/team/#search-teams-with-paging
type SearchTeamParams func(values url.Values)

// WithQuery adds a query parameter
func WithQuery(query string) SearchTeamParams {
	return func(v url.Values) {
		v.Set("query", query)
	}
}

// WithPagesize adds a page size query parameter
func WithPagesize(size uint) SearchTeamParams {
	return func(v url.Values) {
		v.Set("perpage", strconv.FormatUint(uint64(size),10))
	}
}

// WithPage adds a page number query parameter
func WithPage(page uint) SearchTeamParams {
	return func(v url.Values) {
		v.Set("page", strconv.FormatUint(uint64(page),10))
	}
}

// WithTeam adds a query parameter
func WithTeam(team string) SearchTeamParams {
	return func(v url.Values) {
		v.Set("team", team)
	}
}