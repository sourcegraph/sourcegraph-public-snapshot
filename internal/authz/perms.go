package authz

import (
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/collections"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var ErrPermsNotFound = errors.New("permissions not found")

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

// Permission determines if a user with a specific id
// can read a repository with a specific id
type Permission struct {
	UserID            int32       // The internal database ID of a user
	ExternalAccountID int32       // The internal database ID of a user external account
	RepoID            int32       // The internal database ID of a repo
	CreatedAt         time.Time   // The creation time
	UpdatedAt         time.Time   // The last updated time
	Source            PermsSource // source of the permission
}

// A struct that holds the entity we are updating the permissions for
// It can be either a user or a repository.
type PermissionEntity struct {
	UserID            int32 // The internal database ID of a user
	ExternalAccountID int32 // The internal database ID of a user external account
	RepoID            int32 // The internal database ID of a repo
}

type UserIDWithExternalAccountID struct {
	UserID            int32
	ExternalAccountID int32
}

type PermsSource string

const (
	SourceRepoSync PermsSource = "repo_sync"
	SourceUserSync PermsSource = "user_sync"
	SourceAPI      PermsSource = "api"
)

func (s PermsSource) ToGraphQL() string { return strings.ToUpper(string(s)) }

func (p *Permission) Attrs() []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.Int("SrcPermissions.UserID", int(p.UserID)),
		attribute.Int("SrcPermissions.RepoID", int(p.RepoID)),
		attribute.Int("SrcPermissions.ExternalAccountID", int(p.ExternalAccountID)),
		attribute.Stringer("SrcPermissions.CreatedAt", p.CreatedAt),
		attribute.Stringer("SrcPermissions.UpdatedAt", p.UpdatedAt),
		attribute.String("SrcPermissions.Source", string(p.Source)),
	}
}

// RepoPermissions declares which users have access to a given repository
type RepoPermissions struct {
	RepoID         int32                  // The internal database ID of a repository
	Perm           Perms                  // The permission set
	UserIDs        collections.Set[int32] // The user IDs
	PendingUserIDs collections.Set[int64] // The pending user IDs
	UpdatedAt      time.Time              // The last updated time
	SyncedAt       time.Time              // The last repo-centric synced time
	Unrestricted   bool                   // Anyone can see the repo, overrides all other permissions
}

func (p *RepoPermissions) Attrs() []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.Int("RepoPermissions.RepoID", int(p.RepoID)),
		attribute.Stringer("RepoPermissions.Perm", p.Perm),
	}

	if p.UserIDs != nil {
		attrs = append(attrs,
			attribute.Int("RepoPermissions.UserIDs.Count", len(p.UserIDs)),
			attribute.Int("RepoPermissions.PendingUserIDs.Count", len(p.PendingUserIDs)),
			attribute.Stringer("RepoPermissions.UpdatedAt", p.UpdatedAt),
			attribute.Stringer("RepoPermissions.SyncedAt", p.SyncedAt),
		)
	}

	return attrs
}

// UserGrantPermissions defines the structure to grant pending permissions to a user.
// See also UserPendingPermissions.
type UserGrantPermissions struct {
	// UserID of the user to grant permissions to.
	UserID int32
	// ID of the user external account that the permissions are from.
	UserExternalAccountID int32
	// The type of the code host as if it would be used as extsvc.AccountSpec.ServiceType
	ServiceType string
	// The ID of the code host as if it would be used as extsvc.AccountSpec.ServiceID
	ServiceID string
	// The account ID of the user external account, that the permissions are from
	AccountID string
}

func (p *UserGrantPermissions) Attrs() []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.Int("UserGrantPermissions.UserID", int(p.UserID)),
		attribute.Int("UserGrantPermissions.UserExternalAccountID", int(p.UserExternalAccountID)),
		attribute.String("UserPendingPermissions.ServiceType", p.ServiceType),
		attribute.String("UserPendingPermissions.ServiceID", p.ServiceID),
		attribute.String("UserPendingPermissions.AccountID", p.AccountID),
	}

	return attrs
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
	IDs collections.Set[int32]
	// The last updated time.
	UpdatedAt time.Time
}

// GenerateSortedIDsSlice returns a sorted slice of the IDs set.
func (p *UserPendingPermissions) GenerateSortedIDsSlice() []int32 {
	return collections.SortedSetValues(p.IDs)
}

func (p *UserPendingPermissions) Attrs() []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.Int64("UserPendingPermissions.ID", p.ID),
		attribute.String("UserPendingPermissions.ServiceType", p.ServiceType),
		attribute.String("UserPendingPermissions.ServiceID", p.ServiceID),
		attribute.String("UserPendingPermissions.BindID", p.BindID),
		attribute.Stringer("UserPendingPermissions.Perm", p.Perm),
		attribute.String("UserPendingPermissions.Type", string(p.Type)),
	}

	if p.IDs != nil {
		attrs = append(attrs,
			attribute.Int("UserPendingPermissions.IDs.Count", len(p.IDs)),
			attribute.Stringer("UserPendingPermissions.UpdatedAt", p.UpdatedAt),
		)
	}

	return attrs
}
