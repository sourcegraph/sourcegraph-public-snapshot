package commands

import (
	"context"
	"errors"
	"time"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"

	serverconfig "github.com/openfga/openfga/internal/server/config"
	"github.com/openfga/openfga/pkg/encoder"
	"github.com/openfga/openfga/pkg/logger"
	serverErrors "github.com/openfga/openfga/pkg/server/errors"
	"github.com/openfga/openfga/pkg/storage"
)

type ReadChangesQuery struct {
	backend       storage.ChangelogBackend
	logger        logger.Logger
	encoder       encoder.Encoder
	horizonOffset time.Duration
}

type ReadChangesQueryOption func(*ReadChangesQuery)

func WithReadChangesQueryLogger(l logger.Logger) ReadChangesQueryOption {
	return func(rq *ReadChangesQuery) {
		rq.logger = l
	}
}

func WithReadChangesQueryEncoder(e encoder.Encoder) ReadChangesQueryOption {
	return func(rq *ReadChangesQuery) {
		rq.encoder = e
	}
}

func WithReadChangeQueryHorizonOffset(horizonOffset int) ReadChangesQueryOption {
	return func(rq *ReadChangesQuery) {
		rq.horizonOffset = time.Duration(horizonOffset) * time.Minute
	}
}

// NewReadChangesQuery creates a ReadChangesQuery with specified `ChangelogBackend`.
func NewReadChangesQuery(backend storage.ChangelogBackend, opts ...ReadChangesQueryOption) *ReadChangesQuery {
	rq := &ReadChangesQuery{
		backend:       backend,
		logger:        logger.NewNoopLogger(),
		encoder:       encoder.NewBase64Encoder(),
		horizonOffset: time.Duration(serverconfig.DefaultChangelogHorizonOffset) * time.Minute,
	}

	for _, opt := range opts {
		opt(rq)
	}
	return rq
}

// Execute the ReadChangesQuery, returning paginated `openfga.TupleChange`(s) and a possibly non-empty continuation token.
func (q *ReadChangesQuery) Execute(ctx context.Context, req *openfgav1.ReadChangesRequest) (*openfgav1.ReadChangesResponse, error) {
	decodedContToken, err := q.encoder.Decode(req.GetContinuationToken())
	if err != nil {
		return nil, serverErrors.InvalidContinuationToken
	}
	paginationOptions := storage.NewPaginationOptions(req.GetPageSize().GetValue(), string(decodedContToken))

	changes, contToken, err := q.backend.ReadChanges(ctx, req.GetStoreId(), req.GetType(), paginationOptions, q.horizonOffset)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return &openfgav1.ReadChangesResponse{
				ContinuationToken: req.GetContinuationToken(),
			}, nil
		}
		return nil, serverErrors.HandleError("", err)
	}

	encodedContToken, err := q.encoder.Encode(contToken)
	if err != nil {
		return nil, serverErrors.HandleError("", err)
	}

	return &openfgav1.ReadChangesResponse{
		Changes:           changes,
		ContinuationToken: encodedContToken,
	}, nil
}
