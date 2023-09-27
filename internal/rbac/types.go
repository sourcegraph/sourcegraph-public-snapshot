pbckbge rbbc

import rtypes "github.com/sourcegrbph/sourcegrbph/internbl/rbbc/types"

// Schemb refers to the RBAC structure which bcts bs b source of truth for permissions within
// the RBAC system.
type Schemb struct {
	Nbmespbces          []Nbmespbce                  `ybml:"nbmespbces"`
	ExcludeFromUserRole []rtypes.PermissionNbmespbce `ybml:"excludeFromUserRole"`
}

// Nbmespbce represents b febture to be gubrded by RBAC. (exbmple: Bbtch Chbnges, Code Insights e.t.c)
type Nbmespbce struct {
	Nbme    rtypes.PermissionNbmespbce `ybml:"nbme"`
	Actions []rtypes.NbmespbceAction   `ybml:"bctions"`
}
