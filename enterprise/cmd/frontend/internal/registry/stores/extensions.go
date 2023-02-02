package stores

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Extension describes an extension in the extension registry.
//
// It is the internal form of github.com/sourcegraph/sourcegraph/internal/registry.Extension (which is
// the external API type). These types should generally be kept in sync, but registry.Extension
// updates require backcompat.
type Extension struct {
	ID        int32
	UUID      string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time

	// NonCanonicalExtensionID is the denormalized fully qualified extension ID
	// ("[registry/]publisher/name" format), using the username/name of the extension's publisher
	// (joined from another table) as of when the query executed. Do not persist this, because the
	// (denormalized) registry and publisher names can change.
	//
	// If this value is obtained directly from a method on RegistryExtensions, this field will never
	// contain the registry name prefix (which is necessary to distinguish local extensions from
	// remote extensions). Call prefixLocalExtensionID to add it. The recommended way to apply this
	// automatically (when needed) is to use registry.GetExtensionByExtensionID instead of
	// (dbExtensions).GetByExtensionID.
	NonCanonicalExtensionID string

	// NonCanonicalRegistry is the denormalized registry name (as of when this field was set). This
	// field is only set by prefixLocalExtensionID and is always empty if this value is obtained
	// directly from a method on RegistryExtensions. Do not persist this value, because the
	// (denormalized) registry name can change.
	NonCanonicalRegistry string
}

// ExtensionNotFoundError occurs when an extension is not found in the extension registry.
type ExtensionNotFoundError struct {
	args []any
}

// NotFound implements errcode.NotFounder.
func (err ExtensionNotFoundError) NotFound() bool { return true }

func (err ExtensionNotFoundError) Error() string {
	return fmt.Sprintf("registry extension not found: %v", err.args)
}

type ExtensionStore interface {
	// Create creates a new extension in the extension registry. Exactly 1 of publisherUserID and publisherOrgID must be nonzero.
	Create(ctx context.Context, publisherUserID, publisherOrgID int32, name string) (id int32, err error)
	// GetByID retrieves the registry extension (if any) given its ID.
	//
	// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this registry extension.
	GetByID(ctx context.Context, id int32) (*Extension, error)
	// GetByUUID retrieves the registry extension (if any) given its UUID.
	//
	// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this registry extension.
	GetByUUID(ctx context.Context, uuid string) (*Extension, error)
	// GetByExtensionID retrieves the registry extension (if any) given its extension ID, which is the
	// concatenation of the publisher name, a slash ("/"), and the extension name.
	//
	// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this registry extension.
	GetByExtensionID(ctx context.Context, extensionID string) (*Extension, error)
	// List lists all registry extensions that satisfy the options.
	//
	// 🚨 SECURITY: The caller must ensure that the actor is permitted to list with the specified
	// options.
	List(ctx context.Context, opt ExtensionsListOptions) ([]*Extension, error)
	// Count counts all registry extensions that satisfy the options (ignoring limit and offset).
	//
	// 🚨 SECURITY: The caller must ensure that the actor is permitted to count the results.
	Count(ctx context.Context, opt ExtensionsListOptions) (int, error)
	// Delete marks an registry extension as deleted.
	Delete(ctx context.Context, id int32) error

	Transact(context.Context) (ExtensionStore, error)
	With(basestore.ShareableStore) ExtensionStore
	basestore.ShareableStore
}

type extensionStore struct {
	*basestore.Store
}

var _ ExtensionStore = (*extensionStore)(nil)

// Extensions instantiates and returns a new ExtensionsStore with prepared statements.
func Extensions(db database.DB) ExtensionStore {
	return &extensionStore{Store: basestore.NewWithHandle(db.Handle())}
}

// ExtensionsWith instantiates and returns a new ExtensionsStore using the other store handle.
func ExtensionsWith(other basestore.ShareableStore) ExtensionStore {
	return &extensionStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *extensionStore) With(other basestore.ShareableStore) ExtensionStore {
	return &extensionStore{Store: s.Store.With(other)}
}

func (s *extensionStore) Transact(ctx context.Context) (ExtensionStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &extensionStore{Store: txBase}, err
}

func (s *extensionStore) Create(ctx context.Context, publisherUserID, publisherOrgID int32, name string) (id int32, err error) {
	if publisherUserID != 0 && publisherOrgID != 0 {
		return 0, errors.New("at most 1 of the publisher user/org may be set")
	}

	newUUID, err := uuid.NewRandom()
	if err != nil {
		return 0, err
	}

	// Include users/orgs table query (with "FOR UPDATE") to ensure that the publisher user/org
	// not been deleted. If it was deleted, the query will return an error.
	q := sqlf.Sprintf(`
INSERT INTO registry_extensions
	(uuid, publisher_user_id, publisher_org_id, name)
VALUES(
	%s,
	(SELECT id FROM users WHERE id = %s AND deleted_at IS NULL FOR UPDATE),
	(SELECT id FROM orgs WHERE id = %s AND deleted_at IS NULL FOR UPDATE),
	%s
)
RETURNING id`,
		newUUID,
		publisherUserID,
		publisherOrgID,
		name,
	)

	if err := s.QueryRow(ctx, q).Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

func (s *extensionStore) GetByID(ctx context.Context, id int32) (*Extension, error) {
	results, err := s.list(ctx, []*sqlf.Query{sqlf.Sprintf("x.id = %d", id)}, nil)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, ExtensionNotFoundError{[]any{id}}
	}

	return results[0], nil
}

func (s *extensionStore) GetByUUID(ctx context.Context, uuid string) (*Extension, error) {
	results, err := s.list(ctx, []*sqlf.Query{sqlf.Sprintf("x.uuid = %d", uuid)}, nil)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, ExtensionNotFoundError{[]any{uuid}}
	}

	return results[0], nil
}

const (
	// extensionPublisherNameExpr is the SQL expression for the extension's publisher's name (using
	// the table aliases created by (*extensionsStore).listCountSQL.
	extensionPublisherNameExpr = "COALESCE(users.username, orgs.name)"

	// extensionIDExpr is the SQL expression for the extension ID (using the table aliases created by
	// (*extensionsStore).listCountSQL.
	extensionIDExpr = "CONCAT(" + extensionPublisherNameExpr + ", '/', x.name)"
)

func (s *extensionStore) GetByExtensionID(ctx context.Context, extensionID string) (*Extension, error) {
	// TODO(sqs): prevent the creation of an org with the same name as a user so that there is no
	// ambiguity as to whether the publisher refers to a user or org by the given name
	// (https://github.com/sourcegraph/sourcegraph/issues/12068).
	parts := strings.SplitN(extensionID, "/", 2)
	if len(parts) < 2 {
		return nil, ExtensionNotFoundError{[]any{fmt.Sprintf("extensionID %q", extensionID)}}
	}
	publisherName := parts[0]
	extensionName := parts[1]

	results, err := s.list(ctx, []*sqlf.Query{
		sqlf.Sprintf("x.name = %s", extensionName),
		sqlf.Sprintf("(users.username = %s OR orgs.name = %s)", publisherName, publisherName),
	}, nil)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, ExtensionNotFoundError{[]any{fmt.Sprintf("extensionID %q", extensionID)}}
	}

	return results[0], nil
}

// ExtensionsListOptions contains options for listing registry extensions.
type ExtensionsListOptions struct {
	// Query matches the extension ID and latest release's manifest's title
	Query        string
	ExtensionIDs []string

	*database.LimitOffset
}

func (o ExtensionsListOptions) sqlConditions() []*sqlf.Query {
	var conds []*sqlf.Query
	if o.Query != "" {
		likePattern := func(value string) string {
			return "%" + strings.ReplaceAll(strings.ToLower(value), " ", "%") + "%"
		}
		queryConds := []*sqlf.Query{
			sqlf.Sprintf(extensionIDExpr+" ILIKE %s", likePattern(o.Query)),
			// BACKCOMPAT: This still reads the title property even though extensions no longer have titles.
			sqlf.Sprintf(`CASE WHEN rer.manifest IS NOT NULL THEN (rer.manifest->>'description' ILIKE %s OR rer.manifest->>'title' ILIKE %s) ELSE false END`, likePattern(o.Query), likePattern(o.Query)),
		}
		conds = append(conds, sqlf.Sprintf("(%s)", sqlf.Join(queryConds, ") OR (")))
	}
	if o.ExtensionIDs != nil {
		ids := make([]*sqlf.Query, len(o.ExtensionIDs)+1)
		for i, id := range o.ExtensionIDs {
			ids[i] = sqlf.Sprintf("%s", id)
		}
		ids[len(o.ExtensionIDs)] = sqlf.Sprintf("NULL")
		conds = append(conds, sqlf.Sprintf(extensionIDExpr+` IN (%s)`, sqlf.Join(ids, ",")))
	}
	if len(conds) == 0 {
		conds = append(conds, sqlf.Sprintf("TRUE"))
	}
	return conds
}

func (s *extensionStore) List(ctx context.Context, opt ExtensionsListOptions) ([]*Extension, error) {
	return s.list(ctx, opt.sqlConditions(), opt.LimitOffset)
}

func (s *extensionStore) listCountSQL(conds []*sqlf.Query) *sqlf.Query {
	return sqlf.Sprintf(`
FROM registry_extensions x
LEFT JOIN users ON users.id = publisher_user_id AND users.deleted_at IS NULL
LEFT JOIN orgs ON orgs.id = publisher_org_id AND orgs.deleted_at IS NULL
LEFT JOIN registry_extension_releases rer ON rer.registry_extension_id = x.id AND rer.deleted_at IS NULL
WHERE (%s)
  -- Join only to latest release from registry_extension_releases.
  AND NOT EXISTS (SELECT 1 FROM registry_extension_releases rer2
                  WHERE rer.registry_extension_id=rer2.registry_extension_id
                    AND rer2.deleted_at IS NULL
                    AND rer2.created_at > rer.created_at
  )
  AND x.deleted_at IS NULL`,
		sqlf.Join(conds, ") AND ("))
}

func (s *extensionStore) list(ctx context.Context, conds []*sqlf.Query, limitOffset *database.LimitOffset) ([]*Extension, error) {
	q := sqlf.Sprintf(`
SELECT x.id, x.uuid, x.name, x.created_at, x.updated_at,
  `+extensionIDExpr+` AS non_canonical_extension_id
%s
ORDER BY x.id ASC
%s`,
		s.listCountSQL(conds),
		limitOffset.SQL(),
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var results []*Extension
	for rows.Next() {
		var t Extension
		if err := rows.Scan(&t.ID, &t.UUID, &t.Name, &t.CreatedAt, &t.UpdatedAt, &t.NonCanonicalExtensionID); err != nil {
			return nil, err
		}
		results = append(results, &t)
	}

	return results, nil
}

func (s *extensionStore) Count(ctx context.Context, opt ExtensionsListOptions) (int, error) {
	q := sqlf.Sprintf(`
SELECT COUNT(*) %s
`, s.listCountSQL(opt.sqlConditions()))

	var count int
	if err := s.QueryRow(ctx, q).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (s *extensionStore) Delete(ctx context.Context, id int32) error {
	res, err := s.ExecResult(ctx, sqlf.Sprintf(`
UPDATE
	registry_extensions
SET
	deleted_at = NOW()
WHERE
	id = %s
	AND
	deleted_at IS NULL
`,
		id,
	))
	if err != nil {
		return err
	}

	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if nrows == 0 {
		return ExtensionNotFoundError{[]any{id}}
	}

	return nil
}
