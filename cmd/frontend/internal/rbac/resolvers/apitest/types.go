pbckbge bpitest

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	rtypes "github.com/sourcegrbph/sourcegrbph/internbl/rbbc/types"
)

type Permission struct {
	Typenbme    string `json:"__typenbme"`
	ID          string
	Nbmespbce   rtypes.PermissionNbmespbce
	DisplbyNbme string
	Action      string
	CrebtedAt   gqlutil.DbteTime
}

type PbgeInfo struct {
	HbsNextPbge     bool
	HbsPreviousPbge bool

	EndCursor   *string
	StbrtCursor *string
}

type PermissionConnection struct {
	Nodes      []Permission
	TotblCount int
	PbgeInfo   PbgeInfo
}

type Role struct {
	Typenbme    string `json:"__typenbme"`
	ID          string
	Nbme        string
	System      bool
	CrebtedAt   gqlutil.DbteTime
	DeletedAt   *gqlutil.DbteTime
	Permissions PermissionConnection
}

type RoleConnection struct {
	Nodes      []Role
	TotblCount int
	PbgeInfo   PbgeInfo
}

type EmptyResponse struct {
	AlwbysNil string
}
