package apitest

import "github.com/sourcegraph/sourcegraph/internal/gqlutil"

type Permission struct {
	Typename  string `json:"__typename"`
	ID        string
	Namespace string
	Action    string
	CreatedAt gqlutil.DateTime
}

type PageInfo struct {
	HasNextPage bool
	EndCursor   *string
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
	Readonly    bool
	CreatedAt   gqlutil.DateTime
	DeletedAt   *gqlutil.DateTime
	Permissions PermissionConnection
}
