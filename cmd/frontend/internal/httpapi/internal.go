package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	"github.com/sourcegraph/sourcegraph/pkg/txemail"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

func serveReposGetByName(w http.ResponseWriter, r *http.Request) error {
	repoName := api.RepoName(mux.Vars(r)["RepoName"])
	repo, err := backend.Repos.GetByName(r.Context(), repoName)
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
		Name:         repo.RepoName,
		Description:  repo.Description,
		Fork:         repo.Fork,
		Archived:     repo.Archived,
		Enabled:      repo.Enabled,
		ExternalRepo: repo.ExternalRepo,
	})
	if err != nil {
		return err
	}
	sgRepo, err := backend.Repos.GetByName(r.Context(), repo.RepoName)
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
	if err := db.Repos.UpdateRepositoryMetadata(r.Context(), repo.RepoName, repo.Description, repo.Fork, repo.Archived); err != nil {
		return errors.Wrap(err, "Repos.UpdateRepositoryMetadata failed")
	}
	return nil
}

func servePhabricatorRepoCreate(w http.ResponseWriter, r *http.Request) error {
	var repo api.PhabricatorRepoCreateRequest
	err := json.NewDecoder(r.Body).Decode(&repo)
	if err != nil {
		return err
	}
	phabRepo, err := db.Phabricator.CreateOrUpdate(r.Context(), repo.Callsign, repo.RepoName, repo.URL)
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

// serveExternalServiceConfigs serves a JSON response that is an array of all
// external service configs that match the requested kind.
func serveExternalServiceConfigs(w http.ResponseWriter, r *http.Request) error {
	var req api.ExternalServiceConfigsRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return err
	}
	services, err := db.ExternalServices.List(r.Context(), db.ExternalServicesListOptions{
		Kinds: []string{req.Kind},
	})
	if err != nil {
		return err
	}

	// Instead of returning an intermediate response type, we directly return
	// the array of configs (which are themselves JSON objects).
	// This makes it possible for the caller to directly unmarshal the response into
	// a slice of connection configurations for this external service kind.
	configs := make([]map[string]interface{}, 0, len(services))
	for _, service := range services {
		var config map[string]interface{}
		// Raw configs may have comments in them so we have to use a json parser
		// that supports comments in json.
		if err := jsonc.Unmarshal(service.Config, &config); err != nil {
			log15.Error(
				"ignoring external service config that has invalid json",
				"id", service.ID,
				"displayName", service.DisplayName,
				"config", service.Config,
				"err", err,
			)
			continue
		}
		configs = append(configs, config)
	}
	return json.NewEncoder(w).Encode(configs)
}

// serveExternalServicesList serves a JSON response that is an array of all external services
// of the given kind
func serveExternalServicesList(w http.ResponseWriter, r *http.Request) error {
	var req api.ExternalServicesListRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return err
	}

	if len(req.Kinds) == 0 {
		req.Kinds = append(req.Kinds, req.Kind)
	}

	services, err := db.ExternalServices.List(r.Context(), db.ExternalServicesListOptions{
		Kinds: req.Kinds,
	})
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(services)
}

func serveConfiguration(w http.ResponseWriter, r *http.Request) error {
	raw, err := globals.ConfigurationServerFrontendOnly.Source.Read(r.Context())
	if err != nil {
		return err
	}
	err = json.NewEncoder(w).Encode(raw)
	if err != nil {
		return errors.Wrap(err, "Encode")
	}
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

	// BACKCOMPAT: Add a "URI" field because zoekt-sourcegraph-indexserver expects one to exist
	// (with the repository name). This is a legacy of the rename from "repo URI" to "repo name".
	type repoWithBackcompatURIField struct {
		URI string
		*types.Repo
	}
	res2 := make([]*repoWithBackcompatURIField, len(res))
	for i, repo := range res {
		res2[i] = &repoWithBackcompatURIField{
			URI:  string(repo.Name),
			Repo: repo,
		}
	}

	data, err := json.Marshal(res2)
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
	settings, err := db.Settings.ListAll(r.Context(), "")
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
	info, err := db.QueryRunnerState.Get(r.Context(), query)
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
	err = db.QueryRunnerState.Set(r.Context(), &db.SavedQueryInfo{
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
	err = db.QueryRunnerState.Delete(r.Context(), query)
	if err != nil {
		return errors.Wrap(err, "SavedQueries.Delete")
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
	return nil
}

func serveSettingsGetForSubject(w http.ResponseWriter, r *http.Request) error {
	var subject api.SettingsSubject
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

func serveExternalURL(w http.ResponseWriter, r *http.Request) error {
	if err := json.NewEncoder(w).Encode(globals.ExternalURL.String()); err != nil {
		return errors.Wrap(err, "Encode")
	}
	return nil
}

func serveGitServerAddrs(w http.ResponseWriter, r *http.Request) error {
	if err := json.NewEncoder(w).Encode(conf.SrcGitServers); err != nil {
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

func serveGitResolveRevision(w http.ResponseWriter, r *http.Request) error {
	// used by zoekt-sourcegraph-mirror
	vars := mux.Vars(r)
	name := api.RepoName(vars["RepoName"])
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
	name := api.RepoName(vars["RepoName"])
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

	repoName := api.RepoName(mux.Vars(r)["RepoName"])
	repo, err := backend.Repos.GetByName(r.Context(), repoName)
	if err != nil {
		return err
	}

	if !repo.Enabled {
		return errors.Errorf("repo is not enabled: %s", repo.Name)
	}

	cmd := gitserver.DefaultClient.Command("git", "upload-pack", "--stateless-rpc", "--advertise-refs", ".")
	cmd.Repo = gitserver.Repo{Name: repo.Name}
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
	repoName := api.RepoName(mux.Vars(r)["RepoName"])
	repo, err := backend.Repos.GetByName(r.Context(), repoName)
	if err != nil {
		return err
	}

	gitserver.DefaultClient.UploadPack(repo.Name, w, r)
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
