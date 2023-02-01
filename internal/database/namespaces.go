package database

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// A Namespace is a username or an organization name. No user may have a username that is equal to
// an organization name, and vice versa. This property means that a username or organization name
// serves as a namespace for other objects that are owned by the user or organization, such as
// batch changes and extensions.
type Namespace struct {
	// Name is the canonical-case name of the namespace (which is unique among all namespace
	// types). For a user, this is the username. For an organization, this is the organization name.
	Name string

	User, Organization int32 // exactly 1 is non-zero
}

var (
	ErrNamespaceMultipleIDs = errors.New("multiple namespace IDs provided")
	ErrNamespaceNoID        = errors.New("no namespace ID provided")
	ErrNamespaceNotFound    = errors.New("namespace not found")
)

type NamespaceStore interface {
	basestore.ShareableStore
	With(other basestore.ShareableStore) NamespaceStore
	WithTransact(context.Context, func(NamespaceStore) error) error
	GetByID(ctx context.Context, orgID, userID int32) (*Namespace, error)
	GetByName(ctx context.Context, name string) (*Namespace, error)
}

type namespaceStore struct {
	*basestore.Store
}

// NamespacesWith instantiates and returns a new NamespaceStore using the other store handle.
func NamespacesWith(other basestore.ShareableStore) NamespaceStore {
	return &namespaceStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *namespaceStore) With(other basestore.ShareableStore) NamespaceStore {
	return &namespaceStore{Store: s.Store.With(other)}
}

func (s *namespaceStore) WithTransact(ctx context.Context, f func(NamespaceStore) error) error {
	return s.Store.WithTransact(ctx, func(tx *basestore.Store) error {
		return f(&namespaceStore{Store: tx})
	})
}

// GetByID looks up the namespace by an ID.
//
// One of orgID and userID must be 0: whichever ID is non-zero will be used to
// look up the namespace. If both are given, ErrNamespaceMultipleIDs is
// returned; if neither are given, ErrNamespaceNoID is returned.
//
// If no namespace is found, ErrNamespaceNotFound is returned.
func (s *namespaceStore) GetByID(
	ctx context.Context,
	orgID, userID int32,
) (*Namespace, error) {
	preds := []*sqlf.Query{}
	if orgID != 0 && userID != 0 {
		return nil, ErrNamespaceMultipleIDs
	} else if orgID != 0 {
		preds = append(preds, sqlf.Sprintf("org_id = %s", orgID))
	} else if userID != 0 {
		preds = append(preds, sqlf.Sprintf("user_id = %s", userID))
	} else {
		return nil, ErrNamespaceNoID
	}

	var n Namespace
	if err := s.getNamespace(ctx, &n, preds); err != nil {
		return nil, err
	}
	return &n, nil
}

// GetByName looks up the namespace by a name. The name is matched
// case-insensitively against all namespaces, which is the set of usernames and
// organization names.
//
// If no namespace is found, ErrNamespaceNotFound is returned.
func (s *namespaceStore) GetByName(
	ctx context.Context,
	name string,
) (*Namespace, error) {
	var n Namespace
	if err := s.getNamespace(ctx, &n, []*sqlf.Query{
		sqlf.Sprintf("name = %s", name),
	}); err != nil {
		return nil, err
	}
	return &n, nil
}

func (s *namespaceStore) getNamespace(ctx context.Context, n *Namespace, preds []*sqlf.Query) error {
	q := getNamespaceQuery(preds)
	err := s.QueryRow(
		ctx,
		q,
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
SELECT
	name,
	COALESCE(user_id, 0) AS user_id,
	COALESCE(org_id, 0) AS org_id
FROM
	names
WHERE %s
`
