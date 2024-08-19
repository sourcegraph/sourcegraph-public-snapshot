package commands

import (
	"context"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"

	"github.com/openfga/openfga/pkg/encoder"
	"github.com/openfga/openfga/pkg/logger"
	serverErrors "github.com/openfga/openfga/pkg/server/errors"
	"github.com/openfga/openfga/pkg/storage"
)

type ListStoresQuery struct {
	storesBackend storage.StoresBackend
	logger        logger.Logger
	encoder       encoder.Encoder
}

type ListStoresQueryOption func(*ListStoresQuery)

func WithListStoresQueryLogger(l logger.Logger) ListStoresQueryOption {
	return func(q *ListStoresQuery) {
		q.logger = l
	}
}

func WithListStoresQueryEncoder(e encoder.Encoder) ListStoresQueryOption {
	return func(q *ListStoresQuery) {
		q.encoder = e
	}
}

func NewListStoresQuery(storesBackend storage.StoresBackend, opts ...ListStoresQueryOption) *ListStoresQuery {
	q := &ListStoresQuery{
		storesBackend: storesBackend,
		logger:        logger.NewNoopLogger(),
		encoder:       encoder.NewBase64Encoder(),
	}

	for _, opt := range opts {
		opt(q)
	}
	return q
}

func (q *ListStoresQuery) Execute(ctx context.Context, req *openfgav1.ListStoresRequest) (*openfgav1.ListStoresResponse, error) {
	decodedContToken, err := q.encoder.Decode(req.GetContinuationToken())
	if err != nil {
		return nil, serverErrors.InvalidContinuationToken
	}

	paginationOptions := storage.NewPaginationOptions(req.GetPageSize().GetValue(), string(decodedContToken))

	stores, continuationToken, err := q.storesBackend.ListStores(ctx, paginationOptions)
	if err != nil {
		return nil, serverErrors.HandleError("", err)
	}

	encodedToken, err := q.encoder.Encode(continuationToken)
	if err != nil {
		return nil, serverErrors.HandleError("", err)
	}

	resp := &openfgav1.ListStoresResponse{
		Stores:            stores,
		ContinuationToken: encodedToken,
	}

	return resp, nil
}
