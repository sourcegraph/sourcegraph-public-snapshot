package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

func (r *schemaResolver) Viewer(ctx context.Context) (*viewerResolver, error) {
	user, err := CurrentUser(ctx, r.db)
	if err != nil {
		return nil, err
	}
	if user != nil {
		return &viewerResolver{user}, nil
	}

	// 🚨 SECURITY: Verify that unauthenticated visitors can use the API.
	if err := auth.CheckUnauthenticatedVisitorAccess(); err != nil {
		return nil, err
	}

	return &viewerResolver{&visitorResolver{}}, nil
}

// viewer is the interface for the GraphQL viewer interface.
type viewer interface {
	AffiliatedNamespaces(ctx context.Context) (gqlutil.SliceConnectionResolver[*NamespaceResolver], error)
}

// viewerResolver resolves the GraphQL Viewer interface to a type.
type viewerResolver struct {
	viewer
}

func (v viewerResolver) ToUser() (*UserResolver, bool) {
	n, ok := v.viewer.(*UserResolver)
	return n, ok
}

func (v viewerResolver) ToVisitor() (*visitorResolver, bool) {
	n, ok := v.viewer.(*visitorResolver)
	return n, ok
}

func (r *schemaResolver) ViewerCanChangeLibraryItemVisibilityToPublic(ctx context.Context) (bool, error) {
	err := ViewerCanChangeLibraryItemVisibilityToPublic(ctx, r.db)
	ok := err == nil
	if err == auth.ErrNotAuthenticated || err == auth.ErrMustBeSiteAdmin {
		err = nil
	}
	return ok, err
}

// ViewerCanChangeLibraryItemVisibilityToPublic returns nil (no error) if the current user can
// change the visibility of saved searches and prompt library items.
func ViewerCanChangeLibraryItemVisibilityToPublic(ctx context.Context, db database.DB) error {
	// 🚨 SECURITY: Only site admins may do this for now until we add in better management and abuse
	// prevention. We don't want someone creating an inappropriate saved search on Sourcegraph.com
	// that all users can see until we're ready to handle that kind of situation. The same applies
	// for customer instances for now.
	return auth.CheckCurrentUserIsSiteAdmin(ctx, db)
}
