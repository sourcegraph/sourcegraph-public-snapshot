package dbstore

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *Store) RequestLanguageSupport(ctx context.Context, userID int, language string) (err error) {
	ctx, _, endObservation := s.operations.requestLanguageSupport.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return s.Exec(ctx, sqlf.Sprintf(requestLanguageSupportQuery, userID, language))
}

const requestLanguageSupportQuery = `
-- source: internal/codeintel/stores/dbstore/support.go:RequestLanguageSupport
INSERT INTO codeintel_langugage_support_requests (user_id, language_id)
VALUES (%s, %s)
ON CONFLICT DO NOTHING
`

func (s *Store) LanguagesRequestedBy(ctx context.Context, userID int) (_ []string, err error) {
	ctx, _, endObservation := s.operations.languagesRequestedBy.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return basestore.ScanStrings(s.Query(ctx, sqlf.Sprintf(languagesRequestedByQuery, userID)))
}

const languagesRequestedByQuery = `
-- source: internal/codeintel/stores/dbstore/support.go:LanguagesRequestedBy
SELECT language_id
FROM codeintel_langugage_support_requests
WHERE user_id = %s
ORDER BY language_id
`
