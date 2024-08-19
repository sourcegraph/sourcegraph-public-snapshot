//
// Copyright 2022, Masahiro Yoshida
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package gitlab

import (
	"fmt"
	"net/http"
	"time"
)

// GroupAccessTokensService handles communication with the
// groups access tokens related methods of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/group_access_tokens.html
type GroupAccessTokensService struct {
	client *Client
}

// GroupAccessToken represents a GitLab group access token.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/group_access_tokens.html
type GroupAccessToken struct {
	ID          int              `json:"id"`
	UserID      int              `json:"user_id"`
	Name        string           `json:"name"`
	Scopes      []string         `json:"scopes"`
	CreatedAt   *time.Time       `json:"created_at"`
	ExpiresAt   *ISOTime         `json:"expires_at"`
	LastUsedAt  *time.Time       `json:"last_used_at"`
	Active      bool             `json:"active"`
	Revoked     bool             `json:"revoked"`
	Token       string           `json:"token"`
	AccessLevel AccessLevelValue `json:"access_level"`
}

func (v GroupAccessToken) String() string {
	return Stringify(v)
}

// ListGroupAccessTokensOptions represents the available options for
// listing variables in a group.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_access_tokens.html#list-group-access-tokens
type ListGroupAccessTokensOptions ListOptions

// ListGroupAccessTokens gets a list of all group access tokens in a group.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_access_tokens.html#list-group-access-tokens
func (s *GroupAccessTokensService) ListGroupAccessTokens(gid interface{}, opt *ListGroupAccessTokensOptions, options ...RequestOptionFunc) ([]*GroupAccessToken, *Response, error) {
	groups, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/access_tokens", PathEscape(groups))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var gats []*GroupAccessToken
	resp, err := s.client.Do(req, &gats)
	if err != nil {
		return nil, resp, err
	}

	return gats, resp, nil
}

// GetGroupAccessToken gets a single group access tokens in a group.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_access_tokens.html#get-a-group-access-token
func (s *GroupAccessTokensService) GetGroupAccessToken(gid interface{}, id int, options ...RequestOptionFunc) (*GroupAccessToken, *Response, error) {
	groups, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/access_tokens/%d", PathEscape(groups), id)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	gat := new(GroupAccessToken)
	resp, err := s.client.Do(req, &gat)
	if err != nil {
		return nil, resp, err
	}

	return gat, resp, nil
}

// CreateGroupAccessTokenOptions represents the available CreateVariable()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_access_tokens.html#create-a-group-access-token
type CreateGroupAccessTokenOptions struct {
	Name        *string           `url:"name,omitempty" json:"name,omitempty"`
	Scopes      *[]string         `url:"scopes,omitempty" json:"scopes,omitempty"`
	AccessLevel *AccessLevelValue `url:"access_level,omitempty" json:"access_level,omitempty"`
	ExpiresAt   *ISOTime          `url:"expires_at,omitempty" json:"expires_at,omitempty"`
}

// CreateGroupAccessToken creates a new group access token.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_access_tokens.html#create-a-group-access-token
func (s *GroupAccessTokensService) CreateGroupAccessToken(gid interface{}, opt *CreateGroupAccessTokenOptions, options ...RequestOptionFunc) (*GroupAccessToken, *Response, error) {
	groups, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/access_tokens", PathEscape(groups))

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	pat := new(GroupAccessToken)
	resp, err := s.client.Do(req, pat)
	if err != nil {
		return nil, resp, err
	}

	return pat, resp, nil
}

// RevokeGroupAccessToken revokes a group access token.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_access_tokens.html#revoke-a-group-access-token
func (s *GroupAccessTokensService) RevokeGroupAccessToken(gid interface{}, id int, options ...RequestOptionFunc) (*Response, error) {
	groups, err := parseID(gid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("groups/%s/access_tokens/%d", PathEscape(groups), id)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
