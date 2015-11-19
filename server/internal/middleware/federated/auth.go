package federated

import (
	"net/url"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/auth/authutil"
	"src.sourcegraph.com/sourcegraph/auth/idkey"
	"src.sourcegraph.com/sourcegraph/client/pkg/oauth2client"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/fed"
	"src.sourcegraph.com/sourcegraph/svc"
)

func CustomAuthGetAccessToken(ctx context.Context, op *sourcegraph.AccessTokenRequest, s sourcegraph.AuthServer) (*sourcegraph.AccessTokenResponse, error) {
	// Route to the Sourcegraph host based on the access token
	// request's TokenURL.
	isCurrentDomain, _, err := isCurrentDomain(ctx, op.TokenURL)
	if err != nil {
		return nil, err
	}
	if !isCurrentDomain {
		// Ensure that ctx is authenticated solely as the client, not
		// as a user. This ensures that the OAuth2 AS knows which
		// client is calling GetAccessToken (which is important
		// because auth codes, etc., are stored per-client ID).
		ctx = sourcegraph.WithCredentials(ctx, idkey.FromContext(ctx).TokenSource(ctx, op.TokenURL))

		// Communicate with the fed root. Assumes that the domain is
		// the fed root, which is not necessarily true.
		//
		// TODO(sqs): Generalize this when we have true generalized
		// federation.
		ctx = fed.Config.NewRemoteContext(ctx)

		// Get access token from remote. The request is authenticated
		// with this client's credentials. The access token will be
		// signed by the remote server but its ClientID will be this
		// client's ID.
		return svc.Auth(ctx).GetAccessToken(ctx, op)
	}

	return s.GetAccessToken(ctx, op)
}

func CustomAuthIdentify(ctx context.Context, op *pbtypes.Void, s sourcegraph.AuthServer) (*sourcegraph.AuthInfo, error) {
	// The Auth.Identify call returns the authed actor in the ctx obtained
	// from verifying the auth token. If the token needs to be parsed and
	// verified remotely, an Auth.Identify call on the remote server
	// will be executed in the GRPCMiddleware wrapper and the correct actor
	// will be set in the context before reaching this wrapper. So, here
	// the actor should already be set in the context, hence federation
	// is not necessary.
	return s.Identify(ctx, op)
}

func CustomAuthGetPermissions(ctx context.Context, op *pbtypes.Void, s sourcegraph.AuthServer) (*sourcegraph.UserPermissions, error) {
	if authutil.ActiveFlags.IsLocal() {
		return s.GetPermissions(ctx, op)
	}

	isCurrentDomain, _, err := isCurrentDomain(ctx, oauth2client.TokenURL())
	if err != nil {
		return nil, err
	}
	if !isCurrentDomain {
		// Communicate with the fed root. Assumes that the domain is
		// the fed root, which is not necessarily true.
		//
		// TODO(sqs): Generalize this when we have true generalized
		// federation.
		ctx = fed.Config.NewRemoteContext(ctx)

		return svc.Auth(ctx).GetPermissions(ctx, op)
	}
	return s.GetPermissions(ctx, op)
}

// isCurrentDomain returns a boolean indicating whether the URL's host
// matches the current host's AppURL. The one exception is that if
// urlStr is blank, it it assumed to refer to the current domain. This
// is so that the zero value does not cause cross-domain requests.
//
// The domain return value is the host of the urlStr.
func isCurrentDomain(ctx context.Context, urlStr string) (cur bool, domain string, err error) {
	if urlStr == "" {
		return true, "", nil
	}
	url, err := url.Parse(urlStr)
	if err != nil {
		return false, "", err
	}
	appURL := conf.AppURL(ctx)
	return appURL.Scheme == url.Scheme && appURL.Host == url.Host, url.Host, nil
}
