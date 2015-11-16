package app

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/sourcegraph/mux"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/app/internal/authutil"
	"src.sourcegraph.com/sourcegraph/app/internal/tmpl"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/errcode"
	"src.sourcegraph.com/sourcegraph/ext"
	"src.sourcegraph.com/sourcegraph/ext/github"
	"src.sourcegraph.com/sourcegraph/ext/github/githubcli"
	"src.sourcegraph.com/sourcegraph/util"
	"src.sourcegraph.com/sourcegraph/util/handlerutil"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
	"src.sourcegraph.com/sourcegraph/util/router_util"
)

type userSettingsCommonData struct {
	User        *sourcegraph.User
	OrgsAndSelf []*sourcegraph.User

	OrgMembershipPerms ExternalPerms // GitHub read:org perms for listing unpublicized org memberships
}

type privateRemoteRepo struct {
	ExistsLocally bool
	*sourcegraph.Repo
}

type gitHubIntegrationData struct {
	URL                string
	Host               string
	PrivateRemoteRepos []*privateRemoteRepo
	TokenIsPresent     bool
	TokenIsValid       bool
}

var errUserSettingsCommonWroteResponse = errors.New("userSettingsCommon already wrote an HTTP response")

// userSettingsCommon should be called at the beginning of each HTTP
// handler that generates or saves settings data. It checks auth and
// fetches common data.
//
// If this function returns the error
// errUserSettingsCommonWroteResponse, callers should return nil and
// stop handling the HTTP request. That means that this function
// already sent an HTTP response (such as a redirect). For example:
//
//  userSpec, cd, err := userSettingsCommon(w, r)
//  if err == errUserSettingsCommonWroteResponse {
//  	return nil
//  } else if err != nil {
//  	return err
//  }
func userSettingsCommon(w http.ResponseWriter, r *http.Request) (sourcegraph.UserSpec, *userSettingsCommonData, error) {
	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	currentUser := handlerutil.UserFromRequest(r)
	if currentUser == nil {
		if err := authutil.RedirectToLogIn(w, r); err != nil {
			return sourcegraph.UserSpec{}, nil, err
		}
		return sourcegraph.UserSpec{}, nil, errUserSettingsCommonWroteResponse // tell caller to not continue handling this req
	}

	if err := userSettingsMeRedirect(w, r, currentUser); err == errUserSettingsCommonWroteResponse {
		return sourcegraph.UserSpec{}, nil, err
	}

	p, userSpec, err := getUser(r)
	if err != nil {
		return sourcegraph.UserSpec{}, nil, err
	}

	// The settings panel should have sections for the user AND for
	// each of the orgs that the user can admin. This list is
	// orgsAndSelf.
	orgs, err := apiclient.Orgs.List(ctx, &sourcegraph.OrgsListOp{Member: sourcegraph.UserSpec{UID: int32(currentUser.UID)}, ListOptions: sourcegraph.ListOptions{PerPage: 100}})
	if errcode.GRPC(err) == codes.Unimplemented {
		orgs = &sourcegraph.OrgList{} // ignore error
	} else if err != nil {
		return *userSpec, nil, err
	}

	orgsAndSelf := []*sourcegraph.User{p}
	for _, org := range orgs.Orgs {
		orgsAndSelf = append(orgsAndSelf, &org.User)
	}

	// The current user can only view their own profile, as well as
	// the profiles of orgs they are an admin for.
	currentUserCanAdminOrg := false
	for _, adminable := range orgsAndSelf {
		if p.UID == adminable.UID {
			currentUserCanAdminOrg = true
			break
		}
	}
	if !currentUserCanAdminOrg {
		return *userSpec, nil, &handlerutil.HTTPErr{
			Status: http.StatusForbidden,
			Err:    errors.New("only a user or an org admin can view/edit profile"),
		}
	}

	return *userSpec, &userSettingsCommonData{
		User:        p,
		OrgsAndSelf: orgsAndSelf,
	}, nil
}

// Redirects "/.me/.settings/*" to "/<u>/.settings/*". Returns errUserSettingsCommonWroteResponse if redirect should
// happen. Otherwise returns nil.
func userSettingsMeRedirect(w http.ResponseWriter, r *http.Request, u *sourcegraph.UserSpec) error {
	if u == nil {
		return nil
	}

	vars := mux.Vars(r)
	if userSpec_, err := sourcegraph.ParseUserSpec(vars["User"]); err == nil && userSpec_.Login == ".me" {
		varsCopy := make(map[string]string)
		for k, v := range vars {
			varsCopy[k] = vars[v]
		}
		varsCopy["User"] = u.Login

		redirectURL := router.Rel.URLTo(httpctx.RouteName(r), router_util.MapToArray(varsCopy)...)
		http.Redirect(w, r, redirectURL.String(), http.StatusFound)
		return errUserSettingsCommonWroteResponse
	}
	return nil
}

func userGitHubIntegrationData(ctx context.Context, apiclient *sourcegraph.Client) (*gitHubIntegrationData, error) {
	gd := &gitHubIntegrationData{
		URL:  githubcli.Config.URL(),
		Host: githubcli.Config.Host() + "/",
	}
	ghRepos := &github.Repos{}
	// TODO(perf) Cache this response or perform the fetch after page load to avoid
	// having to wait for an http round trip to github.com.
	privateGitHubRepos, err := ghRepos.ListPrivate(ctx)
	if err != nil {
		// If the error is caused by something other than the token not existing,
		// ensure the user knows there is a value set for the token but that
		// it is invalid.
		if _, ok := err.(ext.TokenNotFoundError); !ok {
			gd.TokenIsPresent = true
		}
		return gd, nil
	}
	gd.TokenIsPresent, gd.TokenIsValid = true, true

	existingRepos := make(map[string]struct{})
	privateRemoteRepos := make([]*privateRemoteRepo, len(privateGitHubRepos))

	repoOpts := &sourcegraph.RepoListOptions{
		ListOptions: sourcegraph.ListOptions{
			PerPage: 1000,
			Page:    1,
		},
	}
	for {
		repoList, err := apiclient.Repos.List(ctx, repoOpts)
		if err != nil {
			return nil, err
		}
		if len(repoList.Repos) == 0 {
			break
		}

		for _, repo := range repoList.Repos {
			existingRepos[repo.URI] = struct{}{}
		}

		repoOpts.ListOptions.Page += 1
	}

	// Check if a user's remote GitHub repo already exists locally under the
	// same URI. If so, mark it so it's clear that it can't be enabled.
	for i, repo := range privateGitHubRepos {
		if _, ok := existingRepos[repo.URI]; ok {
			privateRemoteRepos[i] = &privateRemoteRepo{ExistsLocally: true, Repo: repo}
		} else {
			privateRemoteRepos[i] = &privateRemoteRepo{ExistsLocally: false, Repo: repo}
		}
	}
	gd.PrivateRemoteRepos = privateRemoteRepos

	return gd, nil
}

func serveUserSettingsProfile(w http.ResponseWriter, r *http.Request) error {
	_, cd, err := userSettingsCommon(w, r)
	if err == errUserSettingsCommonWroteResponse {
		return nil
	} else if err != nil {
		return err
	}

	if r.Method == "POST" {
		user := cd.User
		user.Name = r.PostFormValue("Name")
		user.HomepageURL = r.PostFormValue("HomepageURL")
		user.Company = r.PostFormValue("Company")
		user.Location = r.PostFormValue("Location")
		if _, err := handlerutil.APIClient(r).Accounts.Update(httpctx.FromRequest(r), user); err != nil {
			return err
		}

		http.Redirect(w, r, router.Rel.URLTo(router.UserSettingsProfile, "User", cd.User.Login).String(), http.StatusSeeOther)
		return nil
	}

	return tmpl.Exec(r, w, "user/settings/profile.html", http.StatusOK, nil, &struct {
		userSettingsCommonData
		tmpl.Common
	}{
		userSettingsCommonData: *cd,
	})
}

func serveUserSettingsEmails(w http.ResponseWriter, r *http.Request) error {
	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	userSpec, cd, err := userSettingsCommon(w, r)
	if err == errUserSettingsCommonWroteResponse {
		return nil
	} else if err != nil {
		return err
	}

	if cd.User.IsOrganization {
		return &handlerutil.HTTPErr{
			Status: http.StatusNotFound,
			Err:    errors.New("only users have emails"),
		}
	}

	emails, err := apiclient.Users.ListEmails(ctx, &userSpec)
	if err != nil {
		if grpc.Code(err) == codes.PermissionDenied {
			// We are not allowed to view the emails, so just show
			// an empty list
			emails = &sourcegraph.EmailAddrList{EmailAddrs: []*sourcegraph.EmailAddr{}}
		} else {
			return err
		}
	}

	return tmpl.Exec(r, w, "user/settings/emails.html", http.StatusOK, nil, &struct {
		userSettingsCommonData
		EmailAddrs []*sourcegraph.EmailAddr
		ExternalPerms
		tmpl.Common
	}{
		userSettingsCommonData: *cd,
		EmailAddrs:             emails.EmailAddrs,
	})
}

func serveUserSettingsAuth(w http.ResponseWriter, r *http.Request) error {
	_, cd, err := userSettingsCommon(w, r)
	if err == errUserSettingsCommonWroteResponse {
		return nil
	} else if err != nil {
		return err
	}

	return tmpl.Exec(r, w, "user/settings/auth.html", http.StatusOK, nil, &struct {
		userSettingsCommonData
		tmpl.Common
	}{
		userSettingsCommonData: *cd,
	})
}

func serveUserSettingsIntegrations(w http.ResponseWriter, r *http.Request) error {
	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)

	_, cd, err := userSettingsCommon(w, r)
	if err == errUserSettingsCommonWroteResponse {
		return nil
	} else if err != nil {
		return err
	}

	var gd *gitHubIntegrationData

	gd, err = userGitHubIntegrationData(ctx, apiclient)
	if err != nil {
		return err
	}

	return tmpl.Exec(r, w, "user/settings/integrations.html", http.StatusOK, nil, &struct {
		userSettingsCommonData
		GitHub *gitHubIntegrationData
		tmpl.Common
	}{
		userSettingsCommonData: *cd,
		GitHub:                 gd,
	})
}

func serveUserSettingsIntegrationsUpdate(w http.ResponseWriter, r *http.Request) error {
	apiclient := handlerutil.APIClient(r)
	ctx := httpctx.FromRequest(r)
	_, cd, err := userSettingsCommon(w, r)
	if err == errUserSettingsCommonWroteResponse {
		return nil
	} else if err != nil {
		return err
	}

	switch mux.Vars(r)["Integration"] {
	case "github":
		token := r.PostFormValue("Token")
		tokenStore := ext.AccessTokens{}
		err := tokenStore.Set(ctx, githubcli.Config.Host(), token)
		if err != nil {
			return err
		}
	case "enable":
		repoURI := r.PostFormValue("RepoURI")

		// Check repo doesn't already exist, skip if so.
		_, err = apiclient.Repos.Get(ctx, &sourcegraph.RepoSpec{URI: repoURI})
		if grpc.Code(err) != codes.NotFound {
			switch err {
			case nil:
				log15.Warn("repo", repoURI, "already exists")
				http.Redirect(w, r, router.Rel.URLTo(router.UserSettingsIntegrations, "User", cd.User.Login).String(), http.StatusSeeOther)
				return nil
			default:
				return fmt.Errorf("problem getting repo %q: %v", repoURI, err)
			}
		}

		var credentials *sourcegraph.VCSCredentials

		host := util.RepoURIHost(repoURI)
		tokenStore := ext.AccessTokens{}
		token, err := tokenStore.Get(ctx, host)
		if err != nil {
			return fmt.Errorf("could not fetch credentials for host %q: %v", host, err)
		}

		credentials = &sourcegraph.VCSCredentials{
			Pass: token,
		}

		// Perform the following operations locally (non-federated) because it's a private repo and credentials are set.
		_, err = apiclient.Repos.Create(ctx, &sourcegraph.ReposCreateOp{
			URI:      repoURI,
			VCS:      "git",
			CloneURL: "https://" + repoURI + ".git",
			Mirror:   true,
			Private:  true,
		})
		if err != nil {
			return err
		}

		_, err = apiclient.MirrorRepos.RefreshVCS(ctx, &sourcegraph.MirrorReposRefreshVCSOp{
			Repo:        sourcegraph.RepoSpec{URI: repoURI},
			Credentials: credentials,
		})
		if err != nil {
			// If there was a problem, rollback to avoid leaving behind an invalid repo.
			_, _ = apiclient.Repos.Delete(ctx, &sourcegraph.RepoSpec{URI: repoURI})

			return err
		}
	}

	http.Redirect(w, r, router.Rel.URLTo(router.UserSettingsIntegrations, "User", cd.User.Login).String(), http.StatusSeeOther)
	return nil
}
