package graphqlbackend

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func TestSiteConfigurationChangeResolverReproducedDiff(t *testing.T) {
	testCases := []struct {
		name     string
		resolver SiteConfigurationChangeResolver
		expected bool
	}{
		{
			name:     "both siteConfig and previousSiteConfig are nil",
			resolver: SiteConfigurationChangeResolver{},
			expected: false,
		},
		{
			name: "siteConfig is nil",
			resolver: SiteConfigurationChangeResolver{
				previousSiteConfig: &database.SiteConfig{},
			},
			expected: false,
		},
		{
			name: "previousSiteConfig is nil",
			resolver: SiteConfigurationChangeResolver{
				siteConfig: &database.SiteConfig{},
			},
			expected: false,
		},

		{
			name: "siteConfig.RedactedContents is non-empty but previousSiteConfig is nil",
			resolver: SiteConfigurationChangeResolver{
				siteConfig: &database.SiteConfig{
					RedactedContents: "foo",
				},
			},
			expected: true,
		},

		{
			name: "both siteConfig.RedactedContents and previousSiteConfig.RedactedContents are empty",
			resolver: SiteConfigurationChangeResolver{
				siteConfig:         &database.SiteConfig{},
				previousSiteConfig: &database.SiteConfig{},
			},
			expected: false,
		},

		{
			name: "siteConfig.RedactedContents is empty",
			resolver: SiteConfigurationChangeResolver{
				siteConfig:         &database.SiteConfig{},
				previousSiteConfig: &database.SiteConfig{RedactedContents: "foo"},
			},
			expected: false,
		},
		{
			name: "previousSiteConfig.RedactedContents is empty",
			resolver: SiteConfigurationChangeResolver{
				siteConfig:         &database.SiteConfig{RedactedContents: "foo"},
				previousSiteConfig: &database.SiteConfig{},
			},
			expected: false,
		},

		{
			name: "both siteConfig.RedactedContents and previousSiteConfig.RedactedContents is non-empty",
			resolver: SiteConfigurationChangeResolver{
				siteConfig:         &database.SiteConfig{RedactedContents: "foo"},
				previousSiteConfig: &database.SiteConfig{RedactedContents: "foo"},
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.resolver.ReproducedDiff() != tc.expected {
				t.Errorf("mismatched value for ReproducedDiff, expected %v, but got %v", tc.expected, tc.resolver.ReproducedDiff())
			}
		})
	}

}

func TestSiteConfigurationDiff(t *testing.T) {
	stubs := setupSiteConfigStubs(t)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: stubs.users[0].ID})
	schemaResolver, err := newSchemaResolver(stubs.db, gitserver.NewClient()).Site().Configuration(ctx)
	if err != nil {
		t.Fatalf("failed to create schemaResolver: %v", err)
	}

	expectedNodes := []struct {
		ID           int32
		AuthorUserID int32
		Diff         *string
	}{
		{
			ID:           5,
			AuthorUserID: 1,
			Diff: stringPtr(`--- ID: 4
+++ ID: 5
@@ -1,4 +1,3 @@
 {
-  "disableAutoGitUpdates": false,
   "auth.Providers": []
 }
\ No newline at end of file
`),
		},
		{
			ID:           4,
			AuthorUserID: 1,
			Diff: stringPtr(`--- ID: 3
+++ ID: 4
@@ -1,4 +1,4 @@
 {
-  "disableAutoGitUpdates": true,
+  "disableAutoGitUpdates": false,
   "auth.Providers": []
 }
\ No newline at end of file
`),
		},
		{
			ID:           3,
			AuthorUserID: 2,
			Diff: stringPtr(`--- ID: 2
+++ ID: 3
@@ -1,3 +1,4 @@
 {
+  "disableAutoGitUpdates": true,
   "auth.Providers": []
 }
\ No newline at end of file
`),
		},
		{
			ID:           2,
			AuthorUserID: 0,
			Diff: stringPtr(`--- ID: 1
+++ ID: 2
@@ -1,17 +1,3 @@
 {
-  // The externally accessible URL for Sourcegraph (i.e., what you type into your browser)
-  // This is required to be configured for Sourcegraph to work correctly.
-  // "externalURL": "https://sourcegraph.example.com",
-  // The authentication provider to use for identifying and signing in users.
-  // Only one entry is supported.
-  //
-  // The builtin auth provider with signup disallowed (shown below) means that
-  // after the initial site admin signs in, all other users must be invited.
-  //
-  // Other providers are documented at https://docs.sourcegraph.com/admin/auth.
-  "auth.providers": [
-    {
-      "type": "builtin"
-    }
-  ],
+  "auth.Providers": []
 }
\ No newline at end of file
`),
		},
		{
			ID:           1,
			AuthorUserID: 0,
			Diff: stringPtr(`--- ID: 0
+++ ID: 1
@@ -1 +1,17 @@
+{
+  // The externally accessible URL for Sourcegraph (i.e., what you type into your browser)
+  // This is required to be configured for Sourcegraph to work correctly.
+  // "externalURL": "https://sourcegraph.example.com",
+  // The authentication provider to use for identifying and signing in users.
+  // Only one entry is supported.
+  //
+  // The builtin auth provider with signup disallowed (shown below) means that
+  // after the initial site admin signs in, all other users must be invited.
+  //
+  // Other providers are documented at https://docs.sourcegraph.com/admin/auth.
+  "auth.providers": [
+    {
+      "type": "builtin"
+    }
+  ],
+}
\ No newline at end of file
`),
		},
	}

	testCases := []struct {
		name string
		args *graphqlutil.ConnectionResolverArgs
	}{
		// We have tests for pagination so we can skip that here and just check for the diff for all
		// the nodes in both the directions.
		{
			name: "first: 10",
			args: &graphqlutil.ConnectionResolverArgs{First: int32Ptr(10)},
		},
		{
			name: "last: 10",
			args: &graphqlutil.ConnectionResolverArgs{Last: int32Ptr(10)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			connectionResolver, err := schemaResolver.History(ctx, tc.args)
			if err != nil {
				t.Fatalf("failed to get history: %v", err)
			}

			nodes, err := connectionResolver.Nodes(ctx)
			if err != nil {
				t.Fatalf("failed to get nodes: %v", err)
			}

			totalNodes, totalExpectedNodes := len(nodes), len(expectedNodes)
			if totalNodes != totalExpectedNodes {
				t.Fatalf("mismatched number of nodes, expected %d, got: %d", totalExpectedNodes, totalNodes)
			}

			for i := 0; i < totalNodes; i++ {
				siteConfig, expectedNode := nodes[i].siteConfig, expectedNodes[i]

				if siteConfig.ID != expectedNode.ID {
					t.Errorf("mismatched node ID, expected: %d, but got: %d", siteConfig.ID, expectedNode.ID)
				}

				if siteConfig.AuthorUserID != expectedNode.AuthorUserID {
					t.Errorf("mismatched node AuthorUserID, expected: %d, but got: %d", siteConfig.ID, expectedNode.ID)
				}

				if !nodes[i].ReproducedDiff() {
					t.Fatal("expected reproducedDiff to be true but got false")
				}

				if diff := cmp.Diff(*expectedNode.Diff, *nodes[i].Diff()); diff != "" {
					t.Errorf("mismatched node diff (-want, +got):\n%s ", diff)
				}
			}
		})
	}
}
