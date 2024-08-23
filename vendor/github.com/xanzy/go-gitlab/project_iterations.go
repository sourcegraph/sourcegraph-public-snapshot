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

// IterationsAPI handles communication with the project iterations related
// methods of the GitLab API
//
// GitLab API docs: https://docs.gitlab.com/ee/api/iterations.html
type ProjectIterationsService struct {
	client *Client
}

// ProjectIteration represents a GitLab project iteration.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/iterations.html
type ProjectIteration struct {
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

func (i ProjectIteration) String() string {
	return Stringify(i)
}

// ListProjectIterationsOptions contains the available ListProjectIterations()
// options
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/iterations.html#list-project-iterations
type ListProjectIterationsOptions struct {
	ListOptions
	State            *string `url:"state,omitempty" json:"state,omitempty"`
	Search           *string `url:"search,omitempty" json:"search,omitempty"`
	IncludeAncestors *bool   `url:"include_ancestors,omitempty" json:"include_ancestors,omitempty"`
}

// ListProjectIterations returns a list of projects iterations.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/iterations.html#list-project-iterations
func (i *ProjectIterationsService) ListProjectIterations(pid interface{}, opt *ListProjectIterationsOptions, options ...RequestOptionFunc) ([]*ProjectIteration, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/iterations", PathEscape(project))

	req, err := i.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var pis []*ProjectIteration
	resp, err := i.client.Do(req, &pis)
	if err != nil {
		return nil, resp, err
	}

	return pis, resp, nil
}
