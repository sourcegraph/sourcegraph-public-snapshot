package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/keegancsmith/sqlf"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db/dbconn"
)

// RegistryExtension describes an extension in the extension registry.
//
// It is the internal form of github.com/sourcegraph/sourcegraph/pkg/registry.Extension (which is
// the external API type). These types should generally be kept in sync, but registry.Extension
// updates require backcompat.
type RegistryExtension struct {
	ID        int32
	UUID      string
	Publisher RegistryPublisher
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
	// remote extensions). Call backend.PrefixLocalExtensionID to add it. The recommended way to
	// apply this automatically (when needed) is to use backend.GetExtensionByExtensionID instead of
	// RegistryExtensions.GetByExtensionID.
	NonCanonicalExtensionID string

	// NonCanonicalRegistry is the denormalized registry name (as of when this field was set). This
	// field is only set by backend.PrefixLocalExtensionID and is always empty if this value is
	// obtained directly from a method on RegistryExtensions. Do not persist this value, because the
	// (denormalized) registry name can change.
	NonCanonicalRegistry string
}

type registryExtensions struct{}

// RegistryExtensionNotFoundError occurs when an extension is not found in the extension registry.
type RegistryExtensionNotFoundError struct {
	args []interface{}
}

// NotFound implements errcode.NotFounder.
func (err RegistryExtensionNotFoundError) NotFound() bool { return true }

func (err RegistryExtensionNotFoundError) Error() string {
	return fmt.Sprintf("registry extension not found: %v", err.args)
}

// Create creates a new extension in the extension registry. Exactly 1 of publisherUserID and publisherOrgID must be nonzero.
func (s *registryExtensions) Create(ctx context.Context, publisherUserID, publisherOrgID int32, name string) (id int32, err error) {
	if Mocks.RegistryExtensions.Create != nil {
		return Mocks.RegistryExtensions.Create(publisherUserID, publisherOrgID, name)
	}

	if publisherUserID != 0 && publisherOrgID != 0 {
		return 0, errors.New("at most 1 of the publisher user/org may be set")
	}

	uuid, err := uuid.NewRandom()
	if err != nil {
		return 0, err
	}

	if err := dbconn.Global.QueryRowContext(ctx,
		// Include users/orgs table query (with "FOR UPDATE") to ensure that the publisher user/org
		// not been deleted. If it was deleted, the query will return an error.
		`
INSERT INTO registry_extensions(uuid, publisher_user_id, publisher_org_id, name)
VALUES(
  $1,
  (SELECT id FROM users WHERE id=$2 AND deleted_at IS NULL FOR UPDATE),
  (SELECT id FROM orgs WHERE id=$3 AND deleted_at IS NULL FOR UPDATE),
  $4
)
RETURNING id
`,
		uuid, publisherUserID, publisherOrgID, name,
	).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// GetByID retrieves the registry extension (if any) given its ID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this registry extension.
func (s *registryExtensions) GetByID(ctx context.Context, id int32) (*RegistryExtension, error) {
	if Mocks.RegistryExtensions.GetByID != nil {
		return Mocks.RegistryExtensions.GetByID(id)
	}

	results, err := s.list(ctx, []*sqlf.Query{sqlf.Sprintf("x.id=%d", id)}, nil, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, RegistryExtensionNotFoundError{[]interface{}{id}}
	}
	return results[0], nil
}

// GetByUUID retrieves the registry extension (if any) given its UUID.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this registry extension.
func (s *registryExtensions) GetByUUID(ctx context.Context, uuid string) (*RegistryExtension, error) {
	if Mocks.RegistryExtensions.GetByUUID != nil {
		return Mocks.RegistryExtensions.GetByUUID(uuid)
	}

	results, err := s.list(ctx, []*sqlf.Query{sqlf.Sprintf("x.uuid=%d", uuid)}, nil, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, RegistryExtensionNotFoundError{[]interface{}{uuid}}
	}
	return results[0], nil
}

const (
	// extensionPublisherNameExpr is the SQL expression for the extension's publisher's name (using
	// the table aliases created by (registryExtensions).listCountSQL.
	extensionPublisherNameExpr = "COALESCE(users.username, orgs.name)"

	// extensionIDExpr is the SQL expression for the extension ID (using the table aliases created by
	// (registryExtensions).listCountSQL.
	extensionIDExpr = "CONCAT(" + extensionPublisherNameExpr + ", '/', x.name)"
)

// GetByExtensionID retrieves the registry extension (if any) given its extension ID, which is the
// concatenation of the publisher name, a slash ("/"), and the extension name.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to view this registry extension.
func (s *registryExtensions) GetByExtensionID(ctx context.Context, extensionID string) (*RegistryExtension, error) {
	if Mocks.RegistryExtensions.GetByExtensionID != nil {
		return Mocks.RegistryExtensions.GetByExtensionID(extensionID)
	}

	// TODO(sqs): prevent the creation of an org with the same name as a user so that there is no
	// ambiguity as to whether the publisher refers to a user or org by the given name
	// (https://github.com/sourcegraph/sourcegraph/issues/12068).
	parts := strings.SplitN(extensionID, "/", 2)
	if len(parts) < 2 {
		return nil, RegistryExtensionNotFoundError{[]interface{}{fmt.Sprintf("extensionID %q", extensionID)}}
	}
	publisherName := parts[0]
	extensionName := parts[1]

	results, err := s.list(ctx, []*sqlf.Query{
		sqlf.Sprintf("x.name=%s", extensionName),
		sqlf.Sprintf("(users.username=%s OR orgs.name=%s)", publisherName, publisherName),
	}, nil, nil)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, RegistryExtensionNotFoundError{[]interface{}{fmt.Sprintf("extensionID %q", extensionID)}}
	}
	return results[0], nil
}

// RegistryExtensionsListOptions contains options for listing registry extensions.
type RegistryExtensionsListOptions struct {
	Publisher              RegistryPublisher
	Query                  string // matches the extension ID
	PrioritizeExtensionIDs []string
	*LimitOffset
}

func (o RegistryExtensionsListOptions) sqlConditions() []*sqlf.Query {
	var conds []*sqlf.Query
	if o.Publisher.UserID != 0 {
		conds = append(conds, sqlf.Sprintf("x.publisher_user_id=%d", o.Publisher.UserID))
	}
	if o.Publisher.OrgID != 0 {
		conds = append(conds, sqlf.Sprintf("x.publisher_org_id=%d", o.Publisher.OrgID))
	}
	if o.Query != "" {
		conds = append(conds, sqlf.Sprintf(extensionIDExpr+" ILIKE %s", "%"+strings.Replace(strings.ToLower(o.Query), " ", "%", -1)+"%"))
	}
	if len(conds) == 0 {
		conds = append(conds, sqlf.Sprintf("TRUE"))
	}
	return conds
}

func (o RegistryExtensionsListOptions) sqlOrder() []*sqlf.Query {
	ids := make([]*sqlf.Query, len(o.PrioritizeExtensionIDs)+1)
	for i, id := range o.PrioritizeExtensionIDs {
		ids[i] = sqlf.Sprintf("%v", string(id))
	}
	ids[len(o.PrioritizeExtensionIDs)] = sqlf.Sprintf("NULL")
	return []*sqlf.Query{sqlf.Sprintf(extensionIDExpr+` IN (%v) ASC`, sqlf.Join(ids, ","))}
}

// List lists all registry extensions that satisfy the options.
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to list with the specified
// options.
func (s *registryExtensions) List(ctx context.Context, opt RegistryExtensionsListOptions) ([]*RegistryExtension, error) {
	return s.list(ctx, opt.sqlConditions(), opt.sqlOrder(), opt.LimitOffset)
}

func (registryExtensions) listCountSQL(conds []*sqlf.Query) *sqlf.Query {
	return sqlf.Sprintf(`
FROM registry_extensions x
LEFT JOIN users ON users.id=publisher_user_id AND users.deleted_at IS NULL
LEFT JOIN orgs ON orgs.id=publisher_org_id AND orgs.deleted_at IS NULL
WHERE (%s) AND x.deleted_at IS NULL`,
		sqlf.Join(conds, ") AND ("))
}

func (s *registryExtensions) list(ctx context.Context, conds, order []*sqlf.Query, limitOffset *LimitOffset) ([]*RegistryExtension, error) {
	order = append(order, sqlf.Sprintf("TRUE"))
	q := sqlf.Sprintf(`
SELECT x.id, x.uuid, x.publisher_user_id, x.publisher_org_id, x.name, x.created_at, x.updated_at,
  `+extensionIDExpr+` AS non_canonical_extension_id, `+extensionPublisherNameExpr+` AS non_canonical_publisher_name
%s
ORDER BY %s, x.id ASC
%s`,
		s.listCountSQL(conds),
		sqlf.Join(order, ","),
		limitOffset.SQL(),
	)

	rows, err := dbconn.Global.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*RegistryExtension
	for rows.Next() {
		var t RegistryExtension
		var publisherUserID, publisherOrgID sql.NullInt64
		if err := rows.Scan(&t.ID, &t.UUID, &publisherUserID, &publisherOrgID, &t.Name, &t.CreatedAt, &t.UpdatedAt, &t.NonCanonicalExtensionID, &t.Publisher.NonCanonicalName); err != nil {
			return nil, err
		}
		t.Publisher.UserID = int32(publisherUserID.Int64)
		t.Publisher.OrgID = int32(publisherOrgID.Int64)
		results = append(results, &t)
	}
	return results, nil
}

// Count counts all registry extensions that satisfy the options (ignoring limit and offset).
//
// 🚨 SECURITY: The caller must ensure that the actor is permitted to count the results.
func (s *registryExtensions) Count(ctx context.Context, opt RegistryExtensionsListOptions) (int, error) {
	q := sqlf.Sprintf("SELECT COUNT(*) %s", s.listCountSQL(opt.sqlConditions()))
	var count int
	if err := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// Update updates information about the registry extension.
func (*registryExtensions) Update(ctx context.Context, id int32, name *string) error {
	if Mocks.RegistryExtensions.Update != nil {
		return Mocks.RegistryExtensions.Update(id, name)
	}

	res, err := dbconn.Global.ExecContext(ctx,
		"UPDATE registry_extensions SET name=COALESCE($2, name),  updated_at=now() WHERE id=$1 AND deleted_at IS NULL",
		id, name,
	)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return RegistryExtensionNotFoundError{[]interface{}{id}}
	}
	return nil
}

// Delete marks an registry extension as deleted.
func (*registryExtensions) Delete(ctx context.Context, id int32) error {
	if Mocks.RegistryExtensions.Delete != nil {
		return Mocks.RegistryExtensions.Delete(id)
	}

	res, err := dbconn.Global.ExecContext(ctx, "UPDATE registry_extensions SET deleted_at=now() WHERE id=$1 AND deleted_at IS NULL", id)
	if err != nil {
		return err
	}
	nrows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if nrows == 0 {
		return RegistryExtensionNotFoundError{[]interface{}{id}}
	}
	return nil
}

// MockRegistryExtensions mocks the registry extensions store.
type MockRegistryExtensions struct {
	Create           func(publisherUserID, publisherOrgID int32, name string) (int32, error)
	GetByID          func(id int32) (*RegistryExtension, error)
	GetByUUID        func(uuid string) (*RegistryExtension, error)
	GetByExtensionID func(extensionID string) (*RegistryExtension, error)
	Update           func(id int32, name *string) error
	Delete           func(id int32) error
}
