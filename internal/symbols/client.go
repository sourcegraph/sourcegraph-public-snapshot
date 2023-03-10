package symbols

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gobwas/glob"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/sourcegraph/go-ctags"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
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

func defaultEndpoints() *endpoint.Map {
	return endpoint.ConfBased(func(conns conftypes.ServiceConnections) []string {
		return conns.Symbols
	})
}

func LoadConfig() {
	cacheOptions := []TTLCacheOption[string, connAndErr]{
		WithExpirationTime[string, connAndErr](10 * time.Minute),
		WithReapInterval[string, connAndErr](1 * time.Minute),
		WithExpirationFunc(closeGRPCConnection),
	}

	cache := newTTLCache[string, connAndErr](newGRPCConnection, cacheOptions...)
	cache.StartReaper()

	DefaultClient = &Client{
		Endpoints:           defaultEndpoints(),
		gRPCConnCache:       cache,
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
	// Endpoints to symbols service.
	Endpoints *endpoint.Map

	gRPCConnCache *TTLCache[string, connAndErr]

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
	conn, err := c.getGRPCConn(string(repository))
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
	conn, err := c.getGRPCConn(string(args.Repo))
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
	conn, err := c.getGRPCConn(path.Repo)
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
	conn, err := c.getGRPCConn(args.Repo)
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

	symbolsURL, err := c.url(repo)
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

func (c *Client) getGRPCConn(repo string) (*grpc.ClientConn, error) {
	address, err := c.Endpoints.Get(repo)
	if err != nil {
		return nil, errors.Wrapf(err, "getting symbols server address for repo %q", repo)
	}

	connWithErr := c.gRPCConnCache.Get(address)
	return connWithErr.conn, connWithErr.dialErr
}

func (c *Client) url(repo api.RepoName) (string, error) {
	if c.Endpoints == nil {
		return "", errors.New("a symbols service has not been configured")
	}
	return c.Endpoints.Get(string(repo))
}

func newGRPCConnection(address string) connAndErr {
	u, err := url.Parse(address)
	if err != nil {
		return connAndErr{
			dialErr: errors.Wrapf(err, "parsing address %q", address),
		}
	}

	gRPCConn, err := defaults.Dial(u.Host)
	if err != nil {
		return connAndErr{
			dialErr: errors.Wrapf(err, "dialing gRPC connection to %q", address),
		}
	}

	return connAndErr{conn: gRPCConn}
}

func closeGRPCConnection(_ string, conn connAndErr) {
	if conn.dialErr != nil {
		_ = conn.conn.Close()
	}
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

// TTLCacheOption is a function that configures a TTLCache.
type TTLCacheOption[K comparable, V any] func(*TTLCache[K, V])

// WithReapInterval sets the interval at which the cache will reap expired entries.
func WithReapInterval[K comparable, V any](interval time.Duration) TTLCacheOption[K, V] {
	return func(c *TTLCache[K, V]) {
		c.reapInterval = interval
	}
}

// WithExpirationTime sets the expiration time for entries in the cache.
func WithExpirationTime[K comparable, V any](expiration time.Duration) TTLCacheOption[K, V] {
	return func(c *TTLCache[K, V]) {
		c.expirationTime = expiration
	}
}

// WithExpirationFunc sets the callback to be called when an entry expires.
func WithExpirationFunc[K comparable, V any](onExpiration func(K, V)) TTLCacheOption[K, V] {
	return func(c *TTLCache[K, V]) {
		c.expirationFunc = onExpiration
	}
}

// withClock sets the clock to be used by the cache. This is useful for testing.
func withClock[K comparable, V any](clock clock) TTLCacheOption[K, V] {
	return func(c *TTLCache[K, V]) {
		c.clock = clock
	}
}

// newTTLCache returns a new TTLCache with the provided newEntryFunc and options.
//
// newEntryFunc is the routine that runs when a cache miss occurs. The returned value is stored
// in the cache.
//
// By default, the cache will reap expired entries every minute and entries will
// expire after 10 minutes.
func newTTLCache[K comparable, V any](newEntryFunc func(K) V, options ...TTLCacheOption[K, V]) *TTLCache[K, V] {
	ctx, cancel := context.WithCancel(context.Background())

	cache := TTLCache[K, V]{
		reapContext:    ctx,
		reapCancelFunc: cancel,

		reapInterval:   1 * time.Minute,
		expirationTime: 10 * time.Minute,

		newEntryFunc:   newEntryFunc,
		expirationFunc: func(k K, v V) {},

		entries: make(map[K]*ttlEntry[V]),

		clock: productionClock{},
	}

	for _, option := range options {
		option(&cache)
	}

	return &cache
}

// TTLCache is a cache that expires entries after a given expiration time.
type TTLCache[K comparable, V any] struct {
	reapOnce sync.Once // reapOnce ensures that the background reaper is only started once.

	reapContext    context.Context    // reapContext is the context used for the background reaper.
	reapCancelFunc context.CancelFunc // reapCancelFunc is the cancel function for reapContext.

	reapInterval   time.Duration // reapInterval is the interval at which the cache will reap expired entries.
	expirationTime time.Duration // expirationTime is the expiration time for entries in the cache.

	newEntryFunc   func(K) V  // newEntryFunc is the routine that runs when a cache miss occurs.
	expirationFunc func(K, V) // expirationFunc is the callback to be called when an entry expires in the cache.

	mu      sync.RWMutex
	entries map[K]*ttlEntry[V] // entries is the map of entries in the cache.

	clock clock // clock is the clock used to determine the current time.
}

type ttlEntry[V any] struct {
	lastUsed atomic.Pointer[time.Time]
	value    V
}

// Get returns the value for the given key. If the key is not in the cache, it
// will be added using the newEntryFunc and returned to the caller.
func (c *TTLCache[K, V]) Get(key K) V {
	now := c.clock.Now()

	c.mu.RLock()

	// Fast path: check if the entry is already in the cache.
	e, ok := c.entries[key]
	if ok {
		e.lastUsed.Store(&now)
		value := e.value

		c.mu.RUnlock()
		return value
	}
	c.mu.RUnlock()

	// Slow path: lock the entire cache and check again.

	c.mu.Lock()
	defer c.mu.Unlock()

	// Did another goroutine already create the entry?
	e, ok = c.entries[key]
	if ok {
		e.lastUsed.Store(&now)
		return e.value
	}

	// Nobody created one, add a new one.
	e = &ttlEntry[V]{}
	e.lastUsed.Store(&now)
	e.value = c.newEntryFunc(key)

	c.entries[key] = e

	return e.value
}

// StartReaper starts the reaper goroutine. Every reapInterval, the reaper will
// remove entries that have not been accessed since expirationTime.
//
// shutdown can be called to stop the reaper. After shutdown is called, the
// reaper will not be restarted.
func (c *TTLCache[K, V]) StartReaper() {
	c.reapOnce.Do(func() {
		go func() {
			for {
				select {
				case <-c.reapContext.Done():
					return
				case <-time.After(c.reapInterval):
					c.reap()
				}
			}
		}()
	})
}

// reap removes all entries that have not been accessed since expirationTime, and calls
// the expirationFunc for each entry that is removed.
func (c *TTLCache[K, V]) reap() {
	now := c.clock.Now()
	earliestAllowed := now.Add(-c.expirationTime)

	getExpiredKeys := func() []K {
		var expired []K

		for key, entry := range c.entries {
			lastUsed := entry.lastUsed.Load()
			if lastUsed == nil {
				lastUsed = &time.Time{}
			}

			if (*lastUsed).Before(earliestAllowed) {
				expired = append(expired, key)
			}
		}

		return expired
	}

	// First, find all the entries that have expired.
	// We do this under a read lock to avoid blocking other goroutines
	// from accessing the cache.

	var maybeDelete []K

	c.mu.RLock()
	maybeDelete = getExpiredKeys()
	c.mu.RUnlock()

	// If there are no entries to delete, we're done.
	if len(maybeDelete) == 0 {
		return
	}

	// If there are entries to delete, only now do we need to acquire
	// the write lock to delete them.

	c.mu.Lock()
	defer c.mu.Unlock()

	// We need to check again to make sure that the entries are still
	// expired. It's possible that another goroutine has updated the
	// entry since we released the read lock.

	for _, key := range getExpiredKeys() {
		entry := c.entries[key]

		c.expirationFunc(key, entry.value)
		delete(c.entries, key)
	}
}

// Shutdown stops the background reaper. This function has no effect if the cache
// has already been shut down.
func (c *TTLCache[K, V]) Shutdown() {
	c.reapCancelFunc()
}

type clock interface {
	Now() time.Time
}

type productionClock struct{}

func (productionClock) Now() time.Time {
	return time.Now()
}

var _ clock = productionClock{}
