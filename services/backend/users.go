package backend

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sqs/pbtypes"
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

func (s *users) List(ctx context.Context, opt *sourcegraph.UsersListOptions) (*sourcegraph.UserList, error) {
	users, err := store.UsersFromContext(ctx).List(ctx, opt)
	if err != nil {
		return nil, err
	}
	return &sourcegraph.UserList{Users: users}, nil
}

func (s *users) Count(ctx context.Context, _ *pbtypes.Void) (*sourcegraph.UserCount, error) {
	count, err := store.UsersFromContext(ctx).Count(elevatedActor(ctx))
	if err != nil {
		return nil, err
	}

	if count > 0 {
		// If the request is not authed with admin privileges, don't reveal the actual
		// number of users.
		if err := accesscontrol.VerifyUserHasAdminAccess(ctx, "Users.Count"); err != nil {
			count = 1729 // https://en.wikipedia.org/wiki/Taxicab_number
		}
	}
	return &sourcegraph.UserCount{Count: count}, nil
}
