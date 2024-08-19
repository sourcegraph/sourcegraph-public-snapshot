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
	"net/http"
	"time"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

// ClientOptionFunc can be used to customize a new GitLab API client.
type ClientOptionFunc func(*Client) error

// WithBaseURL sets the base URL for API requests to a custom endpoint.
func WithBaseURL(urlStr string) ClientOptionFunc {
	return func(c *Client) error {
		return c.setBaseURL(urlStr)
	}
}

// WithCustomBackoff can be used to configure a custom backoff policy.
func WithCustomBackoff(backoff retryablehttp.Backoff) ClientOptionFunc {
	return func(c *Client) error {
		c.client.Backoff = backoff
		return nil
	}
}

// WithCustomLeveledLogger can be used to configure a custom retryablehttp
// leveled logger.
func WithCustomLeveledLogger(leveledLogger retryablehttp.LeveledLogger) ClientOptionFunc {
	return func(c *Client) error {
		c.client.Logger = leveledLogger
		return nil
	}
}

// WithCustomLimiter injects a custom rate limiter to the client.
func WithCustomLimiter(limiter RateLimiter) ClientOptionFunc {
	return func(c *Client) error {
		c.configureLimiterOnce.Do(func() {})
		c.limiter = limiter
		return nil
	}
}

// WithCustomLogger can be used to configure a custom retryablehttp logger.
func WithCustomLogger(logger retryablehttp.Logger) ClientOptionFunc {
	return func(c *Client) error {
		c.client.Logger = logger
		return nil
	}
}

// WithCustomRetry can be used to configure a custom retry policy.
func WithCustomRetry(checkRetry retryablehttp.CheckRetry) ClientOptionFunc {
	return func(c *Client) error {
		c.client.CheckRetry = checkRetry
		return nil
	}
}

// WithCustomRetryMax can be used to configure a custom maximum number of retries.
func WithCustomRetryMax(retryMax int) ClientOptionFunc {
	return func(c *Client) error {
		c.client.RetryMax = retryMax
		return nil
	}
}

// WithCustomRetryWaitMinMax can be used to configure a custom minimum and
// maximum time to wait between retries.
func WithCustomRetryWaitMinMax(waitMin, waitMax time.Duration) ClientOptionFunc {
	return func(c *Client) error {
		c.client.RetryWaitMin = waitMin
		c.client.RetryWaitMax = waitMax
		return nil
	}
}

// WithErrorHandler can be used to configure a custom error handler.
func WithErrorHandler(handler retryablehttp.ErrorHandler) ClientOptionFunc {
	return func(c *Client) error {
		c.client.ErrorHandler = handler
		return nil
	}
}

// WithHTTPClient can be used to configure a custom HTTP client.
func WithHTTPClient(httpClient *http.Client) ClientOptionFunc {
	return func(c *Client) error {
		c.client.HTTPClient = httpClient
		return nil
	}
}

// WithRequestLogHook can be used to configure a custom request log hook.
func WithRequestLogHook(hook retryablehttp.RequestLogHook) ClientOptionFunc {
	return func(c *Client) error {
		c.client.RequestLogHook = hook
		return nil
	}
}

// WithResponseLogHook can be used to configure a custom response log hook.
func WithResponseLogHook(hook retryablehttp.ResponseLogHook) ClientOptionFunc {
	return func(c *Client) error {
		c.client.ResponseLogHook = hook
		return nil
	}
}

// WithoutRetries disables the default retry logic.
func WithoutRetries() ClientOptionFunc {
	return func(c *Client) error {
		c.disableRetries = true
		return nil
	}
}

// WithRequestOptions can be used to configure default request options applied to every request.
func WithRequestOptions(options ...RequestOptionFunc) ClientOptionFunc {
	return func(c *Client) error {
		c.defaultRequestOptions = append(c.defaultRequestOptions, options...)
		return nil
	}
}
