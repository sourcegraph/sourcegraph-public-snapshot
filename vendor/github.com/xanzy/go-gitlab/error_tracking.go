//
// Copyright 2022, Ryan Glab <ryan.j.glab@gmail.com>
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

// ErrorTrackingService handles communication with the error tracking
// methods of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/error_tracking.html
type ErrorTrackingService struct {
	client *Client
}

// ErrorTrackingClientKey represents an error tracking client key.
//
// GitLab docs:
// https://docs.gitlab.com/ee/api/error_tracking.html#error-tracking-client-keys
type ErrorTrackingClientKey struct {
	ID        int    `json:"id"`
	Active    bool   `json:"active"`
	PublicKey string `json:"public_key"`
	SentryDsn string `json:"sentry_dsn"`
}

func (p ErrorTrackingClientKey) String() string {
	return Stringify(p)
}

// ErrorTrackingSettings represents error tracking settings for a GitLab project.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/error_tracking.html#error-tracking-project-settings
type ErrorTrackingSettings struct {
	Active            bool   `json:"active"`
	ProjectName       string `json:"project_name"`
	SentryExternalURL string `json:"sentry_external_url"`
	APIURL            string `json:"api_url"`
	Integrated        bool   `json:"integrated"`
}

func (p ErrorTrackingSettings) String() string {
	return Stringify(p)
}

// GetErrorTrackingSettings gets error tracking settings.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/error_tracking.html#get-error-tracking-settings
func (s *ErrorTrackingService) GetErrorTrackingSettings(pid interface{}, options ...RequestOptionFunc) (*ErrorTrackingSettings, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/error_tracking/settings", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	ets := new(ErrorTrackingSettings)
	resp, err := s.client.Do(req, ets)
	if err != nil {
		return nil, resp, err
	}

	return ets, resp, nil
}

// EnableDisableErrorTrackingOptions represents the available
// EnableDisableErrorTracking() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/error_tracking.html#enable-or-disable-the-error-tracking-project-settings
type EnableDisableErrorTrackingOptions struct {
	Active     *bool `url:"active,omitempty" json:"active,omitempty"`
	Integrated *bool `url:"integrated,omitempty" json:"integrated,omitempty"`
}

// EnableDisableErrorTracking allows you to enable or disable the error tracking
// settings for a project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/error_tracking.html#enable-or-disable-the-error-tracking-project-settings
func (s *ErrorTrackingService) EnableDisableErrorTracking(pid interface{}, opt *EnableDisableErrorTrackingOptions, options ...RequestOptionFunc) (*ErrorTrackingSettings, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/error_tracking/settings", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPatch, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	ets := new(ErrorTrackingSettings)
	resp, err := s.client.Do(req, &ets)
	if err != nil {
		return nil, resp, err
	}

	return ets, resp, nil
}

// ListClientKeysOptions represents the available ListClientKeys() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/error_tracking.html#list-project-client-keys
type ListClientKeysOptions ListOptions

// ListClientKeys lists error tracking project client keys.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/error_tracking.html#list-project-client-keys
func (s *ErrorTrackingService) ListClientKeys(pid interface{}, opt *ListClientKeysOptions, options ...RequestOptionFunc) ([]*ErrorTrackingClientKey, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/error_tracking/client_keys", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var cks []*ErrorTrackingClientKey
	resp, err := s.client.Do(req, &cks)
	if err != nil {
		return nil, resp, err
	}

	return cks, resp, nil
}

// CreateClientKey creates a new client key for a project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/error_tracking.html#create-a-client-key
func (s *ErrorTrackingService) CreateClientKey(pid interface{}, options ...RequestOptionFunc) (*ErrorTrackingClientKey, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/error_tracking/client_keys", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodPost, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	ck := new(ErrorTrackingClientKey)
	resp, err := s.client.Do(req, ck)
	if err != nil {
		return nil, resp, err
	}

	return ck, resp, nil
}

// DeleteClientKey removes a client key from the project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/error_tracking.html#delete-a-client-key
func (s *ErrorTrackingService) DeleteClientKey(pid interface{}, keyID int, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/error_tracking/client_keys/%d", PathEscape(project), keyID)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
