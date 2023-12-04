package apitest

import (
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	rtypes "github.com/sourcegraph/sourcegraph/internal/rbac/types"
)

type Permission struct {
	Typename    string `json:"__typename"`
	ID          string
	Namespace   rtypes.PermissionNamespace
	DisplayName string
	Action      rtypes.NamespaceAction
	CreatedAt   gqlutil.DateTime
}

type PageInfo struct {
	HasNextPage     bool
	HasPreviousPage bool

	EndCursor   *string
	StartCursor *string
}

type PermissionConnection struct {
	Nodes      []Permission
	TotalCount int
	PageInfo   PageInfo
}

type Role struct {
	Typename    string `json:"__typename"`
	ID          string
	Name        string
	System      bool
	CreatedAt   gqlutil.DateTime
	DeletedAt   *gqlutil.DateTime
	Permissions PermissionConnection
}

type RoleConnection struct {
	Nodes      []Role
	TotalCount int
	PageInfo   PageInfo
}

type User struct {
	ID         string
	DatabaseID int32
	SiteAdmin  bool

	// All permissions associated with the roles that have been assigned to the user.
	Permissions PermissionConnection
	// All roles assigned to this user.
	Roles RoleConnection
}

type EmptyResponse struct {
	AlwaysNil string
}

type GitserverInstance struct {
	Address             string
	FreeDiskSpaceBytes  string
	TotalDiskSpaceBytes string
}

type GitserverInstanceConnection struct {
	Nodes      []GitserverInstance
	TotalCount int
	PageInfo   PageInfo
}
