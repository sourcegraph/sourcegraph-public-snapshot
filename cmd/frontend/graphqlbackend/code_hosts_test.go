package graphqlbackend

import (
	"context"
	"testing"

	mockassert "github.com/derision-test/go-mockgen/testutil/assert"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/stretchr/testify/assert"
)

func TestSchemaResolver_CodeHosts(t *testing.T) {
	first := 1
	after := "Q29kZUhvc3Q6MQ=="

	store := dbmocks.NewMockCodeHostStore()
	store.CountFunc.SetDefaultReturn(3, nil)

	codeHosts := []*types.CodeHost{
		{ID: 1, URL: "github.com", Kind: extsvc.KindGitHub, APIRateLimitQuota: pointers.Ptr(int32(1)), APIRateLimitIntervalSeconds: pointers.Ptr(int32(1)), GitRateLimitIntervalSeconds: pointers.Ptr(int32(1)), GitRateLimitQuota: pointers.Ptr(int32(1))},
		{ID: 2, URL: "gitlab.com", Kind: extsvc.KindGitLab, APIRateLimitQuota: pointers.Ptr(int32(2)), APIRateLimitIntervalSeconds: pointers.Ptr(int32(2)), GitRateLimitIntervalSeconds: pointers.Ptr(int32(2)), GitRateLimitQuota: pointers.Ptr(int32(2))},
		{ID: 3, URL: "bitbucket.com", Kind: extsvc.KindBitbucketServer, APIRateLimitQuota: pointers.Ptr(int32(3)), APIRateLimitIntervalSeconds: pointers.Ptr(int32(3)), GitRateLimitIntervalSeconds: pointers.Ptr(int32(3)), GitRateLimitQuota: pointers.Ptr(int32(3))},
	}
	store.ListFunc.SetDefaultHook(func(ctx context.Context, opts database.ListCodeHostsOpts) ([]*types.CodeHost, int32, error) {
		assert.Equal(t, first, opts.Limit)
		assert.Equal(t, int32(1), opts.Cursor)
		// we are expecting only the second code host to be returned.
		return codeHosts[1:2], 1, nil
	})
	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	eSvcs := []*types.ExternalService{
		{ID: 1, DisplayName: "GITLAB #1"},
		{ID: 2, DisplayName: "GITLAB #2"},
	}
	externalServices := dbmocks.NewMockExternalServiceStore()
	externalServices.ListFunc.SetDefaultHook(func(ctx context.Context, options database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
		assert.Equal(t, options.CodeHostID, int32(2))
		assert.Equal(t, options.Limit, first)
		assert.Equal(t, options.Offset, 0)
		return eSvcs, nil
	})

	ctx := context.Background()
	db := dbmocks.NewMockDB()
	db.CodeHostsFunc.SetDefaultReturn(store)
	db.UsersFunc.SetDefaultReturn(users)
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)

	RunTest(t, &Test{
		Context: ctx,
		Schema:  mustParseGraphQLSchema(t, db),
		Variables: map[string]any{
			"first": first,
			"after": after,
		},
		Query: `query CodeHosts($first: Int $after: String) {
			  codeHosts(first: $first, after: $after) {
				pageInfo{
				  endCursor
				  hasNextPage
				}
				totalCount
				nodes{
				  id
				  kind
				  url
				  apiRateLimitQuota
				  apiRateLimitIntervalSeconds
				  gitRateLimitQuota
				  gitRateLimitIntervalSeconds
                  externalServices(first:$first){
					nodes{
					  id
					  displayName
					}
                  }
				}
			  }
			}`,
		ExpectedResult: `{
			   "codeHosts":{
				  "nodes":[
					 {
						"apiRateLimitIntervalSeconds":2,
						"apiRateLimitQuota":2,
						"externalServices":{
						   "nodes":[
							  {
								 "displayName":"GITLAB #1",
								 "id":"RXh0ZXJuYWxTZXJ2aWNlOjE="
							  },
							  {
								 "displayName":"GITLAB #2",
								 "id":"RXh0ZXJuYWxTZXJ2aWNlOjI="
							  }
						   ]
						},
						"gitRateLimitIntervalSeconds":2,
						"gitRateLimitQuota":2,
						"id":"Q29kZUhvc3Q6Mg==",
						"kind":"GITLAB",
						"url":"gitlab.com"
					 }
				  ],
				  "pageInfo":{
					 "endCursor":"Q29kZUhvc3Q6MQ==",
					 "hasNextPage":true
				  },
				  "totalCount":3
			   }
			}`,
	})

	mockassert.CalledOnce(t, store.CountFunc)
	mockassert.CalledOnce(t, store.ListFunc)
	mockassert.CalledOnce(t, externalServices.ListFunc)
}
