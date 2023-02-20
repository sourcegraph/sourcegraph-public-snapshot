package store

import (
	"context"
	"errors"
	"fmt"

	logger "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/sentinel/shared"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type Store interface {
	Foo(ctx context.Context) error
	InsertVulnerabilities(ctx context.Context, vulnerabilities []shared.Vulnerability) (err error)
}

type store struct {
	db         *basestore.Store
	logger     logger.Logger
	operations *operations
}

// New returns a new sentinel store.
func New(observationCtx *observation.Context, db database.DB) Store {
	return &store{
		db:         basestore.NewWithHandle(db.Handle()),
		logger:     logger.Scoped("sentinel.store", ""),
		operations: newOperations(observationCtx),
	}
}

func (s *store) Foo(ctx context.Context) (err error) {
	// TODO
	return nil
}

func (s *store) InsertVulnerabilities(ctx context.Context, vulnerabilities []shared.Vulnerability) (err error) {
	ctx, _, endObservation := s.operations.insertVulnerabilities.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	for _, v := range vulnerabilities {
		fmt.Printf("INSERT %s...\n", v.SourceID)
	}

	// TODO
	return errors.New("unimplemented")
}
