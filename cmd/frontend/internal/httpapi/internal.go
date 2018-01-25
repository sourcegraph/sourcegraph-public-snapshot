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
	log15 "gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
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
	log15.Info("serveGitoliteUpdateRepos")
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

	log15.Info("serveGitoliteUpdateRepos", "totalCount", len(list), "whitelistCount", len(whitelist))

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
	err = backend.Repos.TryInsertNew(r.Context(), api.InsertRepoOp{URI: repo.RepoURI, Description: repo.Description, Fork: repo.Fork, Enabled: repo.Enabled})
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
	deps, err := backend.Defs.Dependencies(r.Context(), args.RepoID)
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
	inv, err := backend.Repos.GetInventoryUncached(r.Context(), req.Repo, req.CommitID)
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

func serveDefsRefreshIndex(w http.ResponseWriter, r *http.Request) error {
	var args api.DefsRefreshIndexRequest
	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		return err
	}
	err = backend.Defs.RefreshIndex(r.Context(), args.RepoURI, args.CommitID)
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
	err = backend.Pkgs.RefreshIndex(r.Context(), args.RepoURI, args.CommitID)
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
	cmd.Repo = &api.Repo{URI: repo.URI}
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
