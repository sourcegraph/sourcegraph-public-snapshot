package backend

import (
	"fmt"

	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	authpkg "sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/mailchimp"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/mailchimp/chimputil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/svc"
)

var Users sourcegraph.UsersServer = &users{}

type users struct{}

var _ sourcegraph.UsersServer = (*users)(nil)

func (s *users) Get(ctx context.Context, user *sourcegraph.UserSpec) (*sourcegraph.User, error) {
	return store.UsersFromContext(ctx).Get(ctx, *user)
}

// resolveUserSpec fills in the UID and Login fields.
func (s *users) resolveUserSpec(ctx context.Context, userSpec *sourcegraph.UserSpec) error {
	user, err := s.Get(ctx, userSpec)
	if err != nil {
		return err
	}
	*userSpec = user.Spec()
	return nil
}

// ensureUIDPopulated fills in the UID field by looking up the user by
// the Login field, if only the Login field (and not UID) is set.
func (s *users) ensureUIDPopulated(ctx context.Context, userSpec *sourcegraph.UserSpec) error {
	if userSpec.UID != 0 {
		return nil
	}
	return s.resolveUserSpec(ctx, userSpec)
}

func (s *users) GetWithEmail(ctx context.Context, emailAddr *sourcegraph.EmailAddr) (*sourcegraph.User, error) {
	return store.UsersFromContext(ctx).GetWithEmail(ctx, *emailAddr)
}

func (s *users) ListEmails(ctx context.Context, user *sourcegraph.UserSpec) (*sourcegraph.EmailAddrList, error) {
	emails, err := store.UsersFromContext(ctx).ListEmails(ctx, *user)
	if err != nil {
		return nil, err
	}
	return &sourcegraph.EmailAddrList{EmailAddrs: emails}, nil
}

func (s *users) RegisterBeta(ctx context.Context, opt *sourcegraph.BetaRegistration) (*sourcegraph.BetaResponse, error) {
	actor := authpkg.ActorFromContext(ctx)

	userSpec := actor.UserSpec()
	emails, err := svc.Users(ctx).ListEmails(ctx, &userSpec)
	if err != nil {
		return nil, err
	}
	primary, err := emails.Primary()
	if err != nil {
		// User has no primary email, so use the one specified.
		if opt.Email == "" {
			return nil, fmt.Errorf("user has no primary email address, and BetaRegistration.Email is not set")
		}
		_, err := svc.Accounts(ctx).UpdateEmails(ctx, &sourcegraph.UpdateEmailsOp{
			UserSpec: userSpec,
			Add: &sourcegraph.EmailAddrList{
				EmailAddrs: []*sourcegraph.EmailAddr{
					&sourcegraph.EmailAddr{Email: opt.Email, Primary: true},
				},
			},
		})
		if err != nil {
			return nil, err
		}
	} else {
		opt.Email = primary.Email
	}

	// Ensure the BetaRegistered status is properly set for the user.
	user, err := svc.Users(ctx).Get(ctx, &userSpec)
	if err != nil {
		return nil, err
	}
	if !user.BetaRegistered {
		user.BetaRegistered = true
		_, err = svc.Accounts(ctx).Update(ctx, user)
		if err != nil {
			return nil, err
		}
	}

	chimp, err := chimputil.Client()
	if err != nil {
		return nil, err
	}
	_, err = chimp.PutListsMembers(chimputil.SourcegraphBetaListID, mailchimp.SubscriberHash(opt.Email), &mailchimp.PutListsMembersOptions{
		StatusIfNew:  "subscribed",
		EmailAddress: opt.Email,
		MergeFields: map[string]interface{}{
			"FNAME":    opt.FirstName,
			"LNAME":    opt.LastName,
			"LANGUAGE": mailchimp.Array(opt.Languages),
			"EDITOR":   mailchimp.Array(opt.Editors),
			"MESSAGE":  opt.Message,
		},
	})
	if err != nil {
		return nil, err
	}
	return &sourcegraph.BetaResponse{
		EmailAddress: opt.Email,
	}, nil
}
