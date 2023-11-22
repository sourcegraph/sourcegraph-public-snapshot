package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var timeNow = time.Now

func (r *UserResolver) HasVerifiedEmail(ctx context.Context) (bool, error) {
	if deploy.IsApp() {
		return r.hasVerifiedEmailOnDotcom(ctx)
	}

	return r.hasVerifiedEmail(ctx)
}

func (r *UserResolver) hasVerifiedEmail(ctx context.Context) (bool, error) {
	// 🚨 SECURITY: In the UserEmailsService we check that only the
	// authenticated user and site admins can check
	// whether the user has a verified email.
	return backend.NewUserEmailsService(r.db, r.logger).HasVerifiedEmail(ctx, r.user.ID)
}

// hasVerifiedEmailOnDotcom - checks with sourcegraph.com to ensure the app user has verified email.
func (r *UserResolver) hasVerifiedEmailOnDotcom(ctx context.Context) (bool, error) {
	// 🚨 SECURITY: This resolves HasVerifiedEmail only for App by
	// sending the request to dotcom to check if a verified email exists for the user.
	// Dotcom will ensure that only the authenticated user and site admins can check
	if !deploy.IsApp() {
		return false, errors.New("attempt to check dotcom email verified outside of sourcegraph app")
	}

	if envvar.SourcegraphDotComMode() {
		return false, errors.New("attempt to check dotcom email verified outside of sourcegraph app")
	}

	// If app isn't configured with dotcom auth return false immediately
	appConfig := conf.Get().App
	if appConfig == nil {
		return false, nil
	}
	if len(appConfig.DotcomAuthToken) <= 0 {
		return false, nil
	}

	// If we have an app user with a dotcom authtoken ask dotcom if the user has a verified email
	url := "https://sourcegraph.com/.api/graphql?AppHasVerifiedEmailCheck"
	cli := httpcli.ExternalDoer
	payload := strings.NewReader(`{"query":"query AppHasVerifiedEmailCheck{ currentUser { hasVerifiedEmail } }","variables":{}}`)

	// Send GraphQL request to sourcegraph.com to check if email is verified
	req, err := http.NewRequestWithContext(ctx, "POST", url, payload)
	if err != nil {
		r.logger.Warn("failed creating email verification request ", log.Error(err))
		return false, nil
	}
	req.Header.Add("Authorization", fmt.Sprintf("token %s", appConfig.DotcomAuthToken))

	resp, err := cli.Do(req)
	if err != nil {
		r.logger.Warn("email verification request failed", log.Error(err))
		return false, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		r.logger.Warn("email verification failed", log.Int("status", resp.StatusCode))
		return false, nil
	}

	// Get the response
	type Response struct {
		Data struct {
			CurrentUser struct{ HasVerifiedEmail bool }
		}
	}
	var result Response
	b, err := io.ReadAll(io.LimitReader(resp.Body, 1024))
	if err != nil {
		r.logger.Warn("unable to read verified email response", log.Error(err))
		return false, nil
	}
	if err := json.Unmarshal(b, &result); err != nil {
		r.logger.Warn("unable to unmarshal verified email response", log.Error(err))
		return false, nil
	}

	return result.Data.CurrentUser.HasVerifiedEmail, nil
}

func (r *UserResolver) Emails(ctx context.Context) ([]*userEmailResolver, error) {
	// 🚨 SECURITY: Only the authenticated user and site admins can list user's
	// emails on Sourcegraph.com.
	if envvar.SourcegraphDotComMode() {
		if err := auth.CheckSiteAdminOrSameUser(ctx, r.db, r.user.ID); err != nil {
			return nil, err
		}
	}

	userEmails, err := r.db.UserEmails().ListByUser(ctx, database.UserEmailsListOptions{
		UserID: r.user.ID,
	})
	if err != nil {
		return nil, err
	}

	rs := make([]*userEmailResolver, len(userEmails))
	for i, userEmail := range userEmails {
		rs[i] = &userEmailResolver{
			db:        r.db,
			userEmail: *userEmail,
			user:      r,
		}
	}
	return rs, nil
}

func (r *UserResolver) PrimaryEmail(ctx context.Context) (*userEmailResolver, error) {
	// 🚨 SECURITY: Only the authenticated user and site admins can list user's
	// emails on Sourcegraph.com. We don't return an error, but not showing the email
	// either.
	if envvar.SourcegraphDotComMode() {
		if err := auth.CheckSiteAdminOrSameUser(ctx, r.db, r.user.ID); err != nil {
			return nil, nil
		}
	}
	ms, err := r.db.UserEmails().ListByUser(ctx, database.UserEmailsListOptions{
		UserID:       r.user.ID,
		OnlyVerified: true,
	})
	if err != nil {
		return nil, err
	}
	for _, m := range ms {
		if m.Primary {
			return &userEmailResolver{
				db:        r.db,
				userEmail: *m,
				user:      r,
			}, nil
		}
	}
	return nil, nil
}

type userEmailResolver struct {
	db        database.DB
	userEmail database.UserEmail
	user      *UserResolver
}

func (r *userEmailResolver) Email() string { return r.userEmail.Email }

func (r *userEmailResolver) IsPrimary() bool { return r.userEmail.Primary }

func (r *userEmailResolver) Verified() bool { return r.userEmail.VerifiedAt != nil }
func (r *userEmailResolver) VerificationPending() bool {
	return !r.Verified() && conf.EmailVerificationRequired()
}
func (r *userEmailResolver) User() *UserResolver { return r.user }

func (r *userEmailResolver) ViewerCanManuallyVerify(ctx context.Context) (bool, error) {
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err == auth.ErrNotAuthenticated || err == auth.ErrMustBeSiteAdmin {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

type addUserEmailArgs struct {
	User  graphql.ID
	Email string
}

func (r *schemaResolver) AddUserEmail(ctx context.Context, args *addUserEmailArgs) (*EmptyResponse, error) {
	userID, err := UnmarshalUserID(args.User)
	if err != nil {
		return nil, err
	}

	logger := r.logger.Scoped("AddUserEmail").
		With(log.Int32("userID", userID))

	userEmails := backend.NewUserEmailsService(r.db, logger)
	if err := userEmails.Add(ctx, userID, args.Email); err != nil {
		return nil, err
	}

	if conf.CanSendEmail() {
		if err := userEmails.SendUserEmailOnFieldUpdate(ctx, userID, "added an email"); err != nil {
			logger.Warn("Failed to send email to inform user of email addition", log.Error(err))
		}
	}

	return &EmptyResponse{}, nil
}

type removeUserEmailArgs struct {
	User  graphql.ID
	Email string
}

func (r *schemaResolver) RemoveUserEmail(ctx context.Context, args *removeUserEmailArgs) (*EmptyResponse, error) {
	userID, err := UnmarshalUserID(args.User)
	if err != nil {
		return nil, err
	}

	userEmails := backend.NewUserEmailsService(r.db, r.logger)
	if err := userEmails.Remove(ctx, userID, args.Email); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

type setUserEmailPrimaryArgs struct {
	User  graphql.ID
	Email string
}

func (r *schemaResolver) SetUserEmailPrimary(ctx context.Context, args *setUserEmailPrimaryArgs) (*EmptyResponse, error) {
	userID, err := UnmarshalUserID(args.User)
	if err != nil {
		return nil, err
	}

	userEmails := backend.NewUserEmailsService(r.db, r.logger)
	if err := userEmails.SetPrimaryEmail(ctx, userID, args.Email); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

type setUserEmailVerifiedArgs struct {
	User     graphql.ID
	Email    string
	Verified bool
}

func (r *schemaResolver) SetUserEmailVerified(ctx context.Context, args *setUserEmailVerifiedArgs) (*EmptyResponse, error) {
	userID, err := UnmarshalUserID(args.User)
	if err != nil {
		return nil, err
	}

	userEmails := backend.NewUserEmailsService(r.db, r.logger)
	if err := userEmails.SetVerified(ctx, userID, args.Email, args.Verified); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

type resendVerificationEmailArgs struct {
	User  graphql.ID
	Email string
}

func (r *schemaResolver) ResendVerificationEmail(ctx context.Context, args *resendVerificationEmailArgs) (*EmptyResponse, error) {
	userID, err := UnmarshalUserID(args.User)
	if err != nil {
		return nil, err
	}

	userEmails := backend.NewUserEmailsService(r.db, r.logger)
	if err := userEmails.ResendVerificationEmail(ctx, userID, args.Email, timeNow()); err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}
