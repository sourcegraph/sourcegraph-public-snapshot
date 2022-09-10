package store

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

// Store provides the interface for codenav storage.
type Store interface {
	GetLanguagesRequestedBy(ctx context.Context, userID int) (_ []string, err error)
	SetRequestLanguageSupport(ctx context.Context, userID int, language string) (err error)
}

// store manages the codenav store.
type store struct {
	db         *basestore.Store
	operations *operations
}

// New returns a new codenav store.
func New(db database.DB, observationContext *observation.Context) Store {
	return &store{
		db:         basestore.NewWithHandle(db.Handle()),
		operations: newOperations(observationContext),
	}
}

func (s *store) GetLanguagesRequestedBy(ctx context.Context, userID int) (_ []string, err error) {
	ctx, _, endObservation := s.operations.getLanguagesRequestedBy.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return basestore.ScanStrings(s.db.Query(ctx, sqlf.Sprintf(languagesRequestedByQuery, userID)))
}

const languagesRequestedByQuery = `
-- source: internal/codeintel/codenav/internal/store/store.go:GetLanguagesRequestedBy
SELECT language_id
FROM codeintel_langugage_support_requests
WHERE user_id = %s
ORDER BY language_id
`

func (s *store) SetRequestLanguageSupport(ctx context.Context, userID int, language string) (err error) {
	ctx, _, endObservation := s.operations.setRequestLanguageSupport.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.db.Exec(ctx, sqlf.Sprintf(requestLanguageSupportQuery, userID, language))
}

const requestLanguageSupportQuery = `
-- source: internal/codeintel/codenav/internal/store/store.go:SetRequestLanguageSupport
INSERT INTO codeintel_langugage_support_requests (user_id, language_id)
VALUES (%s, %s)
ON CONFLICT DO NOTHING
`
