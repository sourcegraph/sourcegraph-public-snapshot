// Copyright 2022, Daniela Filipe Bento
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package gitlab

import (
	"fmt"
	"net/http"
)

// DeploymentMergeRequestsService handles communication with the deployment's
// merge requests related methods of the GitLab API.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/deployments.html#list-of-merge-requests-associated-with-a-deployment
type DeploymentMergeRequestsService struct {
	client *Client
}

// ListDeploymentMergeRequests get the merge requests associated with deployment.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/deployments.html#list-of-merge-requests-associated-with-a-deployment
func (s *DeploymentMergeRequestsService) ListDeploymentMergeRequests(pid interface{}, deployment int, opts *ListMergeRequestsOptions, options ...RequestOptionFunc) ([]*MergeRequest, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/deployments/%d/merge_requests", PathEscape(project), deployment)

	req, err := s.client.NewRequest(http.MethodGet, u, opts, options)
	if err != nil {
		return nil, nil, err
	}

	var mrs []*MergeRequest
	resp, err := s.client.Do(req, &mrs)
	if err != nil {
		return nil, resp, err
	}

	return mrs, resp, nil
}
