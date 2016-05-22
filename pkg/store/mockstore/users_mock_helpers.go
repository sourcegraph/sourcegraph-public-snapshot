package mockstore

import (
	"testing"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
)

func (s *Users) MockGet(t *testing.T, wantUser string) (called *bool) {
	called = new(bool)
	s.Get_ = func(ctx context.Context, user sourcegraph.UserSpec) (*sourcegraph.User, error) {
		*called = true
		if user.Login != wantUser {
			t.Errorf("got user %q, want %q", user, wantUser)
			return nil, grpc.Errorf(codes.NotFound, "user %v not found", wantUser)
		}
		return &sourcegraph.User{Login: user.Login}, nil
	}
	return
}

func (s *Users) MockGet_Return(t *testing.T, returns *sourcegraph.User) (called *bool) {
	called = new(bool)
	s.Get_ = func(ctx context.Context, user sourcegraph.UserSpec) (*sourcegraph.User, error) {
		*called = true
		if user.Login != returns.Login {
			t.Errorf("got user %q, want %q", user.Login, returns.Login)
			return nil, grpc.Errorf(codes.NotFound, "user %v not found", returns.Login)
		}
		return returns, nil
	}
	return
}

func (s *Users) MockList(t *testing.T, wantUsers ...string) (called *bool) {
	called = new(bool)
	s.List_ = func(ctx context.Context, opt *sourcegraph.UsersListOptions) ([]*sourcegraph.User, error) {
		*called = true
		users := make([]*sourcegraph.User, len(wantUsers))
		for i, user := range wantUsers {
			users[i] = &sourcegraph.User{UID: int32(i + 1), Login: user}
		}
		return users, nil
	}
	return
}

func (s *ExternalAuthTokens) MockGetUserToken(t *testing.T) (called *bool) {
	called = new(bool)
	s.GetUserToken_ = func(ctx context.Context, user int, host, clientID string) (*store.ExternalAuthToken, error) {
		*called = true
		return &store.ExternalAuthToken{}, nil
	}
	return
}

func (s *ExternalAuthTokens) MockSetUserToken(t *testing.T) (called *bool) {
	called = new(bool)
	s.SetUserToken_ = func(ctx context.Context, tok *store.ExternalAuthToken) error {
		*called = true
		return nil
	}
	return
}
