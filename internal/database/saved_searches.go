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

type SavedSearchStore interface {
	Create(_ context.Context, _ *types.SavedSearch) (*types.SavedSearch, error)
	Update(_ context.Context, _ *types.SavedSearch) (*types.SavedSearch, error)
	UpdateOwner(_ context.Context, id int32, newOwner types.Namespace) (*types.SavedSearch, error)
	UpdateVisibility(_ context.Context, id int32, secret bool) (*types.SavedSearch, error)
	Delete(context.Context, int32) error
	GetByID(context.Context, int32) (*types.SavedSearch, error)
	List(context.Context, SavedSearchListArgs, *PaginationArgs) ([]*types.SavedSearch, error)
	Count(context.Context, SavedSearchListArgs) (int, error)
	MarshalToCursor(*types.SavedSearch, OrderBy) (types.MultiCursor, error)
	UnmarshalValuesFromCursor(types.MultiCursor) ([]any, error)
	WithTransact(context.Context, func(SavedSearchStore) error) error
	With(basestore.ShareableStore) SavedSearchStore
	basestore.ShareableStore
}

type savedSearchStore struct {
	*basestore.Store
}

// SavedSearchesWith instantiates and returns a new SavedSearchStore using the other store handle.
func SavedSearchesWith(other basestore.ShareableStore) SavedSearchStore {
	return &savedSearchStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *savedSearchStore) With(other basestore.ShareableStore) SavedSearchStore {
	return &savedSearchStore{Store: s.Store.With(other)}
}

func (s *savedSearchStore) WithTransact(ctx context.Context, f func(SavedSearchStore) error) error {
	return s.Store.WithTransact(ctx, func(tx *basestore.Store) error {
		return f(&savedSearchStore{Store: tx})
	})
}

var (
	errSavedSearchNotFound = resourceNotFoundError{noun: "saved search"}
	savedSearchColumns     = sqlf.Sprintf("description, query, draft, user_id, org_id, visibility_secret, created_by, created_at, updated_by, updated_at")
)

// Create creates a new saved search with the specified parameters. The ID field must be zero, or an
// error will be returned.
//
// 🚨 SECURITY: This method does NOT verify the user's identity or that the user is an admin. It is
// the caller's responsibility to ensure the user has proper permissions to create the saved search.
func (s *savedSearchStore) Create(ctx context.Context, newSavedSearch *types.SavedSearch) (created *types.SavedSearch, err error) {
	if newSavedSearch.ID != 0 {
		return nil, errors.New("newSavedSearch.ID must be zero")
	}

	tr, ctx := trace.New(ctx, "database.SavedSearches.Create")
	defer tr.EndWithErr(&err)

	actorUID := actor.FromContext(ctx).UID

	return scanSavedSearch(
		s.QueryRow(ctx,
			sqlf.Sprintf(`INSERT INTO saved_searches(%v) VALUES(%v, %v, %v, %v, %v, %v, %v, DEFAULT, %v, DEFAULT) RETURNING id, %v`,
				savedSearchColumns,
				newSavedSearch.Description,
				newSavedSearch.Query,
				newSavedSearch.Draft,
				newSavedSearch.Owner.User,
				newSavedSearch.Owner.Org,
				newSavedSearch.VisibilitySecret,
				actorUID,
				actorUID,
				savedSearchColumns,
			),
		))
}

// Update updates an existing saved search.
//
// 🚨 SECURITY: This method does NOT verify the user's identity or that the user is an admin. It is
// the caller's responsibility to ensure the user has proper permissions to perform the update.
func (s *savedSearchStore) Update(ctx context.Context, savedSearch *types.SavedSearch) (updated *types.SavedSearch, err error) {
	tr, ctx := trace.New(ctx, "database.SavedSearches.Update")
	defer tr.EndWithErr(&err)

	return s.update(ctx, savedSearch.ID, []*sqlf.Query{
		sqlf.Sprintf("description=%s", savedSearch.Description),
		sqlf.Sprintf("query=%s", savedSearch.Query),
		sqlf.Sprintf("draft=%v", savedSearch.Draft),
	})
}

// UpdateOwner updates the owner of an existing saved search.
//
// 🚨 SECURITY: This method does NOT verify that the user has permissions to do this. The caller
// MUST do so.
func (s *savedSearchStore) UpdateOwner(ctx context.Context, id int32, newOwner types.Namespace) (updated *types.SavedSearch, err error) {
	tr, ctx := trace.New(ctx, "database.SavedSearches.UpdateOwner")
	defer tr.EndWithErr(&err)
	return s.update(ctx, id, []*sqlf.Query{
		sqlf.Sprintf("user_id=%v", newOwner.User),
		sqlf.Sprintf("org_id=%v", newOwner.Org),
	})
}

// UpdateVisibility updates the visibility state of an existing saved search.
//
// 🚨 SECURITY: This method does NOT verify that the user has permissions to do this. The caller
// MUST do so.
func (s *savedSearchStore) UpdateVisibility(ctx context.Context, id int32, secret bool) (updated *types.SavedSearch, err error) {
	tr, ctx := trace.New(ctx, "database.SavedSearches.UpdateVisibility")
	defer tr.EndWithErr(&err)
	return s.update(ctx, id, []*sqlf.Query{sqlf.Sprintf("visibility_secret=%v", secret)})
}

func (s *savedSearchStore) update(ctx context.Context, id int32, updates []*sqlf.Query) (updated *types.SavedSearch, err error) {
	actorUID := actor.FromContext(ctx).UID
	updates = append(updates, sqlf.Sprintf("updated_at=now()"), sqlf.Sprintf("updated_by=%v", actorUID))
	return scanSavedSearch(
		s.QueryRow(ctx,
			sqlf.Sprintf(
				`UPDATE saved_searches SET %s WHERE id=%v RETURNING id, %v`,
				sqlf.Join(updates, ", "),
				id,
				savedSearchColumns,
			),
		))
}

// Delete hard-deletes an existing saved search.
//
// 🚨 SECURITY: This method does NOT verify the user's identity or that the user is an admin. It is
// the caller's responsibility to ensure the user has proper permissions to perform the delete.
func (s *savedSearchStore) Delete(ctx context.Context, id int32) (err error) {
	tr, ctx := trace.New(ctx, "database.SavedSearches.Delete")
	defer tr.EndWithErr(&err)
	res, err := s.Handle().ExecContext(ctx, `DELETE FROM saved_searches WHERE id=$1`, id)
	if err == nil {
		var nrows int64
		nrows, err = res.RowsAffected()
		if nrows == 0 {
			err = errSavedSearchNotFound
		}
	}
	return err
}

// GetByID returns the saved search with the given ID.
//
// 🚨 SECURITY: This method does NOT verify the user's identity or that the user is an admin. It is
// the caller's responsibility to ensure this response only makes it to users with proper
// permissions to access the saved search.
func (s *savedSearchStore) GetByID(ctx context.Context, id int32) (_ *types.SavedSearch, err error) {
	tr, ctx := trace.New(ctx, "database.SavedSearches.GetByID")
	defer tr.EndWithErr(&err)

	return scanSavedSearch(s.QueryRow(ctx, sqlf.Sprintf(`SELECT id, %v FROM saved_searches WHERE id=%v`, savedSearchColumns, id)))
}

type SavedSearchListArgs struct {
	Query          string
	AffiliatedUser *int32
	PublicOnly     bool
	Owner          *types.Namespace
	HideDrafts     bool
}

type SavedSearchesOrderBy uint8

const (
	SavedSearchesOrderByUpdatedAt SavedSearchesOrderBy = iota // default
	SavedSearchesOrderByDescription
	SavedSearchesOrderByID
)

func (v SavedSearchesOrderBy) ToOptions() (orderBy OrderBy, ascending bool) {
	switch v {
	case SavedSearchesOrderByUpdatedAt:
		orderBy = []OrderByOption{{Field: "updated_at"}}
		ascending = false
	case SavedSearchesOrderByDescription:
		orderBy = []OrderByOption{{Field: "description"}}
		ascending = true
	case SavedSearchesOrderByID:
		orderBy = []OrderByOption{{Field: "id"}}
		ascending = true
	default:
		panic("invalid SavedSearchesOrderBy value")
	}
	return orderBy, ascending
}

func (a SavedSearchListArgs) toSQL() (where []*sqlf.Query, err error) {
	if a.Query != "" {
		queryStr := "%" + a.Query + "%"
		where = append(where, sqlf.Sprintf("(description ILIKE %v OR query ILIKE %v)", queryStr, queryStr))
	}
	if a.AffiliatedUser != nil {
		affiliatedConds := []*sqlf.Query{
			sqlf.Sprintf("user_id=%v", *a.AffiliatedUser),
			sqlf.Sprintf("org_id IN (SELECT org_members.org_id FROM org_members LEFT JOIN orgs ON orgs.id=org_members.org_id WHERE orgs.deleted_at IS NULL AND org_members.user_id=%v)", *a.AffiliatedUser),
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
			where = append(where, sqlf.Sprintf("user_id=%v", *a.Owner.User))
		} else if a.Owner.Org != nil && *a.Owner.Org != 0 {
			where = append(where, sqlf.Sprintf("org_id=%v", *a.Owner.Org))
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

// List lists all saved searches matching the given filter args.
//
// 🚨 SECURITY: This method does NOT perform authorization checks.
func (s *savedSearchStore) List(ctx context.Context, args SavedSearchListArgs, paginationArgs *PaginationArgs) (_ []*types.SavedSearch, err error) {
	tr, ctx := trace.New(ctx, "database.SavedSearches.List")
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

	query := sqlf.Sprintf(`SELECT id, %v FROM saved_searches WHERE (%v)`,
		savedSearchColumns, sqlf.Join(where, ") AND ("),
	)
	query = pg.AppendOrderToQuery(query)
	query = pg.AppendLimitToQuery(query)
	return scanSavedSearches(s.Query(ctx, query))
}

var scanSavedSearches = basestore.NewSliceScanner(scanSavedSearch)

func scanSavedSearch(s dbutil.Scanner) (*types.SavedSearch, error) {
	var row types.SavedSearch
	if err := s.Scan(
		&row.ID,
		&row.Description,
		&row.Query,
		&row.Draft,
		&row.Owner.User,
		&row.Owner.Org,
		&row.VisibilitySecret,
		&row.CreatedByUser,
		&row.CreatedAt,
		&row.UpdatedByUser,
		&row.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, errSavedSearchNotFound
		}
		return nil, errors.Wrap(err, "Scan")
	}
	return &row, nil
}

// Count counts all saved searches matching the given filter args.
//
// 🚨 SECURITY: This method does NOT perform authorization checks.
func (s *savedSearchStore) Count(ctx context.Context, args SavedSearchListArgs) (count int, err error) {
	tr, ctx := trace.New(ctx, "database.SavedSearches.Count")
	defer tr.EndWithErr(&err)

	where, err := args.toSQL()
	if err != nil {
		return 0, err
	}
	query := sqlf.Sprintf(`SELECT COUNT(*) FROM saved_searches WHERE (%v)`, sqlf.Join(where, ") AND ("))
	count, _, err = basestore.ScanFirstInt(s.Query(ctx, query))
	return count, err
}

// MarshalToCursor creates a cursor from the given item. It is used for pagination; see
// ConnectionResolverStore.
func (s *savedSearchStore) MarshalToCursor(item *types.SavedSearch, orderBy OrderBy) (types.MultiCursor, error) {
	var cursors types.MultiCursor
	for _, o := range orderBy {
		c := types.Cursor{Column: o.Field}
		switch o.Field {
		case "id":
			c.Value = strconv.FormatInt(int64(item.ID), 10)
		case "description":
			c.Value = item.Description
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
func (s *savedSearchStore) UnmarshalValuesFromCursor(cursor types.MultiCursor) ([]any, error) {
	values := make([]any, len(cursor))
	for i, c := range cursor {
		switch c.Column {
		case "id":
			id, err := strconv.ParseInt(c.Value, 10, 32)
			if err != nil {
				return nil, errors.Wrap(err, "ParseInt(id)")
			}
			values[i] = int32(id)
		case "description":
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
