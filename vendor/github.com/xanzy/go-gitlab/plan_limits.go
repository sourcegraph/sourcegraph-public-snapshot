//
// Copyright 2021, Igor Varavko
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

import "net/http"

// PlanLimitsService handles communication with the repositories related
// methods of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/plan_limits.html
type PlanLimitsService struct {
	client *Client
}

// PlanLimit represents a GitLab pipeline.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/plan_limits.html
type PlanLimit struct {
	ConanMaxFileSize           int `json:"conan_max_file_size,omitempty"`
	GenericPackagesMaxFileSize int `json:"generic_packages_max_file_size,omitempty"`
	HelmMaxFileSize            int `json:"helm_max_file_size,omitempty"`
	MavenMaxFileSize           int `json:"maven_max_file_size,omitempty"`
	NPMMaxFileSize             int `json:"npm_max_file_size,omitempty"`
	NugetMaxFileSize           int `json:"nuget_max_file_size,omitempty"`
	PyPiMaxFileSize            int `json:"pypi_max_file_size,omitempty"`
	TerraformModuleMaxFileSize int `json:"terraform_module_max_file_size,omitempty"`
}

// GetCurrentPlanLimitsOptions represents the available GetCurrentPlanLimits()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/plan_limits.html#get-current-plan-limits
type GetCurrentPlanLimitsOptions struct {
	PlanName *string `url:"plan_name,omitempty" json:"plan_name,omitempty"`
}

// List the current limits of a plan on the GitLab instance.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/plan_limits.html#get-current-plan-limits
func (s *PlanLimitsService) GetCurrentPlanLimits(opt *GetCurrentPlanLimitsOptions, options ...RequestOptionFunc) (*PlanLimit, *Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "application/plan_limits", opt, options)
	if err != nil {
		return nil, nil, err
	}

	pl := new(PlanLimit)
	resp, err := s.client.Do(req, pl)
	if err != nil {
		return nil, resp, err
	}

	return pl, resp, nil
}

// ChangePlanLimitOptions represents the available ChangePlanLimits() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/plan_limits.html#change-plan-limits
type ChangePlanLimitOptions struct {
	PlanName                   *string `url:"plan_name,omitempty" json:"plan_name,omitempty"`
	ConanMaxFileSize           *int    `url:"conan_max_file_size,omitempty" json:"conan_max_file_size,omitempty"`
	GenericPackagesMaxFileSize *int    `url:"generic_packages_max_file_size,omitempty" json:"generic_packages_max_file_size,omitempty"`
	HelmMaxFileSize            *int    `url:"helm_max_file_size,omitempty" json:"helm_max_file_size,omitempty"`
	MavenMaxFileSize           *int    `url:"maven_max_file_size,omitempty" json:"maven_max_file_size,omitempty"`
	NPMMaxFileSize             *int    `url:"npm_max_file_size,omitempty" json:"npm_max_file_size,omitempty"`
	NugetMaxFileSize           *int    `url:"nuget_max_file_size,omitempty" json:"nuget_max_file_size,omitempty"`
	PyPiMaxFileSize            *int    `url:"pypi_max_file_size,omitempty" json:"pypi_max_file_size,omitempty"`
	TerraformModuleMaxFileSize *int    `url:"terraform_module_max_file_size,omitempty" json:"terraform_module_max_file_size,omitempty"`
}

// ChangePlanLimits modifies the limits of a plan on the GitLab instance.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/plan_limits.html#change-plan-limits
func (s *PlanLimitsService) ChangePlanLimits(opt *ChangePlanLimitOptions, options ...RequestOptionFunc) (*PlanLimit, *Response, error) {
	req, err := s.client.NewRequest(http.MethodPut, "application/plan_limits", opt, options)
	if err != nil {
		return nil, nil, err
	}

	pl := new(PlanLimit)
	resp, err := s.client.Do(req, pl)
	if err != nil {
		return nil, resp, err
	}

	return pl, resp, nil
}
