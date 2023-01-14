package graphqlbackend

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestSiteConfigConnection(t *testing.T) {
	users := database.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{SiteAdmin: true}, nil)

	db := database.NewMockDB()
	db.UsersFunc.SetDefaultReturn(users)
	users.GetByIDFunc.SetDefaultHook(func(ctx context.Context, id int32) (*types.User, error) {
		switch id {
		case 1:
			return &types.User{
				ID:          1,
				Username:    "foo",
				DisplayName: "foo user",
			}, nil
		case 2:
			return &types.User{
				ID:          2,
				Username:    "bar",
				DisplayName: "bar user",
			}, nil
		default:
			return nil, errors.Newf("user ID %d not mocked in this test", id)
		}
	})

	now := time.Now()
	expectedNow := now.Format("2006-01-02T15:04:05Z")

	siteConfigs := []*database.SiteConfig{
		{
			ID:        1,
			Contents:  ``,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:           2,
			AuthorUserID: 1,
			Contents: `{
  foo: 1
}`,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:           3,
			AuthorUserID: 2,
			// Newline added.
			Contents: `{
  foo: 1,
  bar: 2
}`,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:           4,
			AuthorUserID: 1,
			// Existing line removed.
			Contents: `{
  bar: 2
}`,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:           5,
			AuthorUserID: 1,
			// Existing line changed.
			Contents: `{
  bar: 3
}`,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	conf := database.NewMockConfStore()
	conf.SiteGetLatestFunc.SetDefaultReturn(siteConfigs[len(siteConfigs)-1], nil)
	db.ConfFunc.SetDefaultReturn(conf)

	conf.GetSiteConfigCountFunc.SetDefaultReturn(len(siteConfigs), nil)
	conf.ListSiteConfigsFunc.SetDefaultReturn(siteConfigs, nil)
	conf.ListSiteConfigsFunc.SetDefaultHook(
		func(ctx context.Context, opt database.SiteConfigListOptions) ([]*database.SiteConfig, error) {
			return siteConfigs[:opt.Limit], nil
		})
	// db.Users

	RunTests(t, []*Test{
		{
			Schema: mustParseGraphQLSchema(t, db),
			Label:  "Get all site configuration history",
			Query: `
			{
			  site {
				  id
					configuration {
					  id
						history(first: 2){
							totalCount
							nodes{
								id
								previousID
								author{
									id,
									username,
									displayName
								}
								createdAt
								updatedAt
							}
						}
					}
			  }
			}
		`,
			ExpectedResult: fmt.Sprintf(`
			{
					"site": {
						"id": "U2l0ZToic2l0ZSI=",
						"configuration": {
							"id": 5,
							"history": {
								"totalCount": 5,
								"nodes": [
									{
										"id": "U2l0ZUNvbmZpZ3VyYXRpb25DaGFuZ2U6MQ==",
										"previousID": null,
										"author": null,
										"createdAt": %[1]q,
										"updatedAt": %[1]q
									},
									{
										"id": "U2l0ZUNvbmZpZ3VyYXRpb25DaGFuZ2U6Mg==",
										"previousID": "U2l0ZUNvbmZpZ3VyYXRpb25DaGFuZ2U6MQ==",
										"author": {
											"id": "VXNlcjox",
											"username": "foo",
											"displayName": "foo user"
										},
										"createdAt": %[1]q,
										"updatedAt": %[1]q
									}
								]
							}
						}
					}
			}
		`, expectedNow),
		},
	})
}
