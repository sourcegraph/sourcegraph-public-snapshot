package commands

import (
	"context"
	"fmt"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"

	"github.com/openfga/openfga/pkg/encoder"
	"github.com/openfga/openfga/pkg/logger"
	serverErrors "github.com/openfga/openfga/pkg/server/errors"
	"github.com/openfga/openfga/pkg/storage"
	tupleUtils "github.com/openfga/openfga/pkg/tuple"
)

// A ReadQuery can be used to read one or many tuplesets
// Each tupleset specifies keys of a set of relation tuples.
// The set can include a single tuple key, or all tuples with
// a given object ID or userset in a type, optionally
// constrained by a relation name.
type ReadQuery struct {
	datastore storage.OpenFGADatastore
	logger    logger.Logger
	encoder   encoder.Encoder
}

type ReadQueryOption func(*ReadQuery)

func WithReadQueryLogger(l logger.Logger) ReadQueryOption {
	return func(rq *ReadQuery) {
		rq.logger = l
	}
}

func WithReadQueryEncoder(e encoder.Encoder) ReadQueryOption {
	return func(rq *ReadQuery) {
		rq.encoder = e
	}
}

// NewReadQuery creates a ReadQuery using the provided OpenFGA datastore implementation.
func NewReadQuery(datastore storage.OpenFGADatastore, opts ...ReadQueryOption) *ReadQuery {
	rq := &ReadQuery{
		datastore: datastore,
		logger:    logger.NewNoopLogger(),
		encoder:   encoder.NewBase64Encoder(),
	}

	for _, opt := range opts {
		opt(rq)
	}
	return rq
}

// Execute the ReadQuery, returning paginated `openfga.Tuple`(s) that match the tuple. Return all tuples if the tuple is
// nil or empty.
func (q *ReadQuery) Execute(ctx context.Context, req *openfgav1.ReadRequest) (*openfgav1.ReadResponse, error) {
	store := req.GetStoreId()
	tk := req.GetTupleKey()

	// Restrict our reads due to some compatibility issues in one of our storage implementations.
	if tk != nil {
		objectType, objectID := tupleUtils.SplitObject(tk.GetObject())
		if objectType == "" || (objectID == "" && tk.GetUser() == "") {
			return nil, serverErrors.ValidationError(
				fmt.Errorf("the 'tuple_key' field was provided but the object type field is required and both the object id and user cannot be empty"),
			)
		}
	}

	decodedContToken, err := q.encoder.Decode(req.GetContinuationToken())
	if err != nil {
		return nil, serverErrors.InvalidContinuationToken
	}

	paginationOptions := storage.NewPaginationOptions(req.GetPageSize().GetValue(), string(decodedContToken))

	tuples, contToken, err := q.datastore.ReadPage(ctx, store, tupleUtils.ConvertReadRequestTupleKeyToTupleKey(tk), paginationOptions)
	if err != nil {
		return nil, serverErrors.HandleError("", err)
	}

	encodedContToken, err := q.encoder.Encode(contToken)
	if err != nil {
		return nil, serverErrors.HandleError("", err)
	}

	return &openfgav1.ReadResponse{
		Tuples:            tuples,
		ContinuationToken: encodedContToken,
	}, nil
}
