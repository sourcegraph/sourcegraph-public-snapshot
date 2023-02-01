package graphqlbackend

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go/relay"

	gqlerrors "github.com/graph-gophers/graphql-go/errors"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	actorID = 54123451
)

type fakeTeamsDb struct {
	database.TeamStore
	list       []*types.Team
	lastUsedID int32
}

func (teams *fakeTeamsDb) CreateTeam(_ context.Context, t *types.Team) error {
	teams.lastUsedID++
	u := *t
	u.ID = teams.lastUsedID
	teams.list = append(teams.list, &u)
	return nil
}

func (teams *fakeTeamsDb) UpdateTeam(_ context.Context, t *types.Team) error {
	if t == nil {
		return errors.New("UpdateTeam: team cannot be nil")
	}
	if t.ID == 0 {
		return errors.New("UpdateTeam: team.ID must be set (not 0)")
	}
	for _, u := range teams.list {
		if u.ID == t.ID {
			*u = *t
			return nil
		}
	}
	return errors.Newf("UpdateTeam: cannot find team with ID=%d", t.ID)
}

func (teams *fakeTeamsDb) GetTeamByID(_ context.Context, id int32) (*types.Team, error) {
	for _, t := range teams.list {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, database.TeamNotFoundError{}
}

func (teams *fakeTeamsDb) GetTeamByName(_ context.Context, name string) (*types.Team, error) {
	for _, t := range teams.list {
		if t.Name == name {
			return t, nil
		}
	}
	return nil, database.TeamNotFoundError{}
}

func (teams *fakeTeamsDb) DeleteTeam(_ context.Context, id int32) error {
	for i, t := range teams.list {
		if t.ID == id {
			maxI := len(teams.list) - 1
			teams.list[i], teams.list[maxI] = teams.list[maxI], teams.list[i]
			teams.list = teams.list[:maxI]
			return nil
		}
	}
	return database.TeamNotFoundError{}
}

func setupDB() (*database.MockDB, *fakeTeamsDb) {
	ts := &fakeTeamsDb{}
	db := database.NewMockDB()
	db.TeamsFunc.SetDefaultReturn(ts)
	db.WithTransactFunc.SetDefaultHook(func(_ context.Context, callback func(database.DB) error) error {
		return callback(db)
	})
	return db, ts
}

func TestCreateTeamBare(t *testing.T) {
	db, ts := setupDB()
	ctx, user, _ := fakeUser(t, context.Background(), db, true)
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation CreateTeam($name: String!) {
			createTeam(name: $name) {
				name
			}
		}`,
		ExpectedResult: `{
			"createTeam": {
				"name": "team-name-testing"
			}
		}`,
		Variables: map[string]any{
			"name": "team-name-testing",
		},
	})
	expected := &types.Team{
		ID:        1,
		Name:      "team-name-testing",
		CreatorID: user.ID,
	}
	if diff := cmp.Diff([]*types.Team{expected}, ts.list); diff != "" {
		t.Errorf("unexpected teams in fake database (-want,+got):\n%s", diff)
	}
}

func TestCreateTeamDisplayName(t *testing.T) {
	db, _ := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation CreateTeam($name: String!, $displayName: String!) {
			createTeam(name: $name, displayName: $displayName) {
				displayName
			}
		}`,
		ExpectedResult: `{
			"createTeam": {
				"displayName": "Team Display Name"
			}
		}`,
		Variables: map[string]any{
			"name":        "team-name-testing",
			"displayName": "Team Display Name",
		},
	})
}

func TestCreateTeamReadOnlyDefault(t *testing.T) {
	db, _ := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation CreateTeam($name: String!) {
			createTeam(name: $name) {
				readonly
			}
		}`,
		ExpectedResult: `{
			"createTeam": {
				"readonly": false
			}
		}`,
		Variables: map[string]any{
			"name": "team-name-testing",
		},
	})
}

func TestCreateTeamReadOnlyTrue(t *testing.T) {
	db, _ := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation CreateTeam($name: String!, $readonly: Boolean!) {
			createTeam(name: $name, readonly: $readonly) {
				readonly
			}
		}`,
		ExpectedResult: `{
			"createTeam": {
				"readonly": true
			}
		}`,
		Variables: map[string]any{
			"name":     "team-name-testing",
			"readonly": true,
		},
	})
}

func TestCreateTeamParentByID(t *testing.T) {
	db, ts := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	err := ts.CreateTeam(ctx, &types.Team{
		Name: "team-name-parent",
	})
	if err != nil {
		t.Fatal(err)
	}
	parentTeam, err := ts.GetTeamByName(ctx, "team-name-parent")
	if err != nil {
		t.Fatal(err)
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation CreateTeam($name: String!, $parentTeamID: ID!) {
			createTeam(name: $name, parentTeam: $parentTeamID) {
				parentTeam {
					name
				}
			}
		}`,
		ExpectedResult: `{
			"createTeam": {
				"parentTeam": {
					"name": "team-name-parent"
				}
			}
		}`,
		Variables: map[string]any{
			"name":         "team-name-testing",
			"parentTeamID": string(relay.MarshalID("Team", parentTeam.ID)),
		},
	})
}

func TestCreateTeamParentByName(t *testing.T) {
	db, ts := setupDB()
	parentTeam := types.Team{Name: "team-name-parent"}
	if err := ts.CreateTeam(context.Background(), &parentTeam); err != nil {
		t.Fatal(err)
	}
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation CreateTeam($name: String!, $parentTeamName: String!) {
			createTeam(name: $name, parentTeamName: $parentTeamName) {
				parentTeam {
					name
				}
			}
		}`,
		ExpectedResult: `{
			"createTeam": {
				"parentTeam": {
					"name": "team-name-parent"
				}
			}
		}`,
		Variables: map[string]any{
			"name":           "team-name-testing",
			"parentTeamName": "team-name-parent",
		},
	})
}

func TestUpdateTeamByID(t *testing.T) {
	db, ts := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	if err := ts.CreateTeam(ctx, &types.Team{
		Name:        "team-name-testing",
		DisplayName: "Display Name",
	}); err != nil {
		t.Fatalf("failed to create a team: %s", err)
	}
	team, err := ts.GetTeamByName(ctx, "team-name-testing")
	if err != nil {
		t.Fatalf("failed to get fake team: %s", err)
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation UpdateTeam($id: ID!, $newDisplayName: String!) {
			updateTeam(id: $id, displayName: $newDisplayName) {
				displayName
			}
		}`,
		ExpectedResult: `{
			"updateTeam": {
				"displayName": "Updated Display Name"
			}
		}`,
		Variables: map[string]any{
			"id":             string(relay.MarshalID("Team", team.ID)),
			"newDisplayName": "Updated Display Name",
		},
	})
	wantTeams := []*types.Team{
		{
			ID:          1,
			Name:        "team-name-testing",
			DisplayName: "Updated Display Name",
		},
	}
	if diff := cmp.Diff(wantTeams, ts.list); diff != "" {
		t.Errorf("fake teams storage (-want,+got):\n%s", diff)
	}
}

func TestUpdateTeamByName(t *testing.T) {
	db, ts := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	if err := ts.CreateTeam(ctx, &types.Team{
		Name:        "team-name-testing",
		DisplayName: "Display Name",
	}); err != nil {
		t.Fatalf("failed to create a team: %s", err)
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation UpdateTeam($name: String!, $newDisplayName: String!) {
			updateTeam(name: $name, displayName: $newDisplayName) {
				displayName
			}
		}`,
		ExpectedResult: `{
			"updateTeam": {
				"displayName": "Updated Display Name"
			}
		}`,
		Variables: map[string]any{
			"name":           "team-name-testing",
			"newDisplayName": "Updated Display Name",
		},
	})
	wantTeams := []*types.Team{
		{
			ID:          1,
			Name:        "team-name-testing",
			DisplayName: "Updated Display Name",
		},
	}
	if diff := cmp.Diff(wantTeams, ts.list); diff != "" {
		t.Errorf("fake teams storage (-want,+got):\n%s", diff)
	}
}

func TestUpdateTeamErrorBothNameAndID(t *testing.T) {
	db, ts := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	if err := ts.CreateTeam(ctx, &types.Team{
		Name:        "team-name-testing",
		DisplayName: "Display Name",
	}); err != nil {
		t.Fatalf("failed to create a team: %s", err)
	}
	team, err := ts.GetTeamByName(ctx, "team-name-testing")
	if err != nil {
		t.Fatalf("failed to get fake team: %s", err)
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation UpdateTeam($name: String!, $id: ID!, $newDisplayName: String!) {
			updateTeam(name: $name, id: $id, displayName: $newDisplayName) {
				displayName
			}
		}`,
		ExpectedResult: "null",
		ExpectedErrors: []*gqlerrors.QueryError{
			{
				Message: "team to update is identified by either id or name, but both were specified",
				Path:    []any{"updateTeam"},
			},
		},
		Variables: map[string]any{
			"id":             string(relay.MarshalID("Team", team.ID)),
			"name":           "team-name-testing",
			"newDisplayName": "Updated Display Name",
		},
	})
}

func TestUpdateParentByID(t *testing.T) {
	db, ts := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	if err := ts.CreateTeam(ctx, &types.Team{Name: "parent"}); err != nil {
		t.Fatalf("failed to create parent team: %s", err)
	}
	parentTeam, err := ts.GetTeamByName(ctx, "parent")
	if err != nil {
		t.Fatalf("failed to fetch fake parent team: %s", err)
	}
	if err := ts.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
		t.Fatalf("failed to create a team: %s", err)
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation UpdateTeam($name: String!, $newParentID: ID!) {
			updateTeam(name: $name, parentTeam: $newParentID) {
				parentTeam {
					name
				}
			}
		}`,
		ExpectedResult: `{
			"updateTeam": {
				"parentTeam": {
					"name": "parent"
				}
			}
		}`,
		Variables: map[string]any{
			"name":        "team",
			"newParentID": string(relay.MarshalID("Team", parentTeam.ID)),
		},
	})
}

func TestUpdateParentByName(t *testing.T) {
	db, ts := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	if err := ts.CreateTeam(ctx, &types.Team{Name: "parent"}); err != nil {
		t.Fatalf("failed to create parent team: %s", err)
	}
	if err := ts.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
		t.Fatalf("failed to create a team: %s", err)
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation UpdateTeam($name: String!, $newParentName: String!) {
			updateTeam(name: $name, parentTeamName: $newParentName) {
				parentTeam {
					name
				}
			}
		}`,
		ExpectedResult: `{
			"updateTeam": {
				"parentTeam": {
					"name": "parent"
				}
			}
		}`,
		Variables: map[string]any{
			"name":          "team",
			"newParentName": "parent",
		},
	})
}

func TestUpdateParentErrorBothNameAndID(t *testing.T) {
	db, ts := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	if err := ts.CreateTeam(ctx, &types.Team{Name: "parent"}); err != nil {
		t.Fatalf("failed to create parent team: %s", err)
	}
	parentTeam, err := ts.GetTeamByName(ctx, "parent")
	if err != nil {
		t.Fatalf("failed to fetch fake parent team: %s", err)
	}
	if err := ts.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
		t.Fatalf("failed to create a team: %s", err)
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation UpdateTeam($name: String!, $newParentID: ID!, $newParentName: String!) {
			updateTeam(name: $name, parentTeam: $newParentID, parentTeamName: $newParentName) {
				parentTeam {
					name
				}
			}
		}`,
		ExpectedResult: "null",
		ExpectedErrors: []*gqlerrors.QueryError{
			{
				Message: "parent team is identified by either id or name, but both were specified",
				Path:    []any{"updateTeam"},
			},
		},
		Variables: map[string]any{
			"name":          "team",
			"newParentID":   string(relay.MarshalID("Team", parentTeam.ID)),
			"newParentName": parentTeam.Name,
		},
	})
}

func TestDeleteTeamByID(t *testing.T) {
	db, ts := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	if err := ts.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
		t.Fatalf("failed to create a team: %s", err)
	}
	team, err := ts.GetTeamByName(ctx, "team")
	if err != nil {
		t.Fatalf("cannot find fake team: %s", err)
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation DeleteTeam($id: ID!) {
			deleteTeam(id: $id) {
				alwaysNil
			}
		}`,
		ExpectedResult: `{
			"deleteTeam": {
				"alwaysNil": null
			}
		}`,
		Variables: map[string]any{
			"id": string(relay.MarshalID("Team", team.ID)),
		},
	})
	if diff := cmp.Diff([]*types.Team{}, ts.list); diff != "" {
		t.Errorf("expected no teams in fake db after deleting, (-want,+got):\n%s", diff)
	}
}

func TestDeleteTeamByName(t *testing.T) {
	db, ts := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	if err := ts.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
		t.Fatalf("failed to create a team: %s", err)
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation DeleteTeam($name: String!) {
			deleteTeam(name: $name) {
				alwaysNil
			}
		}`,
		ExpectedResult: `{
			"deleteTeam": {
				"alwaysNil": null
			}
		}`,
		Variables: map[string]any{
			"name": "team",
		},
	})
	if diff := cmp.Diff([]*types.Team{}, ts.list); diff != "" {
		t.Errorf("expected no teams in fake db after deleting, (-want,+got):\n%s", diff)
	}
}

func TestDeleteTeamErrorBothIDAndNameGiven(t *testing.T) {
	db, ts := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	if err := ts.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
		t.Fatalf("failed to create a team: %s", err)
	}
	team, err := ts.GetTeamByName(ctx, "team")
	if err != nil {
		t.Fatalf("cannot find fake team: %s", err)
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation DeleteTeam($id: ID!, $name: String!) {
			deleteTeam(id: $id, name: $name) {
				alwaysNil
			}
		}`,
		ExpectedResult: `{
			"deleteTeam": null
		}`,
		ExpectedErrors: []*gqlerrors.QueryError{
			{
				Message: "team to delete is identified by either id or name, but both were specified",
				Path:    []any{"deleteTeam"},
			},
		},
		Variables: map[string]any{
			"id":   string(relay.MarshalID("Team", team.ID)),
			"name": "team",
		},
	})
}

func TestDeleteTeamNoIdentifierGiven(t *testing.T) {
	db, _ := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation DeleteTeam() {
			deleteTeam() {
				alwaysNil
			}
		}`,
		ExpectedResult: `{
			"deleteTeam": null
		}`,
		ExpectedErrors: []*gqlerrors.QueryError{
			{
				Message: "team to delete is identified by either id or name, but neither was specified",
				Path:    []any{"deleteTeam"},
			},
		},
	})
}

func TestDeleteTeamNotFound(t *testing.T) {
	db, _ := setupDB()
	ctx, _, _ := fakeUser(t, context.Background(), db, true)
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation DeleteTeam($name: String!) {
			deleteTeam(name: $name) {
				alwaysNil
			}
		}`,
		ExpectedResult: `{
			"deleteTeam": null
		}`,
		ExpectedErrors: []*gqlerrors.QueryError{
			{
				Message: `team name="does-not-exist" not found: team not found: <nil>`,
				Path:    []any{"deleteTeam"},
			},
		},
		Variables: map[string]any{
			"name": "does-not-exist",
		},
	})
}

func TestDeleteTeamUnauthorized(t *testing.T) {
	db, ts := setupDB()
	// false in the next line indicates not-site-admin
	ctx, _, _ := fakeUser(t, context.Background(), db, false)
	if err := ts.CreateTeam(ctx, &types.Team{Name: "team"}); err != nil {
		t.Fatalf("failed to create a team: %s", err)
	}
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: ctx,
		Query: `mutation DeleteTeam($name: String!) {
			deleteTeam(name: $name) {
				alwaysNil
			}
		}`,
		ExpectedResult: `{
			"deleteTeam": null
		}`,
		ExpectedErrors: []*gqlerrors.QueryError{
			{
				Message: "only site admins can delete teams",
				Path:    []any{"deleteTeam"},
			},
		},
		Variables: map[string]any{
			"name": "team",
		},
	})
}
