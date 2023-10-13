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
	Action      string
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

type EmptyResponse struct {
	AlwaysNil string
}
