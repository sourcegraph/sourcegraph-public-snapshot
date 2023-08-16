package graphqlbackend

import (
	"context"
	"strconv"
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
	t.Parallel()

	codeHosts := []*types.CodeHost{
		{ID: 1, URL: "github.com", Kind: extsvc.KindGitHub, APIRateLimitQuota: pointers.Ptr(int32(1)), APIRateLimitIntervalSeconds: pointers.Ptr(int32(1)), GitRateLimitIntervalSeconds: pointers.Ptr(int32(1)), GitRateLimitQuota: pointers.Ptr(int32(1))},
		{ID: 2, URL: "gitlab.com", Kind: extsvc.KindGitLab, APIRateLimitQuota: pointers.Ptr(int32(2)), APIRateLimitIntervalSeconds: pointers.Ptr(int32(2)), GitRateLimitIntervalSeconds: pointers.Ptr(int32(2)), GitRateLimitQuota: pointers.Ptr(int32(2))},
		{ID: 3, URL: "bitbucket-cloud.com", Kind: extsvc.KindBitbucketCloud},
		{ID: 4, URL: "bitbucket.com", Kind: extsvc.KindBitbucketServer, APIRateLimitQuota: pointers.Ptr(int32(4)), APIRateLimitIntervalSeconds: pointers.Ptr(int32(4)), GitRateLimitIntervalSeconds: pointers.Ptr(int32(4)), GitRateLimitQuota: pointers.Ptr(int32(4))},
	}

	tests := []struct {
		name  string
		first int
		after int32
	}{
		{
			name:  "only first",
			first: 1,
			after: 0,
		},
		{
			name:  "first 1 and after 1",
			first: 1,
			after: 1,
		},
		{
			name:  "first 1 and after 2",
			first: 1,
			after: 2,
		},
		{
			name:  "first 1 and after 3",
			first: 1,
			after: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := dbmocks.NewMockCodeHostStore()
			store.CountFunc.SetDefaultReturn(4, nil)
			codeHost := codeHosts[tc.after]

			store.ListFunc.SetDefaultHook(func(ctx context.Context, opts database.ListCodeHostsOpts) ([]*types.CodeHost, int32, error) {
				assert.Equal(t, tc.first, opts.Limit)
				assert.Equal(t, tc.after, opts.Cursor)
				next := tc.after + int32(tc.first)
				if int(next) >= len(codeHosts) {
					next = 0
				}

				return codeHosts[tc.after : tc.after+int32(tc.first)], next, nil
			})
			users := dbmocks.NewMockUserStore()
			users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

			eSvcs := []*types.ExternalService{
				{ID: 1, DisplayName: "GITLAB #1"},
				{ID: 2, DisplayName: "GITLAB #2"},
			}
			externalServices := dbmocks.NewMockExternalServiceStore()
			externalServices.ListFunc.SetDefaultHook(func(ctx context.Context, options database.ExternalServicesListOptions) ([]*types.ExternalService, error) {
				assert.Equal(t, options.CodeHostID, codeHost.ID)
				assert.Equal(t, options.Limit, tc.first)
				assert.Equal(t, options.Offset, 0)
				return eSvcs, nil
			})

			ctx := context.Background()
			db := dbmocks.NewMockDB()
			db.CodeHostsFunc.SetDefaultReturn(store)
			db.UsersFunc.SetDefaultReturn(users)
			db.ExternalServicesFunc.SetDefaultReturn(externalServices)
			variables := map[string]any{
				"first": tc.first,
			}

			gqlAfterID := MarshalCodeHostID(tc.after)
			if tc.after != 0 {
				variables["after"] = gqlAfterID
			}
			newAfter := "null"
			hasNext := "false"
			if int(tc.after+1) < len(codeHosts) {
				newAfter = `"` + string(MarshalCodeHostID(tc.after+1)) + `"`
				hasNext = "true"
			}
			rateLimitValueOrNull := func(value *int32) string {
				if value == nil {
					return "null"
				}
				return strconv.Itoa(int(*value))
			}
			RunTest(t, &Test{
				Context:   ctx,
				Schema:    mustParseGraphQLSchema(t, db),
				Variables: variables,
				Query: `query CodeHosts($first: Int $after: String){
			  codeHosts(first:$first, after:$after) {
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
                  externalServices(first:1){
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
						"apiRateLimitIntervalSeconds":` + rateLimitValueOrNull(codeHost.APIRateLimitIntervalSeconds) + `,
						"apiRateLimitQuota":` + rateLimitValueOrNull(codeHost.APIRateLimitQuota) + `,
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
						"gitRateLimitIntervalSeconds":` + rateLimitValueOrNull(codeHost.GitRateLimitIntervalSeconds) + `,
						"gitRateLimitQuota":` + rateLimitValueOrNull(codeHost.GitRateLimitIntervalSeconds) + `,
						"id":"` + string(MarshalCodeHostID(codeHost.ID)) + `",
						"kind":"` + codeHost.Kind + `",
						"url":"` + codeHost.URL + `"
					 }
				  ],
				  "pageInfo":{
					 "endCursor":` + newAfter + `,
					 "hasNextPage":` + hasNext + `
				  },
				  "totalCount":4
			   }
			}`,
			})

			mockassert.CalledOnce(t, store.CountFunc)
			mockassert.CalledOnce(t, store.ListFunc)
			mockassert.CalledOnce(t, externalServices.ListFunc)
		})
	}
}

func TestCodeHostByID(t *testing.T) {
	codeHost := &types.CodeHost{ID: 2, URL: "github.com", Kind: extsvc.KindGitHub, APIRateLimitQuota: pointers.Ptr(int32(1)), APIRateLimitIntervalSeconds: pointers.Ptr(int32(1)), GitRateLimitIntervalSeconds: pointers.Ptr(int32(1)), GitRateLimitQuota: pointers.Ptr(int32(1))}

	store := dbmocks.NewMockCodeHostStore()
	store.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.CodeHost, error) {
		assert.Equal(t, id, codeHost.ID)
		return codeHost, nil
	})

	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	ctx := context.Background()
	db := dbmocks.NewMockDB()
	db.CodeHostsFunc.SetDefaultReturn(store)
	db.UsersFunc.SetDefaultReturn(users)

	variables := map[string]any{}

	RunTest(t, &Test{
		Context:   ctx,
		Schema:    mustParseGraphQLSchema(t, db),
		Variables: variables,
		Query: `query CodeHostByID() {
		  node(id:"Q29kZUhvc3Q6Mg=="){
			id
			__typename
			... on CodeHost {
						  kind
					url
						}
		  }
		}`,
		ExpectedResult: `{
			"node": {
			  "id": "Q29kZUhvc3Q6Mg==",
			  "__typename": "CodeHost",
			  "kind": "GITHUB",
			  "url": "github.com"
			}
		}`,
	})

	mockassert.CalledOnce(t, store.GetByIDFunc)
}
