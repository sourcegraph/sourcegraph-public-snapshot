package load

import (
	"net/http"
	"net/url"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/app/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/test/e2e/e2etestuser"
)

const testUserName = e2etestuser.Prefix + "loadtest"
const testPassword = "e2etest"

type authedCookie struct {
	// HeaderValue is the cookie serialized for the Cookie header
	HeaderValue string
	// Expires is when the cookie will expire
	Expires time.Duration
}

func getAuthedCookie(endpoint *url.URL, username, password string) (*authedCookie, error) {
	ctx := sourcegraph.WithGRPCEndpoint(context.Background(), endpoint)
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return nil, err
	}
	tok, err := cl.Auth.GetAccessToken(ctx, &sourcegraph.AccessTokenRequest{
		AuthorizationGrant: &sourcegraph.AccessTokenRequest_ResourceOwnerPassword{
			ResourceOwnerPassword: &sourcegraph.LoginCredentials{
				Login:    username,
				Password: password,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	cookie, err := auth.NewSessionCookie(auth.Session{AccessToken: tok.AccessToken})
	if err != nil {
		return nil, err
	}
	// If only Name and Value are set, then Cookie.String returns the
	// serialization of the cookie for use in a Cookie header.
	cookie = &http.Cookie{Name: cookie.Name, Value: cookie.Value}

	// Say the token expires 5 minutes earlier so we have time to refresh it
	expires := (time.Duration(tok.ExpiresInSec) * time.Second) - (5 * time.Minute)
	return &authedCookie{
		HeaderValue: cookie.String(),
		Expires:     expires,
	}, nil
}

func createLoadTestUser(endpoint *url.URL) error {
	ctx := sourcegraph.WithGRPCEndpoint(context.Background(), endpoint)
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return err
	}
	_, err = cl.Accounts.Create(ctx, &sourcegraph.NewAccount{
		Login:    testUserName,
		Email:    testUserName + "@sourcegraph.com",
		Password: testPassword,
	})
	if err != nil && grpc.Code(err) != codes.AlreadyExists {
		return err
	}
	return nil
}
