package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	log15 "gopkg.in/inconshreveable/log15.v2"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
)

func serveReposGetByURI(w http.ResponseWriter, r *http.Request) error {
	uri, _ := mux.Vars(r)["RepoURI"]
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
	list, err := gitserver.DefaultClient.List()
	if err != nil {
		return err
	}

	for _, uri := range list {
		err := backend.Repos.TryInsertNew(r.Context(), uri, "", false, false)
		if err != nil {
			log15.Warn("TryInsertNew failed on repos-update", "uri", uri, "err", err)
		}
	}

	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte("OK"))
	return nil
}

func serveReposCreateIfNotExists(w http.ResponseWriter, r *http.Request) error {
	var repo sourcegraph.RepoCreateOrUpdateRequest
	err := json.NewDecoder(r.Body).Decode(&repo)
	if err != nil {
		return err
	}
	err = backend.Repos.TryInsertNew(r.Context(), repo.URI, repo.Description, repo.Fork, repo.Private)
	if err != nil {
		return err
	}
	sgRepo, err := backend.Repos.GetByURI(r.Context(), repo.URI)
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
	var repo sourcegraph.RepoUpdateIndexRequest
	err := json.NewDecoder(r.Body).Decode(&repo)
	if err != nil {
		return err
	}
	if err := localstore.Repos.UpdateIndexedRevision(r.Context(), repo.RepoID, repo.Revision); err != nil {
		return errors.Wrap(err, "Repos.UpdateIndexedRevision failed")
	}
	if err := localstore.Repos.UpdateLanguage(r.Context(), repo.RepoID, repo.Language); err != nil {
		return fmt.Errorf("Repos.UpdateLanguage failed: %s", err)
	}
	return nil
}

func servePhabricatorRepoCreate(w http.ResponseWriter, r *http.Request) error {
	var repo sourcegraph.PhabricatorRepoCreateRequest
	err := json.NewDecoder(r.Body).Decode(&repo)
	if err != nil {
		return err
	}
	phabRepo, err := localstore.Phabricator.CreateIfNotExists(r.Context(), repo.Callsign, repo.URI, repo.URL)
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
	var args sourcegraph.RepoUnindexedDependenciesRequest
	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		return err
	}
	deps, err := backend.Defs.Dependencies(r.Context(), args.RepoID)
	if err != nil {
		return fmt.Errorf("Defs.DependencyReferences failed: %s", err)
	}

	// Filter out already-indexed dependencies
	var unfetchedDeps []*sourcegraph.DependencyReference
	for _, dep := range deps {
		pkgs, err := backend.Pkgs.ListPackages(r.Context(), &sourcegraph.ListPackagesOp{Lang: args.Language, PkgQuery: depReferenceToPkgQuery(args.Language, dep), Limit: 1})
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
	var revSpec sourcegraph.RepoRevSpec
	err := json.NewDecoder(r.Body).Decode(&revSpec)
	if err != nil {
		return err
	}
	inv, err := backend.Repos.GetInventoryUncached(r.Context(), &revSpec)
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
	var opt sourcegraph.RepoListOptions
	err := json.NewDecoder(r.Body).Decode(&opt)
	if err != nil {
		return err
	}
	res, err := backend.Repos.List(r.Context(), &opt)
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
	var args sourcegraph.DefsRefreshIndexRequest
	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		return err
	}
	err = backend.Defs.RefreshIndex(r.Context(), args.URI, args.Revision)
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

	uri, _ := mux.Vars(r)["RepoURI"]
	repo, err := backend.Repos.GetByURI(r.Context(), uri)
	if err != nil {
		return err
	}

	cmd := gitserver.DefaultClient.Command("git", "upload-pack", "--stateless-rpc", "--advertise-refs", ".")
	cmd.Repo = repo
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
	uri, _ := mux.Vars(r)["RepoURI"]
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
func depReferenceToPkgQuery(lang string, dep *sourcegraph.DependencyReference) map[string]interface{} {
	switch lang {
	case "Java":
		return map[string]interface{}{"id": dep.DepData["id"]}
	default:
		return nil
	}
}
