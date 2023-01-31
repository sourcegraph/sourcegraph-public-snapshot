package graphqlbackend

import (
	"context"
	"fmt"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func TestSiteConfigurationDiff(t *testing.T) {
	stubs := setupSiteConfigStubs(t)

	ctx := actor.WithActor(context.Background(), &actor.Actor{UID: stubs.users[0].ID})
	schemaResolver, err := newSchemaResolver(stubs.db, gitserver.NewClient()).Site().Configuration(ctx)
	if err != nil {
		t.Fatalf("failed to create schemaResolver: %v", err)
	}

	testCases := []struct {
		name string
		args *graphqlutil.ConnectionResolverArgs
		// TODO: expectedNodes
	}{
		{
			name: "first: 1",
			args: &graphqlutil.ConnectionResolverArgs{First: int32Ptr(2)},
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

			for _, node := range nodes {
				fmt.Printf("%v\n", node.ID())
				diff := node.Diff()
				fmt.Printf("%T\n", diff)
				fmt.Printf("%v\n", len(*diff))
				fmt.Printf("%s", *diff)
			}
		})
	}
}
