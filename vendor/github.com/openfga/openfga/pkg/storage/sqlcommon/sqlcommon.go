package sqlcommon

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/go-sql-driver/mysql"
	"github.com/oklog/ulid/v2"
	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"github.com/pressly/goose/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/openfga/openfga/internal/build"
	"github.com/openfga/openfga/pkg/logger"
	"github.com/openfga/openfga/pkg/storage"
	tupleUtils "github.com/openfga/openfga/pkg/tuple"
)

// Config defines the configuration parameters
// for setting up and managing a sql connection.
type Config struct {
	Username               string
	Password               string
	Logger                 logger.Logger
	MaxTuplesPerWriteField int
	MaxTypesPerModelField  int

	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxIdleTime time.Duration
	ConnMaxLifetime time.Duration

	ExportMetrics bool
}

// DatastoreOption defines a function type
// used for configuring a Config object.
type DatastoreOption func(*Config)

// WithUsername returns a DatastoreOption that sets the username in the Config.
func WithUsername(username string) DatastoreOption {
	return func(config *Config) {
		config.Username = username
	}
}

// WithPassword returns a DatastoreOption that sets the password in the Config.
func WithPassword(password string) DatastoreOption {
	return func(config *Config) {
		config.Password = password
	}
}

// WithLogger returns a DatastoreOption that sets the Logger in the Config.
func WithLogger(l logger.Logger) DatastoreOption {
	return func(cfg *Config) {
		cfg.Logger = l
	}
}

// WithMaxTuplesPerWrite returns a DatastoreOption that sets
// the maximum number of tuples per write in the Config.
func WithMaxTuplesPerWrite(maxTuples int) DatastoreOption {
	return func(cfg *Config) {
		cfg.MaxTuplesPerWriteField = maxTuples
	}
}

// WithMaxTypesPerAuthorizationModel returns a DatastoreOption that sets
// the maximum number of types per authorization model in the Config.
func WithMaxTypesPerAuthorizationModel(maxTypes int) DatastoreOption {
	return func(cfg *Config) {
		cfg.MaxTypesPerModelField = maxTypes
	}
}

// WithMaxOpenConns returns a DatastoreOption that sets the
// maximum number of open connections in the Config.
func WithMaxOpenConns(c int) DatastoreOption {
	return func(cfg *Config) {
		cfg.MaxOpenConns = c
	}
}

// WithMaxIdleConns returns a DatastoreOption that sets the
// maximum number of idle connections in the Config.
func WithMaxIdleConns(c int) DatastoreOption {
	return func(cfg *Config) {
		cfg.MaxIdleConns = c
	}
}

// WithConnMaxIdleTime returns a DatastoreOption that sets
// the maximum idle time for a connection in the Config.
func WithConnMaxIdleTime(d time.Duration) DatastoreOption {
	return func(cfg *Config) {
		cfg.ConnMaxIdleTime = d
	}
}

// WithConnMaxLifetime returns a DatastoreOption that sets
// the maximum lifetime for a connection in the Config.
func WithConnMaxLifetime(d time.Duration) DatastoreOption {
	return func(cfg *Config) {
		cfg.ConnMaxLifetime = d
	}
}

// WithMetrics returns a DatastoreOption that
// enables the export of metrics in the Config.
func WithMetrics() DatastoreOption {
	return func(cfg *Config) {
		cfg.ExportMetrics = true
	}
}

// NewConfig creates a new Config instance with default values
// and applies any provided DatastoreOption modifications.
func NewConfig(opts ...DatastoreOption) *Config {
	cfg := &Config{}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.Logger == nil {
		cfg.Logger = logger.NewNoopLogger()
	}

	if cfg.MaxTuplesPerWriteField == 0 {
		cfg.MaxTuplesPerWriteField = storage.DefaultMaxTuplesPerWrite
	}

	if cfg.MaxTypesPerModelField == 0 {
		cfg.MaxTypesPerModelField = storage.DefaultMaxTypesPerAuthorizationModel
	}

	return cfg
}

// ContToken represents a continuation token structure used in pagination.
type ContToken struct {
	Ulid       string `json:"ulid"`
	ObjectType string `json:"ObjectType"`
}

// NewContToken creates a new instance of ContToken
// with the provided ULID and object type.
func NewContToken(ulid, objectType string) *ContToken {
	return &ContToken{
		Ulid:       ulid,
		ObjectType: objectType,
	}
}

// UnmarshallContToken takes a string representation of a continuation
// token and attempts to unmarshal it into a ContToken struct.
func UnmarshallContToken(from string) (*ContToken, error) {
	var token ContToken
	if err := json.Unmarshal([]byte(from), &token); err != nil {
		return nil, storage.ErrInvalidContinuationToken
	}
	return &token, nil
}

// SQLTupleIterator is a struct that implements the storage.TupleIterator
// interface for iterating over tuples fetched from a SQL database.
type SQLTupleIterator struct {
	rows     *sql.Rows
	resultCh chan *storage.TupleRecord
	errCh    chan error
}

// Ensures that SQLTupleIterator implements the TupleIterator interface.
var _ storage.TupleIterator = (*SQLTupleIterator)(nil)

// NewSQLTupleIterator returns a SQL tuple iterator.
func NewSQLTupleIterator(rows *sql.Rows) *SQLTupleIterator {
	return &SQLTupleIterator{
		rows:     rows,
		resultCh: make(chan *storage.TupleRecord, 1),
		errCh:    make(chan error, 1),
	}
}

func (t *SQLTupleIterator) next() (*storage.TupleRecord, error) {
	if !t.rows.Next() {
		if err := t.rows.Err(); err != nil {
			return nil, err
		}
		return nil, storage.ErrIteratorDone
	}

	var conditionName sql.NullString
	var conditionContext []byte
	var record storage.TupleRecord
	err := t.rows.Scan(
		&record.Store,
		&record.ObjectType,
		&record.ObjectID,
		&record.Relation,
		&record.User,
		&conditionName,
		&conditionContext,
		&record.Ulid,
		&record.InsertedAt,
	)
	if err != nil {
		return nil, err
	}

	record.ConditionName = conditionName.String

	if conditionContext != nil {
		var conditionContextStruct structpb.Struct
		if err := proto.Unmarshal(conditionContext, &conditionContextStruct); err != nil {
			return nil, err
		}
		record.ConditionContext = &conditionContextStruct
	}

	return &record, nil
}

// ToArray converts the tupleIterator to an []*openfgav1.Tuple and a possibly empty continuation token.
// If the continuation token exists it is the ulid of the last element of the returned array.
func (t *SQLTupleIterator) ToArray(
	opts storage.PaginationOptions,
) ([]*openfgav1.Tuple, []byte, error) {
	var res []*openfgav1.Tuple
	for i := 0; i < opts.PageSize; i++ {
		tupleRecord, err := t.next()
		if err != nil {
			if err == storage.ErrIteratorDone {
				return res, nil, nil
			}
			return nil, nil, err
		}
		res = append(res, tupleRecord.AsTuple())
	}

	// Check if we are at the end of the iterator.
	// If we are then we do not need to return a continuation token.
	// This is why we have LIMIT+1 in the query.
	tupleRecord, err := t.next()
	if err != nil {
		if errors.Is(err, storage.ErrIteratorDone) {
			return res, nil, nil
		}
		return nil, nil, err
	}

	contToken, err := json.Marshal(NewContToken(tupleRecord.Ulid, ""))
	if err != nil {
		return nil, nil, err
	}

	return res, contToken, nil
}

// Next will return the next available item.
func (t *SQLTupleIterator) Next(ctx context.Context) (*openfgav1.Tuple, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	record, err := t.next()
	if err != nil {
		return nil, err
	}

	return record.AsTuple(), nil
}

// Stop terminates iteration.
func (t *SQLTupleIterator) Stop() {
	t.rows.Close()
}

// HandleSQLError processes an SQL error and converts it into a more
// specific error type based on the nature of the SQL error.
func HandleSQLError(err error, args ...interface{}) error {
	if errors.Is(err, sql.ErrNoRows) {
		return storage.ErrNotFound
	} else if errors.Is(err, storage.ErrIteratorDone) {
		return err
	} else if strings.Contains(err.Error(), "duplicate key value") { // Postgres.
		if len(args) > 0 {
			if tk, ok := args[0].(*openfgav1.TupleKey); ok {
				return storage.InvalidWriteInputError(tk, openfgav1.TupleOperation_TUPLE_OPERATION_WRITE)
			}
		}
		return storage.ErrCollision
	} else if me, ok := err.(*mysql.MySQLError); ok && me.Number == 1062 {
		if len(args) > 0 {
			if tk, ok := args[0].(*openfgav1.TupleKey); ok {
				return storage.InvalidWriteInputError(tk, openfgav1.TupleOperation_TUPLE_OPERATION_WRITE)
			}
		}
		return storage.ErrCollision
	}

	return fmt.Errorf("sql error: %w", err)
}

// DBInfo encapsulates DB information for use in common method.
type DBInfo struct {
	db      *sql.DB
	stbl    sq.StatementBuilderType
	sqlTime interface{}
}

// NewDBInfo constructs a [DBInfo] object.
func NewDBInfo(db *sql.DB, stbl sq.StatementBuilderType, sqlTime interface{}) *DBInfo {
	return &DBInfo{
		db:      db,
		stbl:    stbl,
		sqlTime: sqlTime,
	}
}

// Write provides the common method for writing to database across sql storage.
func Write(
	ctx context.Context,
	dbInfo *DBInfo,
	store string,
	deletes storage.Deletes,
	writes storage.Writes,
	now time.Time,
) error {
	txn, err := dbInfo.db.BeginTx(ctx, nil)
	if err != nil {
		return HandleSQLError(err)
	}
	defer func() {
		_ = txn.Rollback()
	}()

	changelogBuilder := dbInfo.stbl.
		Insert("changelog").
		Columns(
			"store", "object_type", "object_id", "relation", "_user",
			"condition_name", "condition_context", "operation", "ulid", "inserted_at",
		)

	deleteBuilder := dbInfo.stbl.Delete("tuple")

	for _, tk := range deletes {
		id := ulid.MustNew(ulid.Timestamp(now), ulid.DefaultEntropy()).String()
		objectType, objectID := tupleUtils.SplitObject(tk.GetObject())

		res, err := deleteBuilder.
			Where(sq.Eq{
				"store":       store,
				"object_type": objectType,
				"object_id":   objectID,
				"relation":    tk.GetRelation(),
				"_user":       tk.GetUser(),
				"user_type":   tupleUtils.GetUserTypeFromUser(tk.GetUser()),
			}).
			RunWith(txn). // Part of a txn.
			ExecContext(ctx)
		if err != nil {
			return HandleSQLError(err, tk)
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return HandleSQLError(err)
		}

		if rowsAffected != 1 {
			return storage.InvalidWriteInputError(
				tk,
				openfgav1.TupleOperation_TUPLE_OPERATION_DELETE,
			)
		}

		changelogBuilder = changelogBuilder.Values(
			store, objectType, objectID,
			tk.GetRelation(), tk.GetUser(),
			"", nil, // Redact condition info for deletes since we only need the base triplet (object, relation, user).
			openfgav1.TupleOperation_TUPLE_OPERATION_DELETE,
			id, dbInfo.sqlTime,
		)
	}

	insertBuilder := dbInfo.stbl.
		Insert("tuple").
		Columns(
			"store", "object_type", "object_id", "relation", "_user", "user_type",
			"condition_name", "condition_context", "ulid", "inserted_at",
		)

	for _, tk := range writes {
		id := ulid.MustNew(ulid.Timestamp(now), ulid.DefaultEntropy()).String()
		objectType, objectID := tupleUtils.SplitObject(tk.GetObject())

		conditionName, conditionContext, err := marshalRelationshipCondition(tk.GetCondition())
		if err != nil {
			return err
		}

		_, err = insertBuilder.
			Values(
				store,
				objectType,
				objectID,
				tk.GetRelation(),
				tk.GetUser(),
				tupleUtils.GetUserTypeFromUser(tk.GetUser()),
				conditionName,
				conditionContext,
				id,
				dbInfo.sqlTime,
			).
			RunWith(txn). // Part of a txn.
			ExecContext(ctx)
		if err != nil {
			return HandleSQLError(err, tk)
		}

		changelogBuilder = changelogBuilder.Values(
			store,
			objectType,
			objectID,
			tk.GetRelation(),
			tk.GetUser(),
			conditionName,
			conditionContext,
			openfgav1.TupleOperation_TUPLE_OPERATION_WRITE,
			id,
			dbInfo.sqlTime,
		)
	}

	if len(writes) > 0 || len(deletes) > 0 {
		_, err := changelogBuilder.RunWith(txn).ExecContext(ctx) // Part of a txn.
		if err != nil {
			return HandleSQLError(err)
		}
	}

	if err := txn.Commit(); err != nil {
		return HandleSQLError(err)
	}

	return nil
}

// WriteAuthorizationModel writes an authorization model for the given store.
func WriteAuthorizationModel(
	ctx context.Context,
	dbInfo *DBInfo,
	store string,
	model *openfgav1.AuthorizationModel,
) error {
	schemaVersion := model.GetSchemaVersion()
	typeDefinitions := model.GetTypeDefinitions()

	if len(typeDefinitions) < 1 {
		return nil
	}

	pbdata, err := proto.Marshal(model)
	if err != nil {
		return err
	}

	_, err = dbInfo.stbl.
		Insert("authorization_model").
		Columns("store", "authorization_model_id", "schema_version", "type", "type_definition", "serialized_protobuf").
		Values(store, model.GetId(), schemaVersion, "", nil, pbdata).
		ExecContext(ctx)
	if err != nil {
		return HandleSQLError(err)
	}

	return nil
}

func constructAuthorizationModelFromSQLRows(rows *sql.Rows) (*openfgav1.AuthorizationModel, error) {
	var modelID string
	var schemaVersion string
	var typeDefs []*openfgav1.TypeDefinition
	for rows.Next() {
		var typeName string
		var marshalledTypeDef []byte
		var marshalledModel []byte
		err := rows.Scan(&modelID, &schemaVersion, &typeName, &marshalledTypeDef, &marshalledModel)
		if err != nil {
			return nil, HandleSQLError(err)
		}

		if len(marshalledModel) > 0 {
			// Prefer building an authorization model from the first row that has it available.
			var model openfgav1.AuthorizationModel
			if err := proto.Unmarshal(marshalledModel, &model); err != nil {
				return nil, err
			}

			return &model, nil
		}

		var typeDef openfgav1.TypeDefinition
		if err := proto.Unmarshal(marshalledTypeDef, &typeDef); err != nil {
			return nil, err
		}

		typeDefs = append(typeDefs, &typeDef)
	}

	if err := rows.Err(); err != nil {
		return nil, HandleSQLError(err)
	}

	if len(typeDefs) == 0 {
		return nil, storage.ErrNotFound
	}

	return &openfgav1.AuthorizationModel{
		SchemaVersion:   schemaVersion,
		Id:              modelID,
		TypeDefinitions: typeDefs,
	}, nil
}

// FindLatestAuthorizationModel reads the latest authorization model corresponding to the store.
func FindLatestAuthorizationModel(
	ctx context.Context,
	dbInfo *DBInfo,
	store string,
) (*openfgav1.AuthorizationModel, error) {
	rows, err := dbInfo.stbl.
		Select("authorization_model_id", "schema_version", "type", "type_definition", "serialized_protobuf").
		From("authorization_model").
		Where(sq.Eq{"store": store}).
		OrderBy("authorization_model_id desc").
		Limit(1).
		QueryContext(ctx)
	if err != nil {
		return nil, HandleSQLError(err)
	}
	defer rows.Close()
	return constructAuthorizationModelFromSQLRows(rows)
}

// ReadAuthorizationModel reads the model corresponding to store and model ID.
func ReadAuthorizationModel(
	ctx context.Context,
	dbInfo *DBInfo,
	store, modelID string,
) (*openfgav1.AuthorizationModel, error) {
	rows, err := dbInfo.stbl.
		Select("authorization_model_id", "schema_version", "type", "type_definition", "serialized_protobuf").
		From("authorization_model").
		Where(sq.Eq{
			"store":                  store,
			"authorization_model_id": modelID,
		}).
		QueryContext(ctx)
	if err != nil {
		return nil, HandleSQLError(err)
	}
	defer rows.Close()
	return constructAuthorizationModelFromSQLRows(rows)
}

// IsReady returns true if the connection to the datastore is successful
// and the datastore has the latest migration applied.
func IsReady(ctx context.Context, db *sql.DB) (storage.ReadinessStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return storage.ReadinessStatus{}, err
	}

	revision, err := goose.GetDBVersion(db)
	if err != nil {
		return storage.ReadinessStatus{}, err
	}

	if revision < build.MinimumSupportedDatastoreSchemaRevision {
		return storage.ReadinessStatus{
			Message: fmt.Sprintf("datastore requires migrations: at revision '%d', but requires '%d'. Run 'openfga migrate'.", revision, build.MinimumSupportedDatastoreSchemaRevision),
			IsReady: false,
		}, nil
	}

	return storage.ReadinessStatus{
		IsReady: true,
	}, nil
}
