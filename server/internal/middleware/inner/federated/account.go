package federated

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/svc"
)

func CustomAccountsUpdate(ctx context.Context, in *sourcegraph.User, s sourcegraph.AccountsServer) (*pbtypes.Void, error) {
	ctx2, err := UserContext(ctx, sourcegraph.UserSpec{UID: in.UID, Login: in.Login, Domain: in.Domain})
	if err != nil {
		return nil, err
	}
	if ctx2 == nil {
		return s.Update(ctx, in)
	}
	ctx = ctx2
	return svc.Accounts(ctx).Update(ctx, in)
}
