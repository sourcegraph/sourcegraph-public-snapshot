package backend

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
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
