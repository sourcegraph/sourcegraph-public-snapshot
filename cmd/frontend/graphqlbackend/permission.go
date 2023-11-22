package graphqlbackend

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type permissionResolver struct {
	permission *types.Permission
}

var _ PermissionResolver = &permissionResolver{}

const permissionIDKind = "Permission"

func MarshalPermissionID(id int32) graphql.ID { return relay.MarshalID(permissionIDKind, id) }

func UnmarshalPermissionID(id graphql.ID) (permissionID int32, err error) {
	err = relay.UnmarshalSpec(id, &permissionID)
	return
}

func (r *permissionResolver) ID() graphql.ID {
	return MarshalPermissionID(r.permission.ID)
}

func (r *permissionResolver) Namespace() (string, error) {
	if r.permission.Namespace.Valid() {
		return r.permission.Namespace.String(), nil
	}
	return "", errors.New("invalid namespace")
}

func (r *permissionResolver) Action() string {
	return r.permission.Action.String()
}

func (r *permissionResolver) DisplayName() string {
	return r.permission.DisplayName()
}

func (r *permissionResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.permission.CreatedAt}
}
