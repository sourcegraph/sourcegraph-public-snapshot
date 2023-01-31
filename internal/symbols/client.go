package symbols

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gobwas/glob"
	"github.com/neelance/parallel"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/go-ctags"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/symbols/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/resetonce"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func defaultEndpoints() *endpoint.Map {
	return endpoint.ConfBased(func(conns conftypes.ServiceConnections) []string {
		return conns.Symbols
	})
}

func LoadConfig() {
	DefaultClient = &Client{
		Endpoints:           defaultEndpoints(),
		HTTPClient:          defaultDoer,
		HTTPLimiter:         parallel.NewRun(500),
		SubRepoPermsChecker: func() authz.SubRepoPermissionChecker { return authz.DefaultSubRepoPermsChecker },
	}
}

// DefaultClient is the default Client. Unless overwritten, it is connected to the server specified by the
// SYMBOLS_URL environment variable.
var DefaultClient *Client

var defaultDoer = func() httpcli.Doer {
	d, err := httpcli.NewInternalClientFactory("symbols").Doer()
	if err != nil {
		panic(err)
	}
	return d
}()

// Client is a symbols service client.
type Client struct {
	// Endpoints to symbols service.
	Endpoints *endpoint.Map

	// HTTP client to use
	HTTPClient httpcli.Doer

	// Limits concurrency of outstanding HTTP posts
	HTTPLimiter *parallel.Run

	// SubRepoPermsChecker is function to return the checker to use. It needs to be a
	// function since we expect the client to be set at runtime once we have a
	// database connection.
	SubRepoPermsChecker func() authz.SubRepoPermissionChecker

	langMappingOnce  resetonce.Once
	langMappingCache map[string][]glob.Glob
}

func (c *Client) url(repo api.RepoName) (string, error) {
	if c.Endpoints == nil {
		return "", errors.New("a symbols service has not been configured")
	}
	return c.Endpoints.Get(string(repo))
}

func (c *Client) ListLanguageMappings(ctx context.Context, repo api.RepoName) (_ map[string][]glob.Glob, err error) {
	c.langMappingOnce.Do(func() {
		var mappings map[string][]string

		if c.isGRPCEnabled(ctx) {
			mappings, err = c.listLanguageMappingsGRPC(ctx, repo)
		} else {
			mappings, err = c.listLanguageMappingsJSON(ctx, repo)
		}

		if err != nil {
			err = errors.Wrap(err, "fetching language mappings")
			return
		}

		globs := make(map[string][]glob.Glob, len(ctags.SupportedLanguages))

		for _, allowedLanguage := range ctags.SupportedLanguages {
			for _, pattern := range mappings[allowedLanguage] {
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

func (c *Client) listLanguageMappingsGRPC(ctx context.Context, repository api.RepoName) (map[string][]string, error) {
	// TODO@ggilmore: This endpoint doesn't need the repository name for anything order than dialing
	// an arbitrary symbols host. We should remove this requirement from this method.

	conn, err := c.dialGRPC(ctx, repository)
	if err != nil {
		return nil, errors.Wrap(err, "dialing symbols service")
	}

	defer conn.Close()

	client := proto.NewSymbolsClient(conn)
	resp, err := client.ListLanguages(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	mappings := make(map[string][]string, len(resp.LanguageFileNameMap))
	for language, fp := range resp.LanguageFileNameMap {
		mappings[language] = fp.Patterns
	}

	return mappings, nil
}

func (c *Client) listLanguageMappingsJSON(ctx context.Context, repository api.RepoName) (map[string][]string, error) {
	// TODO@ggilmore: This endpoint doesn't need the repository name for anything order than dialing
	// an arbitrary symbols host. We should remove this requirement from this method.

	var resp *http.Response
	resp, err := c.httpPost(ctx, "list-languages", repository, nil)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	mapping := make(map[string][]string)
	err = json.NewDecoder(resp.Body).Decode(&mapping)
	return mapping, err
}

// Search performs a symbol search on the symbols service.
func (c *Client) Search(ctx context.Context, args search.SymbolsParameters) (symbols result.Symbols, err error) {
	span, ctx := ot.StartSpanFromContext(ctx, "symbols.Client.Search") //nolint:staticcheck // OT is deprecated
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()
	span.SetTag("Repo", string(args.Repo))
	span.SetTag("CommitID", string(args.CommitID))

	var response search.SymbolsResponse

	if c.isGRPCEnabled(ctx) {
		response, err = c.searchGRPC(ctx, args)
	} else {
		response, err = c.searchJSON(ctx, args)
	}

	if err != nil {
		return nil, errors.Wrap(err, "executing symbols search request")
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

func (c *Client) searchGRPC(ctx context.Context, args search.SymbolsParameters) (search.SymbolsResponse, error) {
	conn, err := c.dialGRPC(ctx, args.Repo)
	if err != nil {
		return search.SymbolsResponse{}, errors.Wrap(err, "dialing GRPC service")
	}

	defer conn.Close()

	grpcClient := proto.NewSymbolsClient(conn)

	var protoArgs proto.SearchRequest
	protoArgs.FromInternal(&args)

	protoResponse, err := grpcClient.Search(ctx, &protoArgs)
	if err != nil {
		return search.SymbolsResponse{}, err
	}

	response := protoResponse.ToInternal()
	return response, nil
}

func (c *Client) searchJSON(ctx context.Context, args search.SymbolsParameters) (search.SymbolsResponse, error) {
	resp, err := c.httpPost(ctx, "search", args.Repo, args)
	if err != nil {
		return search.SymbolsResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// best-effort inclusion of body in error message
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 200))
		return search.SymbolsResponse{}, errors.Errorf(
			"Symbol.Search http status %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	var response search.SymbolsResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return search.SymbolsResponse{}, err
	}
	if response.Err != "" {
		return search.SymbolsResponse{}, errors.New(response.Err)
	}

	return response, nil
}

func (c *Client) LocalCodeIntel(ctx context.Context, args types.RepoCommitPath) (result *types.LocalCodeIntelPayload, err error) {
	span, ctx := ot.StartSpanFromContext(ctx, "squirrel.Client.LocalCodeIntel") //nolint:staticcheck // OT is deprecated
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()
	span.SetTag("Repo", args.Repo)
	span.SetTag("CommitID", args.Commit)

	if c.isGRPCEnabled(ctx) {
		return c.localCodeIntelGRPC(ctx, args)
	}

	return c.localCodeIntelJSON(ctx, args)
}

func (c *Client) localCodeIntelGRPC(ctx context.Context, path types.RepoCommitPath) (result *types.LocalCodeIntelPayload, err error) {
	conn, err := c.dialGRPC(ctx, api.RepoName(path.Repo))
	if err != nil {
		return nil, errors.Wrap(err, "dialing GRPC symbols server endpoint")
	}

	defer conn.Close()

	grpcClient := proto.NewSymbolsClient(conn)

	var rcp proto.RepoCommitPath
	rcp.FromInternal(&path)

	protoArgs := proto.LocalCodeIntelRequest{RepoCommitPath: &rcp}
	protoResponse, err := grpcClient.LocalCodeIntel(ctx, &protoArgs)
	if err != nil {
		return nil, err
	}

	response := protoResponse.ToInternal()
	return &response, nil
}

func (c *Client) localCodeIntelJSON(ctx context.Context, args types.RepoCommitPath) (result *types.LocalCodeIntelPayload, err error) {
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
	span, ctx := ot.StartSpanFromContext(ctx, "squirrel.Client.SymbolInfo") //nolint:staticcheck // OT is deprecated
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()
	span.SetTag("Repo", args.Repo)
	span.SetTag("CommitID", args.Commit)

	if c.isGRPCEnabled(ctx) {
		result, err = c.symbolInfoGRPC(ctx, args)
	} else {
		result, err = c.symbolInfoJSON(ctx, args)
	}

	if err != nil {
		return nil, errors.Wrap(err, "executing symbol info request")
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

func (c *Client) symbolInfoGRPC(ctx context.Context, args types.RepoCommitPathPoint) (result *types.SymbolInfo, err error) {
	conn, err := c.dialGRPC(ctx, api.RepoName(args.Repo))
	if err != nil {
		return nil, errors.Wrap(err, "dialing GRPC symbols server endpoint")
	}

	defer conn.Close()

	client := proto.NewSymbolsClient(conn)

	var rcp proto.RepoCommitPath
	rcp.FromInternal(&args.RepoCommitPath)

	var point proto.Point
	point.FromInternal(&args.Point)

	protoArgs := proto.SymbolInfoRequest{
		RepoCommitPath: &rcp,
		Point:          &point,
	}

	protoResponse, err := client.SymbolInfo(ctx, &protoArgs)
	if err != nil {
		return nil, err
	}

	response := protoResponse.ToInternal()
	return &response, nil
}

func (c *Client) symbolInfoJSON(ctx context.Context, args types.RepoCommitPathPoint) (result *types.SymbolInfo, err error) {
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

	return result, nil
}

func (c *Client) httpPost(
	ctx context.Context,
	method string,
	repo api.RepoName,
	payload any,
) (resp *http.Response, err error) {
	span, ctx := ot.StartSpanFromContext(ctx, "symbols.Client.httpPost") //nolint:staticcheck // OT is deprecated
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()

	repoUrl, err := c.url(repo)
	if err != nil {
		return nil, err
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	if !strings.HasSuffix(repoUrl, "/") {
		repoUrl += "/"
	}
	req, err := http.NewRequest("POST", repoUrl+method, bytes.NewReader(reqBody))
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

// dialGRPC establishes a GRPC connection with the symbols server instance that handles
// the named repository.
func (c *Client) dialGRPC(ctx context.Context, repository api.RepoName) (*grpc.ClientConn, error) {
	rawURL, err := c.url(repository)
	if err != nil {
		return nil, errors.Wrap(err, "getting symbols service URL")
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, errors.Wrap(err, "parsing symbols service URL")
	}

	opts := []grpc.DialOption{
		// 🚨 SECURITY: We use insecure connections to the symbols service. During the
		// grpc prototyping phase - we're leaving TLS authentication out of scope.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	opts = append(opts, defaults.DialOptions()...)
	conn, err := grpc.DialContext(ctx, u.Host, opts...)
	if err != nil {
		return nil, errors.Wrap(err, "dialing symbols GRPC service")
	}

	return conn, nil
}

func (c *Client) isGRPCEnabled(ctx context.Context) bool {
	return featureflag.FromContext(ctx).GetBoolOr("grpc", false)
}
