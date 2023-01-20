package auth

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var ErrMustBeSiteAdmin = errors.New("must be site admin")

// CheckCurrentUserIsSiteAdmin returns an error if the current user is NOT a site admin.
func CheckCurrentUserIsSiteAdmin(ctx context.Context, db database.DB) error {
	_, err := checkCurrentUserIsSiteAdmin(ctx, db)
	return err
}

// CheckCurrentUserIsSiteAdminAndReturn returns an error if the current user is
// NOT a site admin and returns a user otherwise.
func CheckCurrentUserIsSiteAdminAndReturn(ctx context.Context, db database.DB) (*types.User, error) {
	return checkCurrentUserIsSiteAdmin(ctx, db)
}

func checkCurrentUserIsSiteAdmin(ctx context.Context, db database.DB) (*types.User, error) {
	if actor.FromContext(ctx).IsInternal() {
		return nil, nil
	}
	user, err := CurrentUser(ctx, db)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrNotAuthenticated
	}
	if !user.SiteAdmin {
		return nil, ErrMustBeSiteAdmin
	}
	return user, nil
}

// CheckUserIsSiteAdmin returns an error if the user is NOT a site admin.
func CheckUserIsSiteAdmin(ctx context.Context, db database.DB, userID int32) error {
	if actor.FromContext(ctx).IsInternal() {
		return nil
	}
	user, err := db.Users().GetByID(ctx, userID)
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

// InsufficientAuthorizationError is an error that occurs when the authentication is technically valid
// (e.g., the token is not expired) but does not yield a user with privileges to perform a certain
// action.
type InsufficientAuthorizationError struct {
	Message string
}

func (e *InsufficientAuthorizationError) Error() string      { return e.Message }
func (e *InsufficientAuthorizationError) Unauthorized() bool { return true }

// CheckSiteAdminOrSameUser returns an error if the user is NEITHER (1) a
// site admin NOR (2) the user specified by subjectUserID.
//
// It is used when an action on a user can be performed by site admins and the
// user themselves, but nobody else.
//
// Returns an error without the name of the given user.
func CheckSiteAdminOrSameUser(ctx context.Context, db database.DB, subjectUserID int32) error {
	a := actor.FromContext(ctx)
	if a.IsInternal() || (a.IsAuthenticated() && a.UID == subjectUserID) {
		return nil
	}
	isSiteAdminErr := CheckCurrentUserIsSiteAdmin(ctx, db)
	if isSiteAdminErr == nil {
		return nil
	}
	return &InsufficientAuthorizationError{"must be authenticated as the authorized user or site admin"}
}

// CheckSameUser returns an error if the user is not the user specified by
// subjectUserID.
func CheckSameUser(ctx context.Context, subjectUserID int32) error {
	a := actor.FromContext(ctx)
	if a.IsInternal() || (a.IsAuthenticated() && a.UID == subjectUserID) {
		return nil
	}
	return &InsufficientAuthorizationError{Message: fmt.Sprintf("must be authenticated as user with id %d", subjectUserID)}
}

// CurrentUser gets the current authenticated user
// It returns nil, nil if no user is found
func CurrentUser(ctx context.Context, db database.DB) (*types.User, error) {
	user, err := db.Users().GetByCurrentAuthUser(ctx)
	if err != nil {
		if errcode.IsNotFound(err) || err == database.ErrNoCurrentUser {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}
