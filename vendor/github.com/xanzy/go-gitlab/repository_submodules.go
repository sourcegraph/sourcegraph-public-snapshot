//
// Copyright 2021, Sander van Harmelen
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

// RepositorySubmodulesService handles communication with the repository
// submodules related methods of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/repository_submodules.html
type RepositorySubmodulesService struct {
	client *Client
}

// SubmoduleCommit represents a GitLab submodule commit.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/repository_submodules.html
type SubmoduleCommit struct {
	ID             string           `json:"id"`
	ShortID        string           `json:"short_id"`
	Title          string           `json:"title"`
	AuthorName     string           `json:"author_name"`
	AuthorEmail    string           `json:"author_email"`
	CommitterName  string           `json:"committer_name"`
	CommitterEmail string           `json:"committer_email"`
	CreatedAt      *time.Time       `json:"created_at"`
	Message        string           `json:"message"`
	ParentIDs      []string         `json:"parent_ids"`
	CommittedDate  *time.Time       `json:"committed_date"`
	AuthoredDate   *time.Time       `json:"authored_date"`
	Status         *BuildStateValue `json:"status"`
}

func (r SubmoduleCommit) String() string {
	return Stringify(r)
}

// UpdateSubmoduleOptions represents the available UpdateSubmodule() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/repository_submodules.html#update-existing-submodule-reference-in-repository
type UpdateSubmoduleOptions struct {
	Branch        *string `url:"branch,omitempty" json:"branch,omitempty"`
	CommitSHA     *string `url:"commit_sha,omitempty" json:"commit_sha,omitempty"`
	CommitMessage *string `url:"commit_message,omitempty" json:"commit_message,omitempty"`
}

// UpdateSubmodule updates an existing submodule reference.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/repository_submodules.html#update-existing-submodule-reference-in-repository
func (s *RepositorySubmodulesService) UpdateSubmodule(pid interface{}, submodule string, opt *UpdateSubmoduleOptions, options ...RequestOptionFunc) (*SubmoduleCommit, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf(
		"projects/%s/repository/submodules/%s",
		PathEscape(project),
		PathEscape(submodule),
	)

	req, err := s.client.NewRequest(http.MethodPut, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	sc := new(SubmoduleCommit)
	resp, err := s.client.Do(req, sc)
	if err != nil {
		return nil, resp, err
	}

	return sc, resp, nil
}
