package graphqlbackend

import (
	"context"
	"errors"
	"os"
	"sort"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/inventory"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func (srs *searchResultsStats) Languages(ctx context.Context) ([]*languageStatisticsResolver, error) {
	srr, err := srs.getResults(ctx)
	if err != nil {
		return nil, err
	}

	langs, err := searchResultsStatsLanguages(ctx, srr.Results())
	if err != nil {
		return nil, err
	}

	wrapped := make([]*languageStatisticsResolver, len(langs))
	for i, lang := range langs {
		wrapped[i] = &languageStatisticsResolver{lang}
	}
	return wrapped, nil
}

func searchResultsStatsLanguages(ctx context.Context, results []SearchResultResolver) ([]inventory.Lang, error) {
	// Batch our operations by repo-commit.
	type repoCommit struct {
		repo     api.RepoID
		commitID api.CommitID
	}

	// Records the work necessary for a batch (repoCommit).
	type fileStatsWork struct {
		fullFiles    []os.FileInfo     // matched these full files
		partialFiles map[string]uint64 // file with line matches (path) -> count of lines matching
	}

	var (
		repos          = map[api.RepoID]*types.Repo{}
		filesMap       = map[repoCommit]*fileStatsWork{}
		allInventories []inventory.Inventory
	)

	// Track the mapping of repo ID -> repo object as we iterate.
	sawRepo := func(repo *types.Repo) {
		if _, ok := repos[repo.ID]; !ok {
			repos[repo.ID] = repo
		}
	}

	// Only count repo matches if all matches are repo matches. Otherwise, it would get confusing
	// because we might have a match of a repo *and* a file in the repo. We would need to avoid
	// double-counting. In this case, we will just count the matching files.
	hasNonRepoMatches := false
	for _, res := range results {
		if _, ok := res.ToRepository(); !ok {
			hasNonRepoMatches = true
		}
	}

	for _, res := range results {
		if fileMatch, ok := res.ToFileMatch(); ok {
			sawRepo(fileMatch.Repository().repo)
			key := repoCommit{repo: fileMatch.Repository().repo.ID, commitID: fileMatch.CommitID}

			if _, ok := filesMap[key]; !ok {
				filesMap[key] = &fileStatsWork{}
			}

			if fileMatch.File().IsDirectory() {
				repo := gitserver.Repo{Name: fileMatch.Repository().repo.Name}
				treeFiles, err := git.ReadDir(ctx, repo, fileMatch.CommitID, fileMatch.JPath, true)
				if err != nil {
					return nil, err
				}
				filesMap[key].fullFiles = append(filesMap[key].fullFiles, treeFiles...)
			} else {
				if len(fileMatch.LineMatches()) > 0 {
					// Only count matching lines. TODO(sqs): bytes are not counted for these files
					if filesMap[key].partialFiles == nil {
						filesMap[key].partialFiles = map[string]uint64{}
					}
					filesMap[key].partialFiles[fileMatch.JPath] += uint64(len(fileMatch.LineMatches()))
				} else {
					// Count entire file.
					filesMap[key].fullFiles = append(filesMap[key].fullFiles, &fileInfo{
						path:  fileMatch.JPath,
						isDir: fileMatch.File().IsDirectory(),
						size:  1, // fake size 1 to force reading of contents (if size == 0, no read occurs)
					})
				}
			}
		} else if repo, ok := res.ToRepository(); ok && !hasNonRepoMatches {
			sawRepo(repo.repo)
			branchRef, err := repo.DefaultBranch(ctx)
			if err != nil {
				return nil, err
			}
			if branchRef == nil || branchRef.Target() == nil {
				continue
			}
			target, err := branchRef.Target().OID(ctx)
			if err != nil {
				return nil, err
			}
			inv, err := backend.Repos.GetInventory(ctx, repo.repo, api.CommitID(target), true)
			if err != nil {
				return nil, err
			}
			allInventories = append(allInventories, *inv)
		} else if _, ok := res.ToCommitSearchResult(); ok {
			return nil, errors.New("language statistics do not support diff searches")
		}
	}

	for key, work := range filesMap {
		cachedRepo, err := backend.CachedGitRepo(ctx, repos[key.repo])
		if err != nil {
			return nil, err
		}
		invCtx, err := backend.InventoryContext(*cachedRepo, key.commitID, true)
		if err != nil {
			return nil, err
		}

		// Inventory all full-file matches together.
		inv, err := invCtx.Files(ctx, work.fullFiles)
		if err != nil {
			return nil, err
		}
		allInventories = append(allInventories, inv)

		// Separately inventory each partial-file match because we only increment the language lines
		// by the number of matched lines in the file.
		for partialFile, lines := range work.partialFiles {
			inv, err := invCtx.Files(ctx, []os.FileInfo{
				// Fake size 1 to force reading of contents (if size == 0, no read occurs).
				fileInfo{path: partialFile, isDir: false, size: 1},
			})
			if err != nil {
				return nil, err
			}
			for i := range inv.Languages {
				inv.Languages[i].TotalLines = lines
			}
			allInventories = append(allInventories, inv)
		}
	}

	byLang := map[string]inventory.Lang{}
	for _, inv := range allInventories {
		for _, lang := range inv.Languages {
			langInv, ok := byLang[lang.Name]
			if ok {
				langInv.TotalBytes += lang.TotalBytes
				langInv.TotalLines += lang.TotalLines
				byLang[lang.Name] = langInv
			} else {
				byLang[lang.Name] = lang
			}
		}
	}

	langsSorted := make([]inventory.Lang, 0, len(byLang))
	for _, lang := range byLang {
		langsSorted = append(langsSorted, lang)
	}
	sort.Slice(langsSorted, func(i, j int) bool {
		return langsSorted[i].TotalLines > langsSorted[j].TotalLines
	})
	return langsSorted, nil
}
