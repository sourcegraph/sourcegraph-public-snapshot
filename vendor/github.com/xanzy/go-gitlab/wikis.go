//
// Copyright 2021, Stany MARCEL
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

package gitlab

import (
	"fmt"
	"net/http"
	"net/url"
)

// WikisService handles communication with the wikis related methods of
// the Gitlab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/wikis.html
type WikisService struct {
	client *Client
}

// Wiki represents a GitLab wiki.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/wikis.html
type Wiki struct {
	Content  string          `json:"content"`
	Encoding string          `json:"encoding"`
	Format   WikiFormatValue `json:"format"`
	Slug     string          `json:"slug"`
	Title    string          `json:"title"`
}

func (w Wiki) String() string {
	return Stringify(w)
}

// ListWikisOptions represents the available ListWikis options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/wikis.html#list-wiki-pages
type ListWikisOptions struct {
	WithContent *bool `url:"with_content,omitempty" json:"with_content,omitempty"`
}

// ListWikis lists all pages of the wiki of the given project id.
// When with_content is set, it also returns the content of the pages.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/wikis.html#list-wiki-pages
func (s *WikisService) ListWikis(pid interface{}, opt *ListWikisOptions, options ...RequestOptionFunc) ([]*Wiki, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/wikis", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var ws []*Wiki
	resp, err := s.client.Do(req, &ws)
	if err != nil {
		return nil, resp, err
	}

	return ws, resp, nil
}

// GetWikiPageOptions represents options to GetWikiPage
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/wikis.html#get-a-wiki-page
type GetWikiPageOptions struct {
	RenderHTML *bool   `url:"render_html,omitempty" json:"render_html,omitempty"`
	Version    *string `url:"version,omitempty" json:"version,omitempty"`
}

// GetWikiPage gets a wiki page for a given project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/wikis.html#get-a-wiki-page
func (s *WikisService) GetWikiPage(pid interface{}, slug string, opt *GetWikiPageOptions, options ...RequestOptionFunc) (*Wiki, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/wikis/%s", PathEscape(project), url.PathEscape(slug))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	w := new(Wiki)
	resp, err := s.client.Do(req, w)
	if err != nil {
		return nil, resp, err
	}

	return w, resp, nil
}

// CreateWikiPageOptions represents options to CreateWikiPage.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/wikis.html#create-a-new-wiki-page
type CreateWikiPageOptions struct {
	Content *string          `url:"content,omitempty" json:"content,omitempty"`
	Title   *string          `url:"title,omitempty" json:"title,omitempty"`
	Format  *WikiFormatValue `url:"format,omitempty" json:"format,omitempty"`
}

// CreateWikiPage creates a new wiki page for the given repository with
// the given title, slug, and content.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/wikis.html#create-a-new-wiki-page
func (s *WikisService) CreateWikiPage(pid interface{}, opt *CreateWikiPageOptions, options ...RequestOptionFunc) (*Wiki, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/wikis", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPost, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	w := new(Wiki)
	resp, err := s.client.Do(req, w)
	if err != nil {
		return nil, resp, err
	}

	return w, resp, nil
}

// EditWikiPageOptions represents options to EditWikiPage.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/wikis.html#edit-an-existing-wiki-page
type EditWikiPageOptions struct {
	Content *string          `url:"content,omitempty" json:"content,omitempty"`
	Title   *string          `url:"title,omitempty" json:"title,omitempty"`
	Format  *WikiFormatValue `url:"format,omitempty" json:"format,omitempty"`
}

// EditWikiPage Updates an existing wiki page. At least one parameter is
// required to update the wiki page.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/wikis.html#edit-an-existing-wiki-page
func (s *WikisService) EditWikiPage(pid interface{}, slug string, opt *EditWikiPageOptions, options ...RequestOptionFunc) (*Wiki, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/wikis/%s", PathEscape(project), url.PathEscape(slug))

	req, err := s.client.NewRequest(http.MethodPut, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	w := new(Wiki)
	resp, err := s.client.Do(req, w)
	if err != nil {
		return nil, resp, err
	}

	return w, resp, nil
}

// DeleteWikiPage deletes a wiki page with a given slug.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/wikis.html#delete-a-wiki-page
func (s *WikisService) DeleteWikiPage(pid interface{}, slug string, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/wikis/%s", PathEscape(project), url.PathEscape(slug))

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
