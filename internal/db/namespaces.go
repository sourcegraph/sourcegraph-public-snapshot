package db

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"

	"github.com/keegancsmith/sqlf"
)

// A Namespace is a username or an organization name. No user may have a username that is equal to
// an organization name, and vice versa. This property means that a username or organization name
// serves as a namespace for other objects that are owned by the user or organization, such as
// campaigns and extensions.
type Namespace struct {
	// Name is the canonical-case name of the namespace (which is unique among all namespace
	// types). For a user, this is the username. For an organization, this is the organization name.
	Name string

	User, Organization int32 // exactly 1 is non-zero
}

var ErrNamespaceNotFound = errors.New("namespace not found")

type namespaces struct{}

// GetByName looks up the namespace by a name. The name is matched
// case-insensitively against all namespaces, which is the set of usernames and
// organization names.
//
// If no namespace is found, ErrNamespaceNotFound is returned.
func (*namespaces) GetByName(
	ctx context.Context,
	name string,
) (*Namespace, error) {
	if Mocks.Namespaces.GetByName != nil {
		return Mocks.Namespaces.GetByName(ctx, name)
	}

	var n Namespace
	if err := getNamespace(ctx, &n, []*sqlf.Query{
		sqlf.Sprintf("name = %s", name),
	}); err != nil {
		return nil, err
	}
	return &n, nil
}

func getNamespace(ctx context.Context, n *Namespace, preds []*sqlf.Query) error {
	q := getNamespaceQuery(preds)
	err := dbconn.Global.QueryRowContext(
		ctx,
		q.Query(sqlf.PostgresBindVar),
		q.Args()...,
	).Scan(&n.Name, &n.User, &n.Organization)

	if err == sql.ErrNoRows {
		return ErrNamespaceNotFound
	}
	return err
}

func getNamespaceQuery(preds []*sqlf.Query) *sqlf.Query {
	return sqlf.Sprintf(namespaceQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

var namespaceQueryFmtstr = `
-- source: internal/db/namespaces.go:getNamespace
SELECT
	name,
	COALESCE(user_id, 0) AS user_id,
	COALESCE(org_id, 0) AS org_id
FROM
	names
WHERE %s
`
