package authtest

import (
	"testing"
)

func TestSiteAdminEndpoints(t *testing.T) {
	// Create a test user (authtest-user-1) which is not a site admin, the user
	// should receive access denied for site admin endpoints.
	const testUsername = "authtest-user-1"
	testUserID, err := client.CreateUser(testUsername, testUsername+"@sourcegraph.com")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := client.DeleteUser(testUserID, true)
		if err != nil {
			t.Fatal(err)
		}
	}()

	t.Run("HTTP endpoints", func(t *testing.T) {
		// TODO(jchen): Add HTTP endpoints that are site-admin-only
	})

	t.Run("GraphQL queries", func(t *testing.T) {
		// TODO(jchen): Add GraphQL queries that are site-admin-only
	})
}
