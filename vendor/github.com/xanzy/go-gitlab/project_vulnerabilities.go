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

// ProjectVulnerabilitiesService handles communication with the projects
// vulnerabilities related methods of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/project_vulnerabilities.html
type ProjectVulnerabilitiesService struct {
	client *Client
}

// Project represents a GitLab project vulnerability.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/project_vulnerabilities.html
type ProjectVulnerability struct {
	AuthorID                int        `json:"author_id"`
	Confidence              string     `json:"confidence"`
	CreatedAt               *time.Time `json:"created_at"`
	Description             string     `json:"description"`
	DismissedAt             *time.Time `json:"dismissed_at"`
	DismissedByID           int        `json:"dismissed_by_id"`
	DueDate                 *time.Time `json:"due_date"`
	Finding                 *Finding   `json:"finding"`
	ID                      int        `json:"id"`
	LastEditedAt            *time.Time `json:"last_edited_at"`
	LastEditedByID          int        `json:"last_edited_by_id"`
	Project                 *Project   `json:"project"`
	ProjectDefaultBranch    string     `json:"project_default_branch"`
	ReportType              string     `json:"report_type"`
	ResolvedAt              *time.Time `json:"resolved_at"`
	ResolvedByID            int        `json:"resolved_by_id"`
	ResolvedOnDefaultBranch bool       `json:"resolved_on_default_branch"`
	Severity                string     `json:"severity"`
	StartDate               *time.Time `json:"start_date"`
	State                   string     `json:"state"`
	Title                   string     `json:"title"`
	UpdatedAt               *time.Time `json:"updated_at"`
	UpdatedByID             int        `json:"updated_by_id"`
}

// Project represents a GitLab project vulnerability finding.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/project_vulnerabilities.html
type Finding struct {
	Confidence          string     `json:"confidence"`
	CreatedAt           *time.Time `json:"created_at"`
	ID                  int        `json:"id"`
	LocationFingerprint string     `json:"location_fingerprint"`
	MetadataVersion     string     `json:"metadata_version"`
	Name                string     `json:"name"`
	PrimaryIdentifierID int        `json:"primary_identifier_id"`
	ProjectFingerprint  string     `json:"project_fingerprint"`
	ProjectID           int        `json:"project_id"`
	RawMetadata         string     `json:"raw_metadata"`
	ReportType          string     `json:"report_type"`
	ScannerID           int        `json:"scanner_id"`
	Severity            string     `json:"severity"`
	UpdatedAt           *time.Time `json:"updated_at"`
	UUID                string     `json:"uuid"`
	VulnerabilityID     int        `json:"vulnerability_id"`
}

// ListProjectVulnerabilitiesOptions represents the available
// ListProjectVulnerabilities() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/project_vulnerabilities.html#list-project-vulnerabilities
type ListProjectVulnerabilitiesOptions struct {
	ListOptions
}

// ListProjectVulnerabilities gets a list of all project vulnerabilities.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/project_vulnerabilities.html#list-project-vulnerabilities
func (s *ProjectVulnerabilitiesService) ListProjectVulnerabilities(pid interface{}, opt *ListProjectVulnerabilitiesOptions, options ...RequestOptionFunc) ([]*ProjectVulnerability, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/vulnerabilities", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var p []*ProjectVulnerability
	resp, err := s.client.Do(req, &p)
	if err != nil {
		return nil, resp, err
	}

	return p, resp, nil
}

// CreateVulnerabilityOptions represents the available CreateVulnerability()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/project_vulnerabilities.html#new-vulnerability
type CreateVulnerabilityOptions struct {
	FindingID *int `url:"finding_id,omitempty" json:"finding_id,omitempty"`
}

// CreateVulnerability creates a new vulnerability on the selected project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/project_vulnerabilities.html#new-vulnerability
func (s *ProjectVulnerabilitiesService) CreateVulnerability(pid interface{}, opt *CreateVulnerabilityOptions, options ...RequestOptionFunc) (*ProjectVulnerability, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/vulnerabilities", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	p := new(ProjectVulnerability)
	resp, err := s.client.Do(req, p)
	if err != nil {
		return nil, resp, err
	}

	return p, resp, nil
}
