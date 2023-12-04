package graphqlbackend

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestExternalAccountResolver_AccountData(t *testing.T) {
	users := dbmocks.NewMockUserStore()
	users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
		return &types.User{SiteAdmin: id == 1}, nil
	})

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)

	tests := []struct {
		name        string
		ctx         context.Context
		serviceType string
		wantErr     string
	}{
		{
			name:        "github and site admin",
			ctx:         actor.WithActor(context.Background(), &actor.Actor{UID: 1}),
			serviceType: extsvc.TypeGitHub,
			wantErr:     "<nil>",
		},
		{
			name:        "gitlab and site admin",
			ctx:         actor.WithActor(context.Background(), &actor.Actor{UID: 1}),
			serviceType: extsvc.TypeGitLab,
			wantErr:     "<nil>",
		},
		{
			name:        "github and non-site admin",
			ctx:         actor.WithActor(context.Background(), &actor.Actor{UID: 2}),
			serviceType: extsvc.TypeGitHub,
			wantErr:     "<nil>",
		},
		{
			name:        "gitlab and non-site admin",
			ctx:         actor.WithActor(context.Background(), &actor.Actor{UID: 2}),
			serviceType: extsvc.TypeGitLab,
			wantErr:     "<nil>",
		},
		{
			name:        "other and site admin",
			ctx:         actor.WithActor(context.Background(), &actor.Actor{UID: 1}),
			serviceType: extsvc.TypePerforce,
			wantErr:     "<nil>",
		},
		{
			name:        "other and non-site admin",
			ctx:         actor.WithActor(context.Background(), &actor.Actor{UID: 2}),
			serviceType: extsvc.TypePerforce,
			wantErr:     "must be site admin",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := &externalAccountResolver{
				db: db,
				account: extsvc.Account{
					AccountSpec: extsvc.AccountSpec{
						ServiceType: test.serviceType,
					},
				},
			}
			_, err := r.AccountData(test.ctx)
			got := fmt.Sprintf("%v", err)
			if diff := cmp.Diff(test.wantErr, got); diff != "" {
				t.Fatalf("Mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
