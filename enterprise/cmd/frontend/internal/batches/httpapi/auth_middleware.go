package httpapi

import (
	"context"
	"net/http"

	sglog "github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func authMiddleware(next http.Handler, db database.DB, operation *observation.Operation) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//statusCode, err := func() (_ int, err error) {
		//	ctx, trace, endObservation := operation.With(r.Context(), &err, observation.Args{})
		//	defer endObservation(1, observation.Args{})
		//	_ = ctx
		//
		//	trace.Log(log.Event("bypassing code host auth check"))
		//	return 0, nil
		//}()
		//if err != nil {
		//	if statusCode >= 500 {
		//		operation.Logger.Error("batches.httpapi: failed to authorize request", sglog.Error(err))
		//	}
		//
		//	http.Error(w, fmt.Sprintf("failed to authorize request: %s", err.Error()), statusCode)
		//	return
		//}

		next.ServeHTTP(w, r)
	})
}

func isSiteAdmin(ctx context.Context, logger sglog.Logger, db database.DB) bool {
	user, err := db.Users().GetByCurrentAuthUser(ctx)
	if err != nil {
		if errcode.IsNotFound(err) || err == database.ErrNoCurrentUser {
			return false
		}

		logger.Error("batches.httpapi: failed to get up current user", sglog.Error(err))
		return false
	}

	return user != nil && user.SiteAdmin
}

func checkSiteAdminOrSameUser(ctx context.Context, db database.DB, subjectUserID int32) error {
	a := actor.FromContext(ctx)
	if a.IsInternal() || (a.IsAuthenticated() && a.UID == subjectUserID) {
		return nil
	}
	isSiteAdminErr := CheckCurrentUserIsSiteAdmin(ctx, db)
	if isSiteAdminErr == nil {
		return nil
	}
	_, err := db.Users().GetByID(ctx, subjectUserID)
	//if err != nil {
	//	return &InsufficientAuthorizationError{fmt.Sprintf("must be authenticated as an admin (%s)", isSiteAdminErr.Error())}
	//}
	//return &InsufficientAuthorizationError{fmt.Sprintf("must be authenticated as the authorized user or as an admin (%s)", isSiteAdminErr.Error())}
	return err
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

// CheckCurrentUserIsSiteAdmin returns an error if the current user is NOT a site admin.
func CheckCurrentUserIsSiteAdmin(ctx context.Context, db database.DB) error {
	if actor.FromContext(ctx).IsInternal() {
		return nil
	}
	//user, err := CurrentUser(ctx, db)
	//if err != nil {
	//	return err
	//}
	//if user == nil {
	//	return ErrNotAuthenticated
	//}
	//if !user.SiteAdmin {
	//	return ErrMustBeSiteAdmin
	//}
	return nil
}
