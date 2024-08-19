package notionapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"time"
)

const (
	apiURL        = "https://api.notion.com"
	apiVersion    = "v1"
	notionVersion = "2022-06-28"
	maxRetries    = 3
)

type Token string

type errJsonDecodeFunc func(data []byte) error

func (it Token) String() string {
	return string(it)
}

// ClientOption to configure API client
type ClientOption func(*Client)

type Client struct {
	httpClient    *http.Client
	baseUrl       *url.URL
	apiVersion    string
	notionVersion string

	maxRetries int

	Token Token

	// used in Authorization header only for requests that require Basic authentication.
	oauthID     string
	oauthSecret string

	Database       DatabaseService
	Block          BlockService
	Page           PageService
	User           UserService
	Search         SearchService
	Comment        CommentService
	Authentication AuthenticationService
}

func NewClient(token Token, opts ...ClientOption) *Client {
	u, err := url.Parse(apiURL)
	if err != nil {
		panic(err)
	}
	c := &Client{
		httpClient:    http.DefaultClient,
		Token:         token,
		baseUrl:       u,
		apiVersion:    apiVersion,
		notionVersion: notionVersion,
		maxRetries:    maxRetries,
	}

	c.Database = &DatabaseClient{apiClient: c}
	c.Block = &BlockClient{apiClient: c}
	c.Page = &PageClient{apiClient: c}
	c.User = &UserClient{apiClient: c}
	c.Search = &SearchClient{apiClient: c}
	c.Comment = &CommentClient{apiClient: c}
	c.Authentication = &AuthenticationClient{apiClient: c}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// WithHTTPClient overrides the default http.Client
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithVersion overrides the Notion API version
func WithVersion(version string) ClientOption {
	return func(c *Client) {
		c.notionVersion = version
	}
}

// WithRetry overrides the default number of max retry attempts on 429 errors
func WithRetry(retries int) ClientOption {
	return func(c *Client) {
		c.maxRetries = retries
	}
}

// WithOAuthAppCredentials sets the OAuth app ID and secret to use when fetching a token from Notion.
func WithOAuthAppCredentials(id, secret string) ClientOption {
	return func(c *Client) {
		c.oauthID = id
		c.oauthSecret = secret
	}
}

func (c *Client) request(ctx context.Context, method string, urlStr string, queryParams map[string]string, requestBody interface{}) (*http.Response, error) {
	return c.requestImpl(ctx, method, urlStr, queryParams, requestBody, false, decodeClientError)
}

func (c *Client) requestImpl(ctx context.Context, method string, urlStr string, queryParams map[string]string, requestBody interface{}, basicAuth bool, errDecoder errJsonDecodeFunc) (*http.Response, error) {
	u, err := c.baseUrl.Parse(fmt.Sprintf("%s/%s", c.apiVersion, urlStr))
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if requestBody != nil && !reflect.ValueOf(requestBody).IsNil() {
		body, err := json.Marshal(requestBody)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewBuffer(body)
	}

	if len(queryParams) > 0 {
		q := u.Query()
		for k, v := range queryParams {
			q.Add(k, v)
		}
		u.RawQuery = q.Encode()
	}
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	if basicAuth {
		cred := base64.StdEncoding.EncodeToString([]byte(c.oauthID + ":" + c.oauthSecret))
		req.Header.Add("Authorization", fmt.Sprintf("Basic %s", cred))
	} else {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.Token.String()))
	}
	req.Header.Add("Notion-Version", c.notionVersion)
	req.Header.Add("Content-Type", "application/json")

	failedAttempts := 0
	var res *http.Response
	for {
		var err error
		res, err = c.httpClient.Do(req.WithContext(ctx))
		if err != nil {
			return nil, err
		}

		if res.StatusCode != http.StatusTooManyRequests {
			break
		}

		failedAttempts++
		if failedAttempts == c.maxRetries {
			return nil, &RateLimitedError{Message: fmt.Sprintf("Retry request with 429 response failed after %d retries", failedAttempts)}
		}
		// https://developers.notion.com/reference/request-limits#rate-limits
		retryAfterHeader := res.Header["Retry-After"]
		if len(retryAfterHeader) == 0 {
			return nil, &RateLimitedError{Message: "Retry-After header missing from Notion API response headers for 429 response"}
		}
		retryAfter := retryAfterHeader[0]

		waitSeconds, err := strconv.Atoi(retryAfter)
		if err != nil {
			break // should not happen
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(waitSeconds) * time.Second):
		}
	}

	if res.StatusCode != http.StatusOK {
		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		return nil, errDecoder(data)
	}

	return res, nil
}

func decodeClientError(data []byte) error {
	var apiErr Error
	err := json.Unmarshal(data, &apiErr)
	if err != nil {
		return err
	}
	return &apiErr
}

type Pagination struct {
	StartCursor Cursor
	PageSize    int
}

func (p *Pagination) ToQuery() map[string]string {
	if p == nil {
		return nil
	}
	r := map[string]string{}
	if p.StartCursor != "" {
		r["start_cursor"] = p.StartCursor.String()
	}

	if p.PageSize != 0 {
		r["page_size"] = strconv.Itoa(p.PageSize)
	}

	return r
}
