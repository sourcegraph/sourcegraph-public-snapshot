package pgsql

import (
	"github.com/sqs/modl"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/util/dbutil"
)

// userPermission specifies the user's permissions on the client.
type userPermission struct {
	UID      int32
	ClientID string `db:"client_id"`
	Read     bool
	Write    bool
	Admin    bool
}

func init() {
	Schema.Map.AddTableWithName(userPermission{}, "user_permissions").SetKeys(false, "uid", "client_id")
	Schema.CreateSQL = append(Schema.CreateSQL,
		`CREATE INDEX user_permissions_client_id ON user_permissions(client_id);`,
	)
}

// UserPermissions is a DB-backed implementation of the UserPermissions store.
type UserPermissions struct{}

var _ store.UserPermissions = (*UserPermissions)(nil)

func (s *UserPermissions) Get(ctx context.Context, opt *sourcegraph.UserPermissionsOptions) (*sourcegraph.UserPermissions, error) {
	var toks []*userPermission
	err := dbh(ctx).Select(&toks, `SELECT * FROM user_permissions WHERE "uid"=$1 AND "client_id"=$2`, opt.UID, opt.ClientSpec.ID)
	if err != nil {
		return nil, err
	}
	if len(toks) == 0 {
		// no permissions for the given user
		return &sourcegraph.UserPermissions{UID: opt.UID, ClientID: opt.ClientSpec.ID}, nil
	}
	return &sourcegraph.UserPermissions{
		UID:      toks[0].UID,
		ClientID: toks[0].ClientID,
		Read:     toks[0].Read,
		Write:    toks[0].Write,
		Admin:    toks[0].Admin,
	}, nil
}

func (s *UserPermissions) Verify(ctx context.Context, perms *sourcegraph.UserPermissions) (bool, error) {
	dbPerms, err := s.Get(ctx, &sourcegraph.UserPermissionsOptions{
		UID:        perms.UID,
		ClientSpec: &sourcegraph.RegisteredClientSpec{ID: perms.ClientID},
	})
	if err != nil {
		return false, err
	}
	if (perms.Admin && !dbPerms.Admin) ||
		(perms.Write && !dbPerms.Write) ||
		(perms.Read && !dbPerms.Read) {
		return false, nil
	}
	return true, nil
}

func (s *UserPermissions) Set(ctx context.Context, perms *sourcegraph.UserPermissions) error {
	newUser := &userPermission{
		UID:      perms.UID,
		ClientID: perms.ClientID,
		Read:     perms.Read,
		Write:    perms.Write,
		Admin:    perms.Admin,
	}
	return dbutil.Transact(dbh(ctx), func(tx modl.SqlExecutor) error {
		ctx = NewContext(ctx, tx)

		var toks []*userPermission
		err := dbh(ctx).Select(&toks, `SELECT * FROM user_permissions WHERE "uid"=$1 AND "client_id"=$2`, perms.UID, perms.ClientID)
		if err != nil {
			return err
		}
		if len(toks) == 0 {
			return tx.Insert(newUser)
		}
		// user record for this client exists.
		_, err = tx.Update(newUser)
		return err
	})
}

func (s *UserPermissions) List(ctx context.Context, client *sourcegraph.RegisteredClientSpec) (*sourcegraph.UserPermissionsList, error) {
	var toks []*userPermission
	err := dbh(ctx).Select(&toks, `SELECT * FROM user_permissions WHERE "client_id"=$1`, client.ID)
	if err != nil {
		return nil, err
	}
	permsList := &sourcegraph.UserPermissionsList{}
	for _, t := range toks {
		permsList.UserPermissions = append(permsList.UserPermissions, &sourcegraph.UserPermissions{
			UID:      t.UID,
			ClientID: t.ClientID,
			Read:     t.Read,
			Write:    t.Write,
			Admin:    t.Admin,
		})
	}
	return permsList, nil
}
