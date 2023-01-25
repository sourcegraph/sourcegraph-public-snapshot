package database

import (
	"context"
	"fmt"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
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

func (opts ListTeamsOpts) SQL() (where, joins []*sqlf.Query) {
	conds := []*sqlf.Query{
		sqlf.Sprintf("id > %s", opts.Cursor),
	}
	joins = []*sqlf.Query{}

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

	return conds, joins
}

type TeamMemberListCursor struct {
	TeamID int32
	UserID int32
}

type ListTeamMembersOpts struct {
	LimitOffset

	// Only return members past this cursor.
	Cursor TeamMemberListCursor
	// Required. Scopes the list operation to the given team.
	TeamID int32
	// Filter members by search term. Currently, name and displayName of the users
	// are searchable.
	Search string
	// Include members of child teams of TeamID. This is mostly useful in count
	// operations, less so in list operations.
	IncludeChildTeamMembers bool
}

func (opts ListTeamMembersOpts) SQL() (where, joins []*sqlf.Query) {
	conds := []*sqlf.Query{
		sqlf.Sprintf("team_id > %s AND user_id > %s", opts.Cursor.TeamID, opts.Cursor.UserID),
	}
	joins = []*sqlf.Query{}

	if opts.TeamID != 0 {
		conds = append(conds, sqlf.Sprintf("team_members.team_id = %s", opts.TeamID))
	}
	if opts.Search != "" {
		joins = append(joins, sqlf.Sprintf("JOIN users ON users.id = team_members.id"))
		// TODO: Proper encoding.
		conds = append(conds, sqlf.Sprintf("(users.username ILIKE %s OR users.display_name ILIKE %s)", opts.Search, opts.Search))
	}
	if opts.IncludeChildTeamMembers {
		// TODO
	}

	return conds, joins
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
	ListTeamMembers(ctx context.Context, opts ListTeamMembersOpts) ([]*types.TeamMember, *TeamMemberListCursor, error)
	// CountTeamMembers counts teams given the options.
	CountTeamMembers(ctx context.Context, opts ListTeamMembersOpts) (int32, error)
	// CreateTeam creates the given team in the database.
	CreateTeam(ctx context.Context, team *types.Team) error
	// UpdateTeam updates the given team in the database.
	UpdateTeam(ctx context.Context, team *types.Team) error
	// DeleteTeam deletes the given team from the database.
	DeleteTeam(ctx context.Context, team int32) error
	// CreateTeamMember creates the team members in the database. If any of the inserts fail,
	// all inserts are reverted.
	CreateTeamMember(ctx context.Context, members ...*types.TeamMember) error
	// DeleteTeam deletes the given team members from the database.
	DeleteTeamMember(ctx context.Context, members ...*types.TeamMember) error
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

func (s *teamStore) ListTeams(ctx context.Context, opts ListTeamsOpts) (_ []*types.Team, next int32, err error) {
	conds, joins := opts.SQL()

	if opts.Limit > 0 {
		opts.Limit++
	}

	q := sqlf.Sprintf(
		listTeamsQueryFmtstr,
		sqlf.Join(teamColumns, ","),
		sqlf.Join(joins, "\n"),
		sqlf.Join(conds, "AND"),
		opts.LimitOffset.SQL(),
	)

	teams, err := scanTeams(s.Query(ctx, q))
	if err != nil {
		return nil, 0, err
	}

	if opts.Limit > 0 && len(teams) == opts.Limit+1 {
		next = teams[len(teams)-1].ID
		teams = teams[:len(teams)-1]
	}

	return teams, next, nil
}

const listTeamsQueryFmtstr = `
SELECT %s
FROM teams
%s
WHERE %s
%s
ORDER BY
	teams.id ASC
`

func (s *teamStore) CountTeams(ctx context.Context, opts ListTeamsOpts) (int32, error) {
	// Disable cursor for counting.
	opts.Cursor = 0
	conds, joins := opts.SQL()

	q := sqlf.Sprintf(
		countTeamsQueryFmtstr,
		sqlf.Join(joins, "\n"),
		sqlf.Join(conds, "AND"),
		opts.LimitOffset.SQL(),
	)

	count, _, err := basestore.ScanFirstInt(s.Query(ctx, q))
	return int32(count), err
}

const countTeamsQueryFmtstr = `
SELECT COUNT(*)
FROM teams
%s
WHERE %s
`

func (s *teamStore) ListTeamMembers(ctx context.Context, opts ListTeamMembersOpts) (_ []*types.TeamMember, next *TeamMemberListCursor, err error) {
	conds, joins := opts.SQL()

	if opts.Limit > 0 {
		opts.Limit++
	}

	q := sqlf.Sprintf(
		listTeamMembersQueryFmtstr,
		sqlf.Join(teamColumns, ","),
		sqlf.Join(joins, "\n"),
		sqlf.Join(conds, "AND"),
		opts.LimitOffset.SQL(),
	)

	tms, err := scanTeamMembers(s.Query(ctx, q))
	if err != nil {
		return nil, nil, err
	}

	if opts.Limit > 0 && len(tms) == opts.Limit+1 {
		next = &TeamMemberListCursor{
			TeamID: tms[len(tms)-1].TeamID,
			UserID: tms[len(tms)-1].UserID,
		}
		tms = tms[:len(tms)-1]
	}

	return tms, next, nil
}

const listTeamMembersQueryFmtstr = `
SELECT %s
FROM team_members
%s
WHERE %s
%s
ORDER BY
	team_members.team_id ASC, team_members.user_id ASC
`

func (s *teamStore) CountTeamMembers(ctx context.Context, opts ListTeamMembersOpts) (int32, error) {
	// Disable cursor for counting.
	opts.Cursor = TeamMemberListCursor{}
	conds, joins := opts.SQL()

	q := sqlf.Sprintf(
		countTeamMembersQueryFmtstr,
		sqlf.Join(joins, "\n"),
		sqlf.Join(conds, "AND"),
		opts.LimitOffset.SQL(),
	)

	count, _, err := basestore.ScanFirstInt(s.Query(ctx, q))
	return int32(count), err
}

const countTeamMembersQueryFmtstr = `
SELECT COUNT(*)
FROM team_members
%s
WHERE %s
`

func (s *teamStore) CreateTeam(ctx context.Context, team *types.Team) error {
	if team.CreatedAt.IsZero() {
		team.CreatedAt = timeutil.Now()
	}

	if team.UpdatedAt.IsZero() {
		team.UpdatedAt = team.CreatedAt
	}

	q := sqlf.Sprintf(
		createTeamQueryFmtstr,
		sqlf.Join(teamInsertColumns, ","),
		team.Name,
		&dbutil.NullString{S: team.DisplayName},
		team.ReadOnly,
		&dbutil.NullInt32{N: &team.ParentTeamID},
		team.CreatedAt,
		team.UpdatedAt,
		sqlf.Join(teamColumns, ","),
	)

	row := s.QueryRow(ctx, q)
	return scanTeam(row, team)
}

const createTeamQueryFmtstr = `
INSERT INTO teams
(%s)
VALUES (%s, %s, %s, %s, %s, %s)
RETURNING %s
`

func (s *teamStore) UpdateTeam(ctx context.Context, team *types.Team) error {
	team.UpdatedAt = timeutil.Now()

	conds := []*sqlf.Query{
		sqlf.Sprintf("id = %s", team.ID),
	}

	q := sqlf.Sprintf(
		updateTeamQueryFmtstr,
		team.Name,
		&dbutil.NullString{S: team.DisplayName},
		&dbutil.NullInt32{N: &team.ParentTeamID},
		team.UpdatedAt,
		sqlf.Join(conds, "AND"),
		sqlf.Join(teamColumns, ","),
		sqlf.Join(conds, "AND"),
	)

	row := s.QueryRow(ctx, q)
	return scanTeam(row, team)
}

const updateTeamQueryFmtstr = `
WITH update AS (
	UPDATE
		teams
	SET
		name = %s,
		display_name = %s,
		parent_team_id = %s,
		updated_at = %s
	WHERE
		%s
)
SELECT
	%s
FROM
	teams
WHERE
	%s
`

func (s *teamStore) DeleteTeam(ctx context.Context, team int32) error {
	conds := []*sqlf.Query{
		sqlf.Sprintf("teams.id =%s", team),
	}

	q := sqlf.Sprintf(deleteTeamQueryFmtstr, sqlf.Join(conds, "AND"))
	return s.Exec(ctx, q)
}

const deleteTeamQueryFmtstr = `
DELETE FROM
	teams
WHERE %s
`

func (s *teamStore) CreateTeamMember(ctx context.Context, members ...*types.TeamMember) error {
	inserter := func(inserter *batch.Inserter) error {
		for _, m := range members {
			if m.CreatedAt.IsZero() {
				m.CreatedAt = timeutil.Now()
			}

			if m.UpdatedAt.IsZero() {
				m.UpdatedAt = m.CreatedAt
			}

			if err := inserter.Insert(
				ctx,
				m.TeamID,
				m.UserID,
				m.CreatedAt,
				m.UpdatedAt,
			); err != nil {
				return err
			}
		}
		return nil
	}

	i := -1
	return batch.WithInserterWithReturn(
		ctx,
		s.Handle(),
		"team_members",
		batch.MaxNumPostgresParameters,
		teamMemberInsertColumns,
		"",
		teamMemberStringColumns,
		func(sc dbutil.Scanner) error {
			i++
			return scanTeamMember(sc, members[i])
		},
		inserter,
	)
}

func (s *teamStore) DeleteTeamMember(ctx context.Context, members ...*types.TeamMember) error {
	ms := [][2]int32{}
	for _, m := range members {
		ms = append(ms, [2]int32{m.TeamID, m.UserID})
	}
	conds := []*sqlf.Query{
		sqlf.Sprintf("(team_id, user_id) = ANY(%s)", pq.Array(ms)),
	}

	q := sqlf.Sprintf(deleteTeamMemberQueryFmtstr, sqlf.Join(conds, "AND"))
	return s.Exec(ctx, q)
}

const deleteTeamMemberQueryFmtstr = `
DELETE FROM
	team_members
WHERE %s
`

var teamColumns = []*sqlf.Query{
	sqlf.Sprintf("teams.id"),
	sqlf.Sprintf("teams.name"),
	sqlf.Sprintf("teams.display_name"),
	sqlf.Sprintf("teams.read_only"),
	sqlf.Sprintf("teams.parent_team_id"),
	sqlf.Sprintf("teams.created_at"),
	sqlf.Sprintf("teams.updated_at"),
}

var teamInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("name"),
	sqlf.Sprintf("display_name"),
	sqlf.Sprintf("read_only"),
	sqlf.Sprintf("parent_team_id"),
	sqlf.Sprintf("created_at"),
	sqlf.Sprintf("updated_at"),
}

var teamMemberColumns = []*sqlf.Query{
	sqlf.Sprintf("team_members.team_id"),
	sqlf.Sprintf("team_members.user_id"),
	sqlf.Sprintf("team_members.created_at"),
	sqlf.Sprintf("team_members.updated_at"),
}

var teamMemberStringColumns = []string{
	"team_members.team_id",
	"team_members.user_id",
	"team_members.created_at",
	"team_members.updated_at",
}

var teamMemberInsertColumns = []string{
	"team_id",
	"user_id",
	"created_at",
	"updated_at",
}

var scanTeams = basestore.NewSliceScanner(func(s dbutil.Scanner) (t *types.Team, _ error) {
	err := scanTeam(s, t)
	return t, err
})

func scanTeam(sc dbutil.Scanner, t *types.Team) error {
	return sc.Scan(
		&t.ID,
		&t.Name,
		&dbutil.NullString{S: t.DisplayName},
		&t.ReadOnly,
		&dbutil.NullInt32{N: &t.ParentTeamID},
		&t.CreatedAt,
		&t.UpdatedAt,
	)
}

var scanTeamMembers = basestore.NewSliceScanner(func(s dbutil.Scanner) (t *types.TeamMember, _ error) {
	err := scanTeamMember(s, t)
	return t, err
})

func scanTeamMember(sc dbutil.Scanner, tm *types.TeamMember) error {
	return sc.Scan(
		&tm.TeamID,
		&tm.UserID,
		&tm.CreatedAt,
		&tm.UpdatedAt,
	)
}
