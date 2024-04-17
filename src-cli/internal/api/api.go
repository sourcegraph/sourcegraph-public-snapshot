// Package api provides a basic client library for the Sourcegraph API.
package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	ioaux "github.com/jig/teereadcloser"
	"github.com/kballard/go-shellquote"
	"github.com/mattn/go-isatty"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/version"
)

// Client instances provide methods to create API requests.
type Client interface {
	// NewQuery is a convenience method to create a GraphQL request without
	// variables.
	NewQuery(query string) Request

	// NewRequest creates a GraphQL request.
	NewRequest(query string, vars map[string]interface{}) Request

	// NewHTTPRequest creates an http.Request for the Sourcegraph API.
	//
	// path is joined against the API route. For example on Sourcegraph.com this
	// will result the URL: https://sourcegraph.com/.api/path.
	NewHTTPRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error)

	// Do runs an http.Request against the Sourcegraph API.
	Do(req *http.Request) (*http.Response, error)
}

// Request instances represent GraphQL requests.
type Request interface {
	// Do actions the request. Normally, this means that the request is
	// transmitted and the response is unmarshalled into result.
	//
	// If no data was available to be unmarshalled — for example, due to the
	// -get-curl flag being set — then ok will return false.
	Do(ctx context.Context, result interface{}) (ok bool, err error)

	// DoRaw has the same behaviour as Do, with one exception: the result will
	// not be unwrapped, and will include the GraphQL errors. Therefore the
	// structure that is provided as the result should have top level Data and
	// Errors keys for the GraphQL wrapper to be unmarshalled into.
	DoRaw(ctx context.Context, result interface{}) (ok bool, err error)
}

// client is the internal concrete type implementing Client.
type client struct {
	opts       ClientOpts
	httpClient *http.Client
}

// request is the internal concrete type implementing Request.
type request struct {
	client *client
	query  string
	vars   map[string]interface{}
}

// ClientOpts encapsulates the options given to NewClient.
type ClientOpts struct {
	Endpoint          string
	AccessToken       string
	AdditionalHeaders map[string]string

	// Flags are the standard API client flags provided by NewFlags. If nil,
	// default values will be used.
	Flags *Flags

	// Out is the writer that will be used when outputting diagnostics, such as
	// curl commands when -get-curl is enabled.
	Out io.Writer
}

// NewClient creates a new API client.
func NewClient(opts ClientOpts) Client {
	if opts.Out == nil {
		panic("unexpected nil out option")
	}

	flags := opts.Flags
	if flags == nil {
		flags = defaultFlags()
	}

	httpClient := http.DefaultClient
	if flags.insecureSkipVerify != nil && *flags.insecureSkipVerify {
		httpClient = &http.Client{
			Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		}
	}

	return &client{
		opts: ClientOpts{
			Endpoint:          opts.Endpoint,
			AccessToken:       opts.AccessToken,
			AdditionalHeaders: opts.AdditionalHeaders,
			Flags:             flags,
			Out:               opts.Out,
		},
		httpClient: httpClient,
	}
}

func (c *client) NewQuery(query string) Request {
	return c.NewRequest(query, nil)
}

func (c *client) NewRequest(query string, vars map[string]interface{}) Request {
	return &request{
		client: c,
		query:  query,
		vars:   vars,
	}
}

func (c *client) Do(req *http.Request) (*http.Response, error) {
	return c.httpClient.Do(req)
}

func (c *client) NewHTTPRequest(ctx context.Context, method, p string, body io.Reader) (*http.Request, error) {
	req, err := c.createHTTPRequest(ctx, method, p, body)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *client) createHTTPRequest(ctx context.Context, method, p string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(c.opts.Endpoint, "/")+"/"+p, body)
	if err != nil {
		return nil, err
	}
	if c.opts.Flags.UserAgentTelemetry() {
		req.Header.Set("User-Agent", fmt.Sprintf("src-cli/%s %s %s", version.BuildTag, runtime.GOOS, runtime.GOARCH))
	} else {
		req.Header.Set("User-Agent", "src-cli/"+version.BuildTag)
	}
	if c.opts.AccessToken != "" {
		req.Header.Set("Authorization", "token "+c.opts.AccessToken)
	}
	if *c.opts.Flags.trace {
		req.Header.Set("X-Sourcegraph-Should-Trace", "true")
	}
	for k, v := range c.opts.AdditionalHeaders {
		req.Header.Set(k, v)
	}

	return req, nil
}

func (r *request) do(ctx context.Context, result interface{}) (bool, error) {
	if *r.client.opts.Flags.getCurl {
		curl, err := r.curlCmd()
		if err != nil {
			return false, err
		}
		_, err = r.client.opts.Out.Write([]byte(curl + "\n"))
		return false, err
	}

	if *r.client.opts.Flags.dump {
		fmt.Fprintf(r.client.opts.Out, "<-- query:\n%s\n\n", r.query)
		if len(r.vars) > 0 {
			fmt.Fprintln(r.client.opts.Out, "<-- variables:")
			for k, v := range r.vars {
				value, err := json.Marshal(v)
				if err != nil {
					return false, err
				}
				fmt.Fprintf(r.client.opts.Out, "    %s: %s\n", k, string(value))
			}
			fmt.Fprintln(r.client.opts.Out, "")
		}
	}

	// Create the JSON object.
	reqBody, err := json.Marshal(map[string]interface{}{
		"query":     r.query,
		"variables": r.vars,
	})
	if err != nil {
		return false, err
	}

	var bufBody io.Reader = bytes.NewBuffer(reqBody)
	bufBody = gzipReader(bufBody)

	// Create the HTTP request.
	req, err := r.client.NewHTTPRequest(ctx, "POST", ".api/graphql", bufBody)
	if err != nil {
		return false, err
	}

	// Use gzip compression.
	req.Header.Set("Content-Encoding", "gzip")

	// Perform the request.
	resp, err := r.client.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Check trace header before we potentially early exit
	if *r.client.opts.Flags.trace {
		_, err := r.client.opts.Out.Write([]byte(fmt.Sprintf("x-trace: %s\n", resp.Header.Get("x-trace"))))
		if err != nil {
			return false, err
		}
	}

	// Our request may have failed before reaching the GraphQL endpoint, so
	// confirm the status code. You can test this easily with e.g. an invalid
	// endpoint like -endpoint=https://google.com
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized && isatty.IsCygwinTerminal(os.Stdout.Fd()) {
			fmt.Println("You may need to specify or update your access token to use this endpoint.")
			fmt.Println("See https://github.com/sourcegraph/src-cli#readme")
			fmt.Println("")
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return false, err
		}
		return false, errors.Newf("error: %s\n\n%s", resp.Status, body)
	}

	body := resp.Body
	if *r.client.opts.Flags.dump {
		var buf bytes.Buffer
		body = ioaux.TeeReadCloser(resp.Body, &buf)
		defer func() {
			var out bytes.Buffer
			_ = json.Indent(&out, buf.Bytes(), "    ", "    ")
			fmt.Fprintf(r.client.opts.Out, "--> %s\n\n", out.String())
		}()
	}

	// Decode the response.
	if err := json.NewDecoder(body).Decode(result); err != nil {
		return false, err
	}

	return true, nil
}

// Do executes the request. Successful requests will be unmarshalled into the
// given result. If GraphQL errors are returned, then the returned error will be
// an instance of GraphQlErrors. Other errors (such as HTTP or network errors)
// will be returned as-is.
func (r *request) Do(ctx context.Context, result interface{}) (bool, error) {
	raw := rawResult{Data: result}
	ok, err := r.do(ctx, &raw)
	if err != nil {
		return false, err
	} else if !ok {
		return false, nil
	}

	// Handle the case of unpacking errors.
	if raw.Errors != nil {
		errs := GraphQlErrors{}
		for _, err := range raw.Errors {
			errs = append(errs, &GraphQlError{err})
		}
		return false, errs
	}
	return true, nil
}

func (r *request) DoRaw(ctx context.Context, result interface{}) (bool, error) {
	return r.do(ctx, result)
}

type rawResult struct {
	Data   interface{}   `json:"data,omitempty"`
	Errors []interface{} `json:"errors,omitempty"`
}

func (r *request) curlCmd() (string, error) {
	data, err := json.Marshal(map[string]interface{}{
		"query":     r.query,
		"variables": r.vars,
	})
	if err != nil {
		return "", err
	}

	s := "curl \\\n"
	if r.client.opts.AccessToken != "" {
		s += fmt.Sprintf("   %s \\\n", shellquote.Join("-H", "Authorization: token "+r.client.opts.AccessToken))
	}
	for k, v := range r.client.opts.AdditionalHeaders {
		s += fmt.Sprintf("   %s \\\n", shellquote.Join("-H", k+": "+v))
	}
	s += fmt.Sprintf("   %s \\\n", shellquote.Join("-d", string(data)))
	s += fmt.Sprintf("   %s", shellquote.Join(r.client.opts.Endpoint+"/.api/graphql"))
	return s, nil
}
