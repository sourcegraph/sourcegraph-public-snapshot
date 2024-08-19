//
// Copyright 2021, Matthias Simon
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

// ResourceWeightEventsService handles communication with the event related
// methods of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/resource_weight_events.html
type ResourceWeightEventsService struct {
	client *Client
}

// WeightEvent represents a resource weight event.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/resource_weight_events.html
type WeightEvent struct {
	ID           int            `json:"id"`
	User         *BasicUser     `json:"user"`
	CreatedAt    *time.Time     `json:"created_at"`
	ResourceType string         `json:"resource_type"`
	ResourceID   int            `json:"resource_id"`
	State        EventTypeValue `json:"state"`
	IssueID      int            `json:"issue_id"`
	Weight       int            `json:"weight"`
}

// ListWeightEventsOptions represents the options for all resource weight events
// list methods.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/resource_weight_events.html#list-project-issue-weight-events
type ListWeightEventsOptions struct {
	ListOptions
}

// ListIssueWeightEvents retrieves resource weight events for the specified
// project and issue.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/resource_weight_events.html#list-project-issue-weight-events
func (s *ResourceWeightEventsService) ListIssueWeightEvents(pid interface{}, issue int, opt *ListWeightEventsOptions, options ...RequestOptionFunc) ([]*WeightEvent, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/issues/%d/resource_weight_events", PathEscape(project), issue)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var wes []*WeightEvent
	resp, err := s.client.Do(req, &wes)
	if err != nil {
		return nil, resp, err
	}

	return wes, resp, nil
}
