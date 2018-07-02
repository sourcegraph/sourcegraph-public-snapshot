package backend

import (
	"context"
	"errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
	"github.com/sourcegraph/sourcegraph/pkg/errcode"
)

var ErrMustBeSiteAdmin = errors.New("must be site admin")

// CheckCurrentUserIsSiteAdmin returns an error if the current user is NOT a site admin.
func CheckCurrentUserIsSiteAdmin(ctx context.Context) error {
	if hasAuthzBypass(ctx) {
		return nil
	}
	user, err := currentUser(ctx)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrNotAuthenticated
	}
	if !user.SiteAdmin {
		return ErrMustBeSiteAdmin
	}
	return nil
}

// CheckUserIsSiteAdmin returns an error if the user is NOT a site admin.
func CheckUserIsSiteAdmin(ctx context.Context, userID int32) error {
	if hasAuthzBypass(ctx) {
		return nil
	}
	user, err := db.Users.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrNotAuthenticated
	}
	if !user.SiteAdmin {
		return ErrMustBeSiteAdmin
	}
	return nil
}

// CheckSiteAdminOrSameUser returns an error if the user is NEITHER (1) a
// site admin NOR (2) the user specified by subjectUserID.
//
// It is used when an action on a user can be performed by site admins and the
// user themselves, but nobody else.
func CheckSiteAdminOrSameUser(ctx context.Context, subjectUserID int32) error {
	if hasAuthzBypass(ctx) {
		return nil
	}
	actor := actor.FromContext(ctx)
	if actor.IsAuthenticated() && actor.UID == subjectUserID {
		return nil
	}
	return CheckCurrentUserIsSiteAdmin(ctx)
}

func currentUser(ctx context.Context) (*types.User, error) {
	user, err := db.Users.GetByCurrentAuthUser(ctx)
	if err != nil {
		if errcode.IsNotFound(err) || err == db.ErrNoCurrentUser {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// WithAuthzBypass returns a context that backend.CheckXyz funcs report as being a site admin. It
// is used to bypass the backend.CheckXyz access control funcs when needed.
//
// 🚨 SECURITY: The caller MUST ensure that it performs its own access controls or removal of
// sensitive data.
func WithAuthzBypass(ctx context.Context) context.Context {
	return context.WithValue(ctx, authzBypass, struct{}{})
}

func hasAuthzBypass(ctx context.Context) bool {
	return ctx.Value(authzBypass) != nil
}

type contextKey int

const (
	authzBypass contextKey = iota
)
