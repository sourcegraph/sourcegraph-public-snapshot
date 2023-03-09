package symbols

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gobwas/glob"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/go-ctags"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	internalgrpc "github.com/sourcegraph/sourcegraph/internal/grpc"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/limiter"
	"github.com/sourcegraph/sourcegraph/internal/resetonce"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	proto "github.com/sourcegraph/sourcegraph/internal/symbols/v1"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func LoadConfig() {
	DefaultClient = &Client{
		ConnectionSource:    &connectionSourceFromSiteConfig{},
		HTTPClient:          defaultDoer,
		HTTPLimiter:         limiter.New(500),
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
	// ConnectionSource provides the connection to use for the symbols service for a given repository.
	ConnectionSource ConnectionSource

	// HTTP client to use
	HTTPClient httpcli.Doer

	// Limits concurrency of outstanding HTTP posts
	HTTPLimiter limiter.Limiter

	// SubRepoPermsChecker is function to return the checker to use. It needs to be a
	// function since we expect the client to be set at runtime once we have a
	// database connection.
	SubRepoPermsChecker func() authz.SubRepoPermissionChecker

	langMappingOnce  resetonce.Once
	langMappingCache map[string][]glob.Glob
}

func (c *Client) ListLanguageMappings(ctx context.Context, repo api.RepoName) (_ map[string][]glob.Glob, err error) {
	c.langMappingOnce.Do(func() {
		var mappings map[string][]string

		if internalgrpc.IsGRPCEnabled(ctx) {
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
	// TODO@ggilmore: This address doesn't need the repository name for anything order than dialing
	// an arbitrary symbols host. We should remove this requirement from this method.
	conn, err := c.ConnectionSource.GetConn(string(repository))
	if err != nil {
		return nil, errors.Wrap(err, "getting gRPC connection to symbols server")
	}

	client := proto.NewSymbolsServiceClient(conn)
	resp, err := client.ListLanguages(ctx, &proto.ListLanguagesRequest{})
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
	// TODO@ggilmore: This address doesn't need the repository name for anything order than dialing
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

	if internalgrpc.IsGRPCEnabled(ctx) {
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
	conn, err := c.ConnectionSource.GetConn(string(args.Repo))
	if err != nil {
		return search.SymbolsResponse{}, errors.Wrap(err, "getting gRPC connection to symbols server")
	}

	grpcClient := proto.NewSymbolsServiceClient(conn)

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

	if internalgrpc.IsGRPCEnabled(ctx) {
		return c.localCodeIntelGRPC(ctx, args)
	}

	return c.localCodeIntelJSON(ctx, args)
}

func (c *Client) localCodeIntelGRPC(ctx context.Context, path types.RepoCommitPath) (result *types.LocalCodeIntelPayload, err error) {
	conn, err := c.ConnectionSource.GetConn(path.Repo)
	if err != nil {
		return nil, errors.Wrap(err, "getting gRPC connection to symbols server")
	}

	grpcClient := proto.NewSymbolsServiceClient(conn)

	var rcp proto.RepoCommitPath
	rcp.FromInternal(&path)

	protoArgs := proto.LocalCodeIntelRequest{RepoCommitPath: &rcp}
	protoResponse, err := grpcClient.LocalCodeIntel(ctx, &protoArgs)
	if err != nil {
		if status.Code(err) == codes.Unimplemented {
			// This ignores errors from LocalCodeIntel to match the behavior found here:
			// https://sourcegraph.com/github.com/sourcegraph/sourcegraph@a1631d58604815917096acc3356447c55baebf22/-/blob/cmd/symbols/squirrel/http_handlers.go?L57-57
			//
			// This is weird, and maybe not intentional, but things break if we return an error.
			return nil, nil
		}
		return nil, err
	}

	return protoResponse.ToInternal(), nil
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

	if internalgrpc.IsGRPCEnabled(ctx) {
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
	conn, err := c.ConnectionSource.GetConn(args.Repo)
	if err != nil {
		return nil, errors.Wrap(err, "getting gRPC connection to symbols server")
	}

	client := proto.NewSymbolsServiceClient(conn)

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
		if status.Code(err) == codes.Unimplemented {
			// This ignores unimplemented errors from SymbolInfo to match the behavior here:
			// https://sourcegraph.com/github.com/sourcegraph/sourcegraph@b039aa70fbd155b5b1eddc4b5deede739626a978/-/blob/cmd/symbols/squirrel/http_handlers.go?L114-114
			return nil, nil
		}
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

	symbolsURL, err := c.ConnectionSource.GetAddress(string(repo))
	if err != nil {
		return nil, err
	}

	reqBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	if !strings.HasSuffix(symbolsURL, "/") {
		symbolsURL += "/"
	}
	req, err := http.NewRequest("POST", symbolsURL+method, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)

	span.LogKV("event", "Waiting on HTTP limiter")
	c.HTTPLimiter.Acquire()
	defer c.HTTPLimiter.Release()
	span.LogKV("event", "Acquired HTTP limiter")

	req, ht := nethttp.TraceRequest(span.Tracer(), req,
		nethttp.OperationName("Symbols Client"),
		nethttp.ClientTrace(false))
	defer ht.Finish()

	return c.HTTPClient.Do(req)
}

// ConnectionSource is an interface for getting the address of the symbols service that handles a
// given repository.
type ConnectionSource interface {
	// GetAddress returns the address of the symbols service that handles the named repository.
	GetAddress(key string) (string, error)

	// GetConn returns a gRPC connection to the symbols service that handles the named repository.
	GetConn(key string) (*grpc.ClientConn, error)
}

// connectionSourceFromSiteConfig is a ConnectionSource that gets its configuration from the site
// configuration.
//
// connectionsSourceFromSiteConfig manages the lifecycle of gRPC connections to symbols servers.
// Whenever the set of symbols servers changes, it handles the transition gracefully by dialing new
// gRPC connections to the new servers and closing old connections to the old servers.
//
// This implementation is thread-safe. However, it is possible that the set of symbols servers will change
// between the time that an address or connection is retrieved and the time that it is used. Callers
// should be prepared to handle this common case.
type connectionSourceFromSiteConfig struct {
	initializeOnce sync.Once
	connections    atomic.Pointer[connectionsAndEndpoints]
}

func (s *connectionSourceFromSiteConfig) GetAddress(key string) (string, error) {
	s.ensureInitialized()

	connections := s.connections.Load()
	return connections.Get(key)
}

func (s *connectionSourceFromSiteConfig) GetConn(key string) (conn *grpc.ClientConn, err error) {
	s.ensureInitialized()

	connections := s.connections.Load()
	return connections.GetConn(key)
}

func (s *connectionSourceFromSiteConfig) ensureInitialized() {
	s.initializeOnce.Do(func() {
		// This ensures that the zero value of connectionSourceFromSiteConfig
		// is always valid.
		s.connections.Store(&connectionsAndEndpoints{})

		conf.Watch(func() {
			configuration := conf.Get()
			if configuration == nil {
				return
			}

			nextAddresses := configuration.ServiceConnectionConfig.Symbols
			s.update(nextAddresses)
		})
	})
}

func (s *connectionSourceFromSiteConfig) update(newAddresses []string) {
	oldConnections := s.connections.Load()
	if oldConnections == nil {
		oldConnections = &connectionsAndEndpoints{}
	}

	oldAddresses := maps.Keys(oldConnections.conns)

	sort.Strings(oldAddresses)
	sort.Strings(newAddresses)

	if slices.Equal(oldAddresses, newAddresses) {
		// The addresses are the same, so we don't need
		// to do any work.
		return
	}

	// The addresses are different, so we need to dial new
	// gRPC connections and close the old ones.

	newEndpointMap := endpoint.Static(newAddresses...)
	newGRPCConnections := make(map[string]connAndErr, len(newAddresses))

	for _, address := range newAddresses {
		u, err := url.Parse(address)
		if err != nil {
			newGRPCConnections[address] = connAndErr{dialErr: errors.Wrapf(err, "parsing address %q", address)}
			continue
		}

		conn, err := defaults.Dial(u.Host)
		newGRPCConnections[address] = connAndErr{conn: conn, dialErr: err}
	}

	s.connections.Store(&connectionsAndEndpoints{
		Map:   newEndpointMap,
		conns: newGRPCConnections,
	})

	for _, ce := range oldConnections.conns {
		if ce.dialErr != nil {
			_ = ce.conn.Close()
		}
	}
}

type connectionsAndEndpoints struct {
	*endpoint.Map

	// conns is a map from addresses to their associated gRPC connection
	// and error. The error is non-nil if the connection failed to
	// be established.
	conns map[string]connAndErr
}

func (s *connectionsAndEndpoints) GetConn(key string) (conn *grpc.ClientConn, err error) {
	address, err := s.Map.Get(key)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to lookup address for key %q", key)
	}

	ce, ok := s.conns[address]
	if !ok {
		return nil, errors.Errorf("no GRPC connection entry for address %q", address)
	}

	return ce.conn, ce.dialErr
}

// connAndErr is a gRPC connection and its associated error. The error is
// non-nil if the connection failed to be established.
type connAndErr struct {
	// conn is the gRPC connection.
	conn *grpc.ClientConn

	// dialErr is the error returned by grpc.Dial when establishing the
	// connection.
	dialErr error
}

type testConnectionSource struct {
	address  string
	grpcConn *grpc.ClientConn
}

// NewTestConnectionSource returns a ConnectionSource that always returns the given address
// - suitable for use in tests.
func NewTestConnectionSource(t *testing.T, address string) ConnectionSource {
	t.Helper()

	u, err := url.Parse(address)
	if err != nil {
		t.Fatalf("parsing address %q: %v", address, err)
	}

	conn, err := grpc.Dial(u.Host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dialing gRPC connection to %q: %v", u.Host, err)
	}

	t.Cleanup(func() {
		_ = conn.Close()
	})

	return &testConnectionSource{
		address:  address,
		grpcConn: conn,
	}
}

func (t testConnectionSource) GetAddress(_ string) (string, error) {
	return t.address, nil
}

func (t testConnectionSource) GetConn(_ string) (*grpc.ClientConn, error) {
	return t.grpcConn, nil
}

var (
	_ ConnectionSource = &connectionSourceFromSiteConfig{}
	_                  = &testConnectionSource{}
)
