package symbols

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gobwas/glob"
	"github.com/neelance/parallel"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/go-ctags"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/resetonce"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var symbolsURL = env.Get("SYMBOLS_URL", "k8s+http://symbols:3184", "symbols service URL")

var defaultDoer = func() httpcli.Doer {
	d, err := httpcli.NewInternalClientFactory("symbols").Doer()
	if err != nil {
		panic(err)
	}
	return d
}()

// DefaultClient is the default Client. Unless overwritten, it is connected to the server specified by the
// SYMBOLS_URL environment variable.
var DefaultClient = &Client{
	URL:                 symbolsURL,
	HTTPClient:          defaultDoer,
	HTTPLimiter:         parallel.NewRun(500),
	SubRepoPermsChecker: func() authz.SubRepoPermissionChecker { return authz.DefaultSubRepoPermsChecker },
}

// Client is a symbols service client.
type Client struct {
	// URL to symbols service.
	URL string

	// HTTP client to use
	HTTPClient httpcli.Doer

	// Limits concurrency of outstanding HTTP posts
	HTTPLimiter *parallel.Run

	// SubRepoPermsChecker is function to return the checker to use. It needs to be a
	// function since we expect the client to be set at runtime once we have a
	// database connection.
	SubRepoPermsChecker func() authz.SubRepoPermissionChecker

	endpointOnce sync.Once
	endpoint     *endpoint.Map

	langMappingOnce  resetonce.Once
	langMappingCache map[string][]glob.Glob
}

func (c *Client) url(repo api.RepoName) (string, error) {
	c.endpointOnce.Do(func() {
		if len(strings.Fields(c.URL)) == 0 {
			c.endpoint = endpoint.Empty(errors.New("a symbols service has not been configured"))
		} else {
			c.endpoint = endpoint.New(c.URL)
		}
	})
	return c.endpoint.Get(string(repo))
}

func (c *Client) ListLanguageMappings(ctx context.Context, repo api.RepoName) (_ map[string][]glob.Glob, err error) {
	c.langMappingOnce.Do(func() {
		var resp *http.Response
		resp, err = c.httpPost(ctx, "list-languages", repo, nil)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			// best-effort inclusion of body in error message
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
			err = errors.Errorf(
				"Symbol.ListLanguageMappings http status %d: %s",
				resp.StatusCode,
				string(body),
			)
			return
		}

		mapping := make(map[string][]string)
		err = json.NewDecoder(resp.Body).Decode(&mapping)

		globs := make(map[string][]glob.Glob, len(ctags.SupportedLanguages))

		for _, allowedLanguage := range ctags.SupportedLanguages {
			for _, pattern := range mapping[allowedLanguage] {
				var compiled glob.Glob
				compiled, err = glob.Compile(pattern)
				if err != nil {
					return
				}

				globs[allowedLanguage] = append(globs[allowedLanguage], compiled)
			}
		}

		c.langMappingCache = globs
		time.AfterFunc(time.Minute*10, c.langMappingOnce.Reset)
	})

	return c.langMappingCache, nil
}

// Search performs a symbol search on the symbols service.
func (c *Client) Search(ctx context.Context, args search.SymbolsParameters) (symbols result.Symbols, err error) {
	span, ctx := ot.StartSpanFromContext(ctx, "symbols.Client.Search")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()
	span.SetTag("Repo", string(args.Repo))
	span.SetTag("CommitID", string(args.CommitID))

	resp, err := c.httpPost(ctx, "search", args.Repo, args)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// best-effort inclusion of body in error message
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
		return nil, errors.Errorf(
			"Symbol.Search http status %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	var response search.SymbolsResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	if response.Err != "" {
		return nil, errors.New(response.Err)
	}
	symbols = response.Symbols

	// 🚨 SECURITY: We have valid results, so we need to apply sub-repo permissions
	// filtering.
	if c.SubRepoPermsChecker == nil {
		return symbols, err
	}

	checker := c.SubRepoPermsChecker()
	if !authz.SubRepoEnabled(checker) {
		return symbols, err
	}

	a := actor.FromContext(ctx)
	// Filter in place
	filtered := symbols[:0]
	for _, r := range symbols {
		rc := authz.RepoContent{
			Repo: args.Repo,
			Path: r.Path,
		}
		perm, err := authz.ActorPermissions(ctx, checker, a, rc)
		if err != nil {
			return nil, errors.Wrap(err, "checking sub-repo permissions")
		}
		if perm.Include(authz.Read) {
			filtered = append(filtered, r)
		}
	}

	return filtered, nil
}

func (c *Client) LocalCodeIntel(ctx context.Context, args types.RepoCommitPath) (result *types.LocalCodeIntelPayload, err error) {
	span, ctx := ot.StartSpanFromContext(ctx, "squirrel.Client.LocalCodeIntel")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()
	span.SetTag("Repo", args.Repo)
	span.SetTag("CommitID", args.Commit)

	resp, err := c.httpPost(ctx, "localCodeIntel", api.RepoName(args.Repo), args)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// best-effort inclusion of body in error message
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
		return nil, errors.Errorf(
			"Squirrel.LocalCodeIntel http status %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, errors.Wrap(err, "decoding response body")
	}

	return result, nil
}

func (c *Client) SymbolInfo(ctx context.Context, args types.RepoCommitPathPoint) (result *types.SymbolInfo, err error) {
	span, ctx := ot.StartSpanFromContext(ctx, "squirrel.Client.SymbolInfo")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()
	span.SetTag("Repo", args.Repo)
	span.SetTag("CommitID", args.Commit)

	resp, err := c.httpPost(ctx, "symbolInfo", api.RepoName(args.Repo), args)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// best-effort inclusion of body in error message
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
		return nil, errors.Errorf(
			"Squirrel.SymbolInfo http status %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, errors.Wrap(err, "decoding response body")
	}

	// 🚨 SECURITY: We have a valid result, so we need to apply sub-repo permissions filtering.
	if c.SubRepoPermsChecker == nil {
		return result, err
	}

	checker := c.SubRepoPermsChecker()
	if !authz.SubRepoEnabled(checker) {
		return result, err
	}

	a := actor.FromContext(ctx)
	// Filter in place
	rc := authz.RepoContent{
		Repo: api.RepoName(args.Repo),
		Path: args.Path,
	}
	perm, err := authz.ActorPermissions(ctx, checker, a, rc)
	if err != nil {
		return nil, errors.Wrap(err, "checking sub-repo permissions")
	}
	if !perm.Include(authz.Read) {
		return nil, nil
	}

	return result, nil
}

func (c *Client) httpPost(
	ctx context.Context,
	method string,
	repo api.RepoName,
	payload any,
) (resp *http.Response, err error) {
	span, ctx := ot.StartSpanFromContext(ctx, "symbols.Client.httpPost")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()

	url, err := c.url(repo)
	if err != nil {
		return nil, err
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	req, err := http.NewRequest("POST", url+method, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)

	if c.HTTPLimiter != nil {
		span.LogKV("event", "Waiting on HTTP limiter")
		c.HTTPLimiter.Acquire()
		defer c.HTTPLimiter.Release()
		span.LogKV("event", "Acquired HTTP limiter")
	}

	req, ht := nethttp.TraceRequest(span.Tracer(), req,
		nethttp.OperationName("Symbols Client"),
		nethttp.ClientTrace(false))
	defer ht.Finish()

	return c.HTTPClient.Do(req)
}
