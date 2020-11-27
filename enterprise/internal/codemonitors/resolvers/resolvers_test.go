package resolvers

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

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
		createdAt:       r.Now(),
		changedBy:       userID,
		changedAt:       r.Now(),
		description:     "test monitor",
		enabled:         true,
		namespaceUserID: &userID,
		namespaceOrgID:  nil,
	}

	// Create a monitor.
	ctx = actor.WithActor(ctx, actor.FromUser(userID))
	got, err := r.insertTestMonitorWithOpts(ctx, t)
	if err != nil {
		t.Fatal(err)
	}

	want.Resolver = got.(*monitor).Resolver
	if !reflect.DeepEqual(want, got.(*monitor)) {
		t.Fatalf("\ngot:\t %+v,\nwant:\t %+v", got, want)
	}

	// Toggle field enabled from true to false.
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

	// Delete code monitor.
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
	ownerOpt := WithOwner(relay.MarshalID("Org", org.ID))
	m, err := r.insertTestMonitorWithOpts(admContext, t, ownerOpt)
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

func TestIsAllowedToCreate(t *testing.T) {
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

	tests := []struct {
		user    int32
		owner   graphql.ID
		allowed bool
	}{
		{
			user:    member,
			owner:   relay.MarshalID("Org", org.ID),
			allowed: true,
		},
		{
			user:    member,
			owner:   relay.MarshalID("User", notMember),
			allowed: false,
		},

		{
			user:    notMember,
			owner:   relay.MarshalID("Org", org.ID),
			allowed: false,
		},
		{
			user:    siteAdmin,
			owner:   relay.MarshalID("Org", org.ID),
			allowed: true,
		},
		{
			user:    siteAdmin,
			owner:   relay.MarshalID("User", member),
			allowed: true,
		},
		{
			user:    siteAdmin,
			owner:   relay.MarshalID("User", notMember),
			allowed: true,
		},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("user %d", tt.user), func(t *testing.T) {
			ctx := actor.WithActor(context.Background(), actor.FromUser(tt.user))
			if err := r.isAllowedToCreate(ctx, tt.owner); (err != nil) == tt.allowed {
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

	// Create a monitor and make sure the trigger query is enqueued.
	ctx = actor.WithActor(ctx, actor.FromUser(userID))
	postHookOpt := WithPostHooks([]hook{func() error { return r.store.EnqueueTriggerQueries(ctx) }})
	_, err := r.insertTestMonitorWithOpts(ctx, t, postHookOpt)
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

	triggerEventEndCursor := string(relay.MarshalID(monitorTriggerEventKind, 1))
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
					CreatedAt:   marshalDateTime(t, r.Now()),
					Trigger: apitest.Trigger{
						Id:    string(relay.MarshalID(monitorTriggerQueryKind, 1)),
						Query: "repo:foo",
						Events: apitest.TriggerEventConnection{
							Nodes: []apitest.TriggerEvent{
								{
									Id:        string(relay.MarshalID(monitorTriggerEventKind, 1)),
									Status:    "PENDING",
									Timestamp: r.Now().UTC().Format(time.RFC3339),
									Message:   nil,
								},
							},
							TotalCount: 1,
							PageInfo: apitest.PageInfo{
								HasNextPage: true,
								EndCursor:   &triggerEventEndCursor,
							},
						},
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
						events {
							totalCount
							nodes {
								id
								status
								timestamp
								message
							}
							pageInfo {
								hasNextPage
								endCursor
							}
						}
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

	// Create 2 test users.
	user1Name := "cm-user1"
	user1ID := insertTestUser(t, dbconn.Global, user1Name, true)
	ns1 := relay.MarshalID("User", user1ID)

	user2Name := "cm-user2"
	user2ID := insertTestUser(t, dbconn.Global, user2Name, true)
	ns2 := relay.MarshalID("User", user2ID)

	// Create a code monitor with 1 trigger and 2 actions.
	ctx = actor.WithActor(ctx, actor.FromUser(user1ID))
	actionOpt := WithActions([]*graphqlbackend.CreateActionArgs{
		{
			Email: &graphqlbackend.CreateActionEmailArgs{
				Enabled:    true,
				Priority:   "NORMAL",
				Recipients: []graphql.ID{ns1},
				Header:     "header action 1",
			}},
		{
			Email: &graphqlbackend.CreateActionEmailArgs{
				Enabled:    true,
				Priority:   "NORMAL",
				Recipients: []graphql.ID{ns1, ns2},
				Header:     "header action 2",
			},
		},
	})
	_, err := r.insertTestMonitorWithOpts(ctx, t, actionOpt)
	if err != nil {
		t.Fatal(err)
	}

	// Update the code monitor.
	// We update all fields, delete one action, and add a new action.
	schema, err := graphqlbackend.NewSchema(nil, nil, nil, r)
	if err != nil {
		t.Fatal(err)
	}
	updateInput := map[string]interface{}{
		"monitorID": string(relay.MarshalID(monitorKind, 1)),
		"triggerID": string(relay.MarshalID(monitorTriggerQueryKind, 1)),
		"actionID":  string(relay.MarshalID(monitorActionEmailKind, 1)),
		"user1ID":   ns1,
		"user2ID":   ns2,
	}
	got := apitest.UpdateCodeMonitorResponse{}
	campaignApitest.MustExec(ctx, t, schema, updateInput, &got, editMonitor)

	want := apitest.UpdateCodeMonitorResponse{
		UpdateCodeMonitor: apitest.Monitor{
			Id:          string(relay.MarshalID(monitorKind, 1)),
			Description: "updated test monitor",
			Enabled:     false,
			Owner: apitest.UserOrg{
				Name: user1Name,
			},
			CreatedBy: apitest.UserOrg{
				Name: user1Name,
			},
			CreatedAt: marshalDateTime(t, r.store.Now()),
			Trigger: apitest.Trigger{
				Id:    string(relay.MarshalID(monitorTriggerQueryKind, 1)),
				Query: "repo:bar",
			},
			Actions: apitest.ActionConnection{
				Nodes: []apitest.Action{{
					ActionEmail: apitest.ActionEmail{
						Id:       string(relay.MarshalID(monitorActionEmailKind, 1)),
						Enabled:  false,
						Priority: "CRITICAL",
						Recipients: apitest.RecipientsConnection{
							Nodes: []apitest.UserOrg{
								{
									Name: user2Name,
								},
							},
						},
						Header: "updated header action 1",
					}}, {
					ActionEmail: apitest.ActionEmail{
						Id:       string(relay.MarshalID(monitorActionEmailKind, 3)),
						Enabled:  true,
						Priority: "NORMAL",
						Recipients: apitest.RecipientsConnection{
							Nodes: []apitest.UserOrg{
								{
									Name: user1Name,
								},
								{
									Name: user2Name,
								},
							},
						},
						Header: "header action 3",
					}},
				},
			},
		}}

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

mutation ($monitorID: ID!, $triggerID: ID!, $actionID: ID!, $user1ID: ID!, $user2ID: ID!) {
  updateCodeMonitor(
    monitor: {id: $monitorID, update: {description: "updated test monitor", enabled: false, namespace: $user1ID}},
	trigger: {id: $triggerID, update: {query: "repo:bar"}},
	actions: [
	  {email: {id: $actionID, update: {enabled: false, priority: CRITICAL, recipients: [$user2ID], header: "updated header action 1"}}}
	  {email: {update: {enabled: true, priority: NORMAL, recipients: [$user1ID, $user2ID], header: "header action 3"}}}
    ]
  )
  {
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
