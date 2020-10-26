package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/visitor"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

// V4Client is a GitHub API client.
type V4Client struct {
	// apiURL is the base URL of a GitHub API. It must point to the base URL of the GitHub API. This
	// is https://api.github.com for GitHub.com and http[s]://[github-enterprise-hostname]/api for
	// GitHub Enterprise.
	apiURL *url.URL

	// githubDotCom is true if this client connects to github.com.
	githubDotCom bool

	// token is the personal access token used to authenticate requests. May be empty, in which case
	// the default behavior is to make unauthenticated requests.
	// 🚨 SECURITY: Should not be changed after client creation to prevent unauthorized access to the
	// repository cache. Use `WithToken` to create a new client with a different token instead.
	token string

	// httpClient is the HTTP client used to make requests to the GitHub API.
	httpClient httpcli.Doer

	// repoCache is the repository cache associated with the token.
	repoCache *rcache.Cache

	// rateLimitMonitor is the API rate limit monitor.
	rateLimitMonitor *ratelimit.Monitor

	// rateLimit is our self imposed rate limiter
	rateLimit *rate.Limiter
}

// NewV4Client creates a new GitHub API client with an optional default personal access token.
//
// apiURL must point to the base URL of the GitHub API. See the docstring for V4Client.apiURL.
func NewV4Client(apiURL *url.URL, token string, cli httpcli.Doer) *V4Client {
	apiURL = canonicalizedURL(apiURL)
	if gitHubDisable {
		cli = disabledClient{}
	}
	if cli == nil {
		cli = httpcli.ExternalDoer()
	}

	cli = requestCounter.Doer(cli, func(u *url.URL) string {
		// The first component of the Path mostly maps to the type of API
		// request we are making. See `curl https://api.github.com` for the
		// exact mapping
		var category string
		if parts := strings.SplitN(u.Path, "/", 3); len(parts) > 1 {
			category = parts[1]
		}
		return category
	})

	rl := ratelimit.DefaultRegistry.GetOrSet(apiURL.String(), rate.NewLimiter(2000, 10))
	rlm := ratelimit.DefaultMonitorRegistry.GetOrSet(apiURL.String(), token, &ratelimit.Monitor{HeaderPrefix: "X-"})

	return &V4Client{
		apiURL:           apiURL,
		githubDotCom:     urlIsGitHubDotCom(apiURL),
		token:            token,
		httpClient:       cli,
		rateLimitMonitor: rlm,
		repoCache:        newRepoCache(apiURL, token),
		rateLimit:        rl,
	}
}

// WithToken returns a copy of the Client authenticated as the GitHub user with the given token.
func (c *V4Client) WithToken(token string) *V4Client {
	return NewV4Client(c.apiURL, token, c.httpClient)
}

func (c *V4Client) do(ctx context.Context, req *http.Request, result interface{}) (err error) {
	req.URL.Path = path.Join(c.apiURL.Path, req.URL.Path)
	req.URL = c.apiURL.ResolveReference(req.URL)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	if c.token != "" {
		req.Header.Set("Authorization", "bearer "+c.token)
	}

	var resp *http.Response

	span, ctx := ot.StartSpanFromContext(ctx, "GitHub")
	span.SetTag("URL", req.URL.String())
	defer func() {
		if err != nil {
			span.SetTag("error", err.Error())
		}
		if resp != nil {
			span.SetTag("status", resp.Status)
		}
		span.Finish()
	}()

	resp, err = c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	c.rateLimitMonitor.Update(resp.Header)
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		var err APIError
		if body, readErr := ioutil.ReadAll(io.LimitReader(resp.Body, 1<<13)); readErr != nil { // 8kb
			err.Message = fmt.Sprintf("failed to read error response from GitHub API: %v: %q", readErr, string(body))
		} else if decErr := json.Unmarshal(body, &err); decErr != nil {
			err.Message = fmt.Sprintf("failed to decode error response from GitHub API: %v: %q", decErr, string(body))
		}
		err.URL = req.URL.String()
		err.Code = resp.StatusCode
		return &err
	}
	return json.NewDecoder(resp.Body).Decode(result)
}

func (c *V4Client) requestGraphQL(ctx context.Context, query string, vars map[string]interface{}, result interface{}) (err error) {
	reqBody, err := json.Marshal(struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}{
		Query:     query,
		Variables: vars,
	})
	if err != nil {
		return err
	}

	// GitHub.com GraphQL endpoint is api.github.com/graphql. GitHub Enterprise is /api/graphql (the
	// REST endpoint is /api/v3, necessitating the "..").
	graphqlEndpoint := "/graphql"
	if !c.githubDotCom {
		graphqlEndpoint = "../graphql"
	}
	req, err := http.NewRequest("POST", graphqlEndpoint, bytes.NewReader(reqBody))
	if err != nil {
		return err
	}

	// Enable Checks API
	// https://developer.github.com/v4/previews/#checks
	req.Header.Add("Accept", "application/vnd.github.antiope-preview+json")
	var respBody struct {
		Data   json.RawMessage `json:"data"`
		Errors graphqlErrors   `json:"errors"`
	}

	cost, err := estimateGraphQLCost(query)
	if err != nil {
		return errors.Wrap(err, "estimating graphql cost")
	}

	if err := c.rateLimitMonitor.SleepRecommendedTimeForBackgroundOp(ctx, c.rateLimit, cost); err != nil {
		return errors.Wrap(err, "rate limit monitor")
	}

	if err := c.do(ctx, req, &respBody); err != nil {
		return err
	}

	// If the GraphQL response has errors, still attempt to unmarshal the data portion, as some
	// requests may expect errors but have useful responses (e.g., querying a list of repositories,
	// some of which you expect to 404).
	if len(respBody.Errors) > 0 {
		err = respBody.Errors
	}
	if result != nil && respBody.Data != nil {
		if err0 := unmarshal(respBody.Data, result); err0 != nil && err == nil {
			return err0
		}
	}
	return err
}

// RateLimitMonitor exposes the rate limit monitor
func (c *V4Client) RateLimitMonitor() *ratelimit.Monitor {
	return c.rateLimitMonitor
}

// estimateGraphQLCost estimates the cost of the query as described here:
// https://developer.github.com/v4/guides/resource-limitations/#calculating-a-rate-limit-score-before-running-the-call
func estimateGraphQLCost(query string) (int, error) {
	doc, err := parser.Parse(parser.ParseParams{
		Source: query,
	})
	if err != nil {
		return 0, errors.Wrap(err, "parsing query")
	}

	var totalCost int
	for _, def := range doc.Definitions {
		cost := calcDefinitionCost(def)
		totalCost += cost
	}

	// As per the calculation spec, cost should be divided by 100
	totalCost = totalCost / 100
	if totalCost < 1 {
		return 1, nil
	}
	return totalCost, nil
}

type limitDepth struct {
	// The 'first' or 'last' limit
	limit int
	// The depth at which it was added
	depth int
}

func calcDefinitionCost(def ast.Node) int {
	var cost int
	limitStack := make([]limitDepth, 0)

	v := &visitor.VisitorOptions{
		Enter: func(p visitor.VisitFuncParams) (string, interface{}) {
			switch node := p.Node.(type) {
			case *ast.IntValue:
				// We're looking for a 'first' or 'last' param indicating a limit
				parent, ok := p.Parent.(*ast.Argument)
				if !ok {
					return visitor.ActionNoChange, nil
				}
				if parent.Name == nil {
					return visitor.ActionNoChange, nil
				}
				if parent.Name.Value != "first" && parent.Name.Value != "last" {
					return visitor.ActionNoChange, nil
				}

				// Prune anything above our current depth as we may have started walking
				// back down the tree
				currentDepth := len(p.Ancestors)
				limitStack = filterInPlace(limitStack, currentDepth)

				limit, err := strconv.Atoi(node.Value)
				if err != nil {
					return "", errors.Wrap(err, "parsing limit")
				}
				limitStack = append(limitStack, limitDepth{limit: limit, depth: currentDepth})
				// The first item in the tree is always worth 1
				if len(limitStack) == 1 {
					cost++
					return visitor.ActionNoChange, nil
				}
				// The cost of the current item is calculated using the limits of
				// its children
				children := limitStack[:len(limitStack)-1]
				product := 1
				// Multiply them all together
				for _, n := range children {
					product = n.limit * product
				}
				cost += product
			}
			return visitor.ActionNoChange, nil
		},
	}

	_ = visitor.Visit(def, v, nil)

	return cost
}

func filterInPlace(limitStack []limitDepth, depth int) []limitDepth {
	n := 0
	for _, x := range limitStack {
		if depth > x.depth {
			limitStack[n] = x
			n++
		}
	}
	limitStack = limitStack[:n]
	return limitStack
}

// graphqlErrors describes the errors in a GraphQL response. It contains at least 1 element when returned by
// requestGraphQL. See https://graphql.github.io/graphql-spec/June2018/#sec-Errors.
type graphqlErrors []struct {
	Message   string        `json:"message"`
	Type      string        `json:"type"`
	Path      []interface{} `json:"path"`
	Locations []struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"locations,omitempty"`
}

const graphqlErrTypeNotFound = "NOT_FOUND"

func (e graphqlErrors) Error() string {
	return fmt.Sprintf("error in GraphQL response: %s", e[0].Message)
}

// ErrPullRequestAlreadyExists is when the requested GitHub Pull Request already exists.
var ErrPullRequestAlreadyExists = errors.New("GitHub pull request already exists")

// ErrPullRequestNotFound is when the requested GitHub Pull Request doesn't exist.
type ErrPullRequestNotFound int

func (e ErrPullRequestNotFound) Error() string {
	return fmt.Sprintf("GitHub pull request not found: %d", e)
}
