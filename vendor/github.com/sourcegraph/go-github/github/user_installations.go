// Copyright 2013 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"fmt"
)

type CombinedInstallations struct {
	TotalCount    *int `json:"total_count,omitempty"`
	Installations []*Installation
}

type InstallationRepos struct {
	TotalCount   *int `json:"total_count,omitempty"`
	Repositories []*Repository
}

// ListInstallations lists app installations that are acessible to the current user.
//
// GitHub API docs: https://developer.github.com/v3/apps/#list-installations-for-user
func (s *UsersService) ListInstallations(ctx context.Context, opt *ListOptions) ([]*Installation, *Response, error) {
	u := "user/installations"
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypeIntegrationPreview)

	var installations CombinedInstallations
	resp, err := s.client.Do(ctx, req, &installations)
	if err != nil {
		return nil, resp, err
	}

	return installations.Installations, resp, nil
}

// ListInstallations lists respositories that are accessible to the given installation for the current user.
//
// GitHub API docs: https://developer.github.com/v3/apps/installations/#list-repositories-accessible-to-the-user-for-an-installation
func (s *UsersService) ListInstallationRepos(ctx context.Context, installationID int, opt *ListOptions) ([]*Repository, *Response, error) {
	u := fmt.Sprintf("user/installations/%v/repositories", installationID)
	u, err := addOptions(u, opt)
	if err != nil {
		return nil, nil, err
	}

	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, nil, err
	}

	// TODO: remove custom Accept header when this API fully launches.
	req.Header.Set("Accept", mediaTypeIntegrationPreview)

	var repos InstallationRepos
	resp, err := s.client.Do(ctx, req, &repos)
	if err != nil {
		return nil, resp, err
	}

	return repos.Repositories, resp, nil
}
