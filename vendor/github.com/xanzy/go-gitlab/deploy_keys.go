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

// DeployKeysService handles communication with the keys related methods
// of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/deploy_keys.html
type DeployKeysService struct {
	client *Client
}

// InstanceDeployKey represents a GitLab deploy key with the associated
// projects it has write access to.
type InstanceDeployKey struct {
	ID                      int                 `json:"id"`
	Title                   string              `json:"title"`
	CreatedAt               *time.Time          `json:"created_at"`
	Key                     string              `json:"key"`
	Fingerprint             string              `json:"fingerprint"`
	ProjectsWithWriteAccess []*DeployKeyProject `json:"projects_with_write_access"`
}

func (k InstanceDeployKey) String() string {
	return Stringify(k)
}

// DeployKeyProject refers to a project an InstanceDeployKey has write access to.
type DeployKeyProject struct {
	ID                int        `json:"id"`
	Description       string     `json:"description"`
	Name              string     `json:"name"`
	NameWithNamespace string     `json:"name_with_namespace"`
	Path              string     `json:"path"`
	PathWithNamespace string     `json:"path_with_namespace"`
	CreatedAt         *time.Time `json:"created_at"`
}

func (k DeployKeyProject) String() string {
	return Stringify(k)
}

// ProjectDeployKey represents a GitLab project deploy key.
type ProjectDeployKey struct {
	ID        int        `json:"id"`
	Title     string     `json:"title"`
	Key       string     `json:"key"`
	CreatedAt *time.Time `json:"created_at"`
	CanPush   bool       `json:"can_push"`
}

func (k ProjectDeployKey) String() string {
	return Stringify(k)
}

// ListProjectDeployKeysOptions represents the available ListAllDeployKeys()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/deploy_keys.html#list-all-deploy-keys
type ListInstanceDeployKeysOptions struct {
	ListOptions
	Public *bool `url:"public,omitempty" json:"public,omitempty"`
}

// ListAllDeployKeys gets a list of all deploy keys
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/deploy_keys.html#list-all-deploy-keys
func (s *DeployKeysService) ListAllDeployKeys(opt *ListInstanceDeployKeysOptions, options ...RequestOptionFunc) ([]*InstanceDeployKey, *Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "deploy_keys", opt, options)
	if err != nil {
		return nil, nil, err
	}

	var ks []*InstanceDeployKey
	resp, err := s.client.Do(req, &ks)
	if err != nil {
		return nil, resp, err
	}

	return ks, resp, nil
}

// ListProjectDeployKeysOptions represents the available ListProjectDeployKeys()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/deploy_keys.html#list-deploy-keys-for-project
type ListProjectDeployKeysOptions ListOptions

// ListProjectDeployKeys gets a list of a project's deploy keys
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/deploy_keys.html#list-deploy-keys-for-project
func (s *DeployKeysService) ListProjectDeployKeys(pid interface{}, opt *ListProjectDeployKeysOptions, options ...RequestOptionFunc) ([]*ProjectDeployKey, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/deploy_keys", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var ks []*ProjectDeployKey
	resp, err := s.client.Do(req, &ks)
	if err != nil {
		return nil, resp, err
	}

	return ks, resp, nil
}

// GetDeployKey gets a single deploy key.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/deploy_keys.html#get-a-single-deploy-key
func (s *DeployKeysService) GetDeployKey(pid interface{}, deployKey int, options ...RequestOptionFunc) (*ProjectDeployKey, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/deploy_keys/%d", PathEscape(project), deployKey)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	k := new(ProjectDeployKey)
	resp, err := s.client.Do(req, k)
	if err != nil {
		return nil, resp, err
	}

	return k, resp, nil
}

// AddDeployKeyOptions represents the available ADDDeployKey() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/deploy_keys.html#add-deploy-key
type AddDeployKeyOptions struct {
	Title   *string `url:"title,omitempty" json:"title,omitempty"`
	Key     *string `url:"key,omitempty" json:"key,omitempty"`
	CanPush *bool   `url:"can_push,omitempty" json:"can_push,omitempty"`
}

// AddDeployKey creates a new deploy key for a project. If deploy key already
// exists in another project - it will be joined to project but only if
// original one was is accessible by same user.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/deploy_keys.html#add-deploy-key
func (s *DeployKeysService) AddDeployKey(pid interface{}, opt *AddDeployKeyOptions, options ...RequestOptionFunc) (*ProjectDeployKey, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/deploy_keys", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	k := new(ProjectDeployKey)
	resp, err := s.client.Do(req, k)
	if err != nil {
		return nil, resp, err
	}

	return k, resp, nil
}

// DeleteDeployKey deletes a deploy key from a project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/deploy_keys.html#delete-deploy-key
func (s *DeployKeysService) DeleteDeployKey(pid interface{}, deployKey int, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/deploy_keys/%d", PathEscape(project), deployKey)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// EnableDeployKey enables a deploy key.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/deploy_keys.html#enable-a-deploy-key
func (s *DeployKeysService) EnableDeployKey(pid interface{}, deployKey int, options ...RequestOptionFunc) (*ProjectDeployKey, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/deploy_keys/%d/enable", PathEscape(project), deployKey)

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	k := new(ProjectDeployKey)
	resp, err := s.client.Do(req, k)
	if err != nil {
		return nil, resp, err
	}

	return k, resp, nil
}

// UpdateDeployKeyOptions represents the available UpdateDeployKey() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/deploy_keys.html#update-deploy-key
type UpdateDeployKeyOptions struct {
	Title   *string `url:"title,omitempty" json:"title,omitempty"`
	CanPush *bool   `url:"can_push,omitempty" json:"can_push,omitempty"`
}

// UpdateDeployKey updates a deploy key for a project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/deploy_keys.html#update-deploy-key
func (s *DeployKeysService) UpdateDeployKey(pid interface{}, deployKey int, opt *UpdateDeployKeyOptions, options ...RequestOptionFunc) (*ProjectDeployKey, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/deploy_keys/%d", PathEscape(project), deployKey)

	req, err := s.client.NewRequest(http.MethodPut, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	k := new(ProjectDeployKey)
	resp, err := s.client.Do(req, k)
	if err != nil {
		return nil, resp, err
	}

	return k, resp, nil
}
