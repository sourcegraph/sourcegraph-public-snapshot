package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/cenkalti/backoff/v4"
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver.
	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/openfga/openfga/pkg/logger"
	"github.com/openfga/openfga/pkg/storage"
	"github.com/openfga/openfga/pkg/storage/sqlcommon"
	tupleUtils "github.com/openfga/openfga/pkg/tuple"
)

var tracer = otel.Tracer("openfga/pkg/storage/postgres")

// Postgres provides a Postgres based implementation of [storage.OpenFGADatastore].
type Postgres struct {
	stbl                   sq.StatementBuilderType
	db                     *sql.DB
	dbInfo                 *sqlcommon.DBInfo
	logger                 logger.Logger
	dbStatsCollector       prometheus.Collector
	maxTuplesPerWriteField int
	maxTypesPerModelField  int
}

// Ensures that Postgres implements the OpenFGADatastore interface.
var _ storage.OpenFGADatastore = (*Postgres)(nil)

// New creates a new [Postgres] storage.
func New(uri string, cfg *sqlcommon.Config) (*Postgres, error) {
	if cfg.Username != "" || cfg.Password != "" {
		parsed, err := url.Parse(uri)
		if err != nil {
			return nil, fmt.Errorf("parse postgres connection uri: %w", err)
		}

		username := ""
		if cfg.Username != "" {
			username = cfg.Username
		} else if parsed.User != nil {
			username = parsed.User.Username()
		}

		switch {
		case cfg.Password != "":
			parsed.User = url.UserPassword(username, cfg.Password)
		case parsed.User != nil:
			if password, ok := parsed.User.Password(); ok {
				parsed.User = url.UserPassword(username, password)
			} else {
				parsed.User = url.User(username)
			}
		default:
			parsed.User = url.User(username)
		}

		uri = parsed.String()
	}

	db, err := sql.Open("pgx", uri)
	if err != nil {
		return nil, fmt.Errorf("initialize postgres connection: %w", err)
	}
	return NewWithDB(db, cfg)
}

// NewWithDB creates a new [Postgres] storage with the provided database connection.
func NewWithDB(db *sql.DB, cfg *sqlcommon.Config) (*Postgres, error) {
	if cfg.MaxOpenConns != 0 {
		db.SetMaxOpenConns(cfg.MaxOpenConns)
	}

	if cfg.MaxIdleConns != 0 {
		db.SetMaxIdleConns(cfg.MaxIdleConns)
	}

	if cfg.ConnMaxIdleTime != 0 {
		db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	}

	if cfg.ConnMaxLifetime != 0 {
		db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}

	policy := backoff.NewExponentialBackOff()
	policy.MaxElapsedTime = 1 * time.Minute
	attempt := 1
	err := backoff.Retry(func() error {
		err := db.PingContext(context.Background())
		if err != nil {
			cfg.Logger.Info("waiting for postgres", zap.Int("attempt", attempt))
			attempt++
			return err
		}
		return nil
	}, policy)
	if err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}

	var collector prometheus.Collector
	if cfg.ExportMetrics {
		collector = collectors.NewDBStatsCollector(db, "openfga")
		if err := prometheus.Register(collector); err != nil {
			return nil, fmt.Errorf("initialize metrics: %w", err)
		}
	}
	stbl := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(db)
	dbInfo := sqlcommon.NewDBInfo(db, stbl, sq.Expr("NOW()"))

	return &Postgres{
		stbl:                   stbl,
		db:                     db,
		dbInfo:                 dbInfo,
		logger:                 cfg.Logger,
		dbStatsCollector:       collector,
		maxTuplesPerWriteField: cfg.MaxTuplesPerWriteField,
		maxTypesPerModelField:  cfg.MaxTypesPerModelField,
	}, nil
}

// Close see [storage.OpenFGADatastore].Close.
func (p *Postgres) Close() {
	if p.dbStatsCollector != nil {
		prometheus.Unregister(p.dbStatsCollector)
	}
	p.db.Close()
}

// Read see [storage.RelationshipTupleReader].Read.
func (p *Postgres) Read(ctx context.Context, store string, tupleKey *openfgav1.TupleKey) (storage.TupleIterator, error) {
	ctx, span := tracer.Start(ctx, "postgres.Read")
	defer span.End()

	return p.read(ctx, store, tupleKey, nil)
}

// ReadPage see [storage.RelationshipTupleReader].ReadPage.
func (p *Postgres) ReadPage(ctx context.Context, store string, tupleKey *openfgav1.TupleKey, opts storage.PaginationOptions) ([]*openfgav1.Tuple, []byte, error) {
	ctx, span := tracer.Start(ctx, "postgres.ReadPage")
	defer span.End()

	iter, err := p.read(ctx, store, tupleKey, &opts)
	if err != nil {
		return nil, nil, err
	}
	defer iter.Stop()

	return iter.ToArray(opts)
}

func (p *Postgres) read(ctx context.Context, store string, tupleKey *openfgav1.TupleKey, opts *storage.PaginationOptions) (*sqlcommon.SQLTupleIterator, error) {
	ctx, span := tracer.Start(ctx, "postgres.read")
	defer span.End()

	sb := p.stbl.
		Select(
			"store", "object_type", "object_id", "relation", "_user",
			"condition_name", "condition_context", "ulid", "inserted_at",
		).
		From("tuple").
		Where(sq.Eq{"store": store})
	if opts != nil {
		sb = sb.OrderBy("ulid")
	}

	objectType, objectID := tupleUtils.SplitObject(tupleKey.GetObject())
	if objectType != "" {
		sb = sb.Where(sq.Eq{"object_type": objectType})
	}
	if objectID != "" {
		sb = sb.Where(sq.Eq{"object_id": objectID})
	}
	if tupleKey.GetRelation() != "" {
		sb = sb.Where(sq.Eq{"relation": tupleKey.GetRelation()})
	}
	if tupleKey.GetUser() != "" {
		sb = sb.Where(sq.Eq{"_user": tupleKey.GetUser()})
	}
	if opts != nil && opts.From != "" {
		token, err := sqlcommon.UnmarshallContToken(opts.From)
		if err != nil {
			return nil, err
		}
		sb = sb.Where(sq.GtOrEq{"ulid": token.Ulid})
	}
	if opts != nil && opts.PageSize != 0 {
		sb = sb.Limit(uint64(opts.PageSize + 1)) // + 1 is used to determine whether to return a continuation token.
	}

	rows, err := sb.QueryContext(ctx)
	if err != nil {
		return nil, sqlcommon.HandleSQLError(err)
	}

	return sqlcommon.NewSQLTupleIterator(rows), nil
}

// Write see [storage.RelationshipTupleWriter].Write.
func (p *Postgres) Write(ctx context.Context, store string, deletes storage.Deletes, writes storage.Writes) error {
	ctx, span := tracer.Start(ctx, "postgres.Write")
	defer span.End()

	if len(deletes)+len(writes) > p.MaxTuplesPerWrite() {
		return storage.ErrExceededWriteBatchLimit
	}

	now := time.Now().UTC()
	return sqlcommon.Write(ctx, p.dbInfo, store, deletes, writes, now)
}

// ReadUserTuple see [storage.RelationshipTupleReader].ReadUserTuple.
func (p *Postgres) ReadUserTuple(ctx context.Context, store string, tupleKey *openfgav1.TupleKey) (*openfgav1.Tuple, error) {
	ctx, span := tracer.Start(ctx, "postgres.ReadUserTuple")
	defer span.End()

	objectType, objectID := tupleUtils.SplitObject(tupleKey.GetObject())
	userType := tupleUtils.GetUserTypeFromUser(tupleKey.GetUser())

	var conditionName sql.NullString
	var conditionContext []byte
	var record storage.TupleRecord

	err := p.stbl.
		Select(
			"object_type", "object_id", "relation", "_user",
			"condition_name", "condition_context",
		).
		From("tuple").
		Where(sq.Eq{
			"store":       store,
			"object_type": objectType,
			"object_id":   objectID,
			"relation":    tupleKey.GetRelation(),
			"_user":       tupleKey.GetUser(),
			"user_type":   userType,
		}).
		QueryRowContext(ctx).
		Scan(
			&record.ObjectType,
			&record.ObjectID,
			&record.Relation,
			&record.User,
			&conditionName,
			&conditionContext,
		)
	if err != nil {
		return nil, sqlcommon.HandleSQLError(err)
	}

	if conditionName.String != "" {
		record.ConditionName = conditionName.String

		if conditionContext != nil {
			var conditionContextStruct structpb.Struct
			if err := proto.Unmarshal(conditionContext, &conditionContextStruct); err != nil {
				return nil, err
			}
			record.ConditionContext = &conditionContextStruct
		}
	}

	return record.AsTuple(), nil
}

// ReadUsersetTuples see [storage.RelationshipTupleReader].ReadUsersetTuples.
func (p *Postgres) ReadUsersetTuples(ctx context.Context, store string, filter storage.ReadUsersetTuplesFilter) (storage.TupleIterator, error) {
	ctx, span := tracer.Start(ctx, "postgres.ReadUsersetTuples")
	defer span.End()

	sb := p.stbl.
		Select(
			"store", "object_type", "object_id", "relation", "_user",
			"condition_name", "condition_context", "ulid", "inserted_at",
		).
		From("tuple").
		Where(sq.Eq{"store": store}).
		Where(sq.Eq{"user_type": tupleUtils.UserSet})

	objectType, objectID := tupleUtils.SplitObject(filter.Object)
	if objectType != "" {
		sb = sb.Where(sq.Eq{"object_type": objectType})
	}
	if objectID != "" {
		sb = sb.Where(sq.Eq{"object_id": objectID})
	}
	if filter.Relation != "" {
		sb = sb.Where(sq.Eq{"relation": filter.Relation})
	}
	if len(filter.AllowedUserTypeRestrictions) > 0 {
		orConditions := sq.Or{}
		for _, userset := range filter.AllowedUserTypeRestrictions {
			if _, ok := userset.GetRelationOrWildcard().(*openfgav1.RelationReference_Relation); ok {
				orConditions = append(orConditions, sq.Like{"_user": userset.GetType() + ":%#" + userset.GetRelation()})
			}
			if _, ok := userset.GetRelationOrWildcard().(*openfgav1.RelationReference_Wildcard); ok {
				orConditions = append(orConditions, sq.Eq{"_user": userset.GetType() + ":*"})
			}
		}
		sb = sb.Where(orConditions)
	}
	rows, err := sb.QueryContext(ctx)
	if err != nil {
		return nil, sqlcommon.HandleSQLError(err)
	}

	return sqlcommon.NewSQLTupleIterator(rows), nil
}

// ReadStartingWithUser see [storage.RelationshipTupleReader].ReadStartingWithUser.
func (p *Postgres) ReadStartingWithUser(ctx context.Context, store string, opts storage.ReadStartingWithUserFilter) (storage.TupleIterator, error) {
	ctx, span := tracer.Start(ctx, "postgres.ReadStartingWithUser")
	defer span.End()

	var targetUsersArg []string
	for _, u := range opts.UserFilter {
		targetUser := u.GetObject()
		if u.GetRelation() != "" {
			targetUser = strings.Join([]string{u.GetObject(), u.GetRelation()}, "#")
		}
		targetUsersArg = append(targetUsersArg, targetUser)
	}

	rows, err := p.stbl.
		Select(
			"store", "object_type", "object_id", "relation", "_user",
			"condition_name", "condition_context", "ulid", "inserted_at",
		).
		From("tuple").
		Where(sq.Eq{
			"store":       store,
			"object_type": opts.ObjectType,
			"relation":    opts.Relation,
			"_user":       targetUsersArg,
		}).QueryContext(ctx)
	if err != nil {
		return nil, sqlcommon.HandleSQLError(err)
	}

	return sqlcommon.NewSQLTupleIterator(rows), nil
}

// MaxTuplesPerWrite see [storage.RelationshipTupleWriter].MaxTuplesPerWrite.
func (p *Postgres) MaxTuplesPerWrite() int {
	return p.maxTuplesPerWriteField
}

// ReadAuthorizationModel see [storage.AuthorizationModelReadBackend].ReadAuthorizationModel.
func (p *Postgres) ReadAuthorizationModel(ctx context.Context, store string, modelID string) (*openfgav1.AuthorizationModel, error) {
	ctx, span := tracer.Start(ctx, "postgres.ReadAuthorizationModel")
	defer span.End()

	return sqlcommon.ReadAuthorizationModel(ctx, p.dbInfo, store, modelID)
}

// ReadAuthorizationModels see [storage.AuthorizationModelReadBackend].ReadAuthorizationModels.
func (p *Postgres) ReadAuthorizationModels(ctx context.Context, store string, opts storage.PaginationOptions) ([]*openfgav1.AuthorizationModel, []byte, error) {
	ctx, span := tracer.Start(ctx, "postgres.ReadAuthorizationModels")
	defer span.End()

	sb := p.stbl.
		Select("authorization_model_id").
		Distinct().
		From("authorization_model").
		Where(sq.Eq{"store": store}).
		OrderBy("authorization_model_id desc")

	if opts.From != "" {
		token, err := sqlcommon.UnmarshallContToken(opts.From)
		if err != nil {
			return nil, nil, err
		}
		sb = sb.Where(sq.LtOrEq{"authorization_model_id": token.Ulid})
	}
	if opts.PageSize > 0 {
		sb = sb.Limit(uint64(opts.PageSize + 1)) // + 1 is used to determine whether to return a continuation token.
	}

	rows, err := sb.QueryContext(ctx)
	if err != nil {
		return nil, nil, sqlcommon.HandleSQLError(err)
	}
	defer rows.Close()

	var modelIDs []string
	var modelID string

	for rows.Next() {
		err = rows.Scan(&modelID)
		if err != nil {
			return nil, nil, sqlcommon.HandleSQLError(err)
		}

		modelIDs = append(modelIDs, modelID)
	}

	if err := rows.Err(); err != nil {
		return nil, nil, sqlcommon.HandleSQLError(err)
	}

	var token []byte
	numModelIDs := len(modelIDs)
	if len(modelIDs) > opts.PageSize {
		numModelIDs = opts.PageSize
		token, err = json.Marshal(sqlcommon.NewContToken(modelID, ""))
		if err != nil {
			return nil, nil, err
		}
	}

	// TODO: make this concurrent with a maximum of 5 goroutines. This may be helpful:
	// https://stackoverflow.com/questions/25306073/always-have-x-number-of-goroutines-running-at-any-time
	models := make([]*openfgav1.AuthorizationModel, 0, numModelIDs)
	// We use numModelIDs here to avoid retrieving possibly one extra model.
	for i := 0; i < numModelIDs; i++ {
		model, err := p.ReadAuthorizationModel(ctx, store, modelIDs[i])
		if err != nil {
			return nil, nil, err
		}
		models = append(models, model)
	}

	return models, token, nil
}

// FindLatestAuthorizationModel see [storage.AuthorizationModelReadBackend].FindLatestAuthorizationModel.
func (p *Postgres) FindLatestAuthorizationModel(ctx context.Context, store string) (*openfgav1.AuthorizationModel, error) {
	ctx, span := tracer.Start(ctx, "postgres.FindLatestAuthorizationModel")
	defer span.End()

	return sqlcommon.FindLatestAuthorizationModel(ctx, p.dbInfo, store)
}

// MaxTypesPerAuthorizationModel see [storage.TypeDefinitionWriteBackend].MaxTypesPerAuthorizationModel.
func (p *Postgres) MaxTypesPerAuthorizationModel() int {
	return p.maxTypesPerModelField
}

// WriteAuthorizationModel see [storage.TypeDefinitionWriteBackend].WriteAuthorizationModel.
func (p *Postgres) WriteAuthorizationModel(ctx context.Context, store string, model *openfgav1.AuthorizationModel) error {
	ctx, span := tracer.Start(ctx, "postgres.WriteAuthorizationModel")
	defer span.End()

	typeDefinitions := model.GetTypeDefinitions()

	if len(typeDefinitions) > p.MaxTypesPerAuthorizationModel() {
		return storage.ExceededMaxTypeDefinitionsLimitError(p.maxTypesPerModelField)
	}

	return sqlcommon.WriteAuthorizationModel(ctx, p.dbInfo, store, model)
}

// CreateStore adds a new store to the Postgres storage.
func (p *Postgres) CreateStore(ctx context.Context, store *openfgav1.Store) (*openfgav1.Store, error) {
	ctx, span := tracer.Start(ctx, "postgres.CreateStore")
	defer span.End()

	var id, name string
	var createdAt time.Time
	err := p.stbl.
		Insert("store").
		Columns("id", "name", "created_at", "updated_at").
		Values(store.GetId(), store.GetName(), "NOW()", "NOW()").
		Suffix("returning id, name, created_at").
		QueryRowContext(ctx).
		Scan(&id, &name, &createdAt)
	if err != nil {
		return nil, sqlcommon.HandleSQLError(err)
	}

	return &openfgav1.Store{
		Id:        id,
		Name:      name,
		CreatedAt: timestamppb.New(createdAt),
		UpdatedAt: timestamppb.New(createdAt),
	}, nil
}

// GetStore retrieves the details of a specific store from the Postgres using its storeID.
func (p *Postgres) GetStore(ctx context.Context, id string) (*openfgav1.Store, error) {
	ctx, span := tracer.Start(ctx, "postgres.GetStore")
	defer span.End()

	row := p.stbl.
		Select("id", "name", "created_at", "updated_at").
		From("store").
		Where(sq.Eq{
			"id":         id,
			"deleted_at": nil,
		}).
		QueryRowContext(ctx)

	var storeID, name string
	var createdAt, updatedAt time.Time
	err := row.Scan(&storeID, &name, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, sqlcommon.HandleSQLError(err)
	}

	return &openfgav1.Store{
		Id:        storeID,
		Name:      name,
		CreatedAt: timestamppb.New(createdAt),
		UpdatedAt: timestamppb.New(updatedAt),
	}, nil
}

// ListStores provides a paginated list of all stores present in the Postgres storage.
func (p *Postgres) ListStores(ctx context.Context, opts storage.PaginationOptions) ([]*openfgav1.Store, []byte, error) {
	ctx, span := tracer.Start(ctx, "postgres.ListStores")
	defer span.End()

	sb := p.stbl.
		Select("id", "name", "created_at", "updated_at").
		From("store").
		Where(sq.Eq{"deleted_at": nil}).
		OrderBy("id")

	if opts.From != "" {
		token, err := sqlcommon.UnmarshallContToken(opts.From)
		if err != nil {
			return nil, nil, err
		}
		sb = sb.Where(sq.GtOrEq{"id": token.Ulid})
	}
	if opts.PageSize > 0 {
		sb = sb.Limit(uint64(opts.PageSize + 1)) // + 1 is used to determine whether to return a continuation token.
	}

	rows, err := sb.QueryContext(ctx)
	if err != nil {
		return nil, nil, sqlcommon.HandleSQLError(err)
	}
	defer rows.Close()

	var stores []*openfgav1.Store
	var id string
	for rows.Next() {
		var name string
		var createdAt, updatedAt time.Time
		err := rows.Scan(&id, &name, &createdAt, &updatedAt)
		if err != nil {
			return nil, nil, sqlcommon.HandleSQLError(err)
		}

		stores = append(stores, &openfgav1.Store{
			Id:        id,
			Name:      name,
			CreatedAt: timestamppb.New(createdAt),
			UpdatedAt: timestamppb.New(updatedAt),
		})
	}

	if err := rows.Err(); err != nil {
		return nil, nil, sqlcommon.HandleSQLError(err)
	}

	if len(stores) > opts.PageSize {
		contToken, err := json.Marshal(sqlcommon.NewContToken(id, ""))
		if err != nil {
			return nil, nil, err
		}
		return stores[:opts.PageSize], contToken, nil
	}

	return stores, nil, nil
}

// DeleteStore removes a store from the Postgres storage.
func (p *Postgres) DeleteStore(ctx context.Context, id string) error {
	ctx, span := tracer.Start(ctx, "postgres.DeleteStore")
	defer span.End()

	_, err := p.stbl.
		Update("store").
		Set("deleted_at", "NOW()").
		Where(sq.Eq{"id": id}).
		ExecContext(ctx)
	if err != nil {
		return sqlcommon.HandleSQLError(err)
	}

	return nil
}

// WriteAssertions see [storage.AssertionsBackend].WriteAssertions.
func (p *Postgres) WriteAssertions(ctx context.Context, store, modelID string, assertions []*openfgav1.Assertion) error {
	ctx, span := tracer.Start(ctx, "postgres.WriteAssertions")
	defer span.End()

	marshalledAssertions, err := proto.Marshal(&openfgav1.Assertions{Assertions: assertions})
	if err != nil {
		return err
	}

	_, err = p.stbl.
		Insert("assertion").
		Columns("store", "authorization_model_id", "assertions").
		Values(store, modelID, marshalledAssertions).
		Suffix("ON CONFLICT (store, authorization_model_id) DO UPDATE SET assertions = ?", marshalledAssertions).
		ExecContext(ctx)
	if err != nil {
		return sqlcommon.HandleSQLError(err)
	}

	return nil
}

// ReadAssertions see [storage.AssertionsBackend].ReadAssertions.
func (p *Postgres) ReadAssertions(ctx context.Context, store, modelID string) ([]*openfgav1.Assertion, error) {
	ctx, span := tracer.Start(ctx, "postgres.ReadAssertions")
	defer span.End()

	var marshalledAssertions []byte
	err := p.stbl.
		Select("assertions").
		From("assertion").
		Where(sq.Eq{
			"store":                  store,
			"authorization_model_id": modelID,
		}).
		QueryRowContext(ctx).
		Scan(&marshalledAssertions)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []*openfgav1.Assertion{}, nil
		}
		return nil, sqlcommon.HandleSQLError(err)
	}

	var assertions openfgav1.Assertions
	err = proto.Unmarshal(marshalledAssertions, &assertions)
	if err != nil {
		return nil, err
	}

	return assertions.GetAssertions(), nil
}

// ReadChanges see [storage.ChangelogBackend].ReadChanges.
func (p *Postgres) ReadChanges(
	ctx context.Context,
	store, objectTypeFilter string,
	opts storage.PaginationOptions,
	horizonOffset time.Duration,
) ([]*openfgav1.TupleChange, []byte, error) {
	ctx, span := tracer.Start(ctx, "postgres.ReadChanges")
	defer span.End()

	sb := p.stbl.
		Select(
			"ulid", "object_type", "object_id", "relation", "_user", "operation",
			"condition_name", "condition_context", "inserted_at",
		).
		From("changelog").
		Where(sq.Eq{"store": store}).
		Where(fmt.Sprintf("inserted_at < NOW() - interval '%dms'", horizonOffset.Milliseconds())).
		OrderBy("ulid asc")

	if objectTypeFilter != "" {
		sb = sb.Where(sq.Eq{"object_type": objectTypeFilter})
	}
	if opts.From != "" {
		token, err := sqlcommon.UnmarshallContToken(opts.From)
		if err != nil {
			return nil, nil, err
		}
		if token.ObjectType != objectTypeFilter {
			return nil, nil, storage.ErrMismatchObjectType
		}

		sb = sb.Where(sq.Gt{"ulid": token.Ulid}) // > as we always return a continuation token.
	}
	if opts.PageSize > 0 {
		sb = sb.Limit(uint64(opts.PageSize)) // + 1 is NOT used here as we always return a continuation token.
	}

	rows, err := sb.QueryContext(ctx)
	if err != nil {
		return nil, nil, sqlcommon.HandleSQLError(err)
	}
	defer rows.Close()

	var changes []*openfgav1.TupleChange
	var ulid string
	for rows.Next() {
		var objectType, objectID, relation, user string
		var operation int
		var insertedAt time.Time
		var conditionName sql.NullString
		var conditionContext []byte

		err = rows.Scan(
			&ulid,
			&objectType,
			&objectID,
			&relation,
			&user,
			&operation,
			&conditionName,
			&conditionContext,
			&insertedAt,
		)
		if err != nil {
			return nil, nil, sqlcommon.HandleSQLError(err)
		}

		var conditionContextStruct structpb.Struct
		if conditionName.String != "" {
			if conditionContext != nil {
				if err := proto.Unmarshal(conditionContext, &conditionContextStruct); err != nil {
					return nil, nil, err
				}
			}
		}

		tk := tupleUtils.NewTupleKeyWithCondition(
			tupleUtils.BuildObject(objectType, objectID),
			relation,
			user,
			conditionName.String,
			&conditionContextStruct,
		)

		changes = append(changes, &openfgav1.TupleChange{
			TupleKey:  tk,
			Operation: openfgav1.TupleOperation(operation),
			Timestamp: timestamppb.New(insertedAt.UTC()),
		})
	}

	if len(changes) == 0 {
		return nil, nil, storage.ErrNotFound
	}

	contToken, err := json.Marshal(sqlcommon.NewContToken(ulid, objectTypeFilter))
	if err != nil {
		return nil, nil, err
	}

	return changes, contToken, nil
}

// IsReady see [sqlcommon.IsReady].
func (p *Postgres) IsReady(ctx context.Context) (storage.ReadinessStatus, error) {
	return sqlcommon.IsReady(ctx, p.db)
}
