//
// Copyright 2022, Mai Lapyst
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

// ResourceMilestoneEventsService handles communication with the event related
// methods of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/resource_milestone_events.html
type ResourceMilestoneEventsService struct {
	client *Client
}

// MilestoneEvent represents a resource milestone event.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/resource_milestone_events.html
type MilestoneEvent struct {
	ID           int        `json:"id"`
	User         *BasicUser `json:"user"`
	CreatedAt    *time.Time `json:"created_at"`
	ResourceType string     `json:"resource_type"`
	ResourceID   int        `json:"resource_id"`
	Milestone    *Milestone `json:"milestone"`
	Action       string     `json:"action"`
}

// ListMilestoneEventsOptions represents the options for all resource state events
// list methods.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/resource_milestone_events.html#list-project-issue-milestone-events
type ListMilestoneEventsOptions struct {
	ListOptions
}

// ListIssueMilestoneEvents retrieves resource milestone events for the specified
// project and issue.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/resource_milestone_events.html#list-project-issue-milestone-events
func (s *ResourceMilestoneEventsService) ListIssueMilestoneEvents(pid interface{}, issue int, opt *ListMilestoneEventsOptions, options ...RequestOptionFunc) ([]*MilestoneEvent, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/issues/%d/resource_milestone_events", PathEscape(project), issue)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var mes []*MilestoneEvent
	resp, err := s.client.Do(req, &mes)
	if err != nil {
		return nil, resp, err
	}

	return mes, resp, nil
}

// GetIssueMilestoneEvent gets a single issue milestone event.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/resource_milestone_events.html#get-single-issue-milestone-event
func (s *ResourceMilestoneEventsService) GetIssueMilestoneEvent(pid interface{}, issue int, event int, options ...RequestOptionFunc) (*MilestoneEvent, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/issues/%d/resource_milestone_events/%d", PathEscape(project), issue, event)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	me := new(MilestoneEvent)
	resp, err := s.client.Do(req, me)
	if err != nil {
		return nil, resp, err
	}

	return me, resp, nil
}

// ListMergeMilestoneEvents retrieves resource milestone events for the specified
// project and merge request.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/resource_milestone_events.html#list-project-merge-request-milestone-events
func (s *ResourceMilestoneEventsService) ListMergeMilestoneEvents(pid interface{}, request int, opt *ListMilestoneEventsOptions, options ...RequestOptionFunc) ([]*MilestoneEvent, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/resource_milestone_events", PathEscape(project), request)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var mes []*MilestoneEvent
	resp, err := s.client.Do(req, &mes)
	if err != nil {
		return nil, resp, err
	}

	return mes, resp, nil
}

// GetMergeRequestMilestoneEvent gets a single merge request milestone event.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/resource_milestone_events.html#get-single-merge-request-milestone-event
func (s *ResourceMilestoneEventsService) GetMergeRequestMilestoneEvent(pid interface{}, request int, event int, options ...RequestOptionFunc) (*MilestoneEvent, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/merge_requests/%d/resource_milestone_events/%d", PathEscape(project), request, event)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	me := new(MilestoneEvent)
	resp, err := s.client.Do(req, me)
	if err != nil {
		return nil, resp, err
	}

	return me, resp, nil
}
