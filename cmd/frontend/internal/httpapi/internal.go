package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	"github.com/sourcegraph/sourcegraph/pkg/txemail"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

func serveReposGetByURI(w http.ResponseWriter, r *http.Request) error {
	uri := api.RepoURI(mux.Vars(r)["RepoURI"])
	repo, err := backend.Repos.GetByURI(r.Context(), uri)
	if err != nil {
		return err
	}
	data, err := json.Marshal(repo)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return nil
}

func serveReposCreateIfNotExists(w http.ResponseWriter, r *http.Request) error {
	var repo api.RepoCreateOrUpdateRequest
	err := json.NewDecoder(r.Body).Decode(&repo)
	if err != nil {
		return err
	}
	err = backend.Repos.Upsert(r.Context(), api.InsertRepoOp{
		URI:          repo.RepoURI,
		Description:  repo.Description,
		Fork:         repo.Fork,
		Archived:     repo.Archived,
		Enabled:      repo.Enabled,
		ExternalRepo: repo.ExternalRepo,
	})
	if err != nil {
		return err
	}
	sgRepo, err := backend.Repos.GetByURI(r.Context(), repo.RepoURI)
	if err != nil {
		return err
	}
	data, err := json.Marshal(sgRepo)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return nil
}

func serveReposUpdateMetadata(w http.ResponseWriter, r *http.Request) error {
	var repo api.ReposUpdateMetadataRequest
	err := json.NewDecoder(r.Body).Decode(&repo)
	if err != nil {
		return err
	}
	if err := db.Repos.UpdateRepositoryMetadata(r.Context(), repo.RepoURI, repo.Description, repo.Fork, repo.Archived); err != nil {
		return errors.Wrap(err, "Repos.UpdateRepositoryMetadata failed")
	}
	return nil
}

func serveReposUpdateIndex(w http.ResponseWriter, r *http.Request) error {
	var repo api.RepoUpdateIndexRequest
	err := json.NewDecoder(r.Body).Decode(&repo)
	if err != nil {
		return err
	}
	if err := db.Repos.UpdateIndexedRevision(r.Context(), repo.RepoID, repo.CommitID); err != nil {
		return errors.Wrap(err, "Repos.UpdateIndexedRevision failed")
	}
	if err := db.Repos.UpdateLanguage(r.Context(), repo.RepoID, repo.Language); err != nil {
		return fmt.Errorf("Repos.UpdateLanguage failed: %s", err)
	}
	return nil
}

func servePhabricatorRepoCreate(w http.ResponseWriter, r *http.Request) error {
	var repo api.PhabricatorRepoCreateRequest
	err := json.NewDecoder(r.Body).Decode(&repo)
	if err != nil {
		return err
	}
	phabRepo, err := db.Phabricator.CreateOrUpdate(r.Context(), repo.Callsign, repo.RepoURI, repo.URL)
	if err != nil {
		return err
	}
	data, err := json.Marshal(phabRepo)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return nil
}

func serveReposInventoryUncached(w http.ResponseWriter, r *http.Request) error {
	var req api.ReposGetInventoryUncachedRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return err
	}
	repo, err := backend.Repos.Get(r.Context(), req.Repo)
	if err != nil {
		return err
	}
	inv, err := backend.Repos.GetInventoryUncached(r.Context(), repo, req.CommitID)
	if err != nil {
		return err
	}
	data, err := json.Marshal(inv)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return nil
}

func serveReposInventory(w http.ResponseWriter, r *http.Request) error {
	var req api.ReposGetInventoryRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return err
	}
	repo, err := backend.Repos.Get(r.Context(), req.Repo)
	if err != nil {
		return err
	}
	inv, err := backend.Repos.GetInventory(r.Context(), repo, req.CommitID)
	if err != nil {
		return err
	}
	data, err := json.Marshal(inv)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return nil
}

func serveReposList(w http.ResponseWriter, r *http.Request) error {
	var opt db.ReposListOptions
	err := json.NewDecoder(r.Body).Decode(&opt)
	if err != nil {
		return err
	}
	res, err := backend.Repos.List(r.Context(), opt)
	if err != nil {
		return err
	}
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return nil
}

func serveReposListEnabled(w http.ResponseWriter, r *http.Request) error {
	names, err := db.Repos.ListEnabledNames(r.Context())
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(names)
}

func serveSavedQueriesListAll(w http.ResponseWriter, r *http.Request) error {
	// List settings for all users, orgs, etc.
	settings, err := db.Settings.ListAll(r.Context())
	if err != nil {
		return errors.Wrap(err, "db.Settings.ListAll")
	}

	queries := make([]api.SavedQuerySpecAndConfig, 0, len(settings))
	for _, settings := range settings {
		var config api.PartialConfigSavedQueries
		if err := jsonc.Unmarshal(settings.Contents, &config); err != nil {
			return err
		}
		for _, query := range config.SavedQueries {
			spec := api.SavedQueryIDSpec{Subject: settings.Subject, Key: query.Key}
			queries = append(queries, api.SavedQuerySpecAndConfig{
				Spec:   spec,
				Config: query,
			})
		}
	}

	if err := json.NewEncoder(w).Encode(queries); err != nil {
		return errors.Wrap(err, "Encode")
	}
	return nil
}

func serveSavedQueriesGetInfo(w http.ResponseWriter, r *http.Request) error {
	var query string
	err := json.NewDecoder(r.Body).Decode(&query)
	if err != nil {
		return errors.Wrap(err, "Decode")
	}
	info, err := db.SavedQueries.Get(r.Context(), query)
	if err != nil {
		return errors.Wrap(err, "SavedQueries.Get")
	}
	if err := json.NewEncoder(w).Encode(info); err != nil {
		return errors.Wrap(err, "Encode")
	}
	return nil
}

func serveSavedQueriesSetInfo(w http.ResponseWriter, r *http.Request) error {
	var info *api.SavedQueryInfo
	err := json.NewDecoder(r.Body).Decode(&info)
	if err != nil {
		return errors.Wrap(err, "Decode")
	}
	err = db.SavedQueries.Set(r.Context(), &db.SavedQueryInfo{
		Query:        info.Query,
		LastExecuted: info.LastExecuted,
		LatestResult: info.LatestResult,
		ExecDuration: info.ExecDuration,
	})
	if err != nil {
		return errors.Wrap(err, "SavedQueries.Set")
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	return nil
}

func serveSavedQueriesDeleteInfo(w http.ResponseWriter, r *http.Request) error {
	var query string
	err := json.NewDecoder(r.Body).Decode(&query)
	if err != nil {
		return errors.Wrap(err, "Decode")
	}
	err = db.SavedQueries.Delete(r.Context(), query)
	if err != nil {
		return errors.Wrap(err, "SavedQueries.Delete")
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	return nil
}

func serveSettingsGetForSubject(w http.ResponseWriter, r *http.Request) error {
	var subject api.ConfigurationSubject
	if err := json.NewDecoder(r.Body).Decode(&subject); err != nil {
		return errors.Wrap(err, "Decode")
	}
	settings, err := db.Settings.GetLatest(r.Context(), subject)
	if err != nil {
		return errors.Wrap(err, "Settings.GetLatest")
	}
	if err := json.NewEncoder(w).Encode(settings); err != nil {
		return errors.Wrap(err, "Encode")
	}
	return nil
}

func serveOrgsListUsers(w http.ResponseWriter, r *http.Request) error {
	var orgID int32
	err := json.NewDecoder(r.Body).Decode(&orgID)
	if err != nil {
		return errors.Wrap(err, "Decode")
	}
	orgMembers, err := db.OrgMembers.GetByOrgID(r.Context(), orgID)
	if err != nil {
		return errors.Wrap(err, "OrgMembers.GetByOrgID")
	}
	users := make([]int32, 0, len(orgMembers))
	for _, member := range orgMembers {
		users = append(users, member.UserID)
	}
	if err := json.NewEncoder(w).Encode(users); err != nil {
		return errors.Wrap(err, "Encode")
	}
	return nil
}

func serveOrgsGetByName(w http.ResponseWriter, r *http.Request) error {
	var orgName string
	err := json.NewDecoder(r.Body).Decode(&orgName)
	if err != nil {
		return errors.Wrap(err, "Decode")
	}
	org, err := db.Orgs.GetByName(r.Context(), orgName)
	if err != nil {
		return errors.Wrap(err, "Orgs.GetByName")
	}
	if err := json.NewEncoder(w).Encode(org.ID); err != nil {
		return errors.Wrap(err, "Encode")
	}
	return nil
}

func serveUsersGetByUsername(w http.ResponseWriter, r *http.Request) error {
	var username string
	err := json.NewDecoder(r.Body).Decode(&username)
	if err != nil {
		return errors.Wrap(err, "Decode")
	}
	user, err := db.Users.GetByUsername(r.Context(), username)
	if err != nil {
		return errors.Wrap(err, "Users.GetByUsername")
	}
	if err := json.NewEncoder(w).Encode(user.ID); err != nil {
		return errors.Wrap(err, "Encode")
	}
	return nil
}

func serveUserEmailsGetEmail(w http.ResponseWriter, r *http.Request) error {
	var userID int32
	err := json.NewDecoder(r.Body).Decode(&userID)
	if err != nil {
		return errors.Wrap(err, "Decode")
	}
	email, _, err := db.UserEmails.GetPrimaryEmail(r.Context(), userID)
	if err != nil {
		return errors.Wrap(err, "UserEmails.GetEmail")
	}
	if err := json.NewEncoder(w).Encode(email); err != nil {
		return errors.Wrap(err, "Encode")
	}
	return nil
}

func serveAppURL(w http.ResponseWriter, r *http.Request) error {
	if err := json.NewEncoder(w).Encode(globals.AppURL.String()); err != nil {
		return errors.Wrap(err, "Encode")
	}
	return nil
}

func serveCanSendEmail(w http.ResponseWriter, r *http.Request) error {
	if err := json.NewEncoder(w).Encode(conf.CanSendEmail()); err != nil {
		return errors.Wrap(err, "Encode")
	}
	return nil
}

func serveSendEmail(w http.ResponseWriter, r *http.Request) error {
	var msg txemail.Message
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		return err
	}
	return txemail.Send(r.Context(), msg)
}

func serveDefsRefreshIndex(w http.ResponseWriter, r *http.Request) error {
	var args api.DefsRefreshIndexRequest
	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		return err
	}
	repo, err := backend.Repos.GetByURI(r.Context(), args.RepoURI)
	if err != nil {
		return err
	}
	err = backend.Dependencies.RefreshIndex(r.Context(), repo, args.CommitID)
	if err != nil {
		return nil
	}
	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte("OK"))
	return nil
}

func servePkgsRefreshIndex(w http.ResponseWriter, r *http.Request) error {
	var args api.PkgsRefreshIndexRequest
	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		return err
	}
	repo, err := backend.Repos.GetByURI(r.Context(), args.RepoURI)
	if err != nil {
		return err
	}
	err = backend.Packages.RefreshIndex(r.Context(), repo, args.CommitID)
	if err != nil {
		return nil
	}
	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte("OK"))
	return nil
}

func serveGitResolveRevision(w http.ResponseWriter, r *http.Request) error {
	// used by zoekt-sourcegraph-mirror
	vars := mux.Vars(r)
	name := api.RepoURI(vars["RepoURI"])
	spec := vars["Spec"]

	// Do not to trigger a repo-updater lookup since this is a batch job.
	commitID, err := git.ResolveRevision(r.Context(), gitserver.Repo{Name: name}, nil, spec, nil)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(commitID))
	return nil
}

func serveGitTar(w http.ResponseWriter, r *http.Request) error {
	// used by zoekt-sourcegraph-mirror
	vars := mux.Vars(r)
	name := api.RepoURI(vars["RepoURI"])
	spec := vars["Commit"]

	// Ensure commit exists. Do not want to trigger a repo-updater lookup since this is a batch job.
	repo := gitserver.Repo{Name: name}
	commit, err := git.ResolveRevision(r.Context(), repo, nil, spec, nil)
	if err != nil {
		return err
	}

	src, err := git.Archive(r.Context(), repo, git.ArchiveOptions{Treeish: string(commit), Format: "tar"})
	if err != nil {
		return err
	}
	defer src.Close()

	w.Header().Set("Content-Type", "application/x-tar")
	w.WriteHeader(http.StatusOK)
	_, err = io.Copy(w, src)
	return err
}

func serveGitInfoRefs(w http.ResponseWriter, r *http.Request) error {
	service := r.URL.Query().Get("service")
	if service != "git-upload-pack" {
		return errors.New("only support service git-upload-pack")
	}

	uri := api.RepoURI(mux.Vars(r)["RepoURI"])
	repo, err := backend.Repos.GetByURI(r.Context(), uri)
	if err != nil {
		return err
	}

	if !repo.Enabled {
		return errors.Errorf("repo is not enabled: %s", repo.URI)
	}

	cmd := gitserver.DefaultClient.Command("git", "upload-pack", "--stateless-rpc", "--advertise-refs", ".")
	cmd.Repo = gitserver.Repo{Name: repo.URI}
	refs, err := cmd.Output(r.Context())
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", fmt.Sprintf("application/x-git-upload-pack-advertisement"))
	w.WriteHeader(http.StatusOK)
	w.Write(packetWrite("# service=git-upload-pack\n"))
	w.Write([]byte("0000"))
	w.Write(refs)
	return nil
}

func serveGitUploadPack(w http.ResponseWriter, r *http.Request) error {
	uri := api.RepoURI(mux.Vars(r)["RepoURI"])
	repo, err := backend.Repos.GetByURI(r.Context(), uri)
	if err != nil {
		return err
	}

	gitserver.DefaultClient.UploadPack(repo.URI, w, r)
	return nil
}

func packetWrite(str string) []byte {
	s := strconv.FormatInt(int64(len(str)+4), 16)
	if len(s)%4 != 0 {
		s = strings.Repeat("0", 4-len(s)%4) + s
	}
	return []byte(s + str)
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}
