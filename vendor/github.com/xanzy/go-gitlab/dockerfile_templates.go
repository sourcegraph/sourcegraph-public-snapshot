//
// Copyright 2022, FantasyTeddy
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
	"net/url"
)

// DockerfileTemplatesService handles communication with the Dockerfile
// templates related methods of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/templates/dockerfiles.html
type DockerfileTemplatesService struct {
	client *Client
}

// DockerfileTemplate represents a GitLab Dockerfile template.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/templates/dockerfiles.html
type DockerfileTemplate struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// DockerfileTemplateListItem represents a GitLab Dockerfile template from the list.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/templates/dockerfiles.html
type DockerfileTemplateListItem struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// ListDockerfileTemplatesOptions represents the available ListAllTemplates() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/templates/dockerfiles.html#list-dockerfile-templates
type ListDockerfileTemplatesOptions ListOptions

// ListTemplates get a list of available Dockerfile templates.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/templates/dockerfiles.html#list-dockerfile-templates
func (s *DockerfileTemplatesService) ListTemplates(opt *ListDockerfileTemplatesOptions, options ...RequestOptionFunc) ([]*DockerfileTemplateListItem, *Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "templates/dockerfiles", opt, options)
	if err != nil {
		return nil, nil, err
	}

	var gs []*DockerfileTemplateListItem
	resp, err := s.client.Do(req, &gs)
	if err != nil {
		return nil, resp, err
	}

	return gs, resp, nil
}

// GetTemplate get a single Dockerfile template.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/templates/dockerfiles.html#single-dockerfile-template
func (s *DockerfileTemplatesService) GetTemplate(key string, options ...RequestOptionFunc) (*DockerfileTemplate, *Response, error) {
	u := fmt.Sprintf("templates/dockerfiles/%s", url.PathEscape(key))

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	g := new(DockerfileTemplate)
	resp, err := s.client.Do(req, g)
	if err != nil {
		return nil, resp, err
	}

	return g, resp, nil
}
