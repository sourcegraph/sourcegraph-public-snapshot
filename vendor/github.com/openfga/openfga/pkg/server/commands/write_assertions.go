package commands

import (
	"context"
	"errors"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"

	"github.com/openfga/openfga/internal/validation"
	"github.com/openfga/openfga/pkg/logger"
	serverErrors "github.com/openfga/openfga/pkg/server/errors"
	"github.com/openfga/openfga/pkg/storage"
	tupleUtils "github.com/openfga/openfga/pkg/tuple"
	"github.com/openfga/openfga/pkg/typesystem"
)

type WriteAssertionsCommand struct {
	datastore storage.OpenFGADatastore
	logger    logger.Logger
}

type WriteAssertionsCmdOption func(*WriteAssertionsCommand)

func WithWriteAssertCmdLogger(l logger.Logger) WriteAssertionsCmdOption {
	return func(c *WriteAssertionsCommand) {
		c.logger = l
	}
}

func NewWriteAssertionsCommand(
	datastore storage.OpenFGADatastore, opts ...WriteAssertionsCmdOption) *WriteAssertionsCommand {
	cmd := &WriteAssertionsCommand{
		datastore: datastore,
		logger:    logger.NewNoopLogger(),
	}

	for _, opt := range opts {
		opt(cmd)
	}
	return cmd
}

func (w *WriteAssertionsCommand) Execute(ctx context.Context, req *openfgav1.WriteAssertionsRequest) (*openfgav1.WriteAssertionsResponse, error) {
	store := req.GetStoreId()
	modelID := req.GetAuthorizationModelId()
	assertions := req.GetAssertions()

	model, err := w.datastore.ReadAuthorizationModel(ctx, store, modelID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, serverErrors.AuthorizationModelNotFound(req.GetAuthorizationModelId())
		}

		return nil, serverErrors.HandleError("", err)
	}

	if !typesystem.IsSchemaVersionSupported(model.GetSchemaVersion()) {
		return nil, serverErrors.ValidationError(typesystem.ErrInvalidSchemaVersion)
	}

	typesys := typesystem.New(model)

	for _, assertion := range assertions {
		if err := validation.ValidateUserObjectRelation(typesys, tupleUtils.ConvertAssertionTupleKeyToTupleKey(assertion.GetTupleKey())); err != nil {
			return nil, serverErrors.ValidationError(err)
		}
	}

	err = w.datastore.WriteAssertions(ctx, store, modelID, assertions)
	if err != nil {
		return nil, serverErrors.HandleError("", err)
	}

	return &openfgav1.WriteAssertionsResponse{}, nil
}
