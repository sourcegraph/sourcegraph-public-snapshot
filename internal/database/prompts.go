package database

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type PromptStore interface {
	Create(_ context.Context, _ *types.Prompt) (*types.Prompt, error)
	Update(_ context.Context, _ *types.Prompt) (*types.Prompt, error)
	UpdateOwner(_ context.Context, id int32, newOwner types.Namespace) (*types.Prompt, error)
	UpdateVisibility(_ context.Context, id int32, secret bool) (*types.Prompt, error)
	Delete(context.Context, int32) error
	GetByID(context.Context, int32) (*types.Prompt, error)
	List(context.Context, PromptListArgs, *PaginationArgs) ([]*types.Prompt, error)
	Count(context.Context, PromptListArgs) (int, error)
	MarshalToCursor(*types.Prompt, OrderBy) (types.MultiCursor, error)
	UnmarshalValuesFromCursor(types.MultiCursor) ([]any, error)
	WithTransact(context.Context, func(PromptStore) error) error
	With(basestore.ShareableStore) PromptStore
	basestore.ShareableStore
}

type promptStore struct {
	*basestore.Store
}

// PromptsWith instantiates and returns a new PromptStore using the other store handle.
func PromptsWith(other basestore.ShareableStore) PromptStore {
	return &promptStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *promptStore) With(other basestore.ShareableStore) PromptStore {
	return &promptStore{Store: s.Store.With(other)}
}

func (s *promptStore) WithTransact(ctx context.Context, f func(PromptStore) error) error {
	return s.Store.WithTransact(ctx, func(tx *basestore.Store) error {
		return f(&promptStore{Store: tx})
	})
}

var (
	errPromptNotFound = resourceNotFoundError{noun: "prompt"}
	promptColumns     = sqlf.Sprintf("name, description, definition_text, draft, owner_user_id, owner_org_id, visibility_secret, created_by, created_at, updated_by, updated_at")
)

// Create creates a new prompt with the specified parameters. The ID field must be zero, or an error
// will be returned.
//
// 🚨 SECURITY: This method does NOT verify the user's identity or that the user is an admin. It is
// the caller's responsibility to ensure the user has proper permissions to create the prompt.
func (s *promptStore) Create(ctx context.Context, newPrompt *types.Prompt) (created *types.Prompt, err error) {
	if newPrompt.ID != 0 {
		return nil, errors.New("newPrompt.ID must be zero")
	}

	tr, ctx := trace.New(ctx, "database.Prompts.Create")
	defer tr.EndWithErr(&err)

	actorUID := actor.FromContext(ctx).UID

	return scanPrompt(
		s.QueryRow(ctx,
			sqlf.Sprintf(`
			INSERT INTO prompts(%v)
			VALUES(%v, %v, %v, %v, %v, %v, %v, %v, DEFAULT, %v, DEFAULT)
			RETURNING id, %v, ''::text`,
				promptColumns,
				newPrompt.Name,
				newPrompt.Description,
				newPrompt.DefinitionText,
				newPrompt.Draft,
				newPrompt.Owner.User,
				newPrompt.Owner.Org,
				newPrompt.VisibilitySecret,
				actorUID,
				actorUID,
				promptColumns,
			),
		))
}

// Update updates an existing prompt.
//
// 🚨 SECURITY: This method does NOT verify the user's identity or that the user is an admin. It is
// the caller's responsibility to ensure the user has proper permissions to perform the update.
func (s *promptStore) Update(ctx context.Context, prompt *types.Prompt) (updated *types.Prompt, err error) {
	tr, ctx := trace.New(ctx, "database.Prompts.Update")
	defer tr.EndWithErr(&err)

	return s.update(ctx, prompt.ID, []*sqlf.Query{
		sqlf.Sprintf("name=%v", prompt.Name),
		sqlf.Sprintf("description=%v", prompt.Description),
		sqlf.Sprintf("definition_text=%v", prompt.DefinitionText),
		sqlf.Sprintf("draft=%v", prompt.Draft),
	})
}

// UpdateOwner updates the owner of an existing prompt.
//
// 🚨 SECURITY: This method does NOT verify that the user has permissions to do this. The caller
// MUST do so.
func (s *promptStore) UpdateOwner(ctx context.Context, id int32, newOwner types.Namespace) (updated *types.Prompt, err error) {
	tr, ctx := trace.New(ctx, "database.Prompts.UpdateOwner")
	defer tr.EndWithErr(&err)
	return s.update(ctx, id, []*sqlf.Query{
		sqlf.Sprintf("owner_user_id=%v", newOwner.User),
		sqlf.Sprintf("owner_org_id=%v", newOwner.Org),
	})
}

// UpdateVisibility updates the visibility state of an existing prompt.
//
// 🚨 SECURITY: This method does NOT verify that the user has permissions to do this. The caller
// MUST do so.
func (s *promptStore) UpdateVisibility(ctx context.Context, id int32, secret bool) (updated *types.Prompt, err error) {
	tr, ctx := trace.New(ctx, "database.Prompts.UpdateVisibility")
	defer tr.EndWithErr(&err)
	return s.update(ctx, id, []*sqlf.Query{sqlf.Sprintf("visibility_secret=%v", secret)})
}

func (s *promptStore) update(ctx context.Context, id int32, updates []*sqlf.Query) (updated *types.Prompt, err error) {
	actorUID := actor.FromContext(ctx).UID
	updates = append(updates,
		sqlf.Sprintf("updated_at=now()"),
		sqlf.Sprintf("updated_by=%v", actorUID),
	)
	return scanPrompt(
		s.QueryRow(ctx,
			sqlf.Sprintf(
				`UPDATE prompts SET %s WHERE id=%v RETURNING id, %v, ''::text`,
				sqlf.Join(updates, ", "),
				id,
				promptColumns,
			),
		))
}

// Delete hard-deletes an existing prompt.
//
// 🚨 SECURITY: This method does NOT verify the user's identity or that the user is an admin. It is
// the caller's responsibility to ensure the user has proper permissions to perform the delete.
func (s *promptStore) Delete(ctx context.Context, id int32) (err error) {
	tr, ctx := trace.New(ctx, "database.Prompts.Delete")
	defer tr.EndWithErr(&err)
	res, err := s.Handle().ExecContext(ctx, `DELETE FROM prompts WHERE id=$1`, id)
	if err == nil {
		var nrows int64
		nrows, err = res.RowsAffected()
		if nrows == 0 {
			err = errPromptNotFound
		}
	}
	return err
}

// GetByID returns the prompt with the given ID.
//
// 🚨 SECURITY: This method does NOT verify the user's identity or that the user is an admin. It is
// the caller's responsibility to ensure this response only makes it to users with proper
// permissions to access the prompt.
func (s *promptStore) GetByID(ctx context.Context, id int32) (_ *types.Prompt, err error) {
	tr, ctx := trace.New(ctx, "database.Prompts.GetByID")
	defer tr.EndWithErr(&err)

	return scanPrompt(s.QueryRow(ctx, sqlf.Sprintf(`SELECT id, %v, name_with_owner FROM prompts_view WHERE id=%v`, promptColumns, id)))
}

type PromptListArgs struct {
	Query          string
	AffiliatedUser *int32
	PublicOnly     bool
	Owner          *types.Namespace
	HideDrafts     bool
}

type PromptsOrderBy uint8

const (
	PromptsOrderByID PromptsOrderBy = iota
	PromptsOrderByNameWithOwner
	PromptsOrderByUpdatedAt
)

func (v PromptsOrderBy) ToOptions() (orderBy OrderBy, ascending bool) {
	switch v {
	case PromptsOrderByUpdatedAt:
		orderBy = []OrderByOption{{Field: "updated_at"}}
		ascending = false
	case PromptsOrderByNameWithOwner:
		orderBy = []OrderByOption{{Field: "name_with_owner"}}
		ascending = true
	case PromptsOrderByID:
		orderBy = []OrderByOption{{Field: "id"}}
		ascending = true
	default:
		panic("invalid PromptsOrderBy value")
	}
	return orderBy, ascending
}

func (a PromptListArgs) toSQL() (where []*sqlf.Query, err error) {
	if a.Query != "" {
		queryStr := "%" + a.Query + "%"
		where = append(where, sqlf.Sprintf("(name_with_owner ILIKE %v OR description ILIKE %v OR definition_text ILIKE %v)", queryStr, queryStr, queryStr))
	}
	if a.AffiliatedUser != nil {
		affiliatedConds := []*sqlf.Query{
			sqlf.Sprintf("owner_user_id=%v", *a.AffiliatedUser),
			sqlf.Sprintf("owner_org_id IN (SELECT org_members.org_id FROM org_members LEFT JOIN orgs ON orgs.id=org_members.org_id WHERE orgs.deleted_at IS NULL AND org_members.user_id=%v)", *a.AffiliatedUser),
			sqlf.Sprintf("NOT visibility_secret"), // treat all public items as though they're affiliated with the current user
		}
		where = append(where,
			sqlf.Sprintf("(%v)", sqlf.Join(affiliatedConds, ") OR (")),
		)
	}
	if a.PublicOnly {
		where = append(where, sqlf.Sprintf("NOT visibility_secret"))
	}
	if a.Owner != nil {
		if a.Owner.User != nil && *a.Owner.User != 0 {
			where = append(where, sqlf.Sprintf("owner_user_id=%v", *a.Owner.User))
		} else if a.Owner.Org != nil && *a.Owner.Org != 0 {
			where = append(where, sqlf.Sprintf("owner_org_id=%v", *a.Owner.Org))
		} else {
			return nil, errors.New("invalid owner (no user or org ID)")
		}
	}
	if a.HideDrafts {
		where = append(where, sqlf.Sprintf("NOT draft"))
	}
	if len(where) == 0 {
		where = append(where, sqlf.Sprintf("TRUE"))
	}

	return where, nil
}

// List lists all prompts matching the given filter args.
//
// 🚨 SECURITY: This method does NOT perform authorization checks.
func (s *promptStore) List(ctx context.Context, args PromptListArgs, paginationArgs *PaginationArgs) (_ []*types.Prompt, err error) {
	tr, ctx := trace.New(ctx, "database.Prompts.List")
	defer tr.EndWithErr(&err)

	where, err := args.toSQL()
	if err != nil {
		return nil, err
	}

	if paginationArgs == nil {
		paginationArgs = &PaginationArgs{}
	}
	pg := paginationArgs.SQL()
	if pg.Where != nil {
		where = append(where, pg.Where)
	}

	query := sqlf.Sprintf(`SELECT id, %v, name_with_owner FROM prompts_view WHERE (%v)`,
		promptColumns, sqlf.Join(where, ") AND ("),
	)
	query = pg.AppendOrderToQuery(query)
	query = pg.AppendLimitToQuery(query)
	return scanPrompts(s.Query(ctx, query))
}

var scanPrompts = basestore.NewSliceScanner(scanPrompt)

func scanPrompt(s dbutil.Scanner) (*types.Prompt, error) {
	var row types.Prompt
	if err := s.Scan(
		&row.ID,
		&row.Name,
		&row.Description,
		&row.DefinitionText,
		&row.Draft,
		&row.Owner.User,
		&row.Owner.Org,
		&row.VisibilitySecret,
		&row.CreatedByUser,
		&row.CreatedAt,
		&row.UpdatedByUser,
		&row.UpdatedAt,
		&row.NameWithOwner,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, errPromptNotFound
		}
		return nil, errors.Wrap(err, "Scan")
	}
	return &row, nil
}

// Count counts all prompts matching the given filter args.
//
// 🚨 SECURITY: This method does NOT perform authorization checks.
func (s *promptStore) Count(ctx context.Context, args PromptListArgs) (count int, err error) {
	tr, ctx := trace.New(ctx, "database.Prompts.Count")
	defer tr.EndWithErr(&err)

	where, err := args.toSQL()
	if err != nil {
		return 0, err
	}
	query := sqlf.Sprintf(`SELECT COUNT(*) FROM prompts_view WHERE (%v)`, sqlf.Join(where, ") AND ("))
	count, _, err = basestore.ScanFirstInt(s.Query(ctx, query))
	return count, err
}

// MarshalToCursor creates a cursor from the given item. It is used for pagination; see
// ConnectionResolverStore.
func (s *promptStore) MarshalToCursor(item *types.Prompt, orderBy OrderBy) (types.MultiCursor, error) {
	var cursors types.MultiCursor
	for _, o := range orderBy {
		c := types.Cursor{Column: o.Field}
		switch o.Field {
		case "id":
			c.Value = strconv.FormatInt(int64(item.ID), 10)
		case "name_with_owner":
			if item.NameWithOwner == "" {
				return nil, errors.New("unexpected empty name_with_owner")
			}
			c.Value = item.NameWithOwner
		case "updated_at":
			c.Value = strconv.FormatInt(item.UpdatedAt.UnixNano(), 10)
		default:
			return nil, errors.New("unexpected orderBy column")
		}
		cursors = append(cursors, &c)
	}
	return cursors, nil
}

// UnmarshalValuesFromCursor extracts the DB values from the cursor into a slice, with values at a
// given index corresponding to the cursor's field at that index. It is used for pagination; see
// ConnectionResolverStore.
func (s *promptStore) UnmarshalValuesFromCursor(cursor types.MultiCursor) ([]any, error) {
	values := make([]any, len(cursor))
	for i, c := range cursor {
		switch c.Column {
		case "id":
			id, err := strconv.ParseInt(c.Value, 10, 32)
			if err != nil {
				return nil, errors.Wrap(err, "ParseInt(id)")
			}
			values[i] = int32(id)
		case "name_with_owner":
			values[i] = c.Value
		case "updated_at":
			updatedAt, err := strconv.ParseInt(c.Value, 10, 64)
			if err != nil {
				return nil, errors.Wrap(err, "ParseInt(updated_at)")
			}
			values[i] = time.Unix(0, updatedAt)
		default:
			return nil, errors.New("unexpected orderBy column")
		}
	}
	return values, nil
}
