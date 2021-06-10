package zoekt

import (
	"strings"

	"github.com/google/zoekt"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type IndexedRequestType string

const (
	TextRequest   IndexedRequestType = "text"
	SymbolRequest IndexedRequestType = "symbol"
)

// indexedRepoRevs creates both the Sourcegraph and Zoekt representation of a
// list of repository and refs to search.
type IndexedRepoRevs struct {
	// RepoRevs is the Sourcegraph representation of a the list of RepoRevs
	// repository and revisions to search.
	RepoRevs map[string]*search.RepositoryRevisions

	// RepoBranches will be used when we query zoekt. The order of branches
	// must match that in a reporev such that we can map back results. IE this
	// invariant is maintained:
	//
	//  RepoBranches[reporev.Repo.Name][i] <-> reporev.Revs[i]
	RepoBranches map[string][]string
}

// headBranch is used as a singleton of the indexedRepoRevs.repoBranches to save
// common-case allocations within indexedRepoRevs.Add.
var headBranch = []string{"HEAD"}

// Add will add reporev and repo to the list of repository and branches to
// search if reporev's refs are a subset of repo's branches. It will return
// the revision specifiers it can't add.
func (rb *IndexedRepoRevs) Add(reporev *search.RepositoryRevisions, repo *zoekt.Repository) []search.RevisionSpecifier {
	// A repo should only appear once in revs. However, in case this
	// invariant is broken we will treat later revs as if it isn't
	// indexed.
	if _, ok := rb.RepoBranches[string(reporev.Repo.Name)]; ok {
		return reporev.Revs
	}

	if !reporev.OnlyExplicit() {
		// Contains a RefGlob or ExcludeRefGlob so we can't do indexed
		// search on it.
		//
		// TODO we could only process the explicit revs and return the non
		// explicit ones as unindexed.
		return reporev.Revs
	}

	if len(reporev.Revs) == 1 && repo.Branches[0].Name == "HEAD" && (reporev.Revs[0].RevSpec == "" || reporev.Revs[0].RevSpec == "HEAD") {
		rb.RepoRevs[string(reporev.Repo.Name)] = reporev
		rb.RepoBranches[string(reporev.Repo.Name)] = headBranch
		return nil
	}

	// Assume for large searches they will mostly involve indexed
	// revisions, so just allocate that.
	var unindexed []search.RevisionSpecifier
	indexed := make([]search.RevisionSpecifier, 0, len(reporev.Revs))

	branches := make([]string, 0, len(reporev.Revs))
	for _, rev := range reporev.Revs {
		if rev.RevSpec == "" || rev.RevSpec == "HEAD" {
			// Zoekt convention that first branch is HEAD
			branches = append(branches, repo.Branches[0].Name)
			indexed = append(indexed, rev)
			continue
		}

		found := false
		for _, branch := range repo.Branches {
			if branch.Name == rev.RevSpec {
				branches = append(branches, branch.Name)
				found = true
				break
			}
			// Check if rev is an abbrev commit SHA
			if len(rev.RevSpec) >= 4 && strings.HasPrefix(branch.Version, rev.RevSpec) {
				branches = append(branches, branch.Name)
				found = true
				break
			}
		}

		if found {
			indexed = append(indexed, rev)
		} else {
			unindexed = append(unindexed, rev)
		}
	}

	// We found indexed branches! Track them.
	if len(indexed) > 0 {
		rb.RepoRevs[string(reporev.Repo.Name)] = reporev
		rb.RepoBranches[string(reporev.Repo.Name)] = branches
	}

	return unindexed
}

// GetRepoInputRev returns the repo and inputRev associated with file.
func (rb *IndexedRepoRevs) GetRepoInputRev(file *zoekt.FileMatch) (repo types.RepoName, inputRevs []string) {
	repoRev := rb.RepoRevs[file.Repository]

	inputRevs = make([]string, 0, len(file.Branches))
	for _, branch := range file.Branches {
		for i, b := range rb.RepoBranches[file.Repository] {
			if branch == b {
				// RevSpec is guaranteed to be explicit via zoektIndexedRepos
				inputRevs = append(inputRevs, repoRev.Revs[i].RevSpec)
			}
		}
	}

	if len(inputRevs) == 0 {
		// Did not find a match. This is unexpected, but we can fallback to
		// file.Version to generate correct links.
		inputRevs = append(inputRevs, file.Version)
	}

	return repoRev.Repo, inputRevs
}
