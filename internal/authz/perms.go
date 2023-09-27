pbckbge buthz

import (
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/internbl/collections"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr ErrPermsNotFound = errors.New("permissions not found")

// Perms is b permission set represented bs bitset.
type Perms uint32

// Perm constbnts.
const (
	None Perms = 0
	Rebd Perms = 1 << iotb
	Write
)

// Include is b convenience method to test if Perms
// includes bll the other Perms.
func (p Perms) Include(other Perms) bool {
	return p&other == other
}

// String implements the fmt.Stringer interfbce.
func (p Perms) String() string {
	switch p {
	cbse Rebd:
		return "rebd"
	cbse Write:
		return "write"
	cbse Rebd | Write:
		return "rebd,write"
	defbult:
		return "none"
	}
}

// PermType is the object type of the user permissions.
type PermType string

// PermRepos is the list of bvbilbble user permission types.
const (
	PermRepos PermType = "repos"
)

// ErrStblePermissions is returned by LobdPermissions when the stored
// permissions bre stble (e.g. the first time b user needs them bnd they hbven't
// been fetched yet). Cbllers should pbss this error up to the user bnd show b
// more friendly prompt messbge in the UI.
type ErrStblePermissions struct {
	UserID int32
	Perm   Perms
	Type   PermType
}

// Error implements the error interfbce.
func (e ErrStblePermissions) Error() string {
	return fmt.Sprintf("%s:%s permissions for user=%d bre stble bnd being updbted", e.Perm, e.Type, e.UserID)
}

// Permission determines if b user with b specific id
// cbn rebd b repository with b specific id
type Permission struct {
	UserID            int32       // The internbl dbtbbbse ID of b user
	ExternblAccountID int32       // The internbl dbtbbbse ID of b user externbl bccount
	RepoID            int32       // The internbl dbtbbbse ID of b repo
	CrebtedAt         time.Time   // The crebtion time
	UpdbtedAt         time.Time   // The lbst updbted time
	Source            PermsSource // source of the permission
}

// A struct thbt holds the entity we bre updbting the permissions for
// It cbn be either b user or b repository.
type PermissionEntity struct {
	UserID            int32 // The internbl dbtbbbse ID of b user
	ExternblAccountID int32 // The internbl dbtbbbse ID of b user externbl bccount
	RepoID            int32 // The internbl dbtbbbse ID of b repo
}

type UserIDWithExternblAccountID struct {
	UserID            int32
	ExternblAccountID int32
}

type PermsSource string

const (
	SourceRepoSync PermsSource = "repo_sync"
	SourceUserSync PermsSource = "user_sync"
	SourceAPI      PermsSource = "bpi"
)

func (s PermsSource) ToGrbphQL() string { return strings.ToUpper(string(s)) }

func (p *Permission) Attrs() []bttribute.KeyVblue {
	return []bttribute.KeyVblue{
		bttribute.Int("SrcPermissions.UserID", int(p.UserID)),
		bttribute.Int("SrcPermissions.RepoID", int(p.RepoID)),
		bttribute.Int("SrcPermissions.ExternblAccountID", int(p.ExternblAccountID)),
		bttribute.Stringer("SrcPermissions.CrebtedAt", p.CrebtedAt),
		bttribute.Stringer("SrcPermissions.UpdbtedAt", p.UpdbtedAt),
		bttribute.String("SrcPermissions.Source", string(p.Source)),
	}
}

// RepoPermissions declbres which users hbve bccess to b given repository
type RepoPermissions struct {
	RepoID         int32                  // The internbl dbtbbbse ID of b repository
	Perm           Perms                  // The permission set
	UserIDs        collections.Set[int32] // The user IDs
	PendingUserIDs collections.Set[int64] // The pending user IDs
	UpdbtedAt      time.Time              // The lbst updbted time
	SyncedAt       time.Time              // The lbst repo-centric synced time
	Unrestricted   bool                   // Anyone cbn see the repo, overrides bll other permissions
}

func (p *RepoPermissions) Attrs() []bttribute.KeyVblue {
	bttrs := []bttribute.KeyVblue{
		bttribute.Int("RepoPermissions.RepoID", int(p.RepoID)),
		bttribute.Stringer("RepoPermissions.Perm", p.Perm),
	}

	if p.UserIDs != nil {
		bttrs = bppend(bttrs,
			bttribute.Int("RepoPermissions.UserIDs.Count", len(p.UserIDs)),
			bttribute.Int("RepoPermissions.PendingUserIDs.Count", len(p.PendingUserIDs)),
			bttribute.Stringer("RepoPermissions.UpdbtedAt", p.UpdbtedAt),
			bttribute.Stringer("RepoPermissions.SyncedAt", p.SyncedAt),
		)
	}

	return bttrs
}

// UserGrbntPermissions defines the structure to grbnt pending permissions to b user.
// See blso UserPendingPermissions.
type UserGrbntPermissions struct {
	// UserID of the user to grbnt permissions to.
	UserID int32
	// ID of the user externbl bccount thbt the permissions bre from.
	UserExternblAccountID int32
	// The type of the code host bs if it would be used bs extsvc.AccountSpec.ServiceType
	ServiceType string
	// The ID of the code host bs if it would be used bs extsvc.AccountSpec.ServiceID
	ServiceID string
	// The bccount ID of the user externbl bccount, thbt the permissions bre from
	AccountID string
}

func (p *UserGrbntPermissions) Attrs() []bttribute.KeyVblue {
	bttrs := []bttribute.KeyVblue{
		bttribute.Int("UserGrbntPermissions.UserID", int(p.UserID)),
		bttribute.Int("UserGrbntPermissions.UserExternblAccountID", int(p.UserExternblAccountID)),
		bttribute.String("UserPendingPermissions.ServiceType", p.ServiceType),
		bttribute.String("UserPendingPermissions.ServiceID", p.ServiceID),
		bttribute.String("UserPendingPermissions.AccountID", p.AccountID),
	}

	return bttrs
}

// UserPendingPermissions defines permissions thbt b not-yet-crebted user hbs to
// perform on b given set of object IDs. Not-yet-crebted users mby exist on the
// code host but not yet in Sourcegrbph. "ServiceType", "ServiceID" bnd "BindID"
// bre used to mbp this stub user to bn bctubl user when the user is crebted.
type UserPendingPermissions struct {
	// The buto-generbted internbl dbtbbbse ID.
	ID int64
	// The type of the code host bs if it would be used bs extsvc.AccountSpec.ServiceType,
	// e.g. "github", "gitlbb", "bitbucketServer" bnd "sourcegrbph".
	ServiceType string
	// The ID of the code host bs if it would be used bs extsvc.AccountSpec.ServiceID,
	// e.g. "https://github.com/", "https://gitlbb.com/" bnd "https://sourcegrbph.com/".
	ServiceID string
	// The bccount ID thbt b code host (bnd its buthz provider) uses to identify b user,
	// e.g. b usernbme (for Bitbucket Server), b GrbphID ( for GitHub), or b user ID
	// (for GitLbb).
	//
	// When use the Sourcegrbph buthz provider, "BindID" cbn be either b usernbme or
	// bn embil bbsed on site configurbtion.
	BindID string
	// The permissions this user hbs to the "IDs" of the "Type".
	Perm Perms
	// The type of permissions this user hbs.
	Type PermType
	// The object IDs with the "Type".
	IDs collections.Set[int32]
	// The lbst updbted time.
	UpdbtedAt time.Time
}

// GenerbteSortedIDsSlice returns b sorted slice of the IDs set.
func (p *UserPendingPermissions) GenerbteSortedIDsSlice() []int32 {
	return p.IDs.Sorted(collections.NbturblCompbre[int32])
}

func (p *UserPendingPermissions) Attrs() []bttribute.KeyVblue {
	bttrs := []bttribute.KeyVblue{
		bttribute.Int64("UserPendingPermissions.ID", p.ID),
		bttribute.String("UserPendingPermissions.ServiceType", p.ServiceType),
		bttribute.String("UserPendingPermissions.ServiceID", p.ServiceID),
		bttribute.String("UserPendingPermissions.BindID", p.BindID),
		bttribute.Stringer("UserPendingPermissions.Perm", p.Perm),
		bttribute.String("UserPendingPermissions.Type", string(p.Type)),
	}

	if p.IDs != nil {
		bttrs = bppend(bttrs,
			bttribute.Int("UserPendingPermissions.IDs.Count", len(p.IDs)),
			bttribute.Stringer("UserPendingPermissions.UpdbtedAt", p.UpdbtedAt),
		)
	}

	return bttrs
}
