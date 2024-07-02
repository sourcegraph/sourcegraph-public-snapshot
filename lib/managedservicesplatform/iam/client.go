package iam

import (
	"context"
	"strings"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"github.com/redis/go-redis/v9"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

var iamTracer = otel.Tracer("msp/iam")

// ClientV1 provides helpers to interact with MSP IAM framework v1.
type ClientV1 struct {
	server openfgav1.OpenFGAServiceServer
	// storeID is the OpenFGA-server-generated ID of the store.
	storeID string
	// authorizationModelID is the OpenFGA-server-generated ID of the authorization
	// model.
	authorizationModelID string
}

type ClientV1Config struct {
	// StoreName is the name of the store to create. Changing the name of the store
	// will cause the IAM to create a new store without migrating the data. It is
	// recommended to be machine-friendly, e.g. "enterprise-portal".
	StoreName string
	// AuthorizationModelDSL is the DSL to define the authorization model. See
	// https://openfga.dev/docs/configuration-language for documentation.
	AuthorizationModelDSL string
}

func (opts ClientV1Config) validate() error {
	if opts.StoreName == "" {
		return errors.New("store name is required")
	}
	if len(opts.AuthorizationModelDSL) == 0 {
		return errors.New("authorization model DSL is required")
	}
	return nil
}

// NewClientV1 initializes and returns a new MSP IAM client by initializing with
// the given configuration. The returned `close` function should be called upon
// service shutdown.
func NewClientV1(ctx context.Context, logger log.Logger, contract runtime.Contract, redisClient *redis.Client, opts ClientV1Config) (_ *ClientV1, close func(), err error) {
	ctx, span := iamTracer.Start(ctx, "iam.NewClientV1")
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	if err = opts.validate(); err != nil {
		return nil, nil, err
	}
	opts.AuthorizationModelDSL = strings.TrimSpace(opts.AuthorizationModelDSL)

	sqlDB, err := contract.PostgreSQL.OpenDatabase(ctx, databaseName)
	if err != nil {
		return nil, nil, errors.Wrap(err, "open database")
	}

	metadata, err := migrateAndReconcile(ctx, logger, sqlDB, redisClient)
	if err != nil {
		if !contract.MSP && strings.Contains(err.Error(), "(SQLSTATE 3D000)") {
			return nil, nil, errors.Newf("database '%[1]s' not found, please run 'createdb -h $PGHOST -p $PGPORT -U $PGUSER %[1]s'", databaseName)
		}
		return nil, nil, errors.Wrap(err, "migrate and recon")
	}

	srv, err := newOpenFGAServer(logger, sqlDB)
	defer func() {
		// Proactively close the server if there is an error to be returned.
		if err != nil {
			srv.Close()
		}
	}()

	storeID, authorizationModelID, err := initServerSetup(
		ctx,
		initServerSetupOptions{
			Logger:                logger,
			DB:                    sqlDB,
			Server:                srv,
			StoreName:             opts.StoreName,
			AuthorizationModelDSL: opts.AuthorizationModelDSL,
			Metadata:              metadata,
		},
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "init server setup")
	}

	return &ClientV1{
		server:               srv,
		storeID:              storeID,
		authorizationModelID: authorizationModelID,
	}, func() { srv.Close() }, nil
}

type ListObjectsOptions struct {
	Type     TupleType
	Relation TupleRelation
	Subject  TupleSubject
}

func (l ListObjectsOptions) Validate() error {
	var errs error
	if l.Type == "" {
		errs = errors.Append(errs, errors.New("unknown type"))
	}
	if l.Relation == "" {
		errs = errors.Append(errs, errors.New("unknown relation"))
	}
	if l.Subject == "" {
		errs = errors.Append(errs, errors.New("unknown subject"))
	}
	return errs
}

// ListObjects returns a list of object IDs that satisfy the given options.
func (c *ClientV1) ListObjects(ctx context.Context, opts ListObjectsOptions) ([]string, error) {
	resp, err := c.server.ListObjects(
		ctx,
		&openfgav1.ListObjectsRequest{
			StoreId:              c.storeID,
			AuthorizationModelId: c.authorizationModelID,
			Type:                 string(opts.Type),
			Relation:             string(opts.Relation),
			User:                 string(opts.Subject),
		},
	)
	if err != nil {
		return nil, err
	}
	return resp.GetObjects(), nil
}

type TupleKey struct {
	Object        TupleObject
	TupleRelation TupleRelation
	Subject       TupleSubject
}

type WriteOptions struct {
	Writes  []TupleKey
	Deletes []TupleKey
}

// Write upserts and/or deletes the relation tuples.
func (c *ClientV1) Write(ctx context.Context, opts WriteOptions) error {
	writes := make([]*openfgav1.TupleKey, 0, len(opts.Writes))
	for _, w := range opts.Writes {
		writes = append(
			writes,
			&openfgav1.TupleKey{
				User:     string(w.Subject),
				Relation: string(w.TupleRelation),
				Object:   string(w.Object),
			},
		)
	}

	deletes := make([]*openfgav1.TupleKeyWithoutCondition, 0, len(opts.Deletes))
	for _, d := range opts.Deletes {
		deletes = append(
			deletes,
			&openfgav1.TupleKeyWithoutCondition{
				User:     string(d.Subject),
				Relation: string(d.TupleRelation),
				Object:   string(d.Object),
			},
		)
	}

	var requestWrites *openfgav1.WriteRequestWrites
	var requestDeletes *openfgav1.WriteRequestDeletes
	if len(writes) > 0 {
		requestWrites = &openfgav1.WriteRequestWrites{TupleKeys: writes}
	}
	if len(deletes) > 0 {
		requestDeletes = &openfgav1.WriteRequestDeletes{TupleKeys: deletes}
	}
	if requestWrites == nil && requestDeletes == nil {
		return nil
	}

	_, err := c.server.Write(
		ctx,
		&openfgav1.WriteRequest{
			StoreId:              c.storeID,
			AuthorizationModelId: c.authorizationModelID,
			Writes:               requestWrites,
			Deletes:              requestDeletes,
		},
	)
	if err != nil {
		return err
	}
	return nil
}

type CheckOptions struct {
	TupleKey TupleKey
}

// Check checks whether a relationship exists (thus permission allowed) using
// the given tuple key as the check condition.
func (c *ClientV1) Check(ctx context.Context, opts CheckOptions) (allowed bool, _ error) {
	resp, err := c.server.Check(
		ctx,
		&openfgav1.CheckRequest{
			StoreId:              c.storeID,
			AuthorizationModelId: c.authorizationModelID,
			TupleKey: &openfgav1.CheckRequestTupleKey{
				User:     string(opts.TupleKey.Subject),
				Relation: string(opts.TupleKey.TupleRelation),
				Object:   string(opts.TupleKey.Object),
			},
		},
	)
	if err != nil {
		return false, err
	}
	return resp.GetAllowed(), nil
}
