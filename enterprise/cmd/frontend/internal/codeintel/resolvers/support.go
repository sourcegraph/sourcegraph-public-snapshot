package resolvers

import (
	"context"
)

func (r *resolver) RequestLanguageSupport(ctx context.Context, userID int, language string) error {
	return r.dbStore.RequestLanguageSupport(ctx, userID, language)
}

func (r *resolver) RequestedLanguageSupport(ctx context.Context, userID int) ([]string, error) {
	return r.dbStore.LanguagesRequestedBy(ctx, userID)
}
