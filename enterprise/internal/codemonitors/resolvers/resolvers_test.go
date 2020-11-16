package resolvers

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	campaignApitest "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codemonitors/resolvers/apitest"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func init() {
	dbtesting.DBNameSuffix = "codemonitorsdb"
}

func TestCreateCodeMonitor(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)
	r := newTestResolver(t)

	userID := insertTestUser(t, dbconn.Global, "cm-user1", true)

	want := &monitor{
		id:              1,
		createdBy:       userID,
		createdAt:       r.clock(),
		changedBy:       userID,
		changedAt:       r.clock(),
		description:     "test monitor",
		enabled:         true,
		namespaceUserID: &userID,
		namespaceOrgID:  nil,
	}

	// Create a monitor
	ctx = actor.WithActor(ctx, actor.FromUser(userID))
	ns := relay.MarshalID("User", userID)
	got, err := r.insertTestMonitor(ctx, t, ns)
	if err != nil {
		t.Fatal(err)
	}
	want.Resolver = got.(*monitor).Resolver
	if !reflect.DeepEqual(want, got.(*monitor)) {
		t.Fatalf("\ngot:\t %+v,\nwant:\t %+v", got, want)
	}

	// Toggle field enabled from true to false
	got, err = r.ToggleCodeMonitor(ctx, &graphqlbackend.ToggleCodeMonitorArgs{
		Id:      relay.MarshalID(monitorKind, got.(*monitor).id),
		Enabled: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.(*monitor).enabled {
		t.Fatalf("got enabled=%T, want enabled=%T", got.(*monitor).enabled, false)
	}

	// Delete code monitor
	_, err = r.DeleteCodeMonitor(ctx, &graphqlbackend.DeleteCodeMonitorArgs{Id: got.ID()})
	if err != nil {
		t.Fatal(err)
	}
	_, err = r.monitorForIDInt32(ctx, t, got.(*monitor).id)
	if err == nil {
		t.Fatalf("monitor should have been deleted")
	}
}

func TestIsAllowedToEdit(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	dbtesting.SetupGlobalTestDB(t)

	// Setup users and org
	member := insertTestUser(t, dbconn.Global, "cm-user1", false)
	notMember := insertTestUser(t, dbconn.Global, "cm-user2", false)
	siteAdmin := insertTestUser(t, dbconn.Global, "cm-user3", true)

	admContext := actor.WithActor(context.Background(), actor.FromUser(siteAdmin))
	org, err := db.Orgs.Create(admContext, "cm-test-org", nil)
	if err != nil {
		t.Fatal(err)
	}
	addUserToOrg(t, dbconn.Global, member, org.ID)

	r := newTestResolver(t)

	// Create a monitor and set org as owner.
	ns := relay.MarshalID("Org", org.ID)
	m, err := r.insertTestMonitor(admContext, t, ns)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		user    int32
		allowed bool
	}{
		{
			user:    member,
			allowed: true,
		},
		{
			user:    notMember,
			allowed: false,
		},
		{
			user:    siteAdmin,
			allowed: true,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("user %d", tt.user), func(t *testing.T) {
			ctx := actor.WithActor(context.Background(), actor.FromUser(tt.user))
			if err := r.isAllowedToEdit(ctx, m.ID()); (err != nil) == tt.allowed {
				t.Fatalf("unexpected permissions for user %d", tt.user)
			}
		})
	}
}

func TestQueryMonitor(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)
	r := newTestResolver(t)

	userName := "cm-user1"
	userID := insertTestUser(t, dbconn.Global, userName, true)

	// Create a monitor
	ctx = actor.WithActor(ctx, actor.FromUser(userID))
	ns := relay.MarshalID("User", userID)
	_, err := r.insertTestMonitor(ctx, t, ns)
	if err != nil {
		t.Fatal(err)
	}

	actorCtx := actor.WithActor(ctx, actor.FromUser(userID))
	schema, err := graphqlbackend.NewSchema(nil, nil, nil, r)
	if err != nil {
		t.Fatal(err)
	}
	input := map[string]interface{}{
		"userName": userName,
	}
	response := apitest.Response{}
	campaignApitest.MustExec(actorCtx, t, schema, input, &response, queryMonitor)

	want := apitest.Response{
		User: apitest.User{
			Monitors: apitest.MonitorConnection{
				TotalCount: 1,
				Nodes: []apitest.Monitor{{
					Id:          string(relay.MarshalID(monitorKind, 1)),
					Description: "test monitor",
					Enabled:     true,
					Owner:       apitest.UserOrg{Name: userName},
					CreatedBy:   apitest.UserOrg{Name: userName},
					CreatedAt:   marshalDateTime(t, r.clock()),
					Trigger: apitest.Trigger{
						Id:    string(relay.MarshalID(monitorTriggerQueryKind, 1)),
						Query: "repo:foo",
					},
					Actions: apitest.ActionConnection{
						TotalCount: 1,
						Nodes: []apitest.Action{{
							ActionEmail: apitest.ActionEmail{
								Id:       string(relay.MarshalID(monitorActionEmailKind, 1)),
								Enabled:  true,
								Priority: "NORMAL",
								Recipients: apitest.RecipientsConnection{
									TotalCount: 1,
									Nodes: []apitest.UserOrg{{
										Name: userName,
									}},
								},
								Header: "test header",
							},
						}},
					},
				}},
			},
		},
	}
	if !reflect.DeepEqual(&response, &want) {
		t.Fatalf("\ngot:\t%+v\nwant:\t%+v\n", response, want)
	}
}

const queryMonitor = `
fragment u on User { id, username }
fragment o on Org { id, name }

query($userName: String!){
	user(username:$userName){
		monitors{
			totalCount
			nodes{
				id
				description
				enabled
				owner {
					... on User { ...u }
					... on Org { ...o }
				}
				createdBy { ...u }
				createdAt
				trigger {
					... on MonitorQuery {
						id
						query
					}
				}
				actions{
					totalCount
					nodes{
						... on MonitorEmail{
							id
							priority
							header
							enabled
							recipients {
								totalCount
								nodes {
									... on User { ...u }
									... on Org { ...o }
								}
							}
						}
					}
				}
			}
		}
	}
}
`

func TestEditCodeMonitor(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)
	r := newTestResolver(t)
	userName := "cm-user1"
	userID := insertTestUser(t, dbconn.Global, userName, true)

	// Create a monitor.
	ctx = actor.WithActor(ctx, actor.FromUser(userID))
	ns := relay.MarshalID("User", userID)
	_, err := r.insertTestMonitor(ctx, t, ns)
	if err != nil {
		t.Fatal(err)
	}

	actorCtx := actor.WithActor(ctx, actor.FromUser(userID))
	schema, err := graphqlbackend.NewSchema(nil, nil, nil, r)
	if err != nil {
		t.Fatal(err)
	}

	// Query the monitor we just inserted to get the IDs of the monitor, trigger, and
	// action. We could use the output from insertTestMonitor instead, but going
	// through the GraphQL server makes it much easier to retrieve the deeply nested IDs.
	input := map[string]interface{}{
		"userName": userName,
	}
	queryResponse := apitest.Response{}
	campaignApitest.MustExec(actorCtx, t, schema, input, &queryResponse, queryMonitor)
	input = map[string]interface{}{
		"monitorID": queryResponse.User.Monitors.Nodes[0].Id,
		"triggerID": queryResponse.User.Monitors.Nodes[0].Trigger.Id,
		"actionID":  queryResponse.User.Monitors.Nodes[0].Actions.Nodes[0].Id,
		"userID":    relay.MarshalID("User", userID),
	}

	want := apitest.UpdateCodeMonitorResponse{
		UpdateCodeMonitor: apitest.Monitor{
			Id:          queryResponse.User.Monitors.Nodes[0].Id,
			Description: "updated test monitor",
			Enabled:     false,
			Owner: apitest.UserOrg{
				Name: userName,
			},
			CreatedBy: apitest.UserOrg{
				Name: userName,
			},
			CreatedAt: marshalDateTime(t, r.clock()),
			Trigger: apitest.Trigger{
				Id:    queryResponse.User.Monitors.Nodes[0].Trigger.Id,
				Query: "repo:bar",
			},
			Actions: apitest.ActionConnection{
				Nodes: []apitest.Action{{
					ActionEmail: apitest.ActionEmail{
						Id:       queryResponse.User.Monitors.Nodes[0].Actions.Nodes[0].Id,
						Enabled:  false,
						Priority: "CRITICAL",
						Recipients: apitest.RecipientsConnection{
							Nodes: []apitest.UserOrg{
								{
									Name: userName,
								},
							},
						},
						Header: "updated test header",
					}},
				},
			},
		}}

	// Update the code monitor.
	got := apitest.UpdateCodeMonitorResponse{}
	campaignApitest.MustExec(actorCtx, t, schema, input, &got, editMonitor)

	if !reflect.DeepEqual(&got, &want) {
		t.Fatalf("\ngot:\t%+v\nwant:\t%+v\n", got, want)
	}
}

const editMonitor = `
fragment u on User {
  id
  username
}

fragment o on Org {
  id
  name
}

mutation ($monitorID: ID!, $triggerID: ID!, $actionID: ID!, $userID: ID!) {
  updateCodeMonitor(monitor: {id: $monitorID, update: {description: "updated test monitor", enabled: false, namespace: $userID}}, trigger: {id: $triggerID, update: {query: "repo:bar"}}, actions: [{email: {id: $actionID, update: {enabled: false, priority: CRITICAL, recipients: [$userID], header: "updated test header"}}}]) {
    id
    description
    enabled
    owner {
      ... on User {
        ...u
      }
      ... on Org {
        ...o
      }
    }
    createdBy {
      ...u
    }
    createdAt
    trigger {
      ... on MonitorQuery {
        id
        query
      }
    }
    actions {
      nodes {
        ... on MonitorEmail {
          id
          enabled
          priority
          header
          recipients {
            nodes {
              ... on User {
                username
              }
              ... on Org {
                name
              }
            }
          }
        }
      }
    }
  }
}
`

func TestCreateActionForMonitor(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)
	r := newTestResolver(t)

	user1 := "cm-user1"
	user1ID := insertTestUser(t, dbconn.Global, user1, true)
	user2 := "cm-user2"
	user2ID := insertTestUser(t, dbconn.Global, user2, true)

	// Create a monitor.
	ctx = actor.WithActor(ctx, actor.FromUser(user1ID))
	ns := relay.MarshalID("User", user1ID)
	m, err := r.insertTestMonitor(ctx, t, ns)
	if err != nil {
		t.Fatal(err)
	}

	want := apitest.ActionEmail{
		Id:       string(relay.MarshalID(monitorActionEmailKind, 2)),
		Enabled:  true,
		Priority: "CRITICAL",
		Recipients: apitest.RecipientsConnection{
			Nodes:      []apitest.UserOrg{{user1}, {user2}},
			TotalCount: 2,
		},
		Header: "another action",
	}

	// Add a second action to the monitor we just created.
	res, err := r.CreateCodeMonitorAction(ctx, &graphqlbackend.CreateActionForMonitorArgs{
		Id: m.ID(),
		Action: &graphqlbackend.CreateActionArgs{
			Email: &graphqlbackend.CreateActionEmailArgs{
				Enabled:    want.Enabled,
				Priority:   want.Priority,
				Recipients: []graphql.ID{relay.MarshalID("User", user1ID), relay.MarshalID("User", user2ID)},
				Header:     want.Header,
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Check that we hydrated the returned resolver by checking whether we can get the
	// recipients from the MonitorAction resolver.
	e, ok := res.ToMonitorEmail()
	if !ok {
		t.Fatal("expected monitorEmail")
	}
	rec, err := e.Recipients(ctx, &graphqlbackend.ListRecipientsArgs{
		First: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	nsRes, err := rec.Nodes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got := nsRes[0].NamespaceName(); got != user1 {
		t.Fatalf("got %s, want %s", got, user1)
	}
	if got := nsRes[1].NamespaceName(); got != user2 {
		t.Fatalf("got %s, want %s", got, user2)
	}

	// Query the monitor with all its actions.
	input := map[string]interface{}{
		"userName": user1,
	}
	queryResponse := apitest.Response{}
	schema, err := graphqlbackend.NewSchema(nil, nil, nil, r)
	campaignApitest.MustExec(ctx, t, schema, input, &queryResponse, queryMonitor)

	// Get the second action of the user's first monitor.
	gotMonitors := queryResponse.User.Monitors.Nodes
	if len(gotMonitors) != 1 {
		t.Fatalf("user should have had 1 monitor")
	}
	// First monitor.
	gotActions := gotMonitors[0].Actions.Nodes
	if len(gotActions) != 2 {
		t.Fatalf("user should have had 2 actions")
	}
	// Second action.
	got := gotActions[1].ActionEmail
	if !reflect.DeepEqual(&got, &want) {
		t.Fatalf("\ngot:\t%+v\nwant:\t%+v\n", got, want)
	}
}
