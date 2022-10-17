package authz

import (
	"fmt"
	"sort"
	"time"

	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var ErrPermsNotFound = errors.New("permissions not found")

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

// PermRepos is the list of available user permission types.
const (
	PermRepos PermType = "repos"
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

// ErrStalePermissions is returned by LoadPermissions when the stored
// permissions are stale (e.g. the first time a user needs them and they haven't
// been fetched yet). Callers should pass this error up to the user and show a
// more friendly prompt message in the UI.
type ErrStalePermissions struct {
	UserID int32
	Perm   Perms
	Type   PermType
}

// Error implements the error interface.
func (e ErrStalePermissions) Error() string {
	return fmt.Sprintf("%s:%s permissions for user=%d are stale and being updated", e.Perm, e.Type, e.UserID)
}

// UserPermissions are the permissions of a user to perform an action
// on the given set of object IDs of the defined type.
type UserPermissions struct {
	UserID    int32              // The internal database ID of a user
	Perm      Perms              // The permission set
	Type      PermType           // The type of the permissions
	IDs       map[int32]struct{} // The object IDs
	UpdatedAt time.Time          // The last updated time
	SyncedAt  time.Time          // The last user-centric synced time
}

// Expired returns true if these UserPermissions have elapsed the given ttl.
func (p *UserPermissions) Expired(ttl time.Duration, now time.Time) bool {
	return !now.Before(p.UpdatedAt.Add(ttl))
}

// AuthorizedRepos returns the intersection of the given repository IDs with
// the authorized IDs.
func (p *UserPermissions) AuthorizedRepos(repos []*types.Repo) []RepoPerms {
	// Return directly if it's used for wrong permissions type or no permissions available.
	if p.Type != PermRepos ||
		p.IDs == nil || len(p.IDs) == 0 {
		return []RepoPerms{}
	}

	perms := make([]RepoPerms, 0, len(repos))
	for _, r := range repos {
		_, ok := p.IDs[int32(r.ID)]
		if r.ID != 0 && ok {
			perms = append(perms, RepoPerms{Repo: r, Perms: p.Perm})
		}
	}
	return perms
}

// GenerateSortedIDsSlice returns a sorted slice of the IDs set.
func (p *UserPermissions) GenerateSortedIDsSlice() []int32 {
	return convertMapSetToSortedSlice(p.IDs)
}

// TracingFields returns tracing fields for the opentracing log.
func (p *UserPermissions) TracingFields() []otlog.Field {
	fs := []otlog.Field{
		otlog.Int32("UserPermissions.UserID", p.UserID),
		trace.Stringer("UserPermissions.Perm", p.Perm),
		otlog.String("UserPermissions.Type", string(p.Type)),
	}

	if p.IDs != nil {
		fs = append(fs,
			otlog.Int("UserPermissions.IDs.Count", len(p.IDs)),
			otlog.String("UserPermissions.UpdatedAt", p.UpdatedAt.String()),
			otlog.String("UserPermissions.SyncedAt", p.SyncedAt.String()),
		)
	}

	return fs
}

// RepoPermissions declares which users have access to a given repository
type RepoPermissions struct {
	RepoID         int32              // The internal database ID of a repository
	Perm           Perms              // The permission set
	UserIDs        map[int32]struct{} // The user IDs
	PendingUserIDs map[int64]struct{} // The pending user IDs
	UpdatedAt      time.Time          // The last updated time
	SyncedAt       time.Time          // The last repo-centric synced time
	Unrestricted   bool               // Anyone can see the repo, overrides all other permissions
}

// Expired returns true if these RepoPermissions have elapsed the given ttl.
func (p *RepoPermissions) Expired(ttl time.Duration, now time.Time) bool {
	return !now.Before(p.UpdatedAt.Add(ttl))
}

// GenerateSortedIDsSlice returns a sorted slice of the IDs set.
func (p *RepoPermissions) GenerateSortedIDsSlice() []int32 {
	return convertMapSetToSortedSlice(p.UserIDs)
}

// TracingFields returns tracing fields for the opentracing log.
func (p *RepoPermissions) TracingFields() []otlog.Field {
	fs := []otlog.Field{
		otlog.Int32("RepoPermissions.RepoID", p.RepoID),
		trace.Stringer("RepoPermissions.Perm", p.Perm),
	}

	if p.UserIDs != nil {
		fs = append(fs,
			otlog.Int("RepoPermissions.UserIDs.Count", len(p.UserIDs)),
			otlog.Int("RepoPermissions.PendingUserIDs.Count", len(p.PendingUserIDs)),
			otlog.String("RepoPermissions.UpdatedAt", p.UpdatedAt.String()),
			otlog.String("RepoPermissions.SyncedAt", p.SyncedAt.String()),
		)
	}

	return fs
}

// UserPendingPermissions defines permissions that a not-yet-created user has to
// perform on a given set of object IDs. Not-yet-created users may exist on the
// code host but not yet in Sourcegraph. "ServiceType", "ServiceID" and "BindID"
// are used to map this stub user to an actual user when the user is created.
type UserPendingPermissions struct {
	// The auto-generated internal database ID.
	ID int64
	// The type of the code host as if it would be used as extsvc.AccountSpec.ServiceType,
	// e.g. "github", "gitlab", "bitbucketServer" and "sourcegraph".
	ServiceType string
	// The ID of the code host as if it would be used as extsvc.AccountSpec.ServiceID,
	// e.g. "https://github.com/", "https://gitlab.com/" and "https://sourcegraph.com/".
	ServiceID string
	// The account ID that a code host (and its authz provider) uses to identify a user,
	// e.g. a username (for Bitbucket Server), a GraphID ( for GitHub), or a user ID
	// (for GitLab).
	//
	// When use the Sourcegraph authz provider, "BindID" can be either a username or
	// an email based on site configuration.
	BindID string
	// The permissions this user has to the "IDs" of the "Type".
	Perm Perms
	// The type of permissions this user has.
	Type PermType
	// The object IDs with the "Type".
	IDs map[int32]struct{}
	// The last updated time.
	UpdatedAt time.Time
}

// GenerateSortedIDsSlice returns a sorted slice of the IDs set.
func (p *UserPendingPermissions) GenerateSortedIDsSlice() []int32 {
	return convertMapSetToSortedSlice(p.IDs)
}

// TracingFields returns tracing fields for the opentracing log.
func (p *UserPendingPermissions) TracingFields() []otlog.Field {
	fs := []otlog.Field{
		otlog.Int64("UserPendingPermissions.ID", p.ID),
		otlog.String("UserPendingPermissions.ServiceType", p.ServiceType),
		otlog.String("UserPendingPermissions.ServiceID", p.ServiceID),
		otlog.String("UserPendingPermissions.BindID", p.BindID),
		trace.Stringer("UserPendingPermissions.Perm", p.Perm),
		otlog.String("UserPendingPermissions.Type", string(p.Type)),
	}

	if p.IDs != nil {
		fs = append(fs,
			otlog.Int("UserPendingPermissions.IDs.Count", len(p.IDs)),
			otlog.String("UserPendingPermissions.UpdatedAt", p.UpdatedAt.String()),
		)
	}

	return fs
}

// convertMapSetToSortedSlice converts a map set into a slice of sorted integers
func convertMapSetToSortedSlice(mapSet map[int32]struct{}) []int32 {
	returnSlice := make([]int32, 0, len(mapSet))
	for id := range mapSet {
		returnSlice = append(returnSlice, id)
	}
	sort.Slice(returnSlice, func(i, j int) bool { return returnSlice[i] < returnSlice[j] })
	return returnSlice
}
