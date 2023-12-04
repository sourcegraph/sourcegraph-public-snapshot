package graphqlbackend

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestUser_EventLogs(t *testing.T) {
	db := dbmocks.NewMockDB()
	t.Run("only allowed by authenticated user on Sourcegraph.com", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		db.UsersFunc.SetDefaultReturn(users)

		orig := envvar.SourcegraphDotComMode()
		envvar.MockSourcegraphDotComMode(true)
		defer envvar.MockSourcegraphDotComMode(orig) // reset

		tests := []struct {
			name  string
			ctx   context.Context
			setup func()
		}{
			{
				name: "unauthenticated",
				ctx:  context.Background(),
				setup: func() {
					users.GetByIDFunc.SetDefaultReturn(&types.User{ID: 1}, nil)
				},
			},
			{
				name: "another user",
				ctx:  actor.WithActor(context.Background(), &actor.Actor{UID: 2}),
				setup: func() {
					users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
						return &types.User{ID: id}, nil
					})
				},
			},
			{
				name: "site admin",
				ctx:  actor.WithActor(context.Background(), &actor.Actor{UID: 2}),
				setup: func() {
					users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
						return &types.User{ID: id, SiteAdmin: true}, nil
					})
				},
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				test.setup()

				_, err := NewUserResolver(test.ctx, db, &types.User{ID: 1}).EventLogs(test.ctx, nil)
				got := fmt.Sprintf("%v", err)
				want := "must be authenticated as user with id 1"
				assert.Equal(t, want, got)
			})
		}
	})
}
