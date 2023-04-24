package auth

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestFromConnection(t *testing.T) {
	ctx := context.Background()

	conn := &schema.GitHubConnection{
		Url:   "https://github.com",
		Token: "abc123",
	}

	auther, err := FromConnection(ctx, conn)
	if err != nil {
		t.Fatal(err)
	}

	oat, ok := auther.(*auth.OAuthBearerToken)
	if !ok {
		t.Fatal("expected OAuthBearerToken")
	}

	if oat.Token != "abc123" {
		t.Errorf("want token abc123, got %s", oat.Token)
	}
}
