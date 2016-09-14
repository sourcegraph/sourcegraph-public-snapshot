package oauth2util

import (
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"

	"context"

	"google.golang.org/grpc/metadata"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/auth"
)

func TestGRPCMiddlewareExpiredToken(t *testing.T) {
	tok, err := auth.NewAccessToken(&auth.Actor{}, nil, -time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	_, err = auth.ParseAndVerify(tok)
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
	ctx = metadata.NewContext(ctx, metadata.MD{"authorization": []string{"bearer " + tok}})
	if _, err := GRPCMiddleware(ctx); err != nil {
		t.Errorf("expected no error, got: %s", err)
	}
}
