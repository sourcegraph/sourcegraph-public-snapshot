package mockstore

import (
	"testing"

	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
)

func (s *Users) MockGet(t *testing.T, wantUser int32) (called *bool) {
	called = new(bool)
	s.Get_ = func(ctx context.Context, user sourcegraph.UserSpec) (*sourcegraph.User, error) {
		*called = true
		if user.UID != wantUser {
			t.Errorf("got user %d, want %d", user, wantUser)
			return nil, grpc.Errorf(codes.NotFound, "user %d not found", wantUser)
		}
		return &sourcegraph.User{UID: user.UID}, nil
	}
	return
}

func (s *Users) MockGet_Return(t *testing.T, returns *sourcegraph.User) (called *bool) {
	called = new(bool)
	s.Get_ = func(ctx context.Context, user sourcegraph.UserSpec) (*sourcegraph.User, error) {
		*called = true
		if user.UID != returns.UID {
			t.Errorf("got user %d, want %d", user.UID, returns.UID)
			return nil, grpc.Errorf(codes.NotFound, "user %d not found", returns.UID)
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
