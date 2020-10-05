package graphs

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/graphs"
	"github.com/sourcegraph/sourcegraph/schema"
)

var confWithGraphsEnabled = func() *conf.Unified {
	true := true
	return &conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			ExperimentalFeatures: &schema.ExperimentalFeatures{Graphs: &true},
		},
	}
}()

func TestNullIDResilience(t *testing.T) {
	r := &Resolver{store: NewStore(dbconn.Global)}

	s, err := graphqlbackend.NewSchema(nil, nil, nil, r)
	if err != nil {
		t.Fatal(err)
	}

	ctx := backend.WithAuthzBypass(context.Background())

	conf.Mock(confWithGraphsEnabled)
	defer conf.Mock(nil)

	ids := []graphql.ID{
		marshalGraphID(0),
	}

	for _, id := range ids {
		var response struct{ Node struct{ ID string } }

		query := fmt.Sprintf(`query { node(id: %q) { id } }`, id)
		apitest.MustExec(ctx, t, s, nil, &response, query)

		if have, want := response.Node.ID, ""; have != want {
			t.Fatalf("node has wrong ID. have=%q, want=%q", have, want)
		}
	}

	mutations := []string{
		fmt.Sprintf(`mutation { updateGraph(input: { id: %q, name: "n", description: "d", spec: "s" }) { id } }`, marshalGraphID(0)),
		fmt.Sprintf(`mutation { deleteGraph(graph: %q) { alwaysNil } }`, marshalGraphID(0)),
	}

	for _, m := range mutations {
		var response struct{}
		errs := apitest.Exec(ctx, t, s, nil, &response, m)
		if len(errs) == 0 {
			t.Fatalf("expected errors but none returned (mutation: %q)", m)
		}
		if have, want := errs[0].Error(), fmt.Sprintf("graphql: %s", ErrIDIsZero.Error()); have != want {
			t.Fatalf("wrong errors. have=%s, want=%s (mutation: %q)", have, want, m)
		}
	}
}

func TestCreateGraph(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	dbtesting.SetupGlobalTestDB(t)

	conf.Mock(confWithGraphsEnabled)
	defer conf.Mock(nil)

	userID := insertTestUser(t, dbconn.Global)

	store := NewStore(dbconn.Global)

	r := &Resolver{store: store}
	s, err := graphqlbackend.NewSchema(nil, nil, nil, r)
	if err != nil {
		t.Fatal(err)
	}

	description := "My description"
	graph := graphs.Graph{
		OwnerUserID: userID,
		Name:        "my-graph",
		Description: &description,
		Spec:        "my spec",
	}

	input := map[string]interface{}{
		"owner":       string(graphqlbackend.MarshalUserID(userID)),
		"name":        graph.Name,
		"description": graph.Description,
		"spec":        graph.Spec,
	}

	var response struct{ CreateGraph apitestGraph }
	actorCtx := actor.WithActor(ctx, actor.FromUser(userID))

	apitest.MustExec(actorCtx, t, s, input, &response, mutationCreateGraph)

	if response.CreateGraph.ID == "" {
		t.Fatalf("expected graph to be created, but was not")
	}
}

const mutationCreateGraph = `
mutation($owner: ID!, $name: String!, $description: String, $spec: String!) {
  createGraph(input: { owner: $owner, name: $name, description: $description, spec: $spec }) {
    id
  }
}
`

func TestUpdateGraph(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	dbtesting.SetupGlobalTestDB(t)

	conf.Mock(confWithGraphsEnabled)
	defer conf.Mock(nil)

	userID := insertTestUser(t, dbconn.Global)

	now := time.Now().UTC().Truncate(time.Microsecond)
	clock := func() time.Time {
		return now.UTC().Truncate(time.Microsecond)
	}
	store := NewStoreWithClock(dbconn.Global, clock)

	description := "My description"
	graph := graphs.Graph{
		OwnerUserID: userID,
		Name:        "my-graph",
		Description: &description,
		Spec:        "my spec",
	}
	if err := store.CreateGraph(ctx, &graph); err != nil {
		t.Fatal(err)
	}

	r := &Resolver{store: store}
	s, err := graphqlbackend.NewSchema(nil, nil, nil, r)
	if err != nil {
		t.Fatal(err)
	}

	userAPIID := string(graphqlbackend.MarshalUserID(userID))

	input := map[string]interface{}{
		"id":          marshalGraphID(graph.ID),
		"name":        graph.Name + "-updated",
		"description": description + "-updated",
		"spec":        graph.Spec + "-updated",
	}

	var response struct{ UpdateGraph apitestGraph }
	actorCtx := actor.WithActor(ctx, actor.FromUser(userID))
	apitest.MustExec(actorCtx, t, s, input, &response, mutationUpdateGraph)

	have := response.UpdateGraph
	want := apitestGraph{
		ID: have.ID,
		Owner: apitestGraphOwner{
			ID:         userAPIID,
			DatabaseID: userID,
		},
		Name:        graph.Name + "-updated",
		Description: *graph.Description + "-updated",
		Spec:        graph.Spec + "-updated",
	}

	if diff := cmp.Diff(want, have); diff != "" {
		t.Fatalf("unexpected response (-want +got):\n%s", diff)
	}

	// Now we execute it again and make sure we get the same graph back
	apitest.MustExec(actorCtx, t, s, input, &response, mutationUpdateGraph)
	have2 := response.UpdateGraph
	if diff := cmp.Diff(want, have2); diff != "" {
		t.Fatalf("unexpected response (-want +got):\n%s", diff)
	}

	// Execute it again, but with the wrong graph ID.
	graphID, err := unmarshalGraphID(graphql.ID(have2.ID))
	if err != nil {
		t.Fatal(err)
	}
	input["id"] = marshalGraphID(graphID + 999)
	errs := apitest.Exec(actorCtx, t, s, input, &response, mutationUpdateGraph)
	if len(errs) == 0 {
		t.Fatalf("expected errors, got none")
	}
}

const mutationUpdateGraph = `
fragment u on User { id, databaseID }
fragment o on Org  { id, name }

mutation($id: ID!, $name: String!, $description: String, $spec: String!) {
  updateGraph(input: { id: $id, name: $name, description: $description, spec: $spec }) {
    id
    owner {
        ... on User { ...u }
        ... on Org  { ...o }
    }
    name
    description
    spec
  }
}
`
