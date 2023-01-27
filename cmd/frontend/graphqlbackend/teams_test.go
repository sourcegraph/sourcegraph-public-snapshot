package graphqlbackend

import (
	"context"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
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
	t.ID = teams.lastUsedID
	teams.list = append(teams.list, t)
	return nil
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

func TestCreateTeamBare(t *testing.T) {
	fakeTeams := &fakeTeamsDb{}
	db := database.NewMockDB()
	db.TeamsFunc.SetDefaultReturn(fakeTeams)
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: actor.WithActor(context.Background(), &actor.Actor{UID: actorID}),
		Query: `mutation CreateTeam($name: String!) {
			createTeam(name: $name) {
				id
				name
			}
		}`,
		ExpectedResult: `{
			"createTeam": {
				"id": "1",
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
		CreatorID: actorID,
	}
	if diff := cmp.Diff([]*types.Team{expected}, fakeTeams.list); diff != "" {
		t.Errorf("unexpected teams in fake database (-want,+got):\n%s", diff)
	}
}

func TestCreateTeamDisplayName(t *testing.T) {
	fakeTeams := &fakeTeamsDb{}
	db := database.NewMockDB()
	db.TeamsFunc.SetDefaultReturn(fakeTeams)
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: actor.WithActor(context.Background(), &actor.Actor{UID: actorID}),
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
	fakeTeams := &fakeTeamsDb{}
	db := database.NewMockDB()
	db.TeamsFunc.SetDefaultReturn(fakeTeams)
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: actor.WithActor(context.Background(), &actor.Actor{UID: actorID}),
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
	fakeTeams := &fakeTeamsDb{}
	db := database.NewMockDB()
	db.TeamsFunc.SetDefaultReturn(fakeTeams)
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: actor.WithActor(context.Background(), &actor.Actor{UID: actorID}),
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
	fakeTeams := &fakeTeamsDb{}
	parentTeam := types.Team{Name: "team-name-parent"}
	if err := fakeTeams.CreateTeam(context.Background(), &parentTeam); err != nil {
		t.Fatal(err)
	}
	db := database.NewMockDB()
	db.TeamsFunc.SetDefaultReturn(fakeTeams)
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: actor.WithActor(context.Background(), &actor.Actor{UID: actorID}),
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
			"parentTeamID": strconv.Itoa(int(parentTeam.ID)),
		},
	})
}

func TestCreateTeamParentByName(t *testing.T) {
	fakeTeams := &fakeTeamsDb{}
	parentTeam := types.Team{Name: "team-name-parent"}
	if err := fakeTeams.CreateTeam(context.Background(), &parentTeam); err != nil {
		t.Fatal(err)
	}
	db := database.NewMockDB()
	db.TeamsFunc.SetDefaultReturn(fakeTeams)
	RunTest(t, &Test{
		Schema:  mustParseGraphQLSchema(t, db),
		Context: actor.WithActor(context.Background(), &actor.Actor{UID: actorID}),
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
