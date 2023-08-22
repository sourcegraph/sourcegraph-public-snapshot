package graphqlbackend

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/stretchr/testify/assert"
)

func TestSchemaResolver_SetCodeHostRateLimits_NotASiteAdmin(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbmocks.NewMockDB()
	r := &schemaResolver{logger: logger, db: db}

	usersStore := dbmocks.NewMockUserStore()
	usersStore.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: false}, nil)
	db.UsersFunc.SetDefaultReturn(usersStore)

	_, err := r.SetCodeHostRateLimits(context.Background(), SetCodeHostRateLimitsArgs{
		Input: SetCodeHostRateLimitsInput{},
	})
	assert.NotNil(t, err)
	assert.Equal(t, "must be site admin", err.Error())
}

func TestSchemaResolver_SetCodeHostRateLimits_InvalidConfigs(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbmocks.NewMockDB()
	r := &schemaResolver{logger: logger, db: db}
	ctx := context.Background()
	wantErr := errors.New("rate limit settings must be positive integers")

	tests := []struct {
		name    string
		args    SetCodeHostRateLimitsArgs
		wantErr error
	}{
		{
			name: "Negative APIQuota",
			args: SetCodeHostRateLimitsArgs{
				Input: SetCodeHostRateLimitsInput{
					CodeHostID:                      "Q29kZUhvc3Q6MQ==",
					APIQuota:                        -1,
					APIReplenishmentIntervalSeconds: 1,
					GitQuota:                        1,
					GitReplenishmentIntervalSeconds: 1,
				},
			},
			wantErr: wantErr,
		},
		{
			name: "Negative APIReplenishmentIntervalSeconds",
			args: SetCodeHostRateLimitsArgs{
				Input: SetCodeHostRateLimitsInput{
					CodeHostID:                      "Q29kZUhvc3Q6MQ==",
					APIQuota:                        1,
					APIReplenishmentIntervalSeconds: -1,
					GitQuota:                        1,
					GitReplenishmentIntervalSeconds: 1,
				},
			},
			wantErr: wantErr,
		},
		{
			name: "Negative GitQuota",
			args: SetCodeHostRateLimitsArgs{
				Input: SetCodeHostRateLimitsInput{
					CodeHostID:                      "Q29kZUhvc3Q6MQ==",
					APIQuota:                        1,
					APIReplenishmentIntervalSeconds: 1,
					GitQuota:                        -1,
					GitReplenishmentIntervalSeconds: 1,
				},
			},
			wantErr: wantErr,
		},
		{
			name: "Negative GitReplenishmentIntervalSeconds",
			args: SetCodeHostRateLimitsArgs{
				Input: SetCodeHostRateLimitsInput{
					CodeHostID:                      "Q29kZUhvc3Q6MQ==",
					APIQuota:                        1,
					APIReplenishmentIntervalSeconds: 1,
					GitQuota:                        1,
					GitReplenishmentIntervalSeconds: -1,
				},
			},
			wantErr: wantErr,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			usersStore := dbmocks.NewMockUserStore()
			usersStore.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
			db.UsersFunc.SetDefaultReturn(usersStore)

			_, err := r.SetCodeHostRateLimits(ctx, test.args)
			assert.NotNil(t, err)
			assert.Equal(t, "rate limit settings must be positive integers", err.Error())
		})
	}
}

func TestSchemaResolver_SetCodeHostRateLimits_InvalidCodeHostID(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbmocks.NewMockDB()
	r := &schemaResolver{logger: logger, db: db}
	wantErr := errors.New("test error")

	usersStore := dbmocks.NewMockUserStore()
	usersStore.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
	db.UsersFunc.SetDefaultReturn(usersStore)

	codeHostStore := dbmocks.NewMockCodeHostStore()
	codeHostStore.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.CodeHost, error) {
		return nil, wantErr
	})
	db.CodeHostsFunc.SetDefaultReturn(codeHostStore)

	_, err := r.SetCodeHostRateLimits(context.Background(), SetCodeHostRateLimitsArgs{
		Input: SetCodeHostRateLimitsInput{CodeHostID: ""},
	})
	assert.NotNil(t, err)
	assert.Equal(t, "invalid code host id: invalid graphql.ID", err.Error())
}

func TestSchemaResolver_SetCodeHostRateLimits_GetCodeHostByIDError(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbmocks.NewMockDB()
	r := &schemaResolver{logger: logger, db: db}
	wantErr := errors.New("test error")

	usersStore := dbmocks.NewMockUserStore()
	usersStore.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
	db.UsersFunc.SetDefaultReturn(usersStore)

	codeHostStore := dbmocks.NewMockCodeHostStore()
	codeHostStore.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.CodeHost, error) {
		return nil, wantErr
	})
	db.CodeHostsFunc.SetDefaultReturn(codeHostStore)

	_, err := r.SetCodeHostRateLimits(context.Background(), SetCodeHostRateLimitsArgs{
		Input: SetCodeHostRateLimitsInput{CodeHostID: "Q29kZUhvc3Q6MQ=="},
	})
	assert.NotNil(t, err)
	assert.Equal(t, wantErr.Error(), err.Error())
}

func TestSchemaResolver_SetCodeHostRateLimits_UpdateCodeHostError(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbmocks.NewMockDB()
	r := &schemaResolver{logger: logger, db: db}
	wantErr := errors.New("test error")

	usersStore := dbmocks.NewMockUserStore()
	usersStore.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
	db.UsersFunc.SetDefaultReturn(usersStore)

	codeHostStore := dbmocks.NewMockCodeHostStore()
	codeHostStore.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.CodeHost, error) {
		assert.Equal(t, int32(1), id)
		return &types.CodeHost{ID: 1}, nil
	})
	codeHostStore.UpdateFunc.SetDefaultHook(func(ctx context.Context, host *types.CodeHost) error {
		return wantErr
	})
	db.CodeHostsFunc.SetDefaultReturn(codeHostStore)

	_, err := r.SetCodeHostRateLimits(context.Background(), SetCodeHostRateLimitsArgs{
		Input: SetCodeHostRateLimitsInput{CodeHostID: "Q29kZUhvc3Q6MQ=="},
	})
	assert.NotNil(t, err)
	assert.Equal(t, wantErr.Error(), err.Error())
}

func TestSchemaResolver_SetCodeHostRateLimits_Success(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbmocks.NewMockDB()
	r := &schemaResolver{logger: logger, db: db}
	ctx := context.Background()
	setCodeHostRateLimitsInput := SetCodeHostRateLimitsInput{
		CodeHostID:                      "Q29kZUhvc3Q6MQ==",
		APIQuota:                        1,
		APIReplenishmentIntervalSeconds: 2,
		GitQuota:                        3,
		GitReplenishmentIntervalSeconds: 4,
	}

	usersStore := dbmocks.NewMockUserStore()
	usersStore.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)
	db.UsersFunc.SetDefaultReturn(usersStore)

	codeHostStore := dbmocks.NewMockCodeHostStore()
	codeHostStore.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.CodeHost, error) {
		assert.Equal(t, int32(1), id)
		return &types.CodeHost{ID: 1}, nil
	})
	codeHostStore.UpdateFunc.SetDefaultHook(func(ctx context.Context, host *types.CodeHost) error {
		assert.Equal(t, setCodeHostRateLimitsInput.APIQuota, *(host.APIRateLimitQuota))
		assert.Equal(t, setCodeHostRateLimitsInput.APIReplenishmentIntervalSeconds, *(host.APIRateLimitIntervalSeconds))
		assert.Equal(t, setCodeHostRateLimitsInput.GitQuota, *(host.GitRateLimitQuota))
		assert.Equal(t, setCodeHostRateLimitsInput.GitReplenishmentIntervalSeconds, *(host.GitRateLimitIntervalSeconds))
		return nil
	})
	db.CodeHostsFunc.SetDefaultReturn(codeHostStore)

	variables := map[string]any{
		"input": map[string]any{
			"codeHostID":                      "Q29kZUhvc3Q6MQ==",
			"apiQuota":                        setCodeHostRateLimitsInput.APIQuota,
			"apiReplenishmentIntervalSeconds": setCodeHostRateLimitsInput.APIReplenishmentIntervalSeconds,
			"gitQuota":                        setCodeHostRateLimitsInput.GitQuota,
			"gitReplenishmentIntervalSeconds": setCodeHostRateLimitsInput.GitReplenishmentIntervalSeconds,
		},
	}
	RunTest(t, &Test{
		Context:   ctx,
		Schema:    mustParseGraphQLSchema(t, db),
		Variables: variables,

		Query: `mutation setCodeHostRateLimits($input:SetCodeHostRateLimitsInput!) {
		  setCodeHostRateLimits(input:$input) {
			alwaysNil
		  }
		}`,
		ExpectedResult: `{
			"setCodeHostRateLimits": {
			  "alwaysNil": null
			}
		}`,
	})
	_, err := r.SetCodeHostRateLimits(ctx, SetCodeHostRateLimitsArgs{
		Input: setCodeHostRateLimitsInput,
	})
	assert.Nil(t, err)
}
