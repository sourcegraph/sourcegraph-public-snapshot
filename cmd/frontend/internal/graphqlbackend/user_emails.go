package graphqlbackend

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
)

func (u *userResolver) Emails(ctx context.Context) ([]*userEmailResolver, error) {
	// 🚨 SECURITY: Only the self user and site admins can fetch a user's emails.
	if err := backend.CheckSiteAdminOrSameUser(ctx, u.user.ID); err != nil {
		return nil, err
	}

	userEmails, err := db.UserEmails.ListByUser(ctx, u.user.ID)
	if err != nil {
		return nil, err
	}

	rs := make([]*userEmailResolver, len(userEmails))
	for i, userEmail := range userEmails {
		rs[i] = &userEmailResolver{
			userEmail: *userEmail,
			user:      u,
		}
	}
	return rs, nil
}

type userEmailResolver struct {
	userEmail db.UserEmail
	user      *userResolver
}

func (r *userEmailResolver) Email() string  { return r.userEmail.Email }
func (r *userEmailResolver) Verified() bool { return r.userEmail.VerifiedAt != nil }
func (r *userEmailResolver) VerificationPending() bool {
	return !r.Verified() && conf.EmailVerificationRequired()
}
func (r *userEmailResolver) User() *userResolver { return r.user }
