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
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Provider struct {
	urn              string
	client           client
	codeHost         *extsvc.CodeHost
	projectAccessMap map[string]gerrit.ProjectAccessInfo
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
		urn:              conn.URN,
		client:           gClient,
		codeHost:         extsvc.NewCodeHost(baseURL, extsvc.TypeGerrit),
		projectAccessMap: map[string]gerrit.ProjectAccessInfo{},
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

// FetchUserPerms is a WIP
// TODO: caching for the projects/project access so we don't have to make these calls every time for every user.
// TODO: handle "revision" field of access info
func (p Provider) FetchUserPerms(ctx context.Context, account *extsvc.Account, opts authz.FetchPermsOptions) (*authz.ExternalUserPermissions, error) {
	var (
		// perms tracks repos this user has access to
		perms = &authz.ExternalUserPermissions{
			Exacts: make([]extsvc.RepoID, 0, 10), // TODO: how to determine length?
		}
	)
	// fetch account data from Gerrit by the account id
	user, err := gerrit.GetExternalAccountData(&account.AccountData)
	if err != nil {
		return nil, err
	}

	// Get all groups that this account is a member of
	groups, err := p.client.GetAccountGroups(ctx, user.AccountID)
	if err != nil {
		return nil, err
	}

	// Fetch all projects (in batches?) and determine if the user has access to each
	for {
		// TODO: pagination (how many projects should we list at a time?)
		// List all projects
		page, nextPage, err := p.client.ListProjects(ctx, gerrit.ListProjectsArgs{})
		if err != nil {
			return nil, errors.Wrap(err, "listing projects")
		}
		names := getProjectNamesFromMap(page)
		// Fetch the project access info for all of these projects
		resp, err := p.client.GetProjectAccess(ctx, names...)
		if err != nil {
			return nil, errors.Wrap(err, "getting project access")
		}
		// Based on the project access information determine if the user has access to these projects/repos
		repoAccess, err := p.interpretProjectAccess(ctx, resp, groups)
		if err != nil {
			return perms, err
		}
		addReposToUserPerms(repoAccess, perms)
		if !nextPage {
			break
		}
	}
	return perms, nil
}

func getProjectNamesFromMap(page *gerrit.ListProjectsResponse) []string {
	names := make([]string, 0, len(*page))
	for name := range *page {
		names = append(names, name)
	}
	return names
}

func (p Provider) interpretProjectAccess(ctx context.Context,
	accessResp gerrit.GetProjectAccessResponse,
	groups gerrit.GetAccountGroupsResponse) (map[string]bool, error) {

	repoAccessMap := make(map[string]bool, len(accessResp))
	for name, access := range accessResp {
		// Determine if user has access to this project and add it to the repoAccessMap if they do
		if hasAccess, err := p.userHasAccess(ctx, name, access, groups, 0); hasAccess {
			repoAccessMap[name] = true
		} else if err != nil {
			return repoAccessMap, err
		}
	}
	return repoAccessMap, nil
}

// userHasAccess checks if this user has access to the project based on if it is a member of a group which has access to the project, or if it inherits permissions from a project
// which grants permissions to a group the user has access to. Since the project can inherit access from another project which also has
// inherited access, this function contains recursion.
func (p Provider) userHasAccess(ctx context.Context, projectName string,
	access gerrit.ProjectAccessInfo,
	accountGroups gerrit.GetAccountGroupsResponse,
	counter int) (bool, error) {
	if counter > 10 { // TODO: how do we want to safeguard around this?
		return false, errors.New("no more than 10 levels of inherited access supported.")
	}
	counter++
	// If applicable, fetch inherited access information from the api or the projectAccessMap
	inheritedAccess, err := p.getInheritedAccess(ctx, access.InheritsFrom)
	if err != nil {
		return false, err
	}
	// Determine if user has access to this project and add it to the repoAccessMap if they do
	if len(access.Groups) != 0 {
		// Check if one of the groups for this user has access to this project
		if checkGroupAccess(accountGroups, projectName, access.Groups) {
			return true, nil
		}
	}
	if inheritedAccess != nil {
		if !inheritedAccess.InheritsFrom.IsEmpty() {
			// Need to recurse here and call this function using the inherited access information
			return p.userHasAccess(ctx, access.InheritsFrom.Name, *inheritedAccess, accountGroups, counter)
		}
		if checkGroupAccess(accountGroups, projectName, inheritedAccess.Groups) {
			return true, nil
		}
	}
	return false, nil
}

func checkGroupAccess(accountGroups gerrit.GetAccountGroupsResponse, projectName string, projectAccessGroups map[string]gerrit.GroupInfo) bool {
	for groupID, group := range projectAccessGroups {
		for _, aGroup := range accountGroups {
			if group.ID == aGroup.ID || groupID == aGroup.ID {
				return true
			}
		}
	}
	return false
}

func (p Provider) getInheritedAccess(ctx context.Context, inheritsFrom gerrit.Project) (*gerrit.ProjectAccessInfo, error) {
	if inheritsFrom.ID == "" { // TODO: check name as well?
		return nil, nil
	}

	// check if we've already fetched the access info for this project
	if access, ok := p.projectAccessMap[inheritsFrom.ID]; ok {
		return &access, nil
	}

	// fetch project access for this inherited project
	inheritedAccessResponse, err := p.client.GetProjectAccess(ctx, inheritsFrom.ID)
	if err != nil {
		return nil, err
	}
	if len(inheritedAccessResponse) != 1 {
		return nil, errors.New(fmt.Sprintf("A project can only inherit access from one other project, got %d instead", len(inheritedAccessResponse)))
	}
	for pname, ia := range inheritedAccessResponse {
		p.projectAccessMap[pname] = ia
		// we only have one result here so return immediately
		return &ia, nil
	}
	return nil, nil
}

func addReposToUserPerms(repoAccessMap map[string]bool, userPerms *authz.ExternalUserPermissions) *authz.ExternalUserPermissions {
	for repo := range repoAccessMap {
		userPerms.Exacts = append(userPerms.Exacts, extsvc.RepoID(repo)) // todo: figure out repo id?
	}
	return userPerms
}

func (p Provider) FetchRepoPerms(ctx context.Context, repo *extsvc.Repository, opts authz.FetchPermsOptions) ([]extsvc.AccountID, error) {
	return nil, &authz.ErrUnimplemented{Feature: "gerrit.FetchRepoPerms"}
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

func (p Provider) ValidateConnection(ctx context.Context) (warnings []string) {
	return nil
}
