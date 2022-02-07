package graphqlbackend

import (
	"context"
	"net/url"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/auth/userpasswd"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type randomizeUserPasswordResult struct {
	resetURL *url.URL
}

func (r *randomizeUserPasswordResult) ResetPasswordURL() *string {
	if r.resetURL == nil {
		return nil
	}
	urlStr := globals.ExternalURL().ResolveReference(r.resetURL).String()
	return &urlStr
}

func sendEmail(ctx context.Context, db database.DB, userID int32, resetURL *url.URL) error {
	user, err := db.Users().GetByID(ctx, userID)
	if err != nil {
		return err
	}

	email, _, err := db.UserEmails().GetPrimaryEmail(ctx, userID)
	if err != nil {
		return err
	}

	if err = userpasswd.SendResetPasswordURLEmail(ctx, email, user.Username, resetURL); err != nil {
		return err
	}

	return nil
}

func (r *schemaResolver) RandomizeUserPassword(ctx context.Context, args *struct {
	User graphql.ID
}) (*randomizeUserPasswordResult, error) {
	if !userpasswd.ResetPasswordEnabled() {
		return nil, errors.New("resetting passwords is not enabled")
	}
	if envvar.SourcegraphDotComMode() && !conf.CanSendEmail() {
		return nil, errors.New("unable to reset password because email sending is not configured")
	}
	// 🚨 SECURITY: Only site admins can randomize user passwords.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	userID, err := UnmarshalUserID(args.User)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse user ID")
	}

	if err := r.db.Users().RandomizePasswordAndClearPasswordResetRateLimit(ctx, userID); err != nil {
		return nil, err
	}

	// This method modifies the DB, which is somewhat counterintuitive for a "value" type from an
	// implementation POV. Its behavior is justified because it is convenient and intuitive from the
	// POV of the API consumer.
	resetURL, err := backend.MakePasswordResetURL(ctx, r.db, userID)
	if err != nil {
		return nil, err
	}
	// Send email to the user instead of returning the reset URL on Cloud
	if envvar.SourcegraphDotComMode() {
		if err := sendEmail(ctx, r.db, userID, resetURL); err != nil {
			return nil, err
		}

		// 🚨 SECURITY: Do not return reset URL on Cloud
		resetURL = nil
	}

	return &randomizeUserPasswordResult{resetURL: resetURL}, nil
}
