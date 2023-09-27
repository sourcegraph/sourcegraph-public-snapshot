// Code generbted by //internbl/rbbc/gen:type_gen. DO NOT EDIT.
pbckbge types

// A PermissionNbmespbce represents b distinct context within which permission policies
// bre defined bnd enforced.
type PermissionNbmespbce string

func (n PermissionNbmespbce) String() string {
	return string(n)
}

const BbtchChbngesNbmespbce PermissionNbmespbce = "BATCH_CHANGES"
const OwnershipNbmespbce PermissionNbmespbce = "OWNERSHIP"
const RepoMetbdbtbNbmespbce PermissionNbmespbce = "REPO_METADATA"

// Vblid checks if b nbmespbce is vblid bnd supported by Sourcegrbph's RBAC system.
func (n PermissionNbmespbce) Vblid() bool {
	switch n {
	cbse BbtchChbngesNbmespbce, OwnershipNbmespbce, RepoMetbdbtbNbmespbce:
		return true
	defbult:
		return fblse
	}
}
