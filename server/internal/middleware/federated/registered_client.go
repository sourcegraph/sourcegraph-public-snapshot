package federated

import (
	"errors"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/auth/idkey"
	"src.sourcegraph.com/sourcegraph/fed"
	"src.sourcegraph.com/sourcegraph/svc"
	"src.sourcegraph.com/sourcegraph/svc/middleware/remote"
)

// Only the RegisteredClients.*UserPermissions methods are to be federated, since
// user permissions are handled by the federation root only. Other methods in this
// service should never be federated.
// TODO: improve this code to avoid having to write a trivial custom federated
// implementation for the remaining RegisteredClients methods.

func CustomRegisteredClientsGet(ctx context.Context, in *sourcegraph.RegisteredClientSpec, s sourcegraph.RegisteredClientsServer) (*sourcegraph.RegisteredClient, error) {
	return s.Get(ctx, in)
}

func CustomRegisteredClientsGetCurrent(ctx context.Context, in *pbtypes.Void, s sourcegraph.RegisteredClientsServer) (*sourcegraph.RegisteredClient, error) {
	return s.GetCurrent(ctx, in)
}

func CustomRegisteredClientsCreate(ctx context.Context, in *sourcegraph.RegisteredClient, s sourcegraph.RegisteredClientsServer) (*sourcegraph.RegisteredClient, error) {
	return s.Create(ctx, in)
}

func CustomRegisteredClientsUpdate(ctx context.Context, in *sourcegraph.RegisteredClient, s sourcegraph.RegisteredClientsServer) (*pbtypes.Void, error) {
	return s.Update(ctx, in)
}

func CustomRegisteredClientsDelete(ctx context.Context, in *sourcegraph.RegisteredClientSpec, s sourcegraph.RegisteredClientsServer) (*pbtypes.Void, error) {
	return s.Delete(ctx, in)
}

func CustomRegisteredClientsList(ctx context.Context, in *sourcegraph.RegisteredClientListOptions, s sourcegraph.RegisteredClientsServer) (*sourcegraph.RegisteredClientList, error) {
	return s.List(ctx, in)
}

func CustomRegisteredClientsGetUserPermissions(ctx context.Context, in *sourcegraph.UserPermissionsOptions, s sourcegraph.RegisteredClientsServer) (*sourcegraph.UserPermissions, error) {
	ctx2, clientID, err := getUserPermissionsCtx(ctx)
	if err != nil {
		return nil, err
	}
	if ctx2 == nil {
		return s.GetUserPermissions(ctx, in)
	}
	if in.ClientSpec.ID == "" {
		in.ClientSpec.ID = clientID
	}
	return svc.RegisteredClients(ctx2).GetUserPermissions(ctx2, in)
}

func CustomRegisteredClientsSetUserPermissions(ctx context.Context, in *sourcegraph.UserPermissions, s sourcegraph.RegisteredClientsServer) (*pbtypes.Void, error) {
	ctx2, clientID, err := getUserPermissionsCtx(ctx)
	if err != nil {
		return nil, err
	}
	if ctx2 == nil {
		return s.SetUserPermissions(ctx, in)
	}
	if in.ClientID == "" {
		in.ClientID = clientID
	}
	return svc.RegisteredClients(ctx2).SetUserPermissions(ctx2, in)
}

func CustomRegisteredClientsListUserPermissions(ctx context.Context, in *sourcegraph.RegisteredClientSpec, s sourcegraph.RegisteredClientsServer) (*sourcegraph.UserPermissionsList, error) {
	ctx2, clientID, err := getUserPermissionsCtx(ctx)
	if err != nil {
		return nil, err
	}
	if ctx2 == nil {
		return s.ListUserPermissions(ctx, in)
	}
	if in.ID == "" {
		in.ID = clientID
	}
	return svc.RegisteredClients(ctx2).ListUserPermissions(ctx2, in)
}

func getUserPermissionsCtx(ctx context.Context) (context.Context, string, error) {
	if fed.Config.IsRoot {
		return nil, "", nil
	}

	idKey := idkey.FromContext(ctx)
	if idKey == nil {
		return nil, "", errors.New("UserWhitelist: client id not found in context")
	}
	clientID := idKey.ID

	mothership, err := fed.Config.RootGRPCEndpoint()
	if err != nil {
		return nil, "", err
	}
	ctx = sourcegraph.WithGRPCEndpoint(ctx, mothership)
	return svc.WithServices(ctx, remote.Services), clientID, nil
}
