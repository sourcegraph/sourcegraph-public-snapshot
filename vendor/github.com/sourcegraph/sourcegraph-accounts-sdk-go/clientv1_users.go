package sams

import (
	"context"

	"connectrpc.com/connect"
	"golang.org/x/oauth2"

	clientsv1 "github.com/sourcegraph/sourcegraph-accounts-sdk-go/clients/v1"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/clients/v1/clientsv1connect"
)

// UsersServiceV1 provides client methods to interact with the UsersService API
// v1.
type UsersServiceV1 struct {
	client *ClientV1
}

func (s *UsersServiceV1) newClient(ctx context.Context) clientsv1connect.UsersServiceClient {
	return clientsv1connect.NewUsersServiceClient(
		oauth2.NewClient(ctx, s.client.tokenSource),
		s.client.gRPCURL(),
		connect.WithInterceptors(s.client.defaultInterceptors...),
	)
}

// GetUserByID returns the SAMS user with the given ID. It returns ErrNotFound
// if no such user exists.
//
// Required scope: profile
func (s *UsersServiceV1) GetUserByID(ctx context.Context, id string) (*clientsv1.User, error) {
	req := &clientsv1.GetUserRequest{Id: id}
	client := s.newClient(ctx)
	resp, err := parseResponseAndError(client.GetUser(ctx, connect.NewRequest(req)))
	if err != nil {
		return nil, err
	}
	return resp.Msg.User, nil
}

// GetUserByEmail returns the SAMS user with the given verified email. It returns
// ErrNotFound if no such user exists.
//
// Required scope: profile
func (s *UsersServiceV1) GetUserByEmail(ctx context.Context, email string) (*clientsv1.User, error) {
	req := &clientsv1.GetUserRequest{Email: email}
	client := s.newClient(ctx)
	resp, err := parseResponseAndError(client.GetUser(ctx, connect.NewRequest(req)))
	if err != nil {
		return nil, err
	}
	return resp.Msg.User, nil
}

// GetUsersByIDs returns the list of SAMS users matching the provided IDs.
//
// NOTE: It silently ignores any invalid user IDs, i.e. the length of the return
// slice may be less than the length of the input slice.
//
// Required scopes: profile
func (s *UsersServiceV1) GetUsersByIDs(ctx context.Context, ids []string) ([]*clientsv1.User, error) {
	req := &clientsv1.GetUsersRequest{Ids: ids}
	client := s.newClient(ctx)
	resp, err := parseResponseAndError(client.GetUsers(ctx, connect.NewRequest(req)))
	if err != nil {
		return nil, err
	}
	return resp.Msg.GetUsers(), nil
}

// GetUserRolesByID returns all roles that have been assigned to the SAMS user
// with the given ID and scoped by the service.
//
// Required scopes: sams::user.roles::read
func (s *UsersServiceV1) GetUserRolesByID(ctx context.Context, userID, service string) ([]string, error) {
	req := &clientsv1.GetUserRolesRequest{
		Id:      userID,
		Service: service,
	}
	client := s.newClient(ctx)
	resp, err := parseResponseAndError(client.GetUserRoles(ctx, connect.NewRequest(req)))
	if err != nil {
		return nil, err
	}
	return resp.Msg.GetRoles(), nil
}
