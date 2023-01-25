package database

import (
	"context"
	"fmt"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type ListTeamsOpts struct {
	LimitOffset

	// Only return teams past this cursor.
	Cursor int32
	// List teams of a specific parent team only.
	WithParentID int32
	// Only return root teams (teams that have no parent).
	// This is used on the main overview list of teams.
	RootOnly bool
	// Filter teams by search term. Currently, name and displayName are searchable.
	Search string
	// Include child teams of teams. This is mostly useful in count operations, less
	// so in list operations.
	IncludeChildTeams bool
	// List teams that a specific user is a member of.
	ForUserMember int32
}

type ListTeamMembersOpts struct {
	LimitOffset

	// Only return members past this cursor.
	Cursor int32
	// Required. Scopes the list operation to the given team.
	TeamID int32
	// Filter members by search term. Currently, name and displayName of the users
	// are searchable.
	Search string
	// Include members of child teams of TeamID. This is mostly useful in count
	// operations, less so in list operations.
	IncludeChildTeamMembers bool
}

// TeamNotFoundError is returned when a team cannot be found.
type TeamNotFoundError struct {
	id int32
}

func (err TeamNotFoundError) Error() string {
	return fmt.Sprintf("team not found: id=%d", err.id)
}

func (TeamNotFoundError) NotFound() bool {
	return true
}

// TeamStore provides database methods for interacting with teams and their members.
type TeamStore interface {
	basestore.ShareableStore
	// GetTeam returns the given team by ID. If not found, a NotFounder error is returned.
	GetTeam(ctx context.Context, id int32) (*types.Team, error)
	// ListTeams lists teams given the options. The matching teams, plus the next cursor are
	// returned.
	ListTeams(ctx context.Context, opts ListTeamsOpts) ([]*types.Team, int32, error)
	// CountTeams counts teams given the options.
	CountTeams(ctx context.Context, opts ListTeamsOpts) (int32, error)
	// ListTeamMembers lists team members given the options. The matching teams,
	// plus the next cursor are returned.
	ListTeamMembers(ctx context.Context, opts ListTeamMembersOpts) ([]*types.TeamMember, int32, error)
	// CountTeamMembers counts teams given the options.
	CountTeamMembers(ctx context.Context, opts ListTeamMembersOpts) (int32, error)
	// CreateTeam creates the given teams in the database. If any of the inserts fail,
	// all inserts are reverted.
	CreateTeam(ctx context.Context, teams ...[]*types.Team) error
	// UpdateTeam updates the given teams in the database. If any of the updates fail,
	// all updates are reverted.
	UpdateTeam(ctx context.Context, teams ...[]*types.Team) error
	// DeleteTeam deletes the given teams from the database.
	DeleteTeam(ctx context.Context, teams ...[]int32) error
	// CreateTeamMember creates the team members in the database. If any of the inserts fail,
	// all inserts are reverted.
	CreateTeamMember(ctx context.Context, member ...[]*types.TeamMember) error
	// DeleteTeam deletes the given team members from the database.
	DeleteTeamMember(ctx context.Context, member ...[]*types.TeamMember) error
}

type teamStore struct {
	*basestore.Store
}

// TeamsWith instantiates and returns a new TeamStore using the other store handle.
func TeamsWith(other basestore.ShareableStore) TeamStore {
	return &teamStore{
		Store: basestore.NewWithHandle(other.Handle()),
	}
}

func (s *teamStore) With(other basestore.ShareableStore) TeamStore {
	return &teamStore{
		Store: s.Store.With(other),
	}
}

func (s *teamStore) Transact(ctx context.Context) (TeamStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &teamStore{
		Store: txBase,
	}, err
}

func (s *teamStore) GetTeam(ctx context.Context, id int32) (*types.Team, error) {
	conds := []*sqlf.Query{
		sqlf.Sprintf("id = %s", id),
	}
	q := sqlf.Sprintf(getTeamQueryFmtstr, sqlf.Join(teamColumns, ","), sqlf.Join(conds, "AND"))

	teams, err := scanTeams(s.Query(ctx, q))
	if err != nil {
		return nil, err
	}

	if len(teams) != 1 {
		return nil, TeamNotFoundError{id: id}
	}

	return teams[0], nil
}

const getTeamQueryFmtstr = `
SELECT %s
FROM teams
%s
`

func (s *teamStore) ListTeams(ctx context.Context, opts ListTeamsOpts) ([]*types.Team, int32, error) {
	conds := []*sqlf.Query{
		sqlf.Sprintf("id > %s", opts.Cursor),
	}
	joins := []*sqlf.Query{}

	if opts.WithParentID != 0 {
		conds = append(conds, sqlf.Sprintf("parent_team_id = %s", opts.WithParentID))
	}
	if opts.RootOnly {
		conds = append(conds, sqlf.Sprintf("parent_team_id IS NULL"))
	}
	if opts.Search != "" {
		// TODO: Proper encoding.
		conds = append(conds, sqlf.Sprintf("(team.name ILIKE %s OR team.display_name ILIKE %s)", opts.Search, opts.Search))
	}
	if opts.IncludeChildTeams {
		// TODO
	}
	if opts.ForUserMember != 0 {
		joins = append(joins, sqlf.Sprintf("JOIN team_members ON team_members.team_id = teams.id"))
		conds = append(conds, sqlf.Sprintf("team_members.user_id = %s", opts.ForUserMember))
	}

	q := sqlf.Sprintf(
		getTeamQueryFmtstr,
		sqlf.Join(teamColumns, ","),
		sqlf.Join(joins, "\n"),
		sqlf.Join(conds, "AND"),
		opts.LimitOffset.SQL(),
	)

	teams, err := scanTeams(s.Query(ctx, q))
	if err != nil {
		return nil, 0, err
	}

	// TODO
	var next int32

	return teams, next, nil
}

const listTeamsQueryFmtstr = `
SELECT %s
FROM teams
%s
%s
%s
`

func (s *teamStore) CountTeams(ctx context.Context, opts ListTeamsOpts) (int32, error) {
	return 0, nil
}

func (s *teamStore) ListTeamMembers(ctx context.Context, opts ListTeamMembersOpts) ([]*types.TeamMember, int32, error) {
	return nil, 0, nil
}

func (s *teamStore) CountTeamMembers(ctx context.Context, opts ListTeamMembersOpts) (int32, error) {
	return 0, nil
}

func (s *teamStore) CreateTeam(ctx context.Context, teams ...[]*types.Team) error {
	return nil
}

func (s *teamStore) UpdateTeam(ctx context.Context, teams ...[]*types.Team) error {
	return nil
}

func (s *teamStore) DeleteTeam(ctx context.Context, teams ...[]int32) error {
	return nil
}

func (s *teamStore) CreateTeamMember(ctx context.Context, member ...[]*types.TeamMember) error {
	return nil
}

func (s *teamStore) DeleteTeamMember(ctx context.Context, member ...[]*types.TeamMember) error {
	return nil
}

var teamColumns = []*sqlf.Query{
	sqlf.Sprintf("teams.id"),
	sqlf.Sprintf("teams.name"),
	sqlf.Sprintf("teams.display_name"),
	sqlf.Sprintf("teams.read_only"),
	sqlf.Sprintf("teams.parent_team_id"),
	sqlf.Sprintf("teams.created_at"),
	sqlf.Sprintf("teams.updated_at"),
}

var scanTeams = basestore.NewSliceScanner(func(s dbutil.Scanner) (t *types.Team, _ error) {
	err := s.Scan(
		&t.ID,
		&t.Name,
		&dbutil.NullString{S: t.DisplayName},
		&t.ReadOnly,
		&dbutil.NullInt32{N: &t.ParentTeamID},
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	return t, err
})
