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

type PagesService struct {
	client *Client
}

// UnpublishPages unpublished pages. The user must have admin privileges.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/pages.html#unpublish-pages
func (s *PagesService) UnpublishPages(gid interface{}, options ...RequestOptionFunc) (*Response, error) {
	page, err := parseID(gid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/pages", PathEscape(page))

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
