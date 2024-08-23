// Package server contains the endpoint handlers.
package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"time"

	"github.com/openfga/openfga/internal/throttler/threshold"

	"github.com/openfga/openfga/internal/throttler"

	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/openfga/openfga/internal/build"
	"github.com/openfga/openfga/internal/condition"
	"github.com/openfga/openfga/internal/graph"
	serverconfig "github.com/openfga/openfga/internal/server/config"
	"github.com/openfga/openfga/internal/utils"
	"github.com/openfga/openfga/internal/validation"
	"github.com/openfga/openfga/pkg/encoder"
	"github.com/openfga/openfga/pkg/gateway"
	"github.com/openfga/openfga/pkg/logger"
	httpmiddleware "github.com/openfga/openfga/pkg/middleware/http"
	"github.com/openfga/openfga/pkg/middleware/validator"
	"github.com/openfga/openfga/pkg/server/commands"
	serverErrors "github.com/openfga/openfga/pkg/server/errors"
	"github.com/openfga/openfga/pkg/storage"
	"github.com/openfga/openfga/pkg/storage/storagewrappers"
	"github.com/openfga/openfga/pkg/telemetry"
	"github.com/openfga/openfga/pkg/tuple"
	"github.com/openfga/openfga/pkg/typesystem"
)

type ExperimentalFeatureFlag string

const (
	AuthorizationModelIDHeader                          = "Openfga-Authorization-Model-Id"
	authorizationModelIDKey                             = "authorization_model_id"
	ExperimentalEnableListUsers ExperimentalFeatureFlag = "enable-list-users"
)

var tracer = otel.Tracer("openfga/pkg/server")

var (
	dispatchCountHistogramName = "dispatch_count"

	dispatchCountHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:                       build.ProjectName,
		Name:                            dispatchCountHistogramName,
		Help:                            "The number of dispatches required to resolve a query (e.g. Check).",
		Buckets:                         []float64{1, 5, 20, 50, 100, 150, 225, 400, 500, 750, 1000},
		NativeHistogramBucketFactor:     1.1,
		NativeHistogramMaxBucketNumber:  100,
		NativeHistogramMinResetDuration: time.Hour,
	}, []string{"grpc_service", "grpc_method"})

	datastoreQueryCountHistogramName = "datastore_query_count"

	datastoreQueryCountHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:                       build.ProjectName,
		Name:                            datastoreQueryCountHistogramName,
		Help:                            "The number of database queries required to resolve a query (e.g. Check, ListObjects or ListUsers).",
		Buckets:                         []float64{1, 5, 20, 50, 100, 150, 225, 400, 500, 750, 1000},
		NativeHistogramBucketFactor:     1.1,
		NativeHistogramMaxBucketNumber:  100,
		NativeHistogramMinResetDuration: time.Hour,
	}, []string{"grpc_service", "grpc_method"})

	requestDurationHistogramName = "request_duration_ms"

	requestDurationHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace:                       build.ProjectName,
		Name:                            requestDurationHistogramName,
		Help:                            "The request duration (in ms) labeled by method and buckets of datastore query counts and number of dispatches. This allows for reporting percentiles based on the number of datastore queries and number of dispatches required to resolve the request.",
		Buckets:                         []float64{1, 5, 10, 25, 50, 80, 100, 150, 200, 300, 1000, 2000, 5000},
		NativeHistogramBucketFactor:     1.1,
		NativeHistogramMaxBucketNumber:  100,
		NativeHistogramMinResetDuration: time.Hour,
	}, []string{"grpc_service", "grpc_method", "datastore_query_count", "dispatch_count"})
)

// A Server implements the OpenFGA service backend as both
// a GRPC and HTTP server.
type Server struct {
	openfgav1.UnimplementedOpenFGAServiceServer

	logger                           logger.Logger
	datastore                        storage.OpenFGADatastore
	encoder                          encoder.Encoder
	transport                        gateway.Transport
	resolveNodeLimit                 uint32
	resolveNodeBreadthLimit          uint32
	changelogHorizonOffset           int
	listObjectsDeadline              time.Duration
	listObjectsMaxResults            uint32
	listUsersDeadline                time.Duration
	listUsersMaxResults              uint32
	maxConcurrentReadsForListObjects uint32
	maxConcurrentReadsForCheck       uint32
	maxConcurrentReadsForListUsers   uint32
	maxAuthorizationModelCacheSize   int
	maxAuthorizationModelSizeInBytes int
	experimentals                    []ExperimentalFeatureFlag
	serviceName                      string

	// NOTE don't use this directly, use function resolveTypesystem. See https://github.com/openfga/openfga/issues/1527
	typesystemResolver     typesystem.TypesystemResolverFunc
	typesystemResolverStop func()

	checkQueryCacheEnabled bool
	checkQueryCacheLimit   uint32
	checkQueryCacheTTL     time.Duration
	cachedCheckResolver    *graph.CachedCheckResolver

	checkResolver graph.CheckResolver

	requestDurationByQueryHistogramBuckets         []uint
	requestDurationByDispatchCountHistogramBuckets []uint

	checkDispatchThrottlingEnabled          bool
	checkDispatchThrottlingFrequency        time.Duration
	checkDispatchThrottlingDefaultThreshold uint32
	checkDispatchThrottlingMaxThreshold     uint32

	listObjectsDispatchThrottlingEnabled      bool
	listObjectsDispatchThrottlingFrequency    time.Duration
	listObjectsDispatchDefaultThreshold       uint32
	listObjectsDispatchThrottlingMaxThreshold uint32

	dispatchThrottlingCheckResolver *graph.DispatchThrottlingCheckResolver

	listObjectsDispatchThrottler throttler.Throttler
}

type OpenFGAServiceV1Option func(s *Server)

// WithDatastore passes a datastore to the Server.
// You must call [storage.OpenFGADatastore.Close] on it after you have stopped using it.
func WithDatastore(ds storage.OpenFGADatastore) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.datastore = ds
	}
}

// WithAuthorizationModelCacheSize sets the maximum number of authorization models that will be cached in memory.
func WithAuthorizationModelCacheSize(maxAuthorizationModelCacheSize int) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.maxAuthorizationModelCacheSize = maxAuthorizationModelCacheSize
	}
}

func WithLogger(l logger.Logger) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.logger = l
	}
}

func WithTokenEncoder(encoder encoder.Encoder) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.encoder = encoder
	}
}

// WithTransport sets the connection transport.
func WithTransport(t gateway.Transport) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.transport = t
	}
}

// WithResolveNodeLimit sets a limit on the number of recursive calls that one Check, ListObjects or ListUsers call will allow.
// Thinking of a request as a tree of evaluations, this option controls
// how many levels we will evaluate before throwing an error that the authorization model is too complex.
func WithResolveNodeLimit(limit uint32) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.resolveNodeLimit = limit
	}
}

// WithResolveNodeBreadthLimit sets a limit on the number of goroutines that can be created
// when evaluating a subtree of a Check, ListObjects or ListUsers call.
// Thinking of a Check request as a tree of evaluations, this option controls,
// on a given level of the tree, the maximum number of nodes that can be evaluated concurrently (the breadth).
// If your authorization models are very complex (e.g. one relation is a union of many relations, or one relation
// is deeply nested), or if you have lots of users for (object, relation) pairs,
// you should set this option to be a low number (e.g. 1000).
func WithResolveNodeBreadthLimit(limit uint32) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.resolveNodeBreadthLimit = limit
	}
}

// WithChangelogHorizonOffset sets an offset (in minutes) from the current time.
// Changes that occur after this offset will not be included in the response of ReadChanges API.
// If your datastore is eventually consistent or if you have a database with replication delay, we recommend setting this (e.g. 1 minute).
func WithChangelogHorizonOffset(offset int) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.changelogHorizonOffset = offset
	}
}

// WithListObjectsDeadline affect the ListObjects API and Streamed ListObjects API only.
// It sets the maximum amount of time that the server will spend gathering results.
func WithListObjectsDeadline(deadline time.Duration) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.listObjectsDeadline = deadline
	}
}

// WithListObjectsMaxResults affects the ListObjects API only.
// It sets the maximum number of results that this API will return.
func WithListObjectsMaxResults(limit uint32) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.listObjectsMaxResults = limit
	}
}

// WithListUsersDeadline affect the ListUsers API only.
// It sets the maximum amount of time that the server will spend gathering results.
func WithListUsersDeadline(deadline time.Duration) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.listUsersDeadline = deadline
	}
}

// WithListUsersMaxResults affects the ListUsers API only.
// It sets the maximum number of results that this API will return.
// If it's zero, all results will be attempted to be returned.
func WithListUsersMaxResults(limit uint32) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.listUsersMaxResults = limit
	}
}

// WithMaxConcurrentReadsForListObjects sets a limit on the number of datastore reads that can be in flight for a given ListObjects call.
// This number should be set depending on the RPS expected for Check and ListObjects APIs, the number of OpenFGA replicas running,
// and the number of connections the datastore allows.
// E.g. If Datastore.MaxOpenConns = 100 and assuming that each ListObjects call takes 1 second and no traffic to Check API:
// - One OpenFGA replica and expected traffic of 100 RPS => set it to 1.
// - One OpenFGA replica and expected traffic of 1 RPS => set it to 100.
// - Two OpenFGA replicas and expected traffic of 1 RPS => set it to 50.
func WithMaxConcurrentReadsForListObjects(max uint32) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.maxConcurrentReadsForListObjects = max
	}
}

// WithMaxConcurrentReadsForCheck sets a limit on the number of datastore reads that can be in flight for a given Check call.
// This number should be set depending on the RPS expected for Check and ListObjects APIs, the number of OpenFGA replicas running,
// and the number of connections the datastore allows.
// E.g. If Datastore.MaxOpenConns = 100 and assuming that each Check call takes 1 second and no traffic to ListObjects API:
// - One OpenFGA replica and expected traffic of 100 RPS => set it to 1.
// - One OpenFGA replica and expected traffic of 1 RPS => set it to 100.
// - Two OpenFGA replicas and expected traffic of 1 RPS => set it to 50.
func WithMaxConcurrentReadsForCheck(max uint32) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.maxConcurrentReadsForCheck = max
	}
}

// WithMaxConcurrentReadsForListUsers sets a limit on the number of datastore reads that can be in flight for a given ListUsers call.
// This number should be set depending on the RPS expected for all query APIs, the number of OpenFGA replicas running,
// and the number of connections the datastore allows.
// E.g. If Datastore.MaxOpenConns = 100 and assuming that each ListUsers call takes 1 second and no traffic to other query APIs:
// - One OpenFGA replica and expected traffic of 100 RPS => set it to 1.
// - One OpenFGA replica and expected traffic of 1 RPS => set it to 100.
// - Two OpenFGA replicas and expected traffic of 1 RPS => set it to 50.
func WithMaxConcurrentReadsForListUsers(max uint32) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.maxConcurrentReadsForListUsers = max
	}
}

func WithExperimentals(experimentals ...ExperimentalFeatureFlag) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.experimentals = experimentals
	}
}

// WithCheckQueryCacheEnabled enables caching of Check results for the Check and List objects APIs.
// This cache is shared for all requests.
// See also WithCheckQueryCacheLimit and WithCheckQueryCacheTTL.
func WithCheckQueryCacheEnabled(enabled bool) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.checkQueryCacheEnabled = enabled
	}
}

// WithCheckQueryCacheLimit sets the cache size limit (in items)
// Needs WithCheckQueryCacheEnabled set to true.
func WithCheckQueryCacheLimit(limit uint32) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.checkQueryCacheLimit = limit
	}
}

// WithCheckQueryCacheTTL sets the TTL of cached checks and list objects partial results
// Needs WithCheckQueryCacheEnabled set to true.
func WithCheckQueryCacheTTL(ttl time.Duration) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.checkQueryCacheTTL = ttl
	}
}

// WithRequestDurationByQueryHistogramBuckets sets the buckets used in labelling the requestDurationByQueryAndDispatchHistogram.
func WithRequestDurationByQueryHistogramBuckets(buckets []uint) OpenFGAServiceV1Option {
	return func(s *Server) {
		sort.Slice(buckets, func(i, j int) bool { return buckets[i] < buckets[j] })
		s.requestDurationByQueryHistogramBuckets = buckets
	}
}

// WithRequestDurationByDispatchCountHistogramBuckets sets the buckets used in labelling the requestDurationByQueryAndDispatchHistogram.
func WithRequestDurationByDispatchCountHistogramBuckets(buckets []uint) OpenFGAServiceV1Option {
	return func(s *Server) {
		sort.Slice(buckets, func(i, j int) bool { return buckets[i] < buckets[j] })
		s.requestDurationByDispatchCountHistogramBuckets = buckets
	}
}

func WithMaxAuthorizationModelSizeInBytes(size int) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.maxAuthorizationModelSizeInBytes = size
	}
}

// WithDispatchThrottlingCheckResolverEnabled sets whether dispatch throttling is enabled for Check requests.
// Enabling this feature will prioritize dispatched requests requiring less than the configured dispatch
// threshold over requests whose dispatch count exceeds the configured threshold.
func WithDispatchThrottlingCheckResolverEnabled(enabled bool) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.checkDispatchThrottlingEnabled = enabled
	}
}

// WithDispatchThrottlingCheckResolverFrequency defines how frequent dispatch throttling
// will be evaluated for Check requests.
// Frequency controls how frequently throttled dispatch requests are evaluated to determine whether
// it can be processed.
// This value should not be too small (i.e., in the ns ranges) as i) there are limitation in timer resolution
// and ii) very small value will result in a higher frequency of processing dispatches,
// which diminishes the value of the throttling.
func WithDispatchThrottlingCheckResolverFrequency(frequency time.Duration) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.checkDispatchThrottlingFrequency = frequency
	}
}

// WithDispatchThrottlingCheckResolverThreshold define the number of dispatches to be throttled.
// In addition, it will update checkDispatchThrottlingMaxThreshold if required.
func WithDispatchThrottlingCheckResolverThreshold(defaultThreshold uint32) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.checkDispatchThrottlingDefaultThreshold = defaultThreshold
	}
}

// WithDispatchThrottlingCheckResolverMaxThreshold define the maximum threshold values allowed
// It will ensure checkDispatchThrottlingMaxThreshold will never be smaller than threshold.
func WithDispatchThrottlingCheckResolverMaxThreshold(maxThreshold uint32) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.checkDispatchThrottlingMaxThreshold = maxThreshold
	}
}

// MustNewServerWithOpts see NewServerWithOpts.
func MustNewServerWithOpts(opts ...OpenFGAServiceV1Option) *Server {
	s, err := NewServerWithOpts(opts...)
	if err != nil {
		panic(fmt.Errorf("failed to construct the OpenFGA server: %w", err))
	}

	return s
}

func (s *Server) IsExperimentallyEnabled(flag ExperimentalFeatureFlag) bool {
	return slices.Contains(s.experimentals, flag)
}

// WithListObjectsDispatchThrottlingEnabled sets whether dispatch throttling is enabled for List Objects requests.
// Enabling this feature will prioritize dispatched requests requiring less than the configured dispatch
// threshold over requests whose dispatch count exceeds the configured threshold.
func WithListObjectsDispatchThrottlingEnabled(enabled bool) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.listObjectsDispatchThrottlingEnabled = enabled
	}
}

// WithListObjectsDispatchThrottlingFrequency defines how frequent dispatch throttling
// will be evaluated for List Objects requests.
// Frequency controls how frequently throttled dispatch requests are evaluated to determine whether
// it can be processed.
// This value should not be too small (i.e., in the ns ranges) as i) there are limitation in timer resolution
// and ii) very small value will result in a higher frequency of processing dispatches,
// which diminishes the value of the throttling.
func WithListObjectsDispatchThrottlingFrequency(frequency time.Duration) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.listObjectsDispatchThrottlingFrequency = frequency
	}
}

// WithListObjectsDispatchThrottlingThreshold define the number of dispatches to be throttled
// for List Objects requests.
func WithListObjectsDispatchThrottlingThreshold(threshold uint32) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.listObjectsDispatchDefaultThreshold = threshold
	}
}

// WithListObjectsDispatchThrottlingMaxThreshold define the maximum threshold values allowed
// It will ensure listObjectsDispatchThrottlingMaxThreshold will never be smaller than threshold.
func WithListObjectsDispatchThrottlingMaxThreshold(maxThreshold uint32) OpenFGAServiceV1Option {
	return func(s *Server) {
		s.listObjectsDispatchThrottlingMaxThreshold = maxThreshold
	}
}

// NewServerWithOpts returns a new server.
// You must call Close on it after you are done using it.
func NewServerWithOpts(opts ...OpenFGAServiceV1Option) (*Server, error) {
	s := &Server{
		logger:                           logger.NewNoopLogger(),
		encoder:                          encoder.NewBase64Encoder(),
		transport:                        gateway.NewNoopTransport(),
		changelogHorizonOffset:           serverconfig.DefaultChangelogHorizonOffset,
		resolveNodeLimit:                 serverconfig.DefaultResolveNodeLimit,
		resolveNodeBreadthLimit:          serverconfig.DefaultResolveNodeBreadthLimit,
		listObjectsDeadline:              serverconfig.DefaultListObjectsDeadline,
		listObjectsMaxResults:            serverconfig.DefaultListObjectsMaxResults,
		listUsersDeadline:                serverconfig.DefaultListUsersDeadline,
		listUsersMaxResults:              serverconfig.DefaultListUsersMaxResults,
		maxConcurrentReadsForCheck:       serverconfig.DefaultMaxConcurrentReadsForCheck,
		maxConcurrentReadsForListObjects: serverconfig.DefaultMaxConcurrentReadsForListObjects,
		maxConcurrentReadsForListUsers:   serverconfig.DefaultMaxConcurrentReadsForListUsers,
		maxAuthorizationModelSizeInBytes: serverconfig.DefaultMaxAuthorizationModelSizeInBytes,
		maxAuthorizationModelCacheSize:   serverconfig.DefaultMaxAuthorizationModelCacheSize,
		experimentals:                    make([]ExperimentalFeatureFlag, 0, 10),

		checkQueryCacheEnabled: serverconfig.DefaultCheckQueryCacheEnable,
		checkQueryCacheLimit:   serverconfig.DefaultCheckQueryCacheLimit,
		checkQueryCacheTTL:     serverconfig.DefaultCheckQueryCacheTTL,
		checkResolver:          nil,

		requestDurationByQueryHistogramBuckets:         []uint{50, 200},
		requestDurationByDispatchCountHistogramBuckets: []uint{50, 200},
		serviceName: openfgav1.OpenFGAService_ServiceDesc.ServiceName,

		checkDispatchThrottlingEnabled:          serverconfig.DefaultCheckDispatchThrottlingEnabled,
		checkDispatchThrottlingFrequency:        serverconfig.DefaultCheckDispatchThrottlingFrequency,
		checkDispatchThrottlingDefaultThreshold: serverconfig.DefaultCheckDispatchThrottlingDefaultThreshold,

		listObjectsDispatchThrottlingEnabled:      serverconfig.DefaultListObjectsDispatchThrottlingEnabled,
		listObjectsDispatchThrottlingFrequency:    serverconfig.DefaultListObjectsDispatchThrottlingFrequency,
		listObjectsDispatchDefaultThreshold:       serverconfig.DefaultListObjectsDispatchThrottlingDefaultThreshold,
		listObjectsDispatchThrottlingMaxThreshold: serverconfig.DefaultListObjectsDispatchThrottlingMaxThreshold,
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.datastore == nil {
		return nil, fmt.Errorf("a datastore option must be provided")
	}

	if len(s.requestDurationByQueryHistogramBuckets) == 0 {
		return nil, fmt.Errorf("request duration datastore count buckets must not be empty")
	}

	if len(s.requestDurationByDispatchCountHistogramBuckets) == 0 {
		return nil, fmt.Errorf("request duration by dispatch count buckets must not be empty")
	}
	if s.checkDispatchThrottlingEnabled && s.checkDispatchThrottlingMaxThreshold != 0 && s.checkDispatchThrottlingDefaultThreshold > s.checkDispatchThrottlingMaxThreshold {
		return nil, fmt.Errorf("check default dispatch throttling threshold must be equal or smaller than max dispatch threshold for Check")
	}

	if s.listObjectsDispatchThrottlingMaxThreshold != 0 && s.listObjectsDispatchDefaultThreshold > s.listObjectsDispatchThrottlingMaxThreshold {
		return nil, fmt.Errorf("ListObjects default dispatch throttling threshold must be equal or smaller than max dispatch threshold for ListObjects")
	}

	// below this point, don't throw errors or we may leak resources in tests

	cycleDetectionCheckResolver := graph.NewCycleDetectionCheckResolver()
	s.checkResolver = cycleDetectionCheckResolver

	localChecker := graph.NewLocalChecker(
		graph.WithResolveNodeBreadthLimit(s.resolveNodeBreadthLimit),
	)

	cycleDetectionCheckResolver.SetDelegate(localChecker)
	localChecker.SetDelegate(cycleDetectionCheckResolver)

	if s.checkQueryCacheEnabled {
		s.logger.Info("Check query cache is enabled and may lead to stale query results up to the configured query cache TTL",
			zap.Duration("CheckQueryCacheTTL", s.checkQueryCacheTTL),
			zap.Uint32("CheckQueryCacheLimit", s.checkQueryCacheLimit))

		cachedCheckResolver := graph.NewCachedCheckResolver(
			graph.WithMaxCacheSize(int64(s.checkQueryCacheLimit)),
			graph.WithLogger(s.logger),
			graph.WithCacheTTL(s.checkQueryCacheTTL),
		)
		s.cachedCheckResolver = cachedCheckResolver

		cachedCheckResolver.SetDelegate(localChecker)
		cycleDetectionCheckResolver.SetDelegate(cachedCheckResolver)
	}

	if s.checkDispatchThrottlingEnabled {
		s.logger.Info("Enabling Check dispatch throttling",
			zap.Duration("Frequency", s.checkDispatchThrottlingFrequency),
			zap.Uint32("DefaultThreshold", s.checkDispatchThrottlingDefaultThreshold),
			zap.Uint32("MaxThreshold", s.checkDispatchThrottlingMaxThreshold),
		)

		dispatchThrottlingConfig := graph.DispatchThrottlingCheckResolverConfig{
			DefaultThreshold: s.checkDispatchThrottlingDefaultThreshold,
			MaxThreshold:     s.checkDispatchThrottlingMaxThreshold,
		}

		dispatchThrottlingCheckResolver := graph.NewDispatchThrottlingCheckResolver(
			graph.WithDispatchThrottlingCheckResolverConfig(dispatchThrottlingConfig),
			graph.WithThrottler(throttler.NewConstantRateThrottler(s.checkDispatchThrottlingFrequency, "check_dispatch_throttle")),
		)
		dispatchThrottlingCheckResolver.SetDelegate(localChecker)
		s.dispatchThrottlingCheckResolver = dispatchThrottlingCheckResolver

		if s.cachedCheckResolver != nil {
			s.cachedCheckResolver.SetDelegate(dispatchThrottlingCheckResolver)
		} else {
			cycleDetectionCheckResolver.SetDelegate(dispatchThrottlingCheckResolver)
		}
	}

	if s.listObjectsDispatchThrottlingEnabled {
		s.logger.Info("Enabling ListObjects dispatch throttling",
			zap.Duration("Frequency", s.listObjectsDispatchThrottlingFrequency),
			zap.Uint32("DefaultThreshold", s.listObjectsDispatchDefaultThreshold),
			zap.Uint32("MaxThreshold", s.listObjectsDispatchThrottlingMaxThreshold),
		)

		s.listObjectsDispatchThrottler = throttler.NewConstantRateThrottler(s.listObjectsDispatchThrottlingFrequency, "list_objects_dispatch_throttle")
	}

	s.datastore = storagewrappers.NewCachedOpenFGADatastore(storagewrappers.NewContextWrapper(s.datastore), s.maxAuthorizationModelCacheSize)

	s.typesystemResolver, s.typesystemResolverStop = typesystem.MemoizedTypesystemResolverFunc(s.datastore)

	return s, nil
}

// Close releases the server resources.
func (s *Server) Close() {
	if s.dispatchThrottlingCheckResolver != nil {
		s.dispatchThrottlingCheckResolver.Close()
	}

	if s.listObjectsDispatchThrottler != nil {
		s.listObjectsDispatchThrottler.Close()
	}

	if s.cachedCheckResolver != nil {
		s.cachedCheckResolver.Close()
	}

	if s.checkResolver != nil {
		s.checkResolver.Close()
	}
	s.datastore.Close()
	s.typesystemResolverStop()
}

func (s *Server) ListObjects(ctx context.Context, req *openfgav1.ListObjectsRequest) (*openfgav1.ListObjectsResponse, error) {
	start := time.Now()

	targetObjectType := req.GetType()

	ctx, span := tracer.Start(ctx, "ListObjects", trace.WithAttributes(
		attribute.String("object_type", targetObjectType),
		attribute.String("relation", req.GetRelation()),
		attribute.String("user", req.GetUser()),
	))
	defer span.End()

	if !validator.RequestIsValidatedFromContext(ctx) {
		if err := req.Validate(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	const methodName = "listobjects"

	ctx = telemetry.ContextWithRPCInfo(ctx, telemetry.RPCInfo{
		Service: s.serviceName,
		Method:  methodName,
	})

	storeID := req.GetStoreId()

	typesys, err := s.resolveTypesystem(ctx, storeID, req.GetAuthorizationModelId())
	if err != nil {
		return nil, err
	}

	q, err := commands.NewListObjectsQuery(
		s.datastore,
		s.checkResolver,
		commands.WithLogger(s.logger),
		commands.WithListObjectsDeadline(s.listObjectsDeadline),
		commands.WithListObjectsMaxResults(s.listObjectsMaxResults),
		commands.WithDispatchThrottlerConfig(threshold.Config{
			Throttler:    s.listObjectsDispatchThrottler,
			Enabled:      s.listObjectsDispatchThrottlingEnabled,
			Threshold:    s.listObjectsDispatchDefaultThreshold,
			MaxThreshold: s.listObjectsDispatchThrottlingMaxThreshold,
		}),
		commands.WithResolveNodeLimit(s.resolveNodeLimit),
		commands.WithResolveNodeBreadthLimit(s.resolveNodeBreadthLimit),
		commands.WithMaxConcurrentReads(s.maxConcurrentReadsForListObjects),
	)
	if err != nil {
		return nil, serverErrors.NewInternalError("", err)
	}

	result, err := q.Execute(
		typesystem.ContextWithTypesystem(ctx, typesys),
		&openfgav1.ListObjectsRequest{
			StoreId:              storeID,
			ContextualTuples:     req.GetContextualTuples(),
			AuthorizationModelId: typesys.GetAuthorizationModelID(), // the resolved model id
			Type:                 targetObjectType,
			Relation:             req.GetRelation(),
			User:                 req.GetUser(),
			Context:              req.GetContext(),
		},
	)
	if err != nil {
		telemetry.TraceError(span, err)
		if errors.Is(err, condition.ErrEvaluationFailed) {
			return nil, serverErrors.ValidationError(err)
		}

		return nil, err
	}
	datastoreQueryCount := float64(*result.ResolutionMetadata.DatastoreQueryCount)

	grpc_ctxtags.Extract(ctx).Set(datastoreQueryCountHistogramName, datastoreQueryCount)
	span.SetAttributes(attribute.Float64(datastoreQueryCountHistogramName, datastoreQueryCount))
	datastoreQueryCountHistogram.WithLabelValues(
		s.serviceName,
		methodName,
	).Observe(datastoreQueryCount)

	dispatchCount := float64(result.ResolutionMetadata.DispatchCounter.Load())

	grpc_ctxtags.Extract(ctx).Set(dispatchCountHistogramName, dispatchCount)
	span.SetAttributes(attribute.Float64(dispatchCountHistogramName, dispatchCount))
	dispatchCountHistogram.WithLabelValues(
		s.serviceName,
		methodName,
	).Observe(dispatchCount)

	requestDurationHistogram.WithLabelValues(
		s.serviceName,
		methodName,
		utils.Bucketize(uint(*result.ResolutionMetadata.DatastoreQueryCount), s.requestDurationByQueryHistogramBuckets),
		utils.Bucketize(uint(result.ResolutionMetadata.DispatchCounter.Load()), s.requestDurationByDispatchCountHistogramBuckets),
	).Observe(float64(time.Since(start).Milliseconds()))

	return &openfgav1.ListObjectsResponse{
		Objects: result.Objects,
	}, nil
}

func (s *Server) StreamedListObjects(req *openfgav1.StreamedListObjectsRequest, srv openfgav1.OpenFGAService_StreamedListObjectsServer) error {
	start := time.Now()

	ctx := srv.Context()
	ctx, span := tracer.Start(ctx, "StreamedListObjects", trace.WithAttributes(
		attribute.String("object_type", req.GetType()),
		attribute.String("relation", req.GetRelation()),
		attribute.String("user", req.GetUser()),
	))
	defer span.End()

	if !validator.RequestIsValidatedFromContext(ctx) {
		if err := req.Validate(); err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}
	}

	const methodName = "streamedlistobjects"

	ctx = telemetry.ContextWithRPCInfo(ctx, telemetry.RPCInfo{
		Service: s.serviceName,
		Method:  methodName,
	})

	storeID := req.GetStoreId()

	typesys, err := s.resolveTypesystem(ctx, storeID, req.GetAuthorizationModelId())
	if err != nil {
		return err
	}

	q, err := commands.NewListObjectsQuery(
		s.datastore,
		s.checkResolver,
		commands.WithLogger(s.logger),
		commands.WithListObjectsDeadline(s.listObjectsDeadline),
		commands.WithDispatchThrottlerConfig(threshold.Config{
			Throttler:    s.listObjectsDispatchThrottler,
			Enabled:      s.listObjectsDispatchThrottlingEnabled,
			Threshold:    s.listObjectsDispatchDefaultThreshold,
			MaxThreshold: s.listObjectsDispatchThrottlingMaxThreshold,
		}),
		commands.WithListObjectsMaxResults(s.listObjectsMaxResults),
		commands.WithResolveNodeLimit(s.resolveNodeLimit),
		commands.WithResolveNodeBreadthLimit(s.resolveNodeBreadthLimit),
		commands.WithMaxConcurrentReads(s.maxConcurrentReadsForListObjects),
	)
	if err != nil {
		return serverErrors.NewInternalError("", err)
	}

	req.AuthorizationModelId = typesys.GetAuthorizationModelID() // the resolved model id

	resolutionMetadata, err := q.ExecuteStreamed(
		typesystem.ContextWithTypesystem(ctx, typesys),
		req,
		srv,
	)
	if err != nil {
		telemetry.TraceError(span, err)
		return err
	}
	datastoreQueryCount := float64(*resolutionMetadata.DatastoreQueryCount)

	grpc_ctxtags.Extract(ctx).Set(datastoreQueryCountHistogramName, datastoreQueryCount)
	span.SetAttributes(attribute.Float64(datastoreQueryCountHistogramName, datastoreQueryCount))
	datastoreQueryCountHistogram.WithLabelValues(
		s.serviceName,
		methodName,
	).Observe(datastoreQueryCount)

	dispatchCount := float64(resolutionMetadata.DispatchCounter.Load())

	grpc_ctxtags.Extract(ctx).Set(dispatchCountHistogramName, dispatchCount)
	span.SetAttributes(attribute.Float64(dispatchCountHistogramName, dispatchCount))
	dispatchCountHistogram.WithLabelValues(
		s.serviceName,
		methodName,
	).Observe(dispatchCount)

	requestDurationHistogram.WithLabelValues(
		s.serviceName,
		methodName,
		utils.Bucketize(uint(*resolutionMetadata.DatastoreQueryCount), s.requestDurationByQueryHistogramBuckets),
		utils.Bucketize(uint(resolutionMetadata.DispatchCounter.Load()), s.requestDurationByDispatchCountHistogramBuckets),
	).Observe(float64(time.Since(start).Milliseconds()))

	return nil
}

func (s *Server) Read(ctx context.Context, req *openfgav1.ReadRequest) (*openfgav1.ReadResponse, error) {
	tk := req.GetTupleKey()
	ctx, span := tracer.Start(ctx, "Read", trace.WithAttributes(
		attribute.KeyValue{Key: "object", Value: attribute.StringValue(tk.GetObject())},
		attribute.KeyValue{Key: "relation", Value: attribute.StringValue(tk.GetRelation())},
		attribute.KeyValue{Key: "user", Value: attribute.StringValue(tk.GetUser())},
	))
	defer span.End()

	if !validator.RequestIsValidatedFromContext(ctx) {
		if err := req.Validate(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	ctx = telemetry.ContextWithRPCInfo(ctx, telemetry.RPCInfo{
		Service: s.serviceName,
		Method:  "Read",
	})

	q := commands.NewReadQuery(s.datastore,
		commands.WithReadQueryLogger(s.logger),
		commands.WithReadQueryEncoder(s.encoder),
	)
	return q.Execute(ctx, &openfgav1.ReadRequest{
		StoreId:           req.GetStoreId(),
		TupleKey:          tk,
		PageSize:          req.GetPageSize(),
		ContinuationToken: req.GetContinuationToken(),
	})
}

func (s *Server) Write(ctx context.Context, req *openfgav1.WriteRequest) (*openfgav1.WriteResponse, error) {
	ctx, span := tracer.Start(ctx, "Write")
	defer span.End()

	if !validator.RequestIsValidatedFromContext(ctx) {
		if err := req.Validate(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	ctx = telemetry.ContextWithRPCInfo(ctx, telemetry.RPCInfo{
		Service: s.serviceName,
		Method:  "Write",
	})

	storeID := req.GetStoreId()

	typesys, err := s.resolveTypesystem(ctx, storeID, req.GetAuthorizationModelId())
	if err != nil {
		return nil, err
	}

	cmd := commands.NewWriteCommand(
		s.datastore,
		commands.WithWriteCmdLogger(s.logger),
	)
	return cmd.Execute(ctx, &openfgav1.WriteRequest{
		StoreId:              storeID,
		AuthorizationModelId: typesys.GetAuthorizationModelID(), // the resolved model id
		Writes:               req.GetWrites(),
		Deletes:              req.GetDeletes(),
	})
}

func (s *Server) Check(ctx context.Context, req *openfgav1.CheckRequest) (*openfgav1.CheckResponse, error) {
	start := time.Now()

	tk := req.GetTupleKey()
	ctx, span := tracer.Start(ctx, "Check", trace.WithAttributes(
		attribute.KeyValue{Key: "store_id", Value: attribute.StringValue(req.GetStoreId())},
		attribute.KeyValue{Key: "object", Value: attribute.StringValue(tk.GetObject())},
		attribute.KeyValue{Key: "relation", Value: attribute.StringValue(tk.GetRelation())},
		attribute.KeyValue{Key: "user", Value: attribute.StringValue(tk.GetUser())},
	))
	defer span.End()

	if !validator.RequestIsValidatedFromContext(ctx) {
		if err := req.Validate(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	ctx = telemetry.ContextWithRPCInfo(ctx, telemetry.RPCInfo{
		Service: s.serviceName,
		Method:  "Check",
	})

	storeID := req.GetStoreId()

	typesys, err := s.resolveTypesystem(ctx, storeID, req.GetAuthorizationModelId())
	if err != nil {
		return nil, err
	}

	if err := validation.ValidateUserObjectRelation(typesys, tuple.ConvertCheckRequestTupleKeyToTupleKey(tk)); err != nil {
		return nil, serverErrors.ValidationError(err)
	}

	for _, ctxTuple := range req.GetContextualTuples().GetTupleKeys() {
		if err := validation.ValidateTuple(typesys, ctxTuple); err != nil {
			return nil, serverErrors.HandleTupleValidateError(err)
		}
	}

	ctx = typesystem.ContextWithTypesystem(ctx, typesys)
	ctx = storage.ContextWithRelationshipTupleReader(ctx,
		storagewrappers.NewBoundedConcurrencyTupleReader(
			storagewrappers.NewCombinedTupleReader(
				s.datastore,
				req.GetContextualTuples().GetTupleKeys(),
			),
			s.maxConcurrentReadsForCheck,
		),
	)

	checkRequestMetadata := graph.NewCheckRequestMetadata(s.resolveNodeLimit)

	resolveCheckRequest := graph.ResolveCheckRequest{
		StoreID:              req.GetStoreId(),
		AuthorizationModelID: typesys.GetAuthorizationModelID(), // the resolved model id
		TupleKey:             tuple.ConvertCheckRequestTupleKeyToTupleKey(req.GetTupleKey()),
		ContextualTuples:     req.GetContextualTuples().GetTupleKeys(),
		Context:              req.GetContext(),
		RequestMetadata:      checkRequestMetadata,
	}

	resp, err := s.checkResolver.ResolveCheck(ctx, &resolveCheckRequest)
	if err != nil {
		telemetry.TraceError(span, err)
		if errors.Is(err, graph.ErrResolutionDepthExceeded) {
			return nil, serverErrors.AuthorizationModelResolutionTooComplex
		}

		if errors.Is(err, condition.ErrEvaluationFailed) {
			return nil, serverErrors.ValidationError(err)
		}

		// Note for ListObjects:
		// Currently this is not feasible in ListObjects as we return partial results.
		if errors.Is(err, context.DeadlineExceeded) && resolveCheckRequest.GetRequestMetadata().WasThrottled.Load() {
			return nil, serverErrors.ThrottledTimeout
		}

		return nil, serverErrors.HandleError("", err)
	}

	queryCount := float64(resp.GetResolutionMetadata().DatastoreQueryCount)
	const methodName = "check"

	grpc_ctxtags.Extract(ctx).Set(datastoreQueryCountHistogramName, queryCount)
	span.SetAttributes(attribute.Float64(datastoreQueryCountHistogramName, queryCount))
	datastoreQueryCountHistogram.WithLabelValues(
		s.serviceName,
		methodName,
	).Observe(queryCount)

	rawDispatchCount := checkRequestMetadata.DispatchCounter.Load()
	dispatchCount := float64(rawDispatchCount)

	grpc_ctxtags.Extract(ctx).Set(dispatchCountHistogramName, dispatchCount)
	span.SetAttributes(attribute.Float64(dispatchCountHistogramName, dispatchCount))
	dispatchCountHistogram.WithLabelValues(
		s.serviceName,
		methodName,
	).Observe(dispatchCount)

	res := &openfgav1.CheckResponse{
		Allowed: resp.Allowed,
	}

	span.SetAttributes(attribute.KeyValue{Key: "allowed", Value: attribute.BoolValue(res.GetAllowed())})

	requestDurationHistogram.WithLabelValues(
		s.serviceName,
		methodName,
		utils.Bucketize(uint(resp.GetResolutionMetadata().DatastoreQueryCount), s.requestDurationByQueryHistogramBuckets),
		utils.Bucketize(uint(rawDispatchCount), s.requestDurationByDispatchCountHistogramBuckets),
	).Observe(float64(time.Since(start).Milliseconds()))

	return res, nil
}

func (s *Server) Expand(ctx context.Context, req *openfgav1.ExpandRequest) (*openfgav1.ExpandResponse, error) {
	tk := req.GetTupleKey()
	ctx, span := tracer.Start(ctx, "Expand", trace.WithAttributes(
		attribute.KeyValue{Key: "object", Value: attribute.StringValue(tk.GetObject())},
		attribute.KeyValue{Key: "relation", Value: attribute.StringValue(tk.GetRelation())},
	))
	defer span.End()

	if !validator.RequestIsValidatedFromContext(ctx) {
		if err := req.Validate(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	ctx = telemetry.ContextWithRPCInfo(ctx, telemetry.RPCInfo{
		Service: s.serviceName,
		Method:  "Expand",
	})

	storeID := req.GetStoreId()

	typesys, err := s.resolveTypesystem(ctx, storeID, req.GetAuthorizationModelId())
	if err != nil {
		return nil, err
	}

	q := commands.NewExpandQuery(s.datastore, commands.WithExpandQueryLogger(s.logger))
	return q.Execute(ctx, &openfgav1.ExpandRequest{
		StoreId:              storeID,
		AuthorizationModelId: typesys.GetAuthorizationModelID(), // the resolved model id
		TupleKey:             tk,
	})
}

func (s *Server) ReadAuthorizationModel(ctx context.Context, req *openfgav1.ReadAuthorizationModelRequest) (*openfgav1.ReadAuthorizationModelResponse, error) {
	ctx, span := tracer.Start(ctx, "ReadAuthorizationModel", trace.WithAttributes(
		attribute.KeyValue{Key: authorizationModelIDKey, Value: attribute.StringValue(req.GetId())},
	))
	defer span.End()

	if !validator.RequestIsValidatedFromContext(ctx) {
		if err := req.Validate(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	ctx = telemetry.ContextWithRPCInfo(ctx, telemetry.RPCInfo{
		Service: s.serviceName,
		Method:  "ReadAuthorizationModels",
	})

	q := commands.NewReadAuthorizationModelQuery(s.datastore, commands.WithReadAuthModelQueryLogger(s.logger))
	return q.Execute(ctx, req)
}

func (s *Server) WriteAuthorizationModel(ctx context.Context, req *openfgav1.WriteAuthorizationModelRequest) (*openfgav1.WriteAuthorizationModelResponse, error) {
	ctx, span := tracer.Start(ctx, "WriteAuthorizationModel")
	defer span.End()

	if !validator.RequestIsValidatedFromContext(ctx) {
		if err := req.Validate(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	ctx = telemetry.ContextWithRPCInfo(ctx, telemetry.RPCInfo{
		Service: s.serviceName,
		Method:  "WriteAuthorizationModel",
	})

	c := commands.NewWriteAuthorizationModelCommand(s.datastore,
		commands.WithWriteAuthModelLogger(s.logger),
		commands.WithWriteAuthModelMaxSizeInBytes(s.maxAuthorizationModelSizeInBytes),
	)
	res, err := c.Execute(ctx, req)
	if err != nil {
		return nil, err
	}

	s.transport.SetHeader(ctx, httpmiddleware.XHttpCode, strconv.Itoa(http.StatusCreated))

	return res, nil
}

func (s *Server) ReadAuthorizationModels(ctx context.Context, req *openfgav1.ReadAuthorizationModelsRequest) (*openfgav1.ReadAuthorizationModelsResponse, error) {
	ctx, span := tracer.Start(ctx, "ReadAuthorizationModels")
	defer span.End()

	if !validator.RequestIsValidatedFromContext(ctx) {
		if err := req.Validate(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	ctx = telemetry.ContextWithRPCInfo(ctx, telemetry.RPCInfo{
		Service: s.serviceName,
		Method:  "ReadAuthorizationModels",
	})

	c := commands.NewReadAuthorizationModelsQuery(s.datastore,
		commands.WithReadAuthModelsQueryLogger(s.logger),
		commands.WithReadAuthModelsQueryEncoder(s.encoder),
	)
	return c.Execute(ctx, req)
}

func (s *Server) WriteAssertions(ctx context.Context, req *openfgav1.WriteAssertionsRequest) (*openfgav1.WriteAssertionsResponse, error) {
	ctx, span := tracer.Start(ctx, "WriteAssertions")
	defer span.End()

	if !validator.RequestIsValidatedFromContext(ctx) {
		if err := req.Validate(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	ctx = telemetry.ContextWithRPCInfo(ctx, telemetry.RPCInfo{
		Service: s.serviceName,
		Method:  "WriteAssertions",
	})

	storeID := req.GetStoreId()

	typesys, err := s.resolveTypesystem(ctx, storeID, req.GetAuthorizationModelId())
	if err != nil {
		return nil, err
	}

	c := commands.NewWriteAssertionsCommand(s.datastore, commands.WithWriteAssertCmdLogger(s.logger))
	res, err := c.Execute(ctx, &openfgav1.WriteAssertionsRequest{
		StoreId:              storeID,
		AuthorizationModelId: typesys.GetAuthorizationModelID(), // the resolved model id
		Assertions:           req.GetAssertions(),
	})
	if err != nil {
		return nil, err
	}

	s.transport.SetHeader(ctx, httpmiddleware.XHttpCode, strconv.Itoa(http.StatusNoContent))

	return res, nil
}

func (s *Server) ReadAssertions(ctx context.Context, req *openfgav1.ReadAssertionsRequest) (*openfgav1.ReadAssertionsResponse, error) {
	ctx, span := tracer.Start(ctx, "ReadAssertions")
	defer span.End()

	if !validator.RequestIsValidatedFromContext(ctx) {
		if err := req.Validate(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	ctx = telemetry.ContextWithRPCInfo(ctx, telemetry.RPCInfo{
		Service: s.serviceName,
		Method:  "ReadAssertions",
	})

	typesys, err := s.resolveTypesystem(ctx, req.GetStoreId(), req.GetAuthorizationModelId())
	if err != nil {
		return nil, err
	}

	q := commands.NewReadAssertionsQuery(s.datastore, commands.WithReadAssertionsQueryLogger(s.logger))
	return q.Execute(ctx, req.GetStoreId(), typesys.GetAuthorizationModelID())
}

func (s *Server) ReadChanges(ctx context.Context, req *openfgav1.ReadChangesRequest) (*openfgav1.ReadChangesResponse, error) {
	ctx, span := tracer.Start(ctx, "ReadChangesQuery", trace.WithAttributes(
		attribute.KeyValue{Key: "type", Value: attribute.StringValue(req.GetType())},
	))
	defer span.End()

	if !validator.RequestIsValidatedFromContext(ctx) {
		if err := req.Validate(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	ctx = telemetry.ContextWithRPCInfo(ctx, telemetry.RPCInfo{
		Service: s.serviceName,
		Method:  "ReadChanges",
	})

	q := commands.NewReadChangesQuery(s.datastore,
		commands.WithReadChangesQueryLogger(s.logger),
		commands.WithReadChangesQueryEncoder(s.encoder),
		commands.WithReadChangeQueryHorizonOffset(s.changelogHorizonOffset),
	)
	return q.Execute(ctx, req)
}

func (s *Server) CreateStore(ctx context.Context, req *openfgav1.CreateStoreRequest) (*openfgav1.CreateStoreResponse, error) {
	ctx, span := tracer.Start(ctx, "CreateStore")
	defer span.End()

	if !validator.RequestIsValidatedFromContext(ctx) {
		if err := req.Validate(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	ctx = telemetry.ContextWithRPCInfo(ctx, telemetry.RPCInfo{
		Service: s.serviceName,
		Method:  "CreateStore",
	})

	c := commands.NewCreateStoreCommand(s.datastore, commands.WithCreateStoreCmdLogger(s.logger))
	res, err := c.Execute(ctx, req)
	if err != nil {
		return nil, err
	}

	s.transport.SetHeader(ctx, httpmiddleware.XHttpCode, strconv.Itoa(http.StatusCreated))

	return res, nil
}

func (s *Server) DeleteStore(ctx context.Context, req *openfgav1.DeleteStoreRequest) (*openfgav1.DeleteStoreResponse, error) {
	ctx, span := tracer.Start(ctx, "DeleteStore")
	defer span.End()

	if !validator.RequestIsValidatedFromContext(ctx) {
		if err := req.Validate(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	ctx = telemetry.ContextWithRPCInfo(ctx, telemetry.RPCInfo{
		Service: s.serviceName,
		Method:  "DeleteStore",
	})

	cmd := commands.NewDeleteStoreCommand(s.datastore, commands.WithDeleteStoreCmdLogger(s.logger))
	res, err := cmd.Execute(ctx, req)
	if err != nil {
		return nil, err
	}

	s.transport.SetHeader(ctx, httpmiddleware.XHttpCode, strconv.Itoa(http.StatusNoContent))

	return res, nil
}

func (s *Server) GetStore(ctx context.Context, req *openfgav1.GetStoreRequest) (*openfgav1.GetStoreResponse, error) {
	ctx, span := tracer.Start(ctx, "GetStore")
	defer span.End()

	if !validator.RequestIsValidatedFromContext(ctx) {
		if err := req.Validate(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	ctx = telemetry.ContextWithRPCInfo(ctx, telemetry.RPCInfo{
		Service: s.serviceName,
		Method:  "GetStore",
	})

	q := commands.NewGetStoreQuery(s.datastore, commands.WithGetStoreQueryLogger(s.logger))
	return q.Execute(ctx, req)
}

func (s *Server) ListStores(ctx context.Context, req *openfgav1.ListStoresRequest) (*openfgav1.ListStoresResponse, error) {
	ctx, span := tracer.Start(ctx, "ListStores")
	defer span.End()

	if !validator.RequestIsValidatedFromContext(ctx) {
		if err := req.Validate(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	ctx = telemetry.ContextWithRPCInfo(ctx, telemetry.RPCInfo{
		Service: s.serviceName,
		Method:  "ListStores",
	})

	q := commands.NewListStoresQuery(s.datastore,
		commands.WithListStoresQueryLogger(s.logger),
		commands.WithListStoresQueryEncoder(s.encoder),
	)
	return q.Execute(ctx, req)
}

// IsReady reports whether the datastore is ready. Please see the implementation of [[storage.OpenFGADatastore.IsReady]]
// for your datastore.
func (s *Server) IsReady(ctx context.Context) (bool, error) {
	// for now we only depend on the datastore being ready, but in the future
	// server readiness may also depend on other criteria in addition to the
	// datastore being ready.

	status, err := s.datastore.IsReady(ctx)
	if err != nil {
		return false, err
	}

	if status.IsReady {
		return true, nil
	}

	s.logger.WarnWithContext(ctx, "datastore is not ready", zap.Any("status", status.Message))
	return false, nil
}

// resolveTypesystem resolves the underlying TypeSystem given the storeID and modelID and
// it sets some response metadata based on the model resolution.
func (s *Server) resolveTypesystem(ctx context.Context, storeID, modelID string) (*typesystem.TypeSystem, error) {
	ctx, span := tracer.Start(ctx, "resolveTypesystem")
	defer span.End()

	typesys, err := s.typesystemResolver(ctx, storeID, modelID)
	if err != nil {
		if errors.Is(err, typesystem.ErrModelNotFound) {
			if modelID == "" {
				return nil, serverErrors.LatestAuthorizationModelNotFound(storeID)
			}

			return nil, serverErrors.AuthorizationModelNotFound(modelID)
		}

		if errors.Is(err, typesystem.ErrInvalidModel) {
			return nil, serverErrors.ValidationError(err)
		}

		err = serverErrors.HandleError("", err)
		telemetry.TraceError(span, err)
		return nil, err
	}

	resolvedModelID := typesys.GetAuthorizationModelID()

	span.SetAttributes(attribute.KeyValue{Key: authorizationModelIDKey, Value: attribute.StringValue(resolvedModelID)})
	grpc_ctxtags.Extract(ctx).Set(authorizationModelIDKey, resolvedModelID)
	s.transport.SetHeader(ctx, AuthorizationModelIDHeader, resolvedModelID)

	return typesys, nil
}
