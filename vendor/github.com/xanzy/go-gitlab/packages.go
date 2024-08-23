//
// Copyright 2021, Kordian Bruck
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

// PackagesService handles communication with the packages related methods
// of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/packages.html
type PackagesService struct {
	client *Client
}

// Package represents a GitLab package.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/packages.html
type Package struct {
	ID               int           `json:"id"`
	Name             string        `json:"name"`
	Version          string        `json:"version"`
	PackageType      string        `json:"package_type"`
	Status           string        `json:"status"`
	Links            *PackageLinks `json:"_links"`
	CreatedAt        *time.Time    `json:"created_at"`
	LastDownloadedAt *time.Time    `json:"last_downloaded_at"`
	Tags             []PackageTag  `json:"tags"`
}

func (s Package) String() string {
	return Stringify(s)
}

// GroupPackage represents a GitLab group package.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/packages.html
type GroupPackage struct {
	Package
	ProjectID   int    `json:"project_id"`
	ProjectPath string `json:"project_path"`
}

func (s GroupPackage) String() string {
	return Stringify(s)
}

// PackageLinks holds links for itself and deleting.
type PackageLinks struct {
	WebPath       string `json:"web_path"`
	DeleteAPIPath string `json:"delete_api_path"`
}

func (s PackageLinks) String() string {
	return Stringify(s)
}

// PackageTag holds label information about the package
type PackageTag struct {
	ID        int        `json:"id"`
	PackageID int        `json:"package_id"`
	Name      string     `json:"name"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

func (s PackageTag) String() string {
	return Stringify(s)
}

// PackageFile represents one file contained within a package.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/packages.html
type PackageFile struct {
	ID        int         `json:"id"`
	PackageID int         `json:"package_id"`
	CreatedAt *time.Time  `json:"created_at"`
	FileName  string      `json:"file_name"`
	Size      int         `json:"size"`
	FileMD5   string      `json:"file_md5"`
	FileSHA1  string      `json:"file_sha1"`
	Pipeline  *[]Pipeline `json:"pipelines"`
}

func (s PackageFile) String() string {
	return Stringify(s)
}

// ListProjectPackagesOptions represents the available ListProjectPackages()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/packages.html#within-a-project
type ListProjectPackagesOptions struct {
	ListOptions
	OrderBy            *string `url:"order_by,omitempty" json:"order_by,omitempty"`
	Sort               *string `url:"sort,omitempty" json:"sort,omitempty"`
	PackageType        *string `url:"package_type,omitempty" json:"package_type,omitempty"`
	PackageName        *string `url:"package_name,omitempty" json:"package_name,omitempty"`
	IncludeVersionless *bool   `url:"include_versionless,omitempty" json:"include_versionless,omitempty"`
	Status             *string `url:"status,omitempty" json:"status,omitempty"`
}

// ListProjectPackages gets a list of packages in a project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/packages.html#within-a-project
func (s *PackagesService) ListProjectPackages(pid interface{}, opt *ListProjectPackagesOptions, options ...RequestOptionFunc) ([]*Package, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("projects/%s/packages", PathEscape(project))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var ps []*Package
	resp, err := s.client.Do(req, &ps)
	if err != nil {
		return nil, resp, err
	}

	return ps, resp, nil
}

// ListGroupPackagesOptions represents the available ListGroupPackages()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/packages.html#within-a-group
type ListGroupPackagesOptions struct {
	ListOptions
	ExcludeSubGroups   *bool   `url:"exclude_subgroups,omitempty" json:"exclude_subgroups,omitempty"`
	OrderBy            *string `url:"order_by,omitempty" json:"order_by,omitempty"`
	Sort               *string `url:"sort,omitempty" json:"sort,omitempty"`
	PackageType        *string `url:"package_type,omitempty" json:"package_type,omitempty"`
	PackageName        *string `url:"package_name,omitempty" json:"package_name,omitempty"`
	IncludeVersionless *bool   `url:"include_versionless,omitempty" json:"include_versionless,omitempty"`
	Status             *string `url:"status,omitempty" json:"status,omitempty"`
}

// ListGroupPackages gets a list of packages in a group.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/packages.html#within-a-group
func (s *PackagesService) ListGroupPackages(gid interface{}, opt *ListGroupPackagesOptions, options ...RequestOptionFunc) ([]*GroupPackage, *Response, error) {
	group, err := parseID(gid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf("groups/%s/packages", PathEscape(group))

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var ps []*GroupPackage
	resp, err := s.client.Do(req, &ps)
	if err != nil {
		return nil, resp, err
	}

	return ps, resp, nil
}

// ListPackageFilesOptions represents the available ListPackageFiles()
// options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/packages.html#list-package-files
type ListPackageFilesOptions ListOptions

// ListPackageFiles gets a list of files that are within a package
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/packages.html#list-package-files
func (s *PackagesService) ListPackageFiles(pid interface{}, pkg int, opt *ListPackageFilesOptions, options ...RequestOptionFunc) ([]*PackageFile, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf(
		"projects/%s/packages/%d/package_files",
		PathEscape(project),
		pkg,
	)

	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	var pfs []*PackageFile
	resp, err := s.client.Do(req, &pfs)
	if err != nil {
		return nil, resp, err
	}

	return pfs, resp, nil
}

// DeleteProjectPackage deletes a package in a project.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/packages.html#delete-a-project-package
func (s *PackagesService) DeleteProjectPackage(pid interface{}, pkg int, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/packages/%d", PathEscape(project), pkg)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

// DeletePackageFile deletes a file in project package
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/packages.html#delete-a-package-file
func (s *PackagesService) DeletePackageFile(pid interface{}, pkg, file int, options ...RequestOptionFunc) (*Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("projects/%s/packages/%d/package_files/%d", PathEscape(project), pkg, file)

	req, err := s.client.NewRequest(http.MethodDelete, u, nil, options)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}
