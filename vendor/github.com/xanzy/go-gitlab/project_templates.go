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
)

// ProjectTemplatesService handles communication with the project templates
// related methods of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/project_templates.html
type ProjectTemplatesService struct {
	client *Client
}

// ProjectTemplate represents a GitLab ProjectTemplate.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/project_templates.html
type ProjectTemplate struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Nickname    string   `json:"nickname"`
	Popular     bool     `json:"popular"`
	HTMLURL     string   `json:"html_url"`
	SourceURL   string   `json:"source_url"`
	Description string   `json:"description"`
	Conditions  []string `json:"conditions"`
	Permissions []string `json:"permissions"`
	Limitations []string `json:"limitations"`
	Content     string   `json:"content"`
}

func (s ProjectTemplate) String() string {
	return Stringify(s)
}

// ListProjectTemplatesOptions represents the available ListSnippets() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/project_templates.html#get-all-templates-of-a-particular-type
type ListProjectTemplatesOptions struct {
	ListOptions
	ID   *int    `url:"id,omitempty" json:"id,omitempty"`
	Type *string `url:"type,omitempty" json:"type,omitempty"`
}

// ListTemplates gets a list of project templates.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/project_templates.html#get-all-templates-of-a-particular-type
func (s *ProjectTemplatesService) ListTemplates(pid interface{}, templateType string, opt *ListProjectTemplatesOptions, options ...RequestOptionFunc) ([]*ProjectTemplate, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/templates/%s", PathEscape(project), templateType)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var pt []*ProjectTemplate
	resp, err := s.client.Do(req, &pt)
	if err != nil {
		return nil, resp, err
	}

	return pt, resp, nil
}

// GetProjectTemplate gets a single project template.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/project_templates.html#get-one-template-of-a-particular-type
func (s *ProjectTemplatesService) GetProjectTemplate(pid interface{}, templateType string, templateName string, options ...RequestOptionFunc) (*ProjectTemplate, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/templates/%s/%s", PathEscape(project), templateType, templateName)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	ptd := new(ProjectTemplate)
	resp, err := s.client.Do(req, ptd)
	if err != nil {
		return nil, resp, err
	}

	return ptd, resp, nil
}
