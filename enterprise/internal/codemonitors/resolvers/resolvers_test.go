package resolvers

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codemonitors/resolvers/apitest"

	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	campaignApitest "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/resolvers/apitest"
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
