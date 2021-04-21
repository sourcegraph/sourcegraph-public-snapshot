package authtest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
)

func TestSiteAdminEndpoints(t *testing.T) {
	// Create a test user (authtest-user-1) which is not a site admin, the user
	// should receive access denied for site admin endpoints.
	const testUsername = "authtest-user-1"
	userClient, err := gqltestutil.SignUp(*baseURL, testUsername+"@sourcegraph.com", testUsername, "mysecurepassword")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := client.DeleteUser(userClient.AuthenticatedUserID(), true)
		if err != nil {
			t.Fatal(err)
		}
	}()

	t.Run("HTTP endpoints", func(t *testing.T) {
		// TODO(jchen): Add HTTP endpoints that are site-admin-only
	})

	t.Run("GraphQL queries", func(t *testing.T) {
		type gqlTest struct {
			name      string
			query     string
			variables map[string]interface{}
		}
		tests := []gqlTest{
			{
				name: "resetTriggerQueryTimestamps",
				query: `
mutation {
	resetTriggerQueryTimestamps(id: "SUQ6MTIz") {
		alwaysNil
	}
}
`,
			},
			{
				name: "dotcom.productLicenses",
				query: `
{
	dotcom {
		productLicenses {
			__typename
		}
	}
}
`,
			},
			{
				name: "dotcom.createProductSubscription",
				query: `
mutation {
	dotcom {
		createProductSubscription(accountID: "VXNlcjox") {
			id
		}
	}
}
`,
			},
			{
				name: "dotcom.setProductSubscriptionBilling",
				query: `
mutation {
	dotcom {
		setProductSubscriptionBilling(id: "VXNlcjox") {
			alwaysNil
		}
	}
}
`,
			},
			{
				name: "dotcom.archiveProductSubscription",
				query: `
mutation {
	dotcom {
		archiveProductSubscription(id: "VXNlcjox") {
			alwaysNil
		}
	}
}
`,
			},
			{
				name: "dotcom.generateProductLicenseForSubscription",
				query: `
mutation {
	dotcom {
		generateProductLicenseForSubscription(
			productSubscriptionID: "UHJvZHVjdFN1YnNjcmlwdGlvbjox"
			license: {tags: [], userCount: 1, expiresAt: 1}
		) {
			id
		}
	}
}
`,
			},
			{
				name: "updateMirrorRepository",
				query: `
mutation {
	updateMirrorRepository(repository: "UmVwb3NpdG9yeTox") {
		alwaysNil
	}
}
`,
			},
			{
				name: "addUserToOrganization",
				query: `
mutation {
	addUserToOrganization(organization: "T3JnYW5pemF0aW9uOjE=", username: "alice") {
		alwaysNil
	}
}
`,
			},
			{
				name: "site.configuration",
				query: `
{
	site {
		configuration {
			id
		}
	}
}
`,
			},
			{
				name: "site.accessTokens",
				query: `
{
	site {
		accessTokens {
			totalCount
		}
	}
}
`,
			},
			{
				name: "updateSiteConfiguration",
				query: `
mutation {
	updateSiteConfiguration(input: "", lastID: 0)
}
`,
			},
			{
				name: "deleteLSIFUpload",
				query: `
mutation {
	deleteLSIFUpload(id: "TFNJRjox") {
		alwaysNil
	}
}
`,
			},
			{
				name: "outOfBandMigrations",
				query: `
{
	outOfBandMigrations {
		id
	}
}
`,
			},
			{
				name: "createAccessToken.ScopeSiteAdminSudo",
				query: `
mutation CreateAccessToken($userID: ID!) {
	createAccessToken(user: $userID, scopes: ["site-admin:sudo"], note: "") {
		id
	}
}
`,
				variables: map[string]interface{}{
					"userID": userClient.AuthenticatedUserID(),
				},
			},
		}

		// todo: sourcegraph/sourcegraph › cmd/frontend/graphqlbackend/oobmigrations.go

		if *dotcom {
			tests = append(tests,
				gqlTest{
					name: "look up user by email",
					query: `
{
	user(email: "alice@example.com") {
		id
	}
}
`,
				},
			)
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				err := userClient.GraphQL("", test.query, test.variables, nil)
				got := fmt.Sprintf("%v", err)
				if !strings.Contains(got, "must be site admin") {
					t.Fatalf(`Want "must be site admin"" error but got %q`, got)
				}
			})
		}
	})
}
