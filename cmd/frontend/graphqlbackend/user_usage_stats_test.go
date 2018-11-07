package graphqlbackend

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

func TestUsers_UsageStatistics(t *testing.T) {
	ctx := context.Background()
	u := &UserResolver{user: &types.User{}}
	_, err := u.UsageStatistics(ctx)
	if err == nil {
		t.Errorf("Non-admin, non-user can access endpoint")
	}
}
