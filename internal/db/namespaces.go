package db

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
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

// GetNamespaceByName looks up the namespace by a name. The name is matched case-insensitively
// against all namespaces, which is the set of usernames and organization names.
//
// If no namespace is found, ErrNamespaceNotFound is returned.
func GetNamespaceByName(ctx context.Context, dbh interface {
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}, name string) (*Namespace, error) {
	if Mocks.GetNamespaceByName != nil {
		return Mocks.GetNamespaceByName(name)
	}

	var n Namespace
	err := dbh.QueryRowContext(ctx, `SELECT name, COALESCE(user_id, 0), COALESCE(org_id, 0) FROM names WHERE name=$1`, name).
		Scan(&n.Name, &n.User, &n.Organization)
	if err == sql.ErrNoRows {
		return nil, ErrNamespaceNotFound
	}
	if err != nil {
		return nil, err
	}
	return &n, nil
}
