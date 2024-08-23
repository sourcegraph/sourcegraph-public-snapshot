//
// Copyright 2022, Timo Furrer <tuxtimo@gmail.com>
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

// MetadataService handles communication with the GitLab server instance to
// retrieve its metadata information via the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/metadata.html
type MetadataService struct {
	client *Client
}

// Metadata represents a GitLab instance version.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/metadata.html
type Metadata struct {
	Version  string `json:"version"`
	Revision string `json:"revision"`
	KAS      struct {
		Enabled     bool   `json:"enabled"`
		ExternalURL string `json:"externalUrl"`
		Version     string `json:"version"`
	} `json:"kas"`
	Enterprise bool `json:"enterprise"`
}

func (s Metadata) String() string {
	return Stringify(s)
}

// GetMetadata gets a GitLab server instance meteadata.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/metadata.html
func (s *MetadataService) GetMetadata(options ...RequestOptionFunc) (*Metadata, *Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "metadata", nil, options)
	if err != nil {
		return nil, nil, err
	}

	v := new(Metadata)
	resp, err := s.client.Do(req, v)
	if err != nil {
		return nil, resp, err
	}

	return v, resp, nil
}
