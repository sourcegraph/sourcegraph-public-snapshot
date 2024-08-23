package iam

import (
	"context"
	"database/sql"
	"math"
	"time"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"github.com/openfga/language/pkg/go/transformer"
	"github.com/openfga/openfga/pkg/server"
	"github.com/openfga/openfga/pkg/storage/postgres"
	"github.com/openfga/openfga/pkg/storage/sqlcommon"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/codes"
	oteltrace "go.opentelemetry.io/otel/trace"
	"go.uber.org/zap" //nolint:logging // dependencies require direct usage of zap

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func newOpenFGAServer(logger log.Logger, sqlDB *sql.DB) (*server.Server, error) {
	openfgaLogger := &openfgaLoggerShim{Logger: logger}

	// Defaults are all coming from github.com/openfga/internal/server/config/config.go,
	// we cannot directly use them because of "internal" package.
	datastore, err := postgres.NewWithDB(
		sqlDB,
		sqlcommon.NewConfig(
			sqlcommon.WithLogger(openfgaLogger),
			sqlcommon.WithMaxTuplesPerWrite(100),                   // Default
			sqlcommon.WithMaxTypesPerAuthorizationModel(256*1_024), // Default
			sqlcommon.WithMaxOpenConns(30),                         // Default
			sqlcommon.WithMaxIdleConns(10),                         // Default
			sqlcommon.WithConnMaxLifetime(time.Minute),
			sqlcommon.WithMetrics(),
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, "initialize postgres datastore")
	}

	// Defaults are all coming from github.com/openfga/internal/server/config/config.go,
	// we cannot directly use them because of "internal" package.
	srv, err := server.NewServerWithOpts(
		server.WithDatastore(datastore),
		server.WithAuthorizationModelCacheSize(100000), // Default
		server.WithLogger(openfgaLogger),
		server.WithResolveNodeLimit(25),                                            // Default
		server.WithResolveNodeBreadthLimit(100),                                    // Default
		server.WithChangelogHorizonOffset(0),                                       // Default
		server.WithListObjectsDeadline(3*time.Second),                              // Default
		server.WithListObjectsMaxResults(1000),                                     // Default
		server.WithListUsersDeadline(3*time.Second),                                // Default
		server.WithListUsersMaxResults(1000),                                       // Default
		server.WithMaxConcurrentReadsForListObjects(math.MaxUint32),                // Default
		server.WithMaxConcurrentReadsForCheck(math.MaxUint32),                      // Default
		server.WithMaxConcurrentReadsForListUsers(math.MaxUint32),                  // Default
		server.WithCheckQueryCacheEnabled(false),                                   // Default
		server.WithCheckQueryCacheLimit(10000),                                     // Default
		server.WithCheckQueryCacheTTL(10*time.Second),                              // Default
		server.WithRequestDurationByQueryHistogramBuckets([]uint{50, 200}),         // Default
		server.WithRequestDurationByDispatchCountHistogramBuckets([]uint{50, 200}), // Default
		server.WithMaxAuthorizationModelSizeInBytes(256*1_024),                     // Default
		server.WithDispatchThrottlingCheckResolverEnabled(false),                   // Default
		server.WithDispatchThrottlingCheckResolverFrequency(10*time.Microsecond),   // Default
		server.WithDispatchThrottlingCheckResolverThreshold(100),                   // Default
		server.WithDispatchThrottlingCheckResolverMaxThreshold(0),                  // Default, 0 means use the default threshold as max.
		server.WithListObjectsDispatchThrottlingEnabled(false),                     // Default
		server.WithListObjectsDispatchThrottlingFrequency(10*time.Microsecond),     // Default
		server.WithListObjectsDispatchThrottlingThreshold(100),                     // Default
		server.WithListObjectsDispatchThrottlingMaxThreshold(0),                    // Default, 0 means use the default threshold as max.
	)
	if err != nil {
		return nil, errors.Wrap(err, "initialize server")
	}
	return srv, nil
}

type openfgaLoggerShim struct {
	log.Logger
}

func (l *openfgaLoggerShim) Panic(s string, field ...zap.Field) {
	l.Error(s, field...)
}

func (l *openfgaLoggerShim) DebugWithContext(ctx context.Context, s string, field ...zap.Field) {
	l.WithTrace(traceContext(ctx)).Debug(s, field...)
}

func (l *openfgaLoggerShim) InfoWithContext(ctx context.Context, s string, field ...zap.Field) {
	l.WithTrace(traceContext(ctx)).Info(s, field...)
}

func (l *openfgaLoggerShim) WarnWithContext(ctx context.Context, s string, field ...zap.Field) {
	l.WithTrace(traceContext(ctx)).Warn(s, field...)
}

func (l *openfgaLoggerShim) ErrorWithContext(ctx context.Context, s string, field ...zap.Field) {
	l.WithTrace(traceContext(ctx)).Error(s, field...)
}

func (l *openfgaLoggerShim) PanicWithContext(ctx context.Context, s string, field ...zap.Field) {
	l.WithTrace(traceContext(ctx)).Error(s, field...)
}

func (l *openfgaLoggerShim) FatalWithContext(ctx context.Context, s string, field ...zap.Field) {
	l.WithTrace(traceContext(ctx)).Fatal(s, field...)
}

// traceContext retrieves the full trace context, if any, from context - this
// includes both TraceID and SpanID.
func traceContext(ctx context.Context) log.TraceContext {
	if otelSpan := oteltrace.SpanContextFromContext(ctx); otelSpan.IsValid() {
		return log.TraceContext{
			TraceID: otelSpan.TraceID().String(),
			SpanID:  otelSpan.SpanID().String(),
		}
	}

	// no span found
	return log.TraceContext{}
}

type initServerSetupOptions struct {
	Logger                log.Logger
	DB                    *sql.DB
	Server                *server.Server
	StoreName             string
	AuthorizationModelDSL string
	Metadata              *metadata
}

func (opts initServerSetupOptions) validate() error {
	if opts.Logger == nil {
		return errors.New("logger is required")
	}
	if opts.DB == nil {
		return errors.New("database is required")
	}
	if opts.Server == nil {
		return errors.New("server is required")
	}
	if opts.StoreName == "" {
		return errors.New("store name is required")
	}
	if opts.AuthorizationModelDSL == "" {
		return errors.New("authorization model DSL is required")
	}
	if opts.Metadata == nil {
		return errors.New("metadata is required")
	}
	return nil
}

func initServerSetup(ctx context.Context, opts initServerSetupOptions) (storeID, authorizationModelID string, err error) {
	ctx, span := iamTracer.Start(ctx, "iam.initServerSetup")
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	if err = opts.validate(); err != nil {
		return "", "", err
	}

	// Ensure the store exists.
	listStoresResp, err := opts.Server.ListStores(ctx, &openfgav1.ListStoresRequest{})
	if err != nil {
		return "", "", errors.Wrap(err, "list stores")
	}

	for _, store := range listStoresResp.GetStores() {
		if store.GetName() == opts.StoreName {
			storeID = store.GetId()
			break
		}
	}
	if storeID == "" {
		resp, err := opts.Server.CreateStore(
			ctx,
			&openfgav1.CreateStoreRequest{Name: opts.StoreName},
		)
		if err != nil {
			return "", "", errors.Wrap(err, "create store")
		}
		storeID = resp.GetId()
	}

	// Ensure the authorization model is up-to-date.
	if opts.Metadata.AuthorizationModelDSL == opts.AuthorizationModelDSL {
		return storeID, opts.Metadata.AuthorizationModelID, nil
	}

	proto, err := transformer.TransformDSLToProto(opts.AuthorizationModelDSL)
	if err != nil {
		return "", "", errors.Wrap(err, "transform DSL to proto")
	}

	writeAuthorizationModelResp, err := opts.Server.WriteAuthorizationModel(
		ctx,
		&openfgav1.WriteAuthorizationModelRequest{
			StoreId:         storeID,
			TypeDefinitions: proto.GetTypeDefinitions(),
			SchemaVersion:   proto.GetSchemaVersion(),
			Conditions:      nil,
		},
	)
	if err != nil {
		return "", "", errors.Wrap(err, "write authorization model")
	}
	authorizationModelID = writeAuthorizationModelResp.GetAuthorizationModelId()

	_, err = opts.DB.ExecContext(
		ctx,
		"UPDATE metadata SET authorization_model_id = $1, authorization_model_dsl = $2, updated_at = $3",
		authorizationModelID,
		opts.AuthorizationModelDSL,
		time.Now().UTC(),
	)
	if err != nil {
		return "", "", errors.Wrap(err, "update metadata")
	}
	return storeID, authorizationModelID, nil
}
