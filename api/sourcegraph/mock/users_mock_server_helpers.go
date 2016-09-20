package mock

import (
	"testing"

	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func (s *UsersServer) MockGetByUID(t *testing.T, wantUser int32) (called *bool) {
	called = new(bool)
	s.Get_ = func(ctx context.Context, user *sourcegraph.UserSpec) (*sourcegraph.User, error) {
		*called = true
		if user.UID != wantUser {
			t.Errorf("got UID %d, want %d", user.UID, wantUser)
			return nil, grpc.Errorf(codes.NotFound, "user with UID %d not found", wantUser)
		}
		return &sourcegraph.User{UID: wantUser}, nil
	}
	return
}

func (s *UsersServer) MockGet_Return(t *testing.T, returns *sourcegraph.User) (called *bool) {
	called = new(bool)
	s.Get_ = func(ctx context.Context, user *sourcegraph.UserSpec) (*sourcegraph.User, error) {
		*called = true
		if user.UID != returns.UID {
			t.Errorf("got user %d, want %d", user.UID, returns.UID)
			return nil, grpc.Errorf(codes.NotFound, "user %d not found", returns.UID)
		}
		return returns, nil
	}
	return
}

func (s *UsersServer) MockList(t *testing.T, wantUsers ...string) (called *bool) {
	called = new(bool)
	s.List_ = func(ctx context.Context, opt *sourcegraph.UsersListOptions) (*sourcegraph.UserList, error) {
		*called = true
		users := make([]*sourcegraph.User, len(wantUsers))
		for i, user := range wantUsers {
			users[i] = &sourcegraph.User{Login: user}
		}
		return &sourcegraph.UserList{Users: users}, nil
	}
	return
}

func (s *UsersServer) MockListEmails(t *testing.T, wantEmails ...string) (called *bool) {
	called = new(bool)
	s.ListEmails_ = func(ctx context.Context, user *sourcegraph.UserSpec) (*sourcegraph.EmailAddrList, error) {
		*called = true
		emails := make([]*sourcegraph.EmailAddr, len(wantEmails))
		for i, email := range wantEmails {
			emails[i] = &sourcegraph.EmailAddr{Email: email, Primary: i == 0}
		}
		return &sourcegraph.EmailAddrList{EmailAddrs: emails}, nil
	}
	return
}
