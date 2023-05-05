package authtest

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/auth"
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

	t.Run("debug endpoints", func(t *testing.T) {
		tests := []struct {
			name string
			path string
		}{
			{
				name: "debug",
				path: "/-/debug/",
			}, {
				name: "jaeger",
				path: "/-/debug/jaeger/",
			},
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				resp, err := userClient.Get(*baseURL + test.path)
				if err != nil {
					t.Fatal(err)
				}
				defer func() { _ = resp.Body.Close() }()

				if resp.StatusCode != http.StatusForbidden {
					t.Fatalf(`Want status code %d error but got %d`, http.StatusForbidden, resp.StatusCode)
				}
			})
		}
	})

	t.Run("latest ping", func(t *testing.T) {
		resp, err := userClient.Get(*baseURL + "/site-admin/pings/latest")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()

		if want, got := http.StatusUnauthorized, resp.StatusCode; want != got {
			t.Fatalf("Want %d but got %d", want, got)
		}
	})

	t.Run("usage stats archive", func(t *testing.T) {
		resp, err := userClient.Get(*baseURL + "/site-admin/usage-statistics/archive")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()

		if want, got := http.StatusUnauthorized, resp.StatusCode; want != got {
			t.Fatalf("Want %d but got %d", want, got)
		}
	})

	t.Run("GraphQL queries", func(t *testing.T) {
		type gqlTest struct {
			name      string
			errorStr  string
			query     string
			variables map[string]any
		}
		tests := []gqlTest{
			{
				name: "resetTriggerQueryTimestamps",
				query: `
mutation {
	resetTriggerQueryTimestamps(id: "SUQ6MTIz") {
		alwaysNil
	}
}`,
			}, {
				name: "reindexRepository",
				query: `
mutation {
	reindexRepository(repository: "UmVwb3NpdG9yeTox") {
		alwaysNil
	}
}`,
			},
			{
				name: "updateMirrorRepository",
				query: `
mutation {
	updateMirrorRepository(repository: "UmVwb3NpdG9yeTox") {
		alwaysNil
	}
}`,
			}, {
				name: "addUserToOrganization",
				query: `
mutation {
	addUserToOrganization(organization: "T3JnYW5pemF0aW9uOjE=", username: "alice") {
		alwaysNil
	}
}`,
			}, {
				name: "site.configuration",
				query: `
{
	site {
		configuration {
			id
		}
	}
}`,
			}, {
				name: "site.accessTokens",
				query: `
{
	site {
		accessTokens {
			totalCount
		}
	}
}`,
			}, {
				name: "site.externalAccounts",
				query: `
{
	site {
		externalAccounts {
			nodes {
				id
			}
		}
	}
}`,
			}, {
				name: "updateSiteConfiguration",
				query: `
mutation {
	updateSiteConfiguration(input: "", lastID: 0)
}`,
			}, {
				name: "deletePreciseIndex",
				query: `
mutation {
	deletePreciseIndex(id: "TFNJRjox") {
		alwaysNil
	}
}`,
			}, {
				name: "outOfBandMigrations",
				query: `
{
	outOfBandMigrations {
		id
	}
}`,
			}, {
				name: "setMigrationDirection",
				query: `
mutation {
	setMigrationDirection(id: "TWlncmF0aW9uOjE=", applyReverse: false) {
		alwaysNil
	}
}`,
			}, {
				name: "createAccessToken.ScopeSiteAdminSudo",
				query: `
mutation CreateAccessToken($userID: ID!) {
	createAccessToken(user: $userID, scopes: ["site-admin:sudo"], note: "") {
		id
	}
}`,
				variables: map[string]any{
					"userID": userClient.AuthenticatedUserID(),
				},
			}, {
				name: "setRepositoryPermissionsForUsers",
				query: `
mutation {
	setRepositoryPermissionsForUsers(
		repository: "UmVwb3NpdG9yeTox"
		userPermissions: [{bindID: "alice@example.com"}]
	) {
		alwaysNil
	}
}`,
			}, {
				name: "scheduleRepositoryPermissionsSync",
				query: `
mutation {
	scheduleRepositoryPermissionsSync(repository: "UmVwb3NpdG9yeTox") {
		alwaysNil
	}
}`,
			}, {
				name:     "scheduleUserPermissionsSync",
				errorStr: auth.ErrMustBeSiteAdminOrSameUser.Error(),
				query: `
mutation {
	scheduleUserPermissionsSync(user: "VXNlcjox") {
		alwaysNil
	}
}`,
			}, {
				name: "authorizedUserRepositories",
				query: `
{
	authorizedUserRepositories(username: "alice", first: 1) {
		totalCount
	}
}`,
			}, {
				name: "usersWithPendingPermissions",
				query: `
{
	usersWithPendingPermissions
}`,
			}, {
				name: "authorizedUserRepositories",
				query: `
{
	authorizedUserRepositories(username: "alice", first: 1) {
		totalCount
	}
}`,
			}, {
				name: "setUserEmailVerified",
				query: `
mutation {
	setUserEmailVerified(
		user: "VXNlcjox"
		email: "alice@exmaple.com"
		verified: true
	) {
		alwaysNil
	}
}`,
			}, {
				name: "setUserIsSiteAdmin",
				query: `
mutation {
	setUserIsSiteAdmin(userID: "VXNlcjox", siteAdmin: true) {
		alwaysNil
	}
}`,
			}, {
				name: "invalidateSessionsByID",
				query: `
mutation {
	invalidateSessionsByID(userID: "VXNlcjox") {
		alwaysNil
	}
}`,
			}, {
				name: "triggerObservabilityTestAlert",
				query: `
mutation {
	triggerObservabilityTestAlert(level: "critical") {
		alwaysNil
	}
}`,
			}, {
				name: "reloadSite",
				query: `
mutation {
	reloadSite {
		alwaysNil
	}
}`,
			}, {
				name: "organizations",
				query: `
{
	organizations {
		nodes {
			id
		}
	}
}`,
			}, {
				name: "surveyResponses",
				query: `
{
	surveyResponses {
		nodes {
			id
		}
	}
}`,
			}, {
				name: "repositoryStats",
				query: `
{
	repositoryStats {
		gitDirBytes
	}
}`,
			}, {
				name: "createUser",
				query: `
mutation {
	createUser(username: "alice") {
		resetPasswordURL
	}
}`,
			}, {
				name: "deleteUser",
				query: `
mutation {
	deleteUser(user: "VXNlcjox") {
		alwaysNil
	}
}`,
			}, {
				name: "deleteOrganization",
				query: `
mutation {
	deleteOrganization(organization: "T3JnYW5pemF0aW9uOjE=") {
		alwaysNil
	}
}`,
			}, {
				name: "randomizeUserPassword",
				query: `
mutation {
	randomizeUserPassword(user: "VXNlcjox") {
		resetPasswordURL
	}
}`,
			},
		}

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
				gqlTest{
					name: "dotcom.productLicenses",
					query: `
{
	dotcom {
		productLicenses {
			__typename
		}
	}
}`,
				},
				gqlTest{
					name: "dotcom.createProductSubscription",
					query: `
mutation {
	dotcom {
		createProductSubscription(accountID: "VXNlcjox") {
			id
		}
	}
}`,
				},
				gqlTest{
					name: "dotcom.setProductSubscriptionBilling",
					query: `
mutation {
	dotcom {
		setProductSubscriptionBilling(id: "VXNlcjox") {
			alwaysNil
		}
	}
}`,
				},
				gqlTest{
					name: "dotcom.archiveProductSubscription",
					query: `
mutation {
	dotcom {
		archiveProductSubscription(id: "VXNlcjox") {
			alwaysNil
		}
	}
}`,
				},
				gqlTest{
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
}`,
				},
				gqlTest{
					name: "dotcom.setUserBilling",
					query: `
mutation {
	dotcom {
		setUserBilling(user: "VXNlcjox", billingCustomerID: "404") {
			alwaysNil
		}
	}
}`,
				},
			)
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				err := userClient.GraphQL("", test.query, test.variables, nil)
				got := fmt.Sprintf("%v", err)
				expected := auth.ErrMustBeSiteAdmin.Error()
				if test.errorStr != "" {
					expected = test.errorStr
				}
				// check if it's one of errors that we expect
				if !strings.Contains(got, expected) {
					t.Fatalf(`Want "%s" error, but got "%q"`, expected, got)
				}
			})
		}
	})
}
