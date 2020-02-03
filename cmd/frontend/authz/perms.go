package authz

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
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
	None Perms = 0
	Read Perms = 1 << iota
	Write
)

// Include is a convenience method to test if Perms
// includes all the other Perms.
func (p Perms) Include(other Perms) bool {
	return p&other == other
}

// String implements the fmt.Stringer interface.
func (p Perms) String() string {
	switch p {
	case Read:
		return "read"
	case Write:
		return "write"
	case Read | Write:
		return "read,write"
	default:
		return "none"
	}
}

// PermType is the object type of the user permissions.
type PermType string

// The list of available user permission types.
const (
	PermRepos PermType = "repos"
)

// ProviderType is the type of provider implementation for the permissions.
type ProviderType string

// The list of available provider types.
const (
	ProviderBitbucketServer ProviderType = bitbucketserver.ServiceType
	ProviderSourcegraph     ProviderType = "sourcegraph"
)

// RepoPermsSort sorts a slice of RepoPerms to guarantee a stable ordering.
type RepoPermsSort []RepoPerms

func (s RepoPermsSort) Len() int      { return len(s) }
func (s RepoPermsSort) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s RepoPermsSort) Less(i, j int) bool {
	if s[i].Repo.ID != s[j].Repo.ID {
		return s[i].Repo.ID < s[j].Repo.ID
	}
	if s[i].Repo.ExternalRepo.ID != s[j].Repo.ExternalRepo.ID {
		return s[i].Repo.ExternalRepo.ID < s[j].Repo.ExternalRepo.ID
	}
	return s[i].Repo.Name < s[j].Repo.Name
}
