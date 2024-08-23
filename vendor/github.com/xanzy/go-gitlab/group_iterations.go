//
// Copyright 2022, Daniel Steinke
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

// IterationsAPI handles communication with the iterations related methods
// of the GitLab API
//
// GitLab API docs: https://docs.gitlab.com/ee/api/group_iterations.html
type GroupIterationsService struct {
	client *Client
}

// GroupInteration represents a GitLab iteration.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/group_iterations.html
type GroupIteration struct {
	ID          int        `json:"id"`
	IID         int        `json:"iid"`
	Sequence    int        `json:"sequence"`
	GroupID     int        `json:"group_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	State       int        `json:"state"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
	DueDate     *ISOTime   `json:"due_date"`
	StartDate   *ISOTime   `json:"start_date"`
	WebURL      string     `json:"web_url"`
}

func (i GroupIteration) String() string {
	return Stringify(i)
}

// ListGroupIterationsOptions contains the available ListGroupIterations()
// options
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_iterations.html#list-group-iterations
type ListGroupIterationsOptions struct {
	ListOptions
	State            *string `url:"state,omitempty" json:"state,omitempty"`
	Search           *string `url:"search,omitempty" json:"search,omitempty"`
	IncludeAncestors *bool   `url:"include_ancestors,omitempty" json:"include_ancestors,omitempty"`
}

// ListGroupIterations returns a list of group iterations.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/group_iterations.html#list-group-iterations
func (s *GroupIterationsService) ListGroupIterations(gid interface{}, opt *ListGroupIterationsOptions, options ...RequestOptionFunc) ([]*GroupIteration, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/iterations", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var gis []*GroupIteration
	resp, err := s.client.Do(req, &gis)
	if err != nil {
		return nil, nil, err
	}

	return gis, resp, nil
}
