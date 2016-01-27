package store

import (
	"errors"
	"fmt"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// Waitlist defines the interface for getting and listing users and
// GitHub orgs that have been waitlisted for access to the private repos feature.
type Waitlist interface {
	AddUser(ctx context.Context, uid int32) error
	GetUser(ctx context.Context, uid int32) (*sourcegraph.WaitlistedUser, error)
	GrantUser(ctx context.Context, uid int32) error
	ListUsers(context.Context) ([]*sourcegraph.WaitlistedUser, error)

	AddOrg(ctx context.Context, orgName string) error
	GetOrg(ctx context.Context, orgName string) (*sourcegraph.WaitlistedOrg, error)
	GrantOrg(ctx context.Context, orgName string) error
	ListOrgs(context.Context) ([]*sourcegraph.WaitlistedOrg, error)
}

// WaitlistedUserNotFoundError occurs when a WaitlistedUser is not
// found with the given arguments.
type WaitlistedUserNotFoundError struct {
	UID int32
}

func (e *WaitlistedUserNotFoundError) Error() string {
	s := fmt.Sprintf("no such waitlisted user with UID %q", e.UID)
	return s
}

// WaitlistedOrgNotFoundError occurs when a WaitlistedUser is not
// found with the given arguments.
type WaitlistedOrgNotFoundError struct {
	OrgName string
}

func (e *WaitlistedOrgNotFoundError) Error() string {
	s := fmt.Sprintf("no such waitlisted org with name %q", e.OrgName)
	return s
}

var (
	// ErrWaitlistedUserExists occurs when a WaitlistedUser with
	// the given UID already exists.
	ErrWaitlistedUserExists = errors.New("waitlisted user already exists with given UID")

	// ErrWaitlistedOrgExists occurs when a WaitlistedOrg with
	// the given name already exists.
	ErrWaitlistedOrgExists = errors.New("waitlisted org already exists with given name")
)
