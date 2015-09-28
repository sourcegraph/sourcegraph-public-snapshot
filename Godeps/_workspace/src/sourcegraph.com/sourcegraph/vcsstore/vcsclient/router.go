package vcsclient

import (
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/google/go-querystring/query"
	muxpkg "github.com/sourcegraph/mux"
	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/vcsstore/git"
)

const (
	// Route names
	RouteRepo                   = "vcs:repo"
	RouteRepoBlameFile          = "vcs:repo.blame-file"
	RouteRepoBranch             = "vcs:repo.branch"
	RouteRepoBranches           = "vcs:repo.branches"
	RouteRepoCommit             = "vcs:repo.commit"
	RouteRepoCommits            = "vcs:repo.commits"
	RouteRepoCommitters         = "vcs:repo.committers"
	RouteRepoCreateOrUpdate     = "vcs:repo.create-or-update"
	RouteRepoDiff               = "vcs:repo.diff"
	RouteRepoCrossRepoDiff      = "vcs:repo.cross-repo-diff"
	RouteRepoMergeBase          = "vcs:repo.merge-base"
	RouteRepoCrossRepoMergeBase = "vcs:repo.cross-repo-merge-base"
	RouteRepoRevision           = "vcs:repo.rev"
	RouteRepoSearch             = "vcs:repo.search"
	RouteRepoTag                = "vcs:repo.tag"
	RouteRepoTags               = "vcs:repo.tags"
	RouteRepoTreeEntry          = "vcs:repo.tree-entry"
	RouteRoot                   = "vcs:root"
)

type Router muxpkg.Router

// NewRouter creates a new router that matches and generates URLs that the HTTP
// handler recognizes.
func NewRouter(parent *muxpkg.Router) *Router {
	if parent == nil {
		parent = muxpkg.NewRouter()
	}

	parent.Path("/").Methods("GET").Name(RouteRoot)

	const repoURIPattern = "(?:[^./][^/]*)(?:/[^./][^/]*)*"

	repoPath := "/{RepoPath:" + repoURIPattern + "}"
	parent.Path(repoPath).Methods("GET").Name(RouteRepo)
	parent.Path(repoPath).Methods("POST").Name(RouteRepoCreateOrUpdate)

	repo := parent.PathPrefix(repoPath).Subrouter()

	// attach git transport endpoints
	repoGit := repo.PathPrefix("/.git").Subrouter()
	git.NewRouter(repoGit)

	repo.Path("/.blame/{Path:.+}").Methods("GET").Name(RouteRepoBlameFile)
	repo.Path("/.diff/{Base}..{Head}").Methods("GET").Name(RouteRepoDiff)
	repo.Path("/.cross-repo-diff/{Base}..{HeadRepoPath:" + repoURIPattern + "}:{Head}").Methods("GET").Name(RouteRepoCrossRepoDiff)
	repo.Path("/.branches").Methods("GET").Name(RouteRepoBranches)
	repo.Path("/.branches/{Branch:.+}").Methods("GET").Name(RouteRepoBranch)
	repo.Path("/.revs/{RevSpec:.+}").Methods("GET").Name(RouteRepoRevision)
	repo.Path("/.tags").Methods("GET").Name(RouteRepoTags)
	repo.Path("/.tags/{Tag:.+}").Methods("GET").Name(RouteRepoTag)
	repo.Path("/.merge-base/{CommitIDA}/{CommitIDB}").Methods("GET").Name(RouteRepoMergeBase)
	repo.Path("/.cross-repo-merge-base/{CommitIDA}/{BRepoPath:" + repoURIPattern + "}/{CommitIDB}").Methods("GET").Name(RouteRepoCrossRepoMergeBase)
	repo.Path("/.committers").Methods("GET").Name(RouteRepoCommitters)
	repo.Path("/.commits").Methods("GET").Name(RouteRepoCommits)
	commitPath := "/.commits/{CommitID}"
	repo.Path(commitPath).Methods("GET").Name(RouteRepoCommit)
	commit := repo.PathPrefix(commitPath).Subrouter()

	// cleanTreeVars modifies the Path route var to be a clean filepath. If it
	// is empty, it is changed to ".".
	cleanTreeVars := func(req *http.Request, match *muxpkg.RouteMatch, r *muxpkg.Route) {
		path := filepath.Clean(strings.TrimPrefix(match.Vars["Path"], "/"))
		if path == "" || path == "." {
			match.Vars["Path"] = "."
		} else {
			match.Vars["Path"] = path
		}
	}
	// prepareTreeVars prepares the Path route var to generate a clean URL.
	prepareTreeVars := func(vars map[string]string) map[string]string {
		if path := vars["Path"]; path == "." {
			vars["Path"] = ""
		} else {
			vars["Path"] = "/" + filepath.Clean(path)
		}
		return vars
	}
	commit.Path("/tree{Path:(?:/.*)*}").Methods("GET").PostMatchFunc(cleanTreeVars).BuildVarsFunc(prepareTreeVars).Name(RouteRepoTreeEntry)
	commit.Path("/search").Methods("GET").Name(RouteRepoSearch)

	return (*Router)(parent)
}

func (r *Router) URLToRepo(repoPath string) *url.URL {
	return r.URLTo(RouteRepo, "RepoPath", repoPath)
}

func (r *Router) URLToRepoBlameFile(repoPath string, path string, opt *vcs.BlameOptions) *url.URL {
	u := r.URLTo(RouteRepoBlameFile, "RepoPath", repoPath, "Path", path)
	if opt != nil {
		q, err := query.Values(opt)
		if err != nil {
			panic(err.Error())
		}
		u.RawQuery = q.Encode()
	}
	return u
}

func (r *Router) URLToRepoDiff(repoPath string, base, head vcs.CommitID, opt *vcs.DiffOptions) *url.URL {
	u := r.URLTo(RouteRepoDiff, "RepoPath", repoPath, "Base", string(base), "Head", string(head))
	if opt != nil {
		q, err := query.Values(opt)
		if err != nil {
			panic(err.Error())
		}
		u.RawQuery = q.Encode()
	}
	return u
}

func (r *Router) URLToRepoCrossRepoDiff(baseRepoPath string, base vcs.CommitID, headRepoPath string, head vcs.CommitID, opt *vcs.DiffOptions) *url.URL {
	u := r.URLTo(RouteRepoCrossRepoDiff, "RepoPath", baseRepoPath, "Base", string(base), "HeadRepoPath", headRepoPath, "Head", string(head))
	if opt != nil {
		q, err := query.Values(opt)
		if err != nil {
			panic(err.Error())
		}
		u.RawQuery = q.Encode()
	}
	return u
}

func (r *Router) URLToRepoBranch(repoPath string, branch string) *url.URL {
	return r.URLTo(RouteRepoBranch, "RepoPath", repoPath, "Branch", branch)
}

func (r *Router) URLToRepoBranches(repoPath string, opt vcs.BranchesOptions) *url.URL {
	u := r.URLTo(RouteRepoBranches, "RepoPath", repoPath)
	q, err := query.Values(opt)
	if err != nil {
		panic(err.Error())
	}
	u.RawQuery = q.Encode()
	return u
}

func (r *Router) URLToRepoRevision(repoPath string, revSpec string) *url.URL {
	return r.URLTo(RouteRepoRevision, "RepoPath", repoPath, "RevSpec", revSpec)
}

func (r *Router) URLToRepoTag(repoPath string, tag string) *url.URL {
	return r.URLTo(RouteRepoTag, "RepoPath", repoPath, "Tag", tag)
}

func (r *Router) URLToRepoTags(repoPath string) *url.URL {
	return r.URLTo(RouteRepoTags, "RepoPath", repoPath)
}

func (r *Router) URLToRepoCommit(repoPath string, commitID vcs.CommitID) *url.URL {
	return r.URLTo(RouteRepoCommit, "RepoPath", repoPath, "CommitID", string(commitID))
}

func (r *Router) URLToRepoCommits(repoPath string, opt vcs.CommitsOptions) *url.URL {
	u := r.URLTo(RouteRepoCommits, "RepoPath", repoPath)
	q, err := query.Values(opt)
	if err != nil {
		panic(err.Error())
	}
	u.RawQuery = q.Encode()
	return u
}

func (r *Router) URLToRepoCommitters(repoPath string, opt vcs.CommittersOptions) *url.URL {
	u := r.URLTo(RouteRepoCommitters, "RepoPath", repoPath)
	q, err := query.Values(opt)
	if err != nil {
		panic(err.Error())
	}
	u.RawQuery = q.Encode()
	return u
}

func (r *Router) URLToRepoTreeEntry(repoPath string, commitID vcs.CommitID, path string) *url.URL {
	return r.URLTo(RouteRepoTreeEntry, "RepoPath", repoPath, "CommitID", string(commitID), "Path", path)
}

func (r *Router) URLToRepoSearch(repoPath string, at vcs.CommitID, opt vcs.SearchOptions) *url.URL {
	u := r.URLTo(RouteRepoSearch, "RepoPath", repoPath, "CommitID", string(at))
	q, err := query.Values(opt)
	if err != nil {
		panic(err.Error())
	}
	u.RawQuery = q.Encode()
	return u
}

func (r *Router) URLToRepoMergeBase(repoPath string, a, b vcs.CommitID) *url.URL {
	return r.URLTo(RouteRepoMergeBase, "RepoPath", repoPath, "CommitIDA", string(a), "CommitIDB", string(b))
}

func (r *Router) URLToRepoCrossRepoMergeBase(repoPath string, a vcs.CommitID, bRepoPath string, b vcs.CommitID) *url.URL {
	return r.URLTo(RouteRepoCrossRepoMergeBase, "RepoPath", repoPath, "CommitIDA", string(a), "BRepoPath", bRepoPath, "CommitIDB", string(b))
}

func (r *Router) URLTo(route string, vars ...string) *url.URL {
	url, err := (*muxpkg.Router)(r).Get(route).URL(vars...)
	if err != nil {
		panic(err.Error())
	}
	return url
}
