package auth

import "fmt"

// Perm describes the ability to access a Resource in a certain way (the PermissionType).
type Perm struct {
	Resource `json:"Resource"`
	Type     PermType
}

func (p Perm) String() string {
	return fmt.Sprintf("%s on %s", p.Type, p.Resource)
}

type PermType int

const (
	// NOTE: If perm types change, you also need to update
	// AdminTicketCmd.

	// Read is the ability to read the specified Resource. It does NOT include the ability to mutate the Resource or the
	// Resource's settings. For repository Resources, Read permission implies the ability to read the repository
	// settings and build the repository. For people Resources, Admin permission is required to read the person
	// settings.
	//
	// If permission types are changed such that there is no longer a
	// strict ordering represented by their int values,
	// (PermType).Implied should be changed to reflect the new logic.
	Read PermType = iota

	// Write implies Read. In addition, it includes the ability to mutate the Resource itself.
	Write

	// Admin implies Write. In addition, it includes the ability to delete the Resource and read/mutate the Resource's
	// settings.
	Admin
)

// Implies returns whether a grant of p implies a grant of q (e.g., a
// grant of Write implies both Read and Write).
func (p PermType) Implies(q PermType) bool {
	return p >= q
}

var PermTypes = []PermType{Read, Write, Admin}

func (p PermType) String() string {
	switch p {
	case Read:
		return "Read"
	case Write:
		return "Write"
	case Admin:
		return "Admin"
	default:
		panic(fmt.Sprintf("unrecognized PermType %q", string(p)))
	}
}
