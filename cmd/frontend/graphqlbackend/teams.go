package graphqlbackend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/own"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func teamByID(ctx context.Context, db database.DB, id graphql.ID) (Node, error) {
	if err := areTeamEndpointsAvailable(); err != nil {
		return nil, err
	}

	team, err := findTeam(ctx, db.Teams(), &id, nil)
	if err != nil {
		return nil, err
	}
	return NewTeamResolver(db, team), nil
}

type ListTeamsArgs struct {
	First  *int32
	After  *string
	Search *string
}

type teamConnectionResolver struct {
	db               database.DB
	parentID         int32
	search           string
	cursor           int32
	limit            int
	once             sync.Once
	teams            []*types.Team
	onlyRootTeams    bool
	exceptAncestorID int32
	pageInfo         *gqlutil.PageInfo
	err              error
}

// applyArgs unmarshals query conditions and limites set in `ListTeamsArgs`
// into `teamConnectionResolver` fields for convenient use in database query.
func (r *teamConnectionResolver) applyArgs(args *ListTeamsArgs) error {
	if args.After != nil {
		cursor, err := gqlutil.DecodeIntCursor(args.After)
		if err != nil {
			return err
		}
		r.cursor = int32(cursor)
		if int(r.cursor) != cursor {
			return errors.Newf("cursor int32 overflow: %d", cursor)
		}
	}
	if args.Search != nil {
		r.search = *args.Search
	}
	if args.First != nil {
		r.limit = int(*args.First)
	}
	return nil
}

// compute resolves teams queried for this resolver.
// The result of running it is setting `teams`, `next` and `err`
// fields on the resolver. This ensures that resolving multiple
// graphQL attributes that require listing (like `pageInfo` and `nodes`)
// results in just one query.
func (r *teamConnectionResolver) compute(ctx context.Context) {
	r.once.Do(func() {
		opts := database.ListTeamsOpts{
			Cursor:           r.cursor,
			WithParentID:     r.parentID,
			Search:           r.search,
			RootOnly:         r.onlyRootTeams,
			ExceptAncestorID: r.exceptAncestorID,
		}
		if r.limit != 0 {
			opts.LimitOffset = &database.LimitOffset{Limit: r.limit}
		}
		teams, next, err := r.db.Teams().ListTeams(ctx, opts)
		if err != nil {
			r.err = err
			return
		}
		r.teams = teams
		if next > 0 {
			r.pageInfo = gqlutil.EncodeIntCursor(&next)
		} else {
			r.pageInfo = gqlutil.HasNextPage(false)
		}
	})
}

func (r *teamConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	// Not taking into account limit or cursor for count.
	opts := database.ListTeamsOpts{
		WithParentID: r.parentID,
		Search:       r.search,
		RootOnly:     r.onlyRootTeams,
	}
	return r.db.Teams().CountTeams(ctx, opts)
}

func (r *teamConnectionResolver) PageInfo(ctx context.Context) (*gqlutil.PageInfo, error) {
	r.compute(ctx)
	return r.pageInfo, r.err
}

func (r *teamConnectionResolver) Nodes(ctx context.Context) ([]*TeamResolver, error) {
	r.compute(ctx)
	if r.err != nil {
		return nil, r.err
	}
	var rs []*TeamResolver
	for _, t := range r.teams {
		rs = append(rs, NewTeamResolver(r.db, t))
	}
	return rs, nil
}

func NewTeamResolver(db database.DB, team *types.Team) *TeamResolver {
	return &TeamResolver{
		db:   db,
		team: team,
	}
}

type TeamResolver struct {
	db   database.DB
	team *types.Team
}

const teamIDKind = "Team"

func MarshalTeamID(id int32) graphql.ID {
	return relay.MarshalID("Team", id)
}

func UnmarshalTeamID(id graphql.ID) (teamID int32, err error) {
	err = relay.UnmarshalSpec(id, &teamID)
	return
}

func (r *TeamResolver) ID() graphql.ID {
	return relay.MarshalID("Team", r.team.ID)
}

func (r *TeamResolver) Name() string {
	return r.team.Name
}

func (r *TeamResolver) URL() string {
	if r.External() {
		return ""
	}
	absolutePath := fmt.Sprintf("/teams/%s", r.team.Name)
	u := &url.URL{Path: absolutePath}
	return u.String()
}

func (r *TeamResolver) AvatarURL() *string {
	return nil
}

func (r *TeamResolver) Creator(ctx context.Context) (*UserResolver, error) {
	if r.team.CreatorID == 0 {
		// User was deleted.
		return nil, nil
	}
	return UserByIDInt32(ctx, r.db, r.team.CreatorID)
}

func (r *TeamResolver) DisplayName() *string {
	if r.team.DisplayName == "" {
		return nil
	}
	return &r.team.DisplayName
}

func (r *TeamResolver) Readonly() bool {
	return r.team.ReadOnly || r.External()
}

func (r *TeamResolver) ParentTeam(ctx context.Context) (*TeamResolver, error) {
	if r.team.ParentTeamID == 0 {
		return nil, nil
	}
	parentTeam, err := r.db.Teams().GetTeamByID(ctx, r.team.ParentTeamID)
	if err != nil {
		return nil, err
	}
	return NewTeamResolver(r.db, parentTeam), nil
}

func (r *TeamResolver) ViewerCanAdminister(ctx context.Context) (bool, error) {
	return canModifyTeam(ctx, r.db, r.team)
}

func (r *TeamResolver) Members(_ context.Context, args *ListTeamMembersArgs) (*teamMemberConnection, error) {
	if r.External() {
		return nil, errors.New("cannot get members of external team")
	}
	c := &teamMemberConnection{
		db:     r.db,
		teamID: r.team.ID,
	}
	if err := c.applyArgs(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (r *TeamResolver) ChildTeams(_ context.Context, args *ListTeamsArgs) (*teamConnectionResolver, error) {
	if r.External() {
		return nil, errors.New("cannot get child teams of external team")
	}
	c := &teamConnectionResolver{
		db:       r.db,
		parentID: r.team.ID,
	}
	if err := c.applyArgs(args); err != nil {
		return nil, err
	}
	return c, nil
}

func (r *TeamResolver) OwnerField() string {
	return EnterpriseResolvers.ownResolver.TeamOwnerField(r)
}

func (r *TeamResolver) External() bool {
	return r.team.ID == 0
}

type ListTeamMembersArgs struct {
	First  *int32
	After  *string
	Search *string
}

type teamMemberConnection struct {
	db       database.DB
	teamID   int32
	cursor   teamMemberListCursor
	search   string
	limit    int
	once     sync.Once
	nodes    []*types.TeamMember
	pageInfo *gqlutil.PageInfo
	err      error
}

type teamMemberListCursor struct {
	TeamID int32 `json:"team,omitempty"`
	UserID int32 `json:"user,omitempty"`
}

// applyArgs unmarshals query conditions and limites set in `ListTeamMembersArgs`
// into `teamMemberConnection` fields for convenient use in database query.
func (r *teamMemberConnection) applyArgs(args *ListTeamMembersArgs) error {
	if args.After != nil && *args.After != "" {
		cursorText, err := gqlutil.DecodeCursor(args.After)
		if err != nil {
			return err
		}
		if err := json.Unmarshal([]byte(cursorText), &r.cursor); err != nil {
			return err
		}
	}
	if args.Search != nil {
		r.search = *args.Search
	}
	if args.First != nil {
		r.limit = int(*args.First)
	}
	return nil
}

// compute resolves team members queried for this resolver.
// The result of running it is setting `nodes`, `pageInfo` and `err`
// fields on the resolver. This ensures that resolving multiple
// graphQL attributes that require listing (like `pageInfo` and `nodes`)
// results in just one query.
func (r *teamMemberConnection) compute(ctx context.Context) {
	r.once.Do(func() {
		opts := database.ListTeamMembersOpts{
			Cursor: database.TeamMemberListCursor{
				TeamID: r.cursor.TeamID,
				UserID: r.cursor.UserID,
			},
			TeamID: r.teamID,
			Search: r.search,
		}
		if r.limit != 0 {
			opts.LimitOffset = &database.LimitOffset{Limit: r.limit}
		}
		nodes, next, err := r.db.Teams().ListTeamMembers(ctx, opts)
		if err != nil {
			r.err = err
			return
		}
		r.nodes = nodes
		if next != nil {
			cursorStruct := teamMemberListCursor{
				TeamID: next.TeamID,
				UserID: next.UserID,
			}
			cursorBytes, err := json.Marshal(&cursorStruct)
			if err != nil {
				r.err = errors.Wrap(err, "error encoding pageInfo")
			}
			cursorString := string(cursorBytes)
			r.pageInfo = gqlutil.EncodeCursor(&cursorString)
		} else {
			r.pageInfo = gqlutil.HasNextPage(false)
		}
	})
}

func (r *teamMemberConnection) TotalCount(ctx context.Context) (int32, error) {
	// Not taking into account limit or cursor for count.
	opts := database.ListTeamMembersOpts{
		TeamID: r.teamID,
		Search: r.search,
	}
	return r.db.Teams().CountTeamMembers(ctx, opts)
}

func (r *teamMemberConnection) PageInfo(ctx context.Context) (*gqlutil.PageInfo, error) {
	r.compute(ctx)
	if r.err != nil {
		return nil, r.err
	}
	return r.pageInfo, nil
}

func (r *teamMemberConnection) Nodes(ctx context.Context) ([]*UserResolver, error) {
	r.compute(ctx)
	if r.err != nil {
		return nil, r.err
	}
	var rs []*UserResolver
	// 🚨 Query in a loop is inefficient: Follow up with another pull request
	// to where team members query joins with users and fetches them in one go.
	for _, n := range r.nodes {
		if n.UserID == 0 {
			// 🚨 At this point only User can be a team member, so user ID should
			// always be present. If not, return a `null` team member.
			rs = append(rs, nil)
			continue
		}
		user, err := r.db.Users().GetByID(ctx, n.UserID)
		if err != nil {
			return nil, err
		}
		rs = append(rs, NewUserResolver(ctx, r.db, user))
	}
	return rs, nil
}

type CreateTeamArgs struct {
	Name           string
	DisplayName    *string
	ReadOnly       bool
	ParentTeam     *graphql.ID
	ParentTeamName *string
}

func (r *schemaResolver) CreateTeam(ctx context.Context, args *CreateTeamArgs) (*TeamResolver, error) {
	if err := areTeamEndpointsAvailable(); err != nil {
		return nil, err
	}

	if args.ReadOnly {
		if !isSiteAdmin(ctx, r.db) {
			return nil, errors.New("only site admins can create read-only teams")
		}
	}

	teams := r.db.Teams()
	var t types.Team
	t.Name = args.Name
	if args.DisplayName != nil {
		t.DisplayName = *args.DisplayName
	}
	t.ReadOnly = args.ReadOnly
	if args.ParentTeam != nil && args.ParentTeamName != nil {
		return nil, errors.New("must specify at most one: ParentTeam or ParentTeamName")
	}
	parentTeam, err := findTeam(ctx, teams, args.ParentTeam, args.ParentTeamName)
	if err != nil {
		return nil, errors.Wrap(err, "parent team")
	}
	if parentTeam != nil {
		t.ParentTeamID = parentTeam.ID
		if ok, err := canModifyTeam(ctx, r.db, parentTeam); err != nil {
			return nil, err
		} else if !ok {
			return nil, ErrNoAccessToTeam
		}
	}
	t.CreatorID = actor.FromContext(ctx).UID
	team, err := teams.CreateTeam(ctx, &t)
	if err != nil {
		return nil, err
	}
	return NewTeamResolver(r.db, team), nil
}

type UpdateTeamArgs struct {
	ID             *graphql.ID
	Name           *string
	DisplayName    *string
	ParentTeam     *graphql.ID
	ParentTeamName *string
	MakeRoot       *bool
}

func (r *schemaResolver) UpdateTeam(ctx context.Context, args *UpdateTeamArgs) (*TeamResolver, error) {
	if err := areTeamEndpointsAvailable(); err != nil {
		return nil, err
	}

	if args.ID == nil && args.Name == nil {
		return nil, errors.New("team to update is identified by either id or name, but neither was specified")
	}
	if args.ID != nil && args.Name != nil {
		return nil, errors.New("team to update is identified by either id or name, but both were specified")
	}
	if (args.ParentTeam != nil || args.ParentTeamName != nil) && args.MakeRoot != nil {
		return nil, errors.New("specifying a parent team contradicts making a team root (no parent team)")
	}
	if args.ParentTeam != nil && args.ParentTeamName != nil {
		return nil, errors.New("parent team is identified by either id or name, but both were specified")
	}
	if args.MakeRoot != nil && !*args.MakeRoot {
		return nil, errors.New("the only possible value for makeRoot is true (if set); to assign a parent team please use parentTeam or parentTeamName parameters")
	}
	var t *types.Team
	err := r.db.WithTransact(ctx, func(tx database.DB) (err error) {
		t, err = findTeam(ctx, tx.Teams(), args.ID, args.Name)
		if err != nil {
			return err
		}

		if ok, err := canModifyTeam(ctx, r.db, t); err != nil {
			return err
		} else if !ok {
			return ErrNoAccessToTeam
		}

		var needsUpdate bool
		if args.DisplayName != nil && *args.DisplayName != t.DisplayName {
			needsUpdate = true
			t.DisplayName = *args.DisplayName
		}
		if args.ParentTeam != nil || args.ParentTeamName != nil {
			parentTeam, err := findTeam(ctx, tx.Teams(), args.ParentTeam, args.ParentTeamName)
			if err != nil {
				return errors.Wrap(err, "cannot find parent team")
			}
			if parentTeam.ID != t.ParentTeamID {
				parentOutsideOfTeamsDescendants, err := tx.Teams().ContainsTeam(ctx, parentTeam.ID, database.ListTeamsOpts{
					ExceptAncestorID: t.ID,
				})
				if err != nil {
					return errors.Newf("could not determine ancestorship on team update: %s", err)
				}
				if !parentOutsideOfTeamsDescendants {
					return errors.Newf("circular dependency: new parent %q is descendant of updated team %q", parentTeam.Name, t.Name)
				}
				needsUpdate = true
				t.ParentTeamID = parentTeam.ID
			}
		}
		if args.MakeRoot != nil && *args.MakeRoot && t.ParentTeamID != 0 {
			needsUpdate = true
			t.ParentTeamID = 0
		}
		if needsUpdate {
			return tx.Teams().UpdateTeam(ctx, t)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return NewTeamResolver(r.db, t), nil
}

// findTeam returns a team by either GraphQL ID or name.
// If both parameters are nil, the result is nil.
func findTeam(ctx context.Context, teams database.TeamStore, graphqlID *graphql.ID, name *string) (*types.Team, error) {
	if graphqlID != nil {
		var id int32
		id, err := UnmarshalTeamID(*graphqlID)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot interpret team id: %q", *graphqlID)
		}
		team, err := teams.GetTeamByID(ctx, id)
		if errcode.IsNotFound(err) {
			return nil, errors.Wrapf(err, "team id=%d not found", id)
		}
		if err != nil {
			return nil, errors.Wrapf(err, "error fetching team id=%d", id)
		}
		return team, nil
	}
	if name != nil {
		team, err := teams.GetTeamByName(ctx, *name)
		if errcode.IsNotFound(err) {
			return nil, errors.Wrapf(err, "team name=%q not found", *name)
		}
		if err != nil {
			return nil, errors.Wrapf(err, "could not fetch team name=%q", *name)
		}
		return team, nil
	}
	return nil, nil
}

type DeleteTeamArgs struct {
	ID   *graphql.ID
	Name *string
}

func (r *schemaResolver) DeleteTeam(ctx context.Context, args *DeleteTeamArgs) (*EmptyResponse, error) {
	if err := areTeamEndpointsAvailable(); err != nil {
		return nil, err
	}

	if args.ID == nil && args.Name == nil {
		return nil, errors.New("team to delete is identified by either id or name, but neither was specified")
	}
	if args.ID != nil && args.Name != nil {
		return nil, errors.New("team to delete is identified by either id or name, but both were specified")
	}
	t, err := findTeam(ctx, r.db.Teams(), args.ID, args.Name)
	if err != nil {
		return nil, err
	}

	if ok, err := canModifyTeam(ctx, r.db, t); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrNoAccessToTeam
	}

	if err := r.db.Teams().DeleteTeam(ctx, t.ID); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}

type TeamMembersArgs struct {
	Team                 *graphql.ID
	TeamName             *string
	Members              []TeamMemberInput
	SkipUnmatchedMembers bool
}

type TeamMemberInput struct {
	UserID                     *graphql.ID
	Username                   *string
	Email                      *string
	ExternalAccountServiceID   *string
	ExternalAccountServiceType *string
	ExternalAccountAccountID   *string
	ExternalAccountLogin       *string
}

func (t TeamMemberInput) String() string {
	conds := []string{}

	if t.UserID != nil {
		conds = append(conds, fmt.Sprintf("ID=%s", string(*t.UserID)))
	}
	if t.Username != nil {
		conds = append(conds, fmt.Sprintf("Username=%s", *t.Username))
	}
	if t.Email != nil {
		conds = append(conds, fmt.Sprintf("Email=%s", *t.Email))
	}
	if t.ExternalAccountServiceID != nil {
		maybeString := func(s *string) string {
			if s == nil {
				return ""
			}
			return *s
		}
		conds = append(conds, fmt.Sprintf(
			"ExternalAccount(ServiceID=%s, ServiceType=%s, AccountID=%s, Login=%s)",
			maybeString(t.ExternalAccountServiceID),
			maybeString(t.ExternalAccountServiceType),
			maybeString(t.ExternalAccountAccountID),
			maybeString(t.ExternalAccountLogin),
		))
	}

	return fmt.Sprintf("team member (%s)", strings.Join(conds, ","))
}

func (r *schemaResolver) AddTeamMembers(ctx context.Context, args *TeamMembersArgs) (*TeamResolver, error) {
	if err := areTeamEndpointsAvailable(); err != nil {
		return nil, err
	}

	if args.Team == nil && args.TeamName == nil {
		return nil, errors.New("team must be identified by either id (team parameter) or name (teamName parameter), none specified")
	}
	if args.Team != nil && args.TeamName != nil {
		return nil, errors.New("team must be identified by either id (team parameter) or name (teamName parameter), both specified")
	}

	users, notFound, err := usersForTeamMembers(ctx, r.db, args.Members)
	if err != nil {
		return nil, err
	}
	if len(notFound) > 0 && !args.SkipUnmatchedMembers {
		var err error
		for _, member := range notFound {
			err = errors.Append(err, errors.Newf("member not found: %s", member))
		}
		return nil, err
	}
	usersMap := make(map[int32]*types.User, len(users))
	for _, user := range users {
		usersMap[user.ID] = user
	}

	team, err := findTeam(ctx, r.db.Teams(), args.Team, args.TeamName)
	if err != nil {
		return nil, err
	}

	if ok, err := canModifyTeam(ctx, r.db, team); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrNoAccessToTeam
	}

	ms := make([]*types.TeamMember, 0, len(users))
	for _, u := range users {
		ms = append(ms, &types.TeamMember{
			UserID: u.ID,
			TeamID: team.ID,
		})
	}
	if err := r.db.Teams().CreateTeamMember(ctx, ms...); err != nil {
		return nil, err
	}

	return NewTeamResolver(r.db, team), nil
}

func (r *schemaResolver) SetTeamMembers(ctx context.Context, args *TeamMembersArgs) (*TeamResolver, error) {
	if err := areTeamEndpointsAvailable(); err != nil {
		return nil, err
	}

	if args.Team == nil && args.TeamName == nil {
		return nil, errors.New("team must be identified by either id (team parameter) or name (teamName parameter), none specified")
	}
	if args.Team != nil && args.TeamName != nil {
		return nil, errors.New("team must be identified by either id (team parameter) or name (teamName parameter), both specified")
	}

	users, notFound, err := usersForTeamMembers(ctx, r.db, args.Members)
	if err != nil {
		return nil, err
	}
	if len(notFound) > 0 && !args.SkipUnmatchedMembers {
		var err error
		for _, member := range notFound {
			err = errors.Append(err, errors.Newf("member not found: %s", member))
		}
		return nil, err
	}
	usersMap := make(map[int32]*types.User, len(users))
	for _, user := range users {
		usersMap[user.ID] = user
	}

	team, err := findTeam(ctx, r.db.Teams(), args.Team, args.TeamName)
	if err != nil {
		return nil, err
	}

	if ok, err := canModifyTeam(ctx, r.db, team); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrNoAccessToTeam
	}

	if err := r.db.WithTransact(ctx, func(tx database.DB) error {
		var membersToRemove []*types.TeamMember
		listOpts := database.ListTeamMembersOpts{
			TeamID: team.ID,
		}
		for {
			existingMembers, cursor, err := tx.Teams().ListTeamMembers(ctx, listOpts)
			if err != nil {
				return err
			}
			for _, m := range existingMembers {
				if _, ok := usersMap[m.UserID]; ok {
					delete(usersMap, m.UserID)
				} else {
					membersToRemove = append(membersToRemove, &types.TeamMember{
						UserID: m.UserID,
						TeamID: team.ID,
					})
				}
			}
			if cursor == nil {
				break
			}
			listOpts.Cursor = *cursor
		}
		var membersToAdd []*types.TeamMember
		for _, user := range users {
			membersToAdd = append(membersToAdd, &types.TeamMember{
				UserID: user.ID,
				TeamID: team.ID,
			})
		}
		if len(membersToRemove) > 0 {
			if err := tx.Teams().DeleteTeamMember(ctx, membersToRemove...); err != nil {
				return err
			}
		}
		if len(membersToAdd) > 0 {
			if err := tx.Teams().CreateTeamMember(ctx, membersToAdd...); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return NewTeamResolver(r.db, team), nil
}

func (r *schemaResolver) RemoveTeamMembers(ctx context.Context, args *TeamMembersArgs) (*TeamResolver, error) {
	if err := areTeamEndpointsAvailable(); err != nil {
		return nil, err
	}

	if args.Team == nil && args.TeamName == nil {
		return nil, errors.New("team must be identified by either id (team parameter) or name (teamName parameter), none specified")
	}
	if args.Team != nil && args.TeamName != nil {
		return nil, errors.New("team must be identified by either id (team parameter) or name (teamName parameter), both specified")
	}

	users, notFound, err := usersForTeamMembers(ctx, r.db, args.Members)
	if err != nil {
		return nil, err
	}
	if len(notFound) > 0 && !args.SkipUnmatchedMembers {
		var err error
		for _, member := range notFound {
			err = errors.Append(err, errors.Newf("member not found: %s", member))
		}
		return nil, err
	}

	team, err := findTeam(ctx, r.db.Teams(), args.Team, args.TeamName)
	if err != nil {
		return nil, err
	}
	if ok, err := canModifyTeam(ctx, r.db, team); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrNoAccessToTeam
	}
	var membersToRemove []*types.TeamMember
	for _, user := range users {
		membersToRemove = append(membersToRemove, &types.TeamMember{
			UserID: user.ID,
			TeamID: team.ID,
		})
	}
	if len(membersToRemove) > 0 {
		if err := r.db.Teams().DeleteTeamMember(ctx, membersToRemove...); err != nil {
			return nil, err
		}
	}
	return NewTeamResolver(r.db, team), nil
}

type QueryTeamsArgs struct {
	ListTeamsArgs
	ExceptAncestor    *graphql.ID
	IncludeChildTeams *bool
}

func (r *schemaResolver) Teams(ctx context.Context, args *QueryTeamsArgs) (*teamConnectionResolver, error) {
	if err := areTeamEndpointsAvailable(); err != nil {
		return nil, err
	}
	c := &teamConnectionResolver{db: r.db}
	if err := c.applyArgs(&args.ListTeamsArgs); err != nil {
		return nil, err
	}
	if args.ExceptAncestor != nil {
		id, err := UnmarshalTeamID(*args.ExceptAncestor)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot interpret exceptAncestor id: %q", *args.ExceptAncestor)
		}
		c.exceptAncestorID = id
	}
	c.onlyRootTeams = true
	if args.IncludeChildTeams != nil && *args.IncludeChildTeams {
		c.onlyRootTeams = false
	}
	return c, nil
}

type TeamArgs struct {
	Name string
}

func (r *schemaResolver) Team(ctx context.Context, args *TeamArgs) (*TeamResolver, error) {
	if err := areTeamEndpointsAvailable(); err != nil {
		return nil, err
	}

	t, err := r.db.Teams().GetTeamByName(ctx, args.Name)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return NewTeamResolver(r.db, t), nil
}

func (r *UserResolver) Teams(ctx context.Context, args *ListTeamsArgs) (*teamConnectionResolver, error) {
	if err := areTeamEndpointsAvailable(); err != nil {
		return nil, err
	}

	c := &teamConnectionResolver{db: r.db}
	if err := c.applyArgs(args); err != nil {
		return nil, err
	}
	c.onlyRootTeams = true
	return c, nil
}

// usersForTeamMembers returns the matching users for the given slice of TeamMemberInput.
// For each input, we look at ID, Username, Email, and then External Account in this precedence
// order. If one field is specified, it is used. If not found, under that predicate, the
// next one is tried. If the record doesn't match a user entirely, it is skipped. (As opposed
// to an error being returned. This is more convenient for ingestion as it allows us to
// skip over users for now.) We might want to revisit this later.
func usersForTeamMembers(ctx context.Context, db database.DB, members []TeamMemberInput) (users []*types.User, noMatch []TeamMemberInput, err error) {
	// First, look at IDs.
	ids := []int32{}
	members, err = extractMembers(members, func(m TeamMemberInput) (drop bool, err error) {
		// If ID is specified for the member, we try to find the user by ID.
		if m.UserID == nil {
			return false, nil
		}
		id, err := UnmarshalUserID(*m.UserID)
		if err != nil {
			return false, err
		}
		ids = append(ids, id)
		return true, nil
	})
	if err != nil {
		return nil, nil, err
	}
	if len(ids) > 0 {
		users, err = db.Users().List(ctx, &database.UsersListOptions{UserIDs: ids})
		if err != nil {
			return nil, nil, err
		}
	}

	// Now, look at all that have username set.
	usernames := []string{}
	members, err = extractMembers(members, func(m TeamMemberInput) (drop bool, err error) {
		if m.Username == nil {
			return false, nil
		}
		usernames = append(usernames, *m.Username)
		return true, nil
	})
	if err != nil {
		return nil, nil, err
	}
	if len(usernames) > 0 {
		us, err := db.Users().List(ctx, &database.UsersListOptions{Usernames: usernames})
		if err != nil {
			return nil, nil, err
		}
		users = append(users, us...)
	}

	// Next up: Email.
	members, err = extractMembers(members, func(m TeamMemberInput) (drop bool, err error) {
		if m.Email == nil {
			return false, nil
		}
		user, err := db.Users().GetByVerifiedEmail(ctx, *m.Email)
		if err != nil {
			return false, err
		}
		users = append(users, user)
		return true, nil
	})
	if err != nil {
		return nil, nil, err
	}

	// Next up: ExternalAccount.
	members, err = extractMembers(members, func(m TeamMemberInput) (drop bool, err error) {
		if m.ExternalAccountServiceID == nil || m.ExternalAccountServiceType == nil {
			return false, nil
		}

		eas, err := db.UserExternalAccounts().List(ctx, database.ExternalAccountsListOptions{
			ServiceType: *m.ExternalAccountServiceType,
			ServiceID:   *m.ExternalAccountServiceID,
		})
		if err != nil {
			return false, err
		}
		for _, ea := range eas {
			if m.ExternalAccountAccountID != nil {
				if ea.AccountID == *m.ExternalAccountAccountID {
					u, err := db.Users().GetByID(ctx, ea.UserID)
					if err != nil {
						return false, err
					}
					users = append(users, u)
					return true, nil
				}
				continue
			}
			if m.ExternalAccountLogin != nil {
				if ea.PublicAccountData.Login == *m.ExternalAccountAccountID {
					u, err := db.Users().GetByID(ctx, ea.UserID)
					if err != nil {
						return false, err
					}
					users = append(users, u)
					return true, nil
				}
				continue
			}
		}
		return false, nil
	})

	return users, members, err
}

// extractMembers calls pred on each member, and returns a new slice of members
// for which the predicate was falsey.
func extractMembers(members []TeamMemberInput, pred func(member TeamMemberInput) (drop bool, err error)) ([]TeamMemberInput, error) {
	remaining := []TeamMemberInput{}
	for _, member := range members {
		ok, err := pred(member)
		if err != nil {
			return nil, err
		}
		if !ok {
			remaining = append(remaining, member)
		}
	}
	return remaining, nil
}

var ErrNoAccessToTeam = errors.New("user cannot modify team")

func areTeamEndpointsAvailable() error {
	if dotcom.SourcegraphDotComMode() {
		return errors.New("teams are not available on sourcegraph.com")
	}
	if !own.IsEnabled() {
		return errors.New("teams are disabled")
	}
	return nil
}

func canModifyTeam(ctx context.Context, db database.DB, team *types.Team) (bool, error) {
	if team.ID == 0 {
		return false, nil
	}
	if isSiteAdmin(ctx, db) {
		return true, nil
	}
	if team.ReadOnly {
		return false, nil
	}
	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return false, auth.ErrNotAuthenticated
	}
	// The creator can always modify a team.
	if team.CreatorID != 0 && team.CreatorID == a.UID {
		return true, nil
	}
	return db.Teams().IsTeamMember(ctx, team.ID, a.UID)
}

func isSiteAdmin(ctx context.Context, db database.DB) bool {
	err := auth.CheckCurrentUserIsSiteAdmin(ctx, db)
	return err == nil
}
