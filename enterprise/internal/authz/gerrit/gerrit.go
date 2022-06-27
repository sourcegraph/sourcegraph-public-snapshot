package gerrit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	jsoniter "github.com/json-iterator/go"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gerrit"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	adminGroupName = "Administrators"
	readPermission = "read"
	allowAccess    = "ALLOW"
	denyAccess     = "DENY"
	blockAccess    = "BLOCK"
	allRefs        = "/refs/*"
	headRefsRegex  = lazyregexp.New(`^\/refs\/heads\/\*$`)
)

type Provider struct {
	urn      string
	client   client
	codeHost *extsvc.CodeHost
}

func NewProvider(conn *types.GerritConnection) (*Provider, error) {
	baseURL, err := url.Parse(conn.Url)
	if err != nil {
		return nil, err
	}
	gClient, err := gerrit.NewClient(conn.URN, conn.GerritConnection, nil)
	if err != nil {
		return nil, err
	}
	return &Provider{
		urn:      conn.URN,
		client:   gClient,
		codeHost: extsvc.NewCodeHost(baseURL, extsvc.TypeGerrit),
	}, nil
}

func (p Provider) FetchAccount(ctx context.Context, user *types.User, current []*extsvc.Account, verifiedEmails []string) (*extsvc.Account, error) {
	// First try to fetch Gerrit account for this username
	accts, err := p.client.ListAccountsByUsername(ctx, user.Username)
	if err != nil {
		return nil, err
	}
	// Check that this account from Gerrit correlates to a verified email
	if acct, found, err := p.checkAccountsAgainstVerifiedEmails(accts, user, verifiedEmails); found && err == nil {
		return acct, nil
	}

	// If no account was found via the user's Sourcegraph username, attempt to find an account via one of the verified emails.
	for _, email := range verifiedEmails {
		accts, err := p.client.ListAccountsByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
		for _, acct := range accts {
			return p.buildExtsvcAccount(acct, user, email)
		}
	}

	return nil, nil
}

func (p Provider) checkAccountsAgainstVerifiedEmails(accts gerrit.ListAccountsResponse, user *types.User, verifiedEmails []string) (*extsvc.Account, bool, error) {
	if accts == nil || len(accts) == 0 {
		return nil, false, nil
	}
	for _, email := range verifiedEmails {
		for _, acct := range accts {
			if acct.Email == email && acct.Username == user.Username {
				foundAcct, err := p.buildExtsvcAccount(acct, user, email)
				return foundAcct, true, err
			}
		}
	}
	return nil, false, nil
}

func (p Provider) buildExtsvcAccount(acct gerrit.Account, user *types.User, email string) (*extsvc.Account, error) {
	acctData, err := marshalAccountData(acct.Username, acct.Email, acct.ID)
	if err != nil {
		return nil, errors.Wrap(err, "marshaling account data")
	}
	return &extsvc.Account{
		UserID: user.ID,
		AccountSpec: extsvc.AccountSpec{
			ServiceType: p.codeHost.ServiceType,
			ServiceID:   p.codeHost.ServiceID,
			AccountID:   email,
		},
		AccountData: extsvc.AccountData{
			Data: acctData,
		},
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func marshalAccountData(username, email string, acctID int32) (*json.RawMessage, error) {
	accountData, err := jsoniter.Marshal(
		gerrit.AccountData{
			Username:  username,
			Email:     email,
			AccountID: acctID,
		},
	)
	if err != nil {
		return nil, err
	}
	return (*json.RawMessage)(&accountData), nil
}

func (p Provider) FetchUserPerms(ctx context.Context, account *extsvc.Account, opts authz.FetchPermsOptions) (*authz.ExternalUserPermissions, error) {
	return nil, &authz.ErrUnimplemented{Feature: "gerrit.FetchUserPerms"}
}

func (p Provider) FetchRepoPerms(ctx context.Context, repo *extsvc.Repository, opts authz.FetchPermsOptions) ([]extsvc.AccountID, error) {
	if repo == nil {
		return nil, errors.New("no project provided")
	} else if !extsvc.IsHostOfRepo(p.codeHost, &repo.ExternalRepoSpec) {
		return nil, errors.Errorf("not a code host of the repository: want %q but have %q",
			repo.ServiceID, p.codeHost.ServiceID)
	}
	groupsAccessPermissions, err := p.getGroupsAccessPermissionsToProject(ctx, repo.ID, readPermission)
	if err != nil {
		return nil, errors.Wrapf(err, "error fetching permissions for Gerrit project: %s", repo.ID)
	}

	// Parse through all the members
	var accounts []extsvc.AccountID
	accountsMap := make(map[int32]struct{})
	accountsBlockedMap := make(map[int32]struct{})

	for group, perm := range groupsAccessPermissions {
		// if the permission is DENY we can just ignore it
		if perm != denyAccess {
			members, err := p.client.ListGroupMembers(ctx, group)
			if err != nil {
				return nil, errors.Wrapf(err, "error getting group members for Gerrit group: %s, while syncing permissions for project: %s", group, repo.ID)
			}
			if perm == allowAccess {
				for _, member := range members {
					if _, ok := accountsBlockedMap[member.ID]; ok {
						continue
					}
					accountsMap[member.ID] = struct{}{}
				}
			} else if perm == blockAccess {
				// if a member is in a group that is blocked, remove them from allowed accounts and mark them as blocked
				for _, member := range members {
					delete(accountsMap, member.ID)
					accountsBlockedMap[member.ID] = struct{}{}
				}

			}
		}
	}

	return accounts, nil
}

// getGroupsWithAccessToProject looks at the access permissions for the project and returns a map of Group IDs
// and whether they have access (or explicit denial) of the specified access type.
func (p Provider) getGroupsAccessPermissionsToProject(ctx context.Context, projectID, accessType string) (groupsAccessPermissions map[string]string, err error) {
	groupsAccessPermissions = make(map[string]string)
	pap, err := p.client.GetProjectAccessPermissions(ctx, projectID)
	if err != nil {
		return nil, errors.Wrap(err, "error when calling GetProjectAccessPermissions")
	}

	projectPerms, ok := pap[projectID]
	if !ok {
		return nil, errors.Errorf("could not find project in GetProjectAccessPermissions response: want %q but have %+v",
			projectID, pap)
	}

	// We want to gather the groups that have been given direct project access to the refs we care about
	// 1) /refs/*
	// 2) /refs/heads/*
	// TODO: Need to address sub-repo permissions 3) /refs/heads/<branch-name> (this is mainly for sub-repo permissions)
	for ref := range projectPerms.Local {
		if ref == allRefs || headRefsRegex.MatchString(ref) {
			if projectPerms.Local[ref].Permissions == nil {
				continue
			}
			perms, ok := (*projectPerms.Local[ref].Permissions)[accessType]
			if !ok {
				continue
			}
			for group := range perms.Rules {
				// if the permission is set to exclusive, we always overwrite,
				// if a permission for this group exists from the ref: /refs/* and is not set to BLOCK we overwrite it since it's set to something different in /refs/heads/*
				// see: https://gerrit-review.googlesource.com/Documentation/access-control.html#:~:text=the%20group%20%27X%27.-,%27BLOCK%27%20and%20%27ALLOW%27%20rules%20in%20the%20same%20project%20with%20the%20Exclusive%20flag,-When%20a%20project
				if perms.Exclusive != nil && *perms.Exclusive {
					groupsAccessPermissions[group] = perms.Rules[group].Action
				} else if perm, ok := groupsAccessPermissions[group]; ok {
					if !headRefsRegex.MatchString(ref) || perm == blockAccess {
						continue
					}
				}
				groupsAccessPermissions[group] = perms.Rules[group].Action
			}
		}
	}

	// If permissions for this project are inherited from another project, recurse.
	if projectPerms.InheritsFrom == nil {
		return groupsAccessPermissions, nil
	}

	// This project also inherits its permissions from a different permissions project, have to query that project too
	inheritedGroupPerms, err := p.getGroupsAccessPermissionsToProject(ctx, projectPerms.InheritsFrom.ID, readPermission)
	if err != nil {
		return nil, errors.Wrapf(err, "error when querying Gerrit project for inherited permissions")
	}

	// Consolidate the group permissions from this project's permissions and the inherited project's permissions.
	// If permission is set to Block anywhere, the group is denied access.
	// Otherwise, keep the permissions of the lower level project.
	for group, ifPerm := range inheritedGroupPerms {
		perm, ok := groupsAccessPermissions[group]
		if !ok {
			groupsAccessPermissions[group] = perm
			continue
		}
		if ifPerm == blockAccess {
			groupsAccessPermissions[group] = blockAccess
			continue
		}
	}

	return groupsAccessPermissions, nil
}

func (p Provider) FetchUserPermsByToken(ctx context.Context, token string, opts authz.FetchPermsOptions) (*authz.ExternalUserPermissions, error) {
	return nil, &authz.ErrUnimplemented{Feature: "gerrit.FetchUserPermsByToken"}
}

func (p Provider) ServiceType() string {
	return p.codeHost.ServiceType
}

func (p Provider) ServiceID() string {
	return p.codeHost.ServiceID
}

func (p Provider) URN() string {
	return p.urn
}

// ValidateConnection validates the connection to the Gerrit code host.
// Currently, this is done by querying for the Administrators group and validating that the
// group returned is valid, hence meaning that the given credentials have Admin permissions.
func (p Provider) ValidateConnection(ctx context.Context) (warnings []string) {

	adminGroup, err := p.client.GetGroupByName(ctx, adminGroupName)
	if err != nil {
		return []string{
			fmt.Sprintf("Unable to get %s group: %v", adminGroupName, err),
		}
	}

	if adminGroup.ID == "" || adminGroup.Name != adminGroupName || adminGroup.CreatedOn == "" {
		return []string{
			fmt.Sprintf("Gerrit credentials not sufficent enough to query %s group", adminGroupName),
		}
	}

	return []string{}
}
