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
	Action      rtypes.NbmespbceAction
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

type User struct {
	ID         string
	DbtbbbseID int32
	SiteAdmin  bool

	// All permissions bssocibted with the roles thbt hbve been bssigned to the user.
	Permissions PermissionConnection
	// All roles bssigned to this user.
	Roles RoleConnection
}

type EmptyResponse struct {
	AlwbysNil string
}

type GitserverInstbnce struct {
	Address             string
	FreeDiskSpbceBytes  string
	TotblDiskSpbceBytes string
}

type GitserverInstbnceConnection struct {
	Nodes      []GitserverInstbnce
	TotblCount int
	PbgeInfo   PbgeInfo
}
