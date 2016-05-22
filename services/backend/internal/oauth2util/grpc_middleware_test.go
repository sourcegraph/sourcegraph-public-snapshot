package oauth2util

import (
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"

	"golang.org/x/net/context"

	"google.golang.org/grpc/metadata"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth/accesstoken"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth/idkey"
)

func TestGRPCMiddlewareExpiredToken(t *testing.T) {
	idkey.SetTestEnvironment(512)
	k, err := idkey.Generate()
	if err != nil {
		t.Fatal(err)
	}

	tok, err := accesstoken.New(k, &auth.Actor{}, nil, -time.Hour, true)
	if err != nil {
		t.Fatal(err)
	}

	_, err = accesstoken.ParseAndVerify(k, tok.AccessToken)
	if err == nil {
		t.Fatal("ParseAndVerify: error expected")
	}
	vErr, ok := err.(*jwt.ValidationError)
	if !ok {
		t.Fatal("ParseAndVerify: ValidationError expected")
	}
	if vErr.Errors&jwt.ValidationErrorExpired == 0 {
		t.Fatal("ParseAndVerify: ValidationErrorExpired expected")
	}

	ctx := context.Background()
	ctx = idkey.NewContext(ctx, k)
	ctx = metadata.NewContext(ctx, metadata.MD{"authorization": []string{"bearer " + tok.AccessToken}})
	if _, err := GRPCMiddleware(ctx); err != nil {
		t.Errorf("expected no error, got: %s", err)
	}
}
