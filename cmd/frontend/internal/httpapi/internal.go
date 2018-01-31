package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/sourcegraph/jsonx"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/globals"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
)

var gitoliteRepoBlacklist = compileGitoliteRepoBlacklist()

func compileGitoliteRepoBlacklist() *regexp.Regexp {
	expr := conf.Get().GitoliteRepoBlacklist
	if expr == "" {
		return nil
	}
	r, err := regexp.Compile(expr)
	if err != nil {
		log15.Error("Invalid regexp for gitolite repo blacklist", "expr", expr, "err", err)
		os.Exit(1)
	}
	return r
}

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

func serveGitoliteUpdateRepos(w http.ResponseWriter, r *http.Request) error {
	log15.Debug("serveGitoliteUpdateRepos")
	list, err := gitserver.DefaultClient.ListGitolite(r.Context())
	if err != nil {
		return err
	}

	var whitelist, blacklist []string
	for _, uri := range list {
		if strings.Contains(uri, "..*") || (gitoliteRepoBlacklist != nil && gitoliteRepoBlacklist.MatchString(uri)) {
			blacklist = append(blacklist, uri)
			continue
		}
		whitelist = append(whitelist, uri)
	}

	if len(blacklist) != 0 {
		if len(blacklist) > 5 {
			log15.Info("blacklisted gitolite repos", "size", len(blacklist))
		} else {
			log15.Info("blacklisted gitolite repos", "blacklist", strings.Join(blacklist, ", "))
		}
	}

	log15.Debug("serveGitoliteUpdateRepos", "totalCount", len(list), "whitelistCount", len(whitelist))

	insertRepoOps := make([]api.InsertRepoOp, len(whitelist))
	for i, entry := range whitelist {
		insertRepoOps[i] = api.InsertRepoOp{URI: api.RepoURI(entry), Enabled: true}
	}
	if err := backend.Repos.TryInsertNewBatch(r.Context(), insertRepoOps); err != nil {
		log15.Warn("TryInsertNewBatch failed", "numRepos", len(insertRepoOps), "err", err)
	}

	for i, entry := range whitelist {
		uri := api.RepoURI(entry)
		repo, err := backend.Repos.GetByURI(r.Context(), uri)
		if err != nil {
			log15.Warn("Could not ensure repository updated", "uri", uri, "error", err)
			continue
		}

		// Run a git fetch to kick-off an update or a clone if the repo doesn't already exist.
		cloned, err := gitserver.DefaultClient.IsRepoCloned(r.Context(), uri)
		if err != nil {
			log15.Warn("Could not ensure repository cloned", "uri", uri, "error", err)
			continue
		}
		if !conf.Get().DisableAutoGitUpdates || !cloned {
			log15.Info("fetching Gitolite repo", "repo", uri, "cloned", cloned, "i", i, "total", len(whitelist))
			err := gitserver.DefaultClient.EnqueueRepoUpdate(r.Context(), repo.URI)
			if err != nil {
				log15.Warn("Could not ensure repository cloned", "uri", uri, "error", err)
				continue
			}
		}
	}

	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte("OK"))
	return nil
}

func serveReposCreateIfNotExists(w http.ResponseWriter, r *http.Request) error {
	var repo api.RepoCreateOrUpdateRequest
	err := json.NewDecoder(r.Body).Decode(&repo)
	if err != nil {
		return err
	}
	err = backend.Repos.TryInsertNew(r.Context(), api.InsertRepoOp{
		URI:          repo.RepoURI,
		Description:  repo.Description,
		Fork:         repo.Fork,
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
	phabRepo, err := db.Phabricator.CreateIfNotExists(r.Context(), repo.Callsign, repo.RepoURI, repo.URL)
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

func serveReposUnindexedDependencies(w http.ResponseWriter, r *http.Request) error {
	var args api.RepoUnindexedDependenciesRequest
	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		return err
	}
	repo, err := backend.Repos.Get(r.Context(), args.RepoID)
	if err != nil {
		return err
	}
	deps, err := backend.Defs.Dependencies(r.Context(), repo)
	if err != nil {
		return fmt.Errorf("Defs.DependencyReferences failed: %s", err)
	}

	// Filter out already-indexed dependencies
	var unfetchedDeps []*api.DependencyReference
	for _, dep := range deps {
		pkgs, err := backend.Pkgs.ListPackages(r.Context(), &api.ListPackagesOp{Lang: args.Language, PkgQuery: depReferenceToPkgQuery(args.Language, dep), Limit: 1})
		if err != nil {
			return err
		}
		if len(pkgs) == 0 {
			unfetchedDeps = append(unfetchedDeps, dep)
		}
	}
	data, err := json.Marshal(unfetchedDeps)
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

// normalizeJSON converts JSON with comments, trailing commas, and some types of syntax errors into
// standard JSON.
func normalizeJSON(input string) []byte {
	output, _ := jsonx.Parse(string(input), jsonx.ParseOptions{Comments: true, TrailingCommas: true})
	if len(output) == 0 {
		return []byte("{}")
	}
	return output
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
		_ = json.Unmarshal(normalizeJSON(settings.Contents), &config)
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

func serveOrgsGetSlackWebhooks(w http.ResponseWriter, r *http.Request) error {
	var orgIDs []int32
	err := json.NewDecoder(r.Body).Decode(&orgIDs)
	if err != nil {
		return errors.Wrap(err, "Decode")
	}
	var webhooks []*string
	for _, orgID := range orgIDs {
		org, err := db.Orgs.GetByID(r.Context(), orgID)
		if err != nil {
			return errors.Wrap(err, "Orgs.Get")
		}
		webhooks = append(webhooks, org.SlackWebhookURL)
	}
	if err := json.NewEncoder(w).Encode(webhooks); err != nil {
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
	email, _, err := db.UserEmails.GetEmail(r.Context(), userID)
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
	err = backend.Defs.RefreshIndex(r.Context(), repo, args.CommitID)
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
	err = backend.Pkgs.RefreshIndex(r.Context(), repo, args.CommitID)
	if err != nil {
		return nil
	}
	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte("OK"))
	return nil
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

// depReferenceToPkgQuery maps from a DependencyReference to a package descriptor query that
// uniquely identifies the dependency package (typically discarding version information).  The
// mapping can be different for different languages, so languages are handled case-by-case.
func depReferenceToPkgQuery(lang string, dep *api.DependencyReference) map[string]interface{} {
	switch lang {
	case "Java":
		return map[string]interface{}{"id": dep.DepData["id"]}
	default:
		return nil
	}
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}
