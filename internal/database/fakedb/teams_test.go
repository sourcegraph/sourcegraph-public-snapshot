package fakedb_test

import (
	"context"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/fakedb"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/stretchr/testify/require"
)

func TestAncestors(t *testing.T) {
	fs := fakedb.New()
	eng := fs.AddTeam(&types.Team{Name: "eng"})
	source := fs.AddTeam(&types.Team{Name: "source", ParentTeamID: eng})
	_ = fs.AddTeam(&types.Team{Name: "repo-management", ParentTeamID: source})
	sales := fs.AddTeam(&types.Team{Name: "sales"})
	salesLeads := fs.AddTeam(&types.Team{Name: "sales-leads", ParentTeamID: sales})
	ts, cursor, err := fs.TeamStore.ListTeams(context.Background(), database.ListTeamsOpts{ExceptAncestorID: source})
	require.NoError(t, err)
	require.Zero(t, cursor)
	want := []*types.Team{
		{ID: eng, Name: "eng"},
		{ID: sales, Name: "sales"},
		{ID: salesLeads, Name: "sales-leads", ParentTeamID: sales},
	}
	sort.Slice(ts, func(i, j int) bool { return ts[i].ID < ts[j].ID })
	sort.Slice(want, func(i, j int) bool { return want[i].ID < want[j].ID })
	if diff := cmp.Diff(want, ts); diff != "" {
		t.Errorf("ListTeams{ExceptAncestorID} -want+got: %s", diff)
	}
}
