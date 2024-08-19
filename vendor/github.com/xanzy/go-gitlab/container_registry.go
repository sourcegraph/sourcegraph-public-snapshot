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

// ContainerRegistryService handles communication with the container registry
// related methods of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/container_registry.html
type ContainerRegistryService struct {
	client *Client
}

// RegistryRepository represents a GitLab content registry repository.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/container_registry.html
type RegistryRepository struct {
	ID                     int                      `json:"id"`
	Name                   string                   `json:"name"`
	Path                   string                   `json:"path"`
	ProjectID              int                      `json:"project_id"`
	Location               string                   `json:"location"`
	CreatedAt              *time.Time               `json:"created_at"`
	CleanupPolicyStartedAt *time.Time               `json:"cleanup_policy_started_at"`
	TagsCount              int                      `json:"tags_count"`
	Tags                   []*RegistryRepositoryTag `json:"tags"`
}

func (s RegistryRepository) String() string {
	return Stringify(s)
}

// RegistryRepositoryTag represents a GitLab registry image tag.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/container_registry.html
type RegistryRepositoryTag struct {
	Name          string     `json:"name"`
	Path          string     `json:"path"`
	Location      string     `json:"location"`
	Revision      string     `json:"revision"`
	ShortRevision string     `json:"short_revision"`
	Digest        string     `json:"digest"`
	CreatedAt     *time.Time `json:"created_at"`
	TotalSize     int        `json:"total_size"`
}

func (s RegistryRepositoryTag) String() string {
	return Stringify(s)
}

// ListRegistryRepositoriesOptions represents the available
// ListRegistryRepositories() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/container_registry.html#list-registry-repositories
type ListRegistryRepositoriesOptions struct {
	ListOptions

	// Deprecated: These options are deprecated for ListGroupRegistryRepositories calls. (Removed in GitLab 15.0)
	Tags      *bool `url:"tags,omitempty" json:"tags,omitempty"`
	TagsCount *bool `url:"tags_count,omitempty" json:"tags_count,omitempty"`
}

// ListProjectRegistryRepositories gets a list of registry repositories in a project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/container_registry.html#within-a-project
func (s *ContainerRegistryService) ListProjectRegistryRepositories(pid interface{}, opt *ListRegistryRepositoriesOptions, options ...RequestOptionFunc) ([]*RegistryRepository, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/registry/repositories", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var repos []*RegistryRepository
	resp, err := s.client.Do(req, &repos)
	if err != nil {
		return nil, resp, err
	}

	return repos, resp, nil
}

// ListGroupRegistryRepositories gets a list of registry repositories in a group.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/container_registry.html#within-a-group
func (s *ContainerRegistryService) ListGroupRegistryRepositories(gid interface{}, opt *ListRegistryRepositoriesOptions, options ...RequestOptionFunc) ([]*RegistryRepository, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/registry/repositories", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var repos []*RegistryRepository
	resp, err := s.client.Do(req, &repos)
	if err != nil {
		return nil, resp, err
	}

	return repos, resp, nil
}

// GetSingleRegistryRepositoryOptions represents the available
// GetSingleRegistryRepository() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/container_registry.html#get-details-of-a-single-repository
type GetSingleRegistryRepositoryOptions struct {
	Tags      *bool `url:"tags,omitempty" json:"tags,omitempty"`
	TagsCount *bool `url:"tags_count,omitempty" json:"tags_count,omitempty"`
}

// GetSingleRegistryRepository gets the details of single registry repository.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/container_registry.html#get-details-of-a-single-repository
func (s *ContainerRegistryService) GetSingleRegistryRepository(pid interface{}, opt *GetSingleRegistryRepositoryOptions, options ...RequestOptionFunc) (*RegistryRepository, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("registry/repositories/%s", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	repo := new(RegistryRepository)
	resp, err := s.client.Do(req, repo)
	if err != nil {
		return nil, resp, err
	}

	return repo, resp, nil
}

// DeleteRegistryRepository deletes a repository in a registry.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/container_registry.html#delete-registry-repository
func (s *ContainerRegistryService) DeleteRegistryRepository(pid interface{}, repository int, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/registry/repositories/%d", PathEscape(project), repository)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// ListRegistryRepositoryTagsOptions represents the available
// ListRegistryRepositoryTags() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/container_registry.html#list-registry-repository-tags
type ListRegistryRepositoryTagsOptions ListOptions

// ListRegistryRepositoryTags gets a list of tags for given registry repository.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/container_registry.html#list-registry-repository-tags
func (s *ContainerRegistryService) ListRegistryRepositoryTags(pid interface{}, repository int, opt *ListRegistryRepositoryTagsOptions, options ...RequestOptionFunc) ([]*RegistryRepositoryTag, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/registry/repositories/%d/tags",
		PathEscape(project),
		repository,
	)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var tags []*RegistryRepositoryTag
	resp, err := s.client.Do(req, &tags)
	if err != nil {
		return nil, resp, err
	}

	return tags, resp, nil
}

// GetRegistryRepositoryTagDetail get details of a registry repository tag
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/container_registry.html#get-details-of-a-registry-repository-tag
func (s *ContainerRegistryService) GetRegistryRepositoryTagDetail(pid interface{}, repository int, tagName string, options ...RequestOptionFunc) (*RegistryRepositoryTag, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/registry/repositories/%d/tags/%s",
		PathEscape(project),
		repository,
		tagName,
	)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	tag := new(RegistryRepositoryTag)
	resp, err := s.client.Do(req, &tag)
	if err != nil {
		return nil, resp, err
	}

	return tag, resp, nil
}

// DeleteRegistryRepositoryTag deletes a registry repository tag.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/container_registry.html#delete-a-registry-repository-tag
func (s *ContainerRegistryService) DeleteRegistryRepositoryTag(pid interface{}, repository int, tagName string, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/registry/repositories/%d/tags/%s",
		PathEscape(project),
		repository,
		tagName,
	)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// DeleteRegistryRepositoryTagsOptions represents the available
// DeleteRegistryRepositoryTags() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/container_registry.html#delete-registry-repository-tags-in-bulk
type DeleteRegistryRepositoryTagsOptions struct {
	NameRegexpDelete *string `url:"name_regex_delete,omitempty" json:"name_regex_delete,omitempty"`
	NameRegexpKeep   *string `url:"name_regex_keep,omitempty" json:"name_regex_keep,omitempty"`
	KeepN            *int    `url:"keep_n,omitempty" json:"keep_n,omitempty"`
	OlderThan        *string `url:"older_than,omitempty" json:"older_than,omitempty"`

	// Deprecated: NameRegexp is deprecated in favor of NameRegexpDelete.
	NameRegexp *string `url:"name_regex,omitempty" json:"name_regex,omitempty"`
}

// DeleteRegistryRepositoryTags deletes repository tags in bulk based on
// given criteria.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/container_registry.html#delete-registry-repository-tags-in-bulk
func (s *ContainerRegistryService) DeleteRegistryRepositoryTags(pid interface{}, repository int, opt *DeleteRegistryRepositoryTagsOptions, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/registry/repositories/%d/tags",
		PathEscape(project),
		repository,
	)

	req, err := s.client.NewRequest(http.MethodDelete, u, opt, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
