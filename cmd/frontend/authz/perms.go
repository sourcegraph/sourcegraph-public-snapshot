package authz

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

// RepoPerms contains a repo and the permissions a given user
// has associated with it.
type RepoPerms struct {
	Repo  *types.Repo
	Perms Perms
}

// Perms is a permission set represented as bitset.
type Perms uint32

// Perm constants.
const (
	Read Perms = 1 << (32 - 1 - iota)
	Write
	None Perms = 0
)

// Include is a convenience method to test if Perms
// includes all the other Perms.
func (p Perms) Include(other Perms) bool {
	return other != None && p&other == other
}

// String implements the fmt.Stringer interface.
func (p Perms) String() string {
	var sb strings.Builder

	for mask := Read; mask != 0; mask >>= 1 {
		switch p & mask {
		case Read:
			sb.WriteString("read,")
		case Write:
			sb.WriteString("write,")
		}
	}

	if hi := sb.Len(); hi > 0 {
		return sb.String()[:hi-1]
	}

	return "none"
}
