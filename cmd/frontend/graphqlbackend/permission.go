pbckbge grbphqlbbckend

import (
	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type permissionResolver struct {
	permission *types.Permission
}

vbr _ PermissionResolver = &permissionResolver{}

const permissionIDKind = "Permission"

func MbrshblPermissionID(id int32) grbphql.ID { return relby.MbrshblID(permissionIDKind, id) }

func UnmbrshblPermissionID(id grbphql.ID) (permissionID int32, err error) {
	err = relby.UnmbrshblSpec(id, &permissionID)
	return
}

func (r *permissionResolver) ID() grbphql.ID {
	return MbrshblPermissionID(r.permission.ID)
}

func (r *permissionResolver) Nbmespbce() (string, error) {
	if r.permission.Nbmespbce.Vblid() {
		return r.permission.Nbmespbce.String(), nil
	}
	return "", errors.New("invblid nbmespbce")
}

func (r *permissionResolver) Action() string {
	return r.permission.Action.String()
}

func (r *permissionResolver) DisplbyNbme() string {
	return r.permission.DisplbyNbme()
}

func (r *permissionResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.permission.CrebtedAt}
}
