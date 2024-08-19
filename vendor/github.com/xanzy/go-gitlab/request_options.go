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
	"context"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

// RequestOptionFunc can be passed to all API requests to customize the API request.
type RequestOptionFunc func(*retryablehttp.Request) error

// WithContext runs the request with the provided context
func WithContext(ctx context.Context) RequestOptionFunc {
	return func(req *retryablehttp.Request) error {
		*req = *req.WithContext(ctx)
		return nil
	}
}

// WithHeader takes a header name and value and appends it to the request headers.
func WithHeader(name, value string) RequestOptionFunc {
	return func(req *retryablehttp.Request) error {
		req.Header.Set(name, value)
		return nil
	}
}

// WithHeaders takes a map of header name/value pairs and appends them to the
// request headers.
func WithHeaders(headers map[string]string) RequestOptionFunc {
	return func(req *retryablehttp.Request) error {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		return nil
	}
}

// WithSudo takes either a username or user ID and sets the SUDO request header.
func WithSudo(uid interface{}) RequestOptionFunc {
	return func(req *retryablehttp.Request) error {
		user, err := parseID(uid)
		if err != nil {
			return err
		}
		req.Header.Set("SUDO", user)
		return nil
	}
}

// WithToken takes a token which is then used when making this one request.
func WithToken(authType AuthType, token string) RequestOptionFunc {
	return func(req *retryablehttp.Request) error {
		switch authType {
		case JobToken:
			req.Header.Set("JOB-TOKEN", token)
		case OAuthToken:
			req.Header.Set("Authorization", "Bearer "+token)
		case PrivateToken:
			req.Header.Set("PRIVATE-TOKEN", token)
		}
		return nil
	}
}
