//go:generate mockgen -source storage.go -destination ../../internal/mocks/mock_storage.go -package mocks OpenFGADatastore

package storage

import (
	"context"
	"time"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
)

type ctxKey string

const (
	// DefaultMaxTuplesPerWrite specifies the default maximum number of tuples that can be written
	// in a single write operation. This constant is used to limit the batch size in write operations
	// to maintain performance and avoid overloading the system. The value is set to 100 tuples,
	// which is a balance between efficiency and resource usage.
	DefaultMaxTuplesPerWrite = 100

	// DefaultMaxTypesPerAuthorizationModel defines the default upper limit on the number of distinct
	// types that can be included in a single authorization model. This constraint helps in managing
	// the complexity and ensuring the maintainability of the authorization models. The limit is
	// set to 100 types, providing ample flexibility while keeping the model manageable.
	DefaultMaxTypesPerAuthorizationModel = 100

	// DefaultPageSize sets the default number of items to be returned in a single page when paginating
	// through a set of results. This constant is used to standardize the pagination size across various
	// parts of the system, ensuring a consistent and manageable volume of data per page. The default
	// value is set to 50, balancing detail per page with the overall number of pages.
	DefaultPageSize = 50

	relationshipTupleReaderCtxKey ctxKey = "relationship-tuple-reader-context-key"
)

// ContextWithRelationshipTupleReader sets the provided [[RelationshipTupleReader]]
// in the context. The context returned is a new context derived from the parent
// context provided.
func ContextWithRelationshipTupleReader(
	parent context.Context,
	reader RelationshipTupleReader,
) context.Context {
	return context.WithValue(parent, relationshipTupleReaderCtxKey, reader)
}

// RelationshipTupleReaderFromContext extracts a [[RelationshipTupleReader]] from the
// provided context (if any). If no such value is in the context a boolean false is returned,
// otherwise the RelationshipTupleReader is returned.
func RelationshipTupleReaderFromContext(ctx context.Context) (RelationshipTupleReader, bool) {
	ctxValue := ctx.Value(relationshipTupleReaderCtxKey)

	reader, ok := ctxValue.(RelationshipTupleReader)
	return reader, ok
}

// PaginationOptions should not be instantiated directly. Use NewPaginationOptions.
type PaginationOptions struct {
	PageSize int
	From     string
}

// NewPaginationOptions creates a new [PaginationOptions] instance
// with a specified page size and continuation token. If the input page size is empty,
// it uses DefaultPageSize.
func NewPaginationOptions(ps int32, contToken string) PaginationOptions {
	pageSize := DefaultPageSize
	if ps > 0 {
		pageSize = int(ps)
	}

	return PaginationOptions{
		PageSize: pageSize,
		From:     contToken,
	}
}

// Writes is a typesafe alias for Write arguments.
type Writes = []*openfgav1.TupleKey

// Deletes is a typesafe alias for Delete arguments.
type Deletes = []*openfgav1.TupleKeyWithoutCondition

// A TupleBackend provides a read/write interface for managing tuples.
type TupleBackend interface {
	RelationshipTupleReader
	RelationshipTupleWriter
}

// RelationshipTupleReader is an interface that defines the set of
// methods required to read relationship tuples from a data store.
type RelationshipTupleReader interface {
	// Read the set of tuples associated with `store` and `tupleKey`, which may be nil or partially filled. If nil,
	// Read will return an iterator over all the tuples in the given `store`. If the `tupleKey` is partially filled,
	// it will return an iterator over those tuples which match the `tupleKey`. Note that at least one of `Object`
	// or `User` (or both), must be specified in this case.
	//
	// The caller must be careful to close the [TupleIterator], either by consuming the entire iterator or by closing it.
	// There is NO guarantee on the order of the tuples returned on the iterator.
	Read(ctx context.Context, store string, tupleKey *openfgav1.TupleKey) (TupleIterator, error)

	// ReadPage functions similarly to Read but includes support for pagination. It takes
	// mandatory pagination options. PageSize will always be greater than zero.
	// It returns a slice of tuples along with a continuation token. This token can be used for retrieving subsequent pages of data.
	// There is NO guarantee on the order of the tuples in one page.
	ReadPage(
		ctx context.Context,
		store string,
		tupleKey *openfgav1.TupleKey,
		paginationOptions PaginationOptions,
	) ([]*openfgav1.Tuple, []byte, error)

	// ReadUserTuple tries to return one tuple that matches the provided key exactly.
	// If none is found, it must return [ErrNotFound].
	ReadUserTuple(
		ctx context.Context,
		store string,
		tupleKey *openfgav1.TupleKey,
	) (*openfgav1.Tuple, error)

	// ReadUsersetTuples returns all userset tuples for a specified object and relation.
	// For example, given the following relationship tuples:
	//	document:doc1, viewer, user:*
	//	document:doc1, viewer, group:eng#member
	// and the filter
	//	object=document:1, relation=viewer, allowedTypesForUser=[group#member]
	// this method would return the tuple (document:doc1, viewer, group:eng#member)
	// If allowedTypesForUser is empty, both tuples would be returned.
	// There is NO guarantee on the order returned on the iterator.
	ReadUsersetTuples(
		ctx context.Context,
		store string,
		filter ReadUsersetTuplesFilter,
	) (TupleIterator, error)

	// ReadStartingWithUser performs a reverse read of relationship tuples starting at one or
	// more user(s) or userset(s) and filtered by object type and relation.
	//
	// For example, given the following relationship tuples:
	//   document:doc1, viewer, user:jon
	//   document:doc2, viewer, group:eng#member
	//   document:doc3, editor, user:jon
	//
	// ReverseReadTuples for ['user:jon', 'group:eng#member'] filtered by 'document#viewer' would
	// return ['document:doc1#viewer@user:jon', 'document:doc2#viewer@group:eng#member'].
	// There is NO guarantee on the order returned on the iterator.
	ReadStartingWithUser(
		ctx context.Context,
		store string,
		filter ReadStartingWithUserFilter,
	) (TupleIterator, error)
}

// RelationshipTupleWriter is an interface that defines the set of methods
// required for writing relationship tuples in a data store.
type RelationshipTupleWriter interface {
	// Write updates data in the tuple backend, performing all delete operations in
	// `deletes` before adding new values in `writes`, returning the time of the transaction, or an error.
	// If there are more than MaxTuplesPerWrite, it must return ErrExceededWriteBatchLimit.
	// If two requests attempt to write the same tuple at the same time, it must return ErrTransactionalWriteFailed.
	// If the tuple to be written already existed or the tuple to be deleted didn't exist, it must return ErrInvalidWriteInput.
	Write(ctx context.Context, store string, d Deletes, w Writes) error

	// MaxTuplesPerWrite returns the maximum number of items (writes and deletes combined)
	// allowed in a single write transaction.
	MaxTuplesPerWrite() int
}

// ReadStartingWithUserFilter specifies the filter options that will be used
// to constrain the [RelationshipTupleReader.ReadStartingWithUser] query.
type ReadStartingWithUserFilter struct {
	ObjectType string
	Relation   string
	UserFilter []*openfgav1.ObjectRelation
}

// ReadUsersetTuplesFilter specifies the filter options that
// will be used to constrain the ReadUsersetTuples query.
type ReadUsersetTuplesFilter struct {
	Object                      string                         // Required.
	Relation                    string                         // Required.
	AllowedUserTypeRestrictions []*openfgav1.RelationReference // Optional.
}

// AuthorizationModelReadBackend provides a read interface for managing type definitions.
type AuthorizationModelReadBackend interface {
	// ReadAuthorizationModel reads the model corresponding to store and model ID.
	// If it's not found, it must return ErrNotFound.
	ReadAuthorizationModel(ctx context.Context, store string, id string) (*openfgav1.AuthorizationModel, error)

	// ReadAuthorizationModels reads all models for the supplied store and returns them in descending order of ULID (from newest to oldest).
	ReadAuthorizationModels(ctx context.Context, store string, options PaginationOptions) ([]*openfgav1.AuthorizationModel, []byte, error)

	// FindLatestAuthorizationModel returns the last model for the store.
	// If none were ever written, it must return ErrNotFound.
	FindLatestAuthorizationModel(ctx context.Context, store string) (*openfgav1.AuthorizationModel, error)
}

// TypeDefinitionWriteBackend provides a write interface for managing typed definition.
type TypeDefinitionWriteBackend interface {
	// MaxTypesPerAuthorizationModel returns the maximum number of type definition rows/items per model.
	MaxTypesPerAuthorizationModel() int

	// WriteAuthorizationModel writes an authorization model for the given store.
	WriteAuthorizationModel(ctx context.Context, store string, model *openfgav1.AuthorizationModel) error
}

// AuthorizationModelBackend provides an read/write interface for managing models and their type definitions.
type AuthorizationModelBackend interface {
	AuthorizationModelReadBackend
	TypeDefinitionWriteBackend
}

// StoresBackend is an interface that defines the set of methods required
// for interacting with and managing different types of storage backends.
type StoresBackend interface {
	CreateStore(ctx context.Context, store *openfgav1.Store) (*openfgav1.Store, error)
	DeleteStore(ctx context.Context, id string) error
	GetStore(ctx context.Context, id string) (*openfgav1.Store, error)
	ListStores(ctx context.Context, paginationOptions PaginationOptions) ([]*openfgav1.Store, []byte, error)
}

// AssertionsBackend is an interface that defines the set of methods for reading and writing assertions.
type AssertionsBackend interface {
	// WriteAssertions overwrites the assertions for a store and modelID.
	WriteAssertions(ctx context.Context, store, modelID string, assertions []*openfgav1.Assertion) error

	// ReadAssertions returns the assertions for a store and modelID.
	// If no assertions were ever written, it must return an empty list.
	ReadAssertions(ctx context.Context, store, modelID string) ([]*openfgav1.Assertion, error)
}

// ChangelogBackend is an interface for interacting with and managing changelogs.
type ChangelogBackend interface {
	// ReadChanges returns the writes and deletes that have occurred for tuples within a store,
	// in the order that they occurred.
	// You can optionally provide a filter to filter out changes for objects of a specific type.
	// The horizonOffset should be specified using a unit no more granular than a millisecond
	// and should be interpreted as a millisecond duration.
	// If no changes are found, it should return storage.ErrNotFound and an empty continuation token.
	ReadChanges(
		ctx context.Context,
		store,
		objectType string,
		paginationOptions PaginationOptions,
		horizonOffset time.Duration,
	) ([]*openfgav1.TupleChange, []byte, error)
}

// OpenFGADatastore is an interface that defines a set of methods for interacting
// with and managing data in an OpenFGA (Fine-Grained Authorization) system.
type OpenFGADatastore interface {
	TupleBackend
	AuthorizationModelBackend
	StoresBackend
	AssertionsBackend
	ChangelogBackend

	// IsReady reports whether the datastore is ready to accept traffic.
	IsReady(ctx context.Context) (ReadinessStatus, error)

	// Close closes the datastore and cleans up any residual resources.
	Close()
}

// ReadinessStatus represents the readiness status of the datastore.
type ReadinessStatus struct {
	// Message is a human-friendly status message for the current datastore status.
	Message string

	IsReady bool
}
