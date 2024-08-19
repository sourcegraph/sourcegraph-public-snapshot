//
// Copyright 2021, Sune Keller
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
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"
)

// GenericPackagesService handles communication with the packages related
// methods of the GitLab API.
//
// GitLab docs:
// https://docs.gitlab.com/ee/user/packages/generic_packages/index.html
type GenericPackagesService struct {
	client *Client
}

// GenericPackagesFile represents a GitLab generic package file.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/user/packages/generic_packages/index.html#publish-a-package-file
type GenericPackagesFile struct {
	ID        int        `json:"id"`
	PackageID int        `json:"package_id"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	Size      int        `json:"size"`
	FileStore int        `json:"file_store"`
	FileMD5   string     `json:"file_md5"`
	FileSHA1  string     `json:"file_sha1"`
	FileName  string     `json:"file_name"`
	File      struct {
		URL string `json:"url"`
	} `json:"file"`
	FileSHA256             string     `json:"file_sha256"`
	VerificationRetryAt    *time.Time `json:"verification_retry_at"`
	VerifiedAt             *time.Time `json:"verified_at"`
	VerificationFailure    bool       `json:"verification_failure"`
	VerificationRetryCount int        `json:"verification_retry_count"`
	VerificationChecksum   string     `json:"verification_checksum"`
	VerificationState      int        `json:"verification_state"`
	VerificationStartedAt  *time.Time `json:"verification_started_at"`
	NewFilePath            string     `json:"new_file_path"`
}

// FormatPackageURL returns the GitLab Package Registry URL for the given artifact metadata, without the BaseURL.
// This does not make a GitLab API request, but rather computes it based on their documentation.
func (s *GenericPackagesService) FormatPackageURL(pid interface{}, packageName, packageVersion, fileName string) (string, error) {
	project, err := parseID(pid)
	if err != nil {
		return "", err
	}
	u := fmt.Sprintf(
		"projects/%s/packages/generic/%s/%s/%s",
		PathEscape(project),
		PathEscape(packageName),
		PathEscape(packageVersion),
		PathEscape(fileName),
	)
	return u, nil
}

// PublishPackageFileOptions represents the available PublishPackageFile()
// options.
//
// GitLab docs:
// https://docs.gitlab.com/ee/user/packages/generic_packages/index.html#publish-a-package-file
type PublishPackageFileOptions struct {
	Status *GenericPackageStatusValue `url:"status,omitempty" json:"status,omitempty"`
	Select *GenericPackageSelectValue `url:"select,omitempty" json:"select,omitempty"`
}

// PublishPackageFile uploads a file to a project's package registry.
//
// GitLab docs:
// https://docs.gitlab.com/ee/user/packages/generic_packages/index.html#publish-a-package-file
func (s *GenericPackagesService) PublishPackageFile(pid interface{}, packageName, packageVersion, fileName string, content io.Reader, opt *PublishPackageFileOptions, options ...RequestOptionFunc) (*GenericPackagesFile, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf(
		"projects/%s/packages/generic/%s/%s/%s",
		PathEscape(project),
		PathEscape(packageName),
		PathEscape(packageVersion),
		PathEscape(fileName),
	)

	// We need to create the request as a GET request to make sure the options
	// are set correctly. After the request is created we will overwrite both
	// the method and the body.
	req, err := s.client.NewRequest(http.MethodGet, u, opt, options)
	if err != nil {
		return nil, nil, err
	}

	// Overwrite the method and body.
	req.Method = http.MethodPut
	req.SetBody(content)

	f := new(GenericPackagesFile)
	resp, err := s.client.Do(req, f)
	if err != nil {
		return nil, resp, err
	}

	return f, resp, nil
}

// DownloadPackageFile allows you to download the package file.
//
// GitLab docs:
// https://docs.gitlab.com/ee/user/packages/generic_packages/index.html#download-package-file
func (s *GenericPackagesService) DownloadPackageFile(pid interface{}, packageName, packageVersion, fileName string, options ...RequestOptionFunc) ([]byte, *Response, error) {
	project, err := parseID(pid)
	if err != nil {
		return nil, nil, err
	}
	u := fmt.Sprintf(
		"projects/%s/packages/generic/%s/%s/%s",
		PathEscape(project),
		PathEscape(packageName),
		PathEscape(packageVersion),
		PathEscape(fileName),
	)

	req, err := s.client.NewRequest(http.MethodGet, u, nil, options)
	if err != nil {
		return nil, nil, err
	}

	var f bytes.Buffer
	resp, err := s.client.Do(req, &f)
	if err != nil {
		return nil, resp, err
	}

	return f.Bytes(), resp, err
}
