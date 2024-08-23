package gitlab

import (
	"fmt"
	"net/http"
	"time"
)

// ProjectFeatureFlagService handles operations on gitlab project feature
// flags using the following api:
//
// GitLab API docs: https://docs.gitlab.com/ee/api/feature_flags.html
type ProjectFeatureFlagService struct {
	client *Client
}

// ProjectFeatureFlag represents a GitLab project iteration.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/feature_flags.html
type ProjectFeatureFlag struct {
	Name        string                        `json:"name"`
	Description string                        `json:"description"`
	Active      bool                          `json:"active"`
	Version     string                        `json:"version"`
	CreatedAt   *time.Time                    `json:"created_at"`
	UpdatedAt   *time.Time                    `json:"updated_at"`
	Scopes      []*ProjectFeatureFlagScope    `json:"scopes"`
	Strategies  []*ProjectFeatureFlagStrategy `json:"strategies"`
}

// ProjectFeatureFlagScope defines the scopes of a feature flag
//
// GitLab API docs: https://docs.gitlab.com/ee/api/feature_flags.html
type ProjectFeatureFlagScope struct {
	ID               int    `json:"id"`
	EnvironmentScope string `json:"environment_scope"`
}

// ProjectFeatureFlagStrategy defines the strategy used for a feature flag
//
// GitLab API docs: https://docs.gitlab.com/ee/api/feature_flags.html
type ProjectFeatureFlagStrategy struct {
	ID         int                                  `json:"id"`
	Name       string                               `json:"name"`
	Parameters *ProjectFeatureFlagStrategyParameter `json:"parameters"`
	Scopes     []*ProjectFeatureFlagScope           `json:"scopes"`
}

// ProjectFeatureFlagStrategyParameter is used in updating and creating feature flags
//
// GitLab API docs: https://docs.gitlab.com/ee/api/feature_flags.html
type ProjectFeatureFlagStrategyParameter struct {
	GroupID    string `json:"groupId,omitempty"`
	UserIDs    string `json:"userIds,omitempty"`
	Percentage string `json:"percentage,omitempty"`
}

func (i ProjectFeatureFlag) String() string {
	return Stringify(i)
}

// ListProjectFeatureFlagOptions contains the options for ListProjectFeatureFlags
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/feature_flags.html#list-feature-flags-for-a-project
type ListProjectFeatureFlagOptions struct {
	ListOptions
	Scope *string `url:"scope,omitempty" json:"scope,omitempty"`
}

// ListProjectFeatureFlags returns a list with the feature flags of a project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/feature_flags.html#list-feature-flags-for-a-project
func (s *ProjectFeatureFlagService) ListProjectFeatureFlags(pid interface{}, opt *ListProjectFeatureFlagOptions, options ...RequestOptionFunc) ([]*ProjectFeatureFlag, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/feature_flags", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var pffs []*ProjectFeatureFlag
	resp, err := s.client.Do(req, &pffs)
	if err != nil {
		return nil, resp, err
	}

	return pffs, resp, nil
}

// GetProjectFeatureFlag gets a single feature flag for the specified project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/feature_flags.html#get-a-single-feature-flag
func (s *ProjectFeatureFlagService) GetProjectFeatureFlag(pid interface{}, name string, options ...RequestOptionFunc) (*ProjectFeatureFlag, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/feature_flags/%s", PathEscape(project), name)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	flag := new(ProjectFeatureFlag)
	resp, err := s.client.Do(req, flag)
	if err != nil {
		return nil, resp, err
	}

	return flag, resp, nil
}

// CreateProjectFeatureFlagOptions represents the available
// CreateProjectFeatureFlag() options.
//
// Gitlab API docs:
// https://docs.gitlab.com/ee/api/feature_flags.html#create-a-feature-flag
type CreateProjectFeatureFlagOptions struct {
	Name        *string                        `url:"name,omitempty" json:"name,omitempty"`
	Description *string                        `url:"description,omitempty" json:"description,omitempty"`
	Version     *string                        `url:"version,omitempty" json:"version,omitempty"`
	Active      *bool                          `url:"active,omitempty" json:"active,omitempty"`
	Strategies  *[]*FeatureFlagStrategyOptions `url:"strategies,omitempty" json:"strategies,omitempty"`
}

// FeatureFlagStrategyOptions represents the available feature flag strategy
// options.
//
// Gitlab API docs:
// https://docs.gitlab.com/ee/api/feature_flags.html#create-a-feature-flag
type FeatureFlagStrategyOptions struct {
	ID         *int                                 `url:"id,omitempty" json:"id,omitempty"`
	Name       *string                              `url:"name,omitempty" json:"name,omitempty"`
	Parameters *ProjectFeatureFlagStrategyParameter `url:"parameters,omitempty" json:"parameters,omitempty"`
	Scopes     *[]*ProjectFeatureFlagScope          `url:"scopes,omitempty" json:"scopes,omitempty"`
}

// ProjectFeatureFlagScopeOptions represents the available feature flag scope
// options.
//
// Gitlab API docs:
// https://docs.gitlab.com/ee/api/feature_flags.html#create-a-feature-flag
type ProjectFeatureFlagScopeOptions struct {
	ID               *int    `url:"id,omitempty" json:"id,omitempty"`
	EnvironmentScope *string `url:"id,omitempty" json:"environment_scope,omitempty"`
}

// CreateProjectFeatureFlag creates a feature flag
//
// Gitlab API docs:
// https://docs.gitlab.com/ee/api/feature_flags.html#create-a-feature-flag
func (s *ProjectFeatureFlagService) CreateProjectFeatureFlag(pid interface{}, opt *CreateProjectFeatureFlagOptions, options ...RequestOptionFunc) (*ProjectFeatureFlag, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/feature_flags",
		PathEscape(project),
	)

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	flag := new(ProjectFeatureFlag)
	resp, err := s.client.Do(req, flag)
	if err != nil {
		return flag, resp, err
	}

	return flag, resp, nil
}

// UpdateProjectFeatureFlagOptions represents the available
// UpdateProjectFeatureFlag() options.
//
// Gitlab API docs:
// https://docs.gitlab.com/ee/api/feature_flags.html#update-a-feature-flag
type UpdateProjectFeatureFlagOptions struct {
	Name        *string                        `url:"name,omitempty" json:"name,omitempty"`
	Description *string                        `url:"description,omitempty" json:"description,omitempty"`
	Active      *bool                          `url:"active,omitempty" json:"active,omitempty"`
	Strategies  *[]*FeatureFlagStrategyOptions `url:"strategies,omitempty" json:"strategies,omitempty"`
}

// UpdateProjectFeatureFlag updates a feature flag
//
// Gitlab API docs:
// https://docs.gitlab.com/ee/api/feature_flags.html#update-a-feature-flag
func (s *ProjectFeatureFlagService) UpdateProjectFeatureFlag(pid interface{}, name string, opt *UpdateProjectFeatureFlagOptions, options ...RequestOptionFunc) (*ProjectFeatureFlag, *Response, error) {
	group, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/feature_flags/%s",
		PathEscape(group),
		name,
	)

	req, err := s.client.NewRequest(http.MethodPut, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	flag := new(ProjectFeatureFlag)
	resp, err := s.client.Do(req, flag)
	if err != nil {
		return flag, resp, err
	}

	return flag, resp, nil
}

// DeleteProjectFeatureFlag deletes a feature flag
//
// Gitlab API docs:
// https://docs.gitlab.com/ee/api/feature_flags.html#delete-a-feature-flag
func (s *ProjectFeatureFlagService) DeleteProjectFeatureFlag(pid interface{}, name string, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/feature_flags/%s", PathEscape(project), name)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
