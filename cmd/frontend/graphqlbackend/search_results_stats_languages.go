package graphqlbackend

import (
	"context"
	"os"
	"sort"
	"strings"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/inventory"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

func (srs *searchResultsStats) Languages(ctx context.Context) ([]*LanguageStatistics, error) {
	srr, err := srs.getResults(ctx)
	if err != nil {
		return nil, err
	}

	getFiles := func(ctx context.Context, repo gitserver.Repo, commitID api.CommitID, path string) ([]os.FileInfo, error) {
		return git.ReadDir(ctx, repo, commitID, "", true)
	}

	var (
		files           []os.FileInfo
		repoInventories []*inventory.Inventory
	)
	for _, res := range srr.Results() {
		if fileMatch, ok := res.ToFileMatch(); ok {
			if fileMatch.File().IsDirectory() {
				repo := gitserver.Repo{Name: fileMatch.Repository().repo.Name}
				treeFiles, err := getFiles(ctx, repo, fileMatch.commitID, fileMatch.JPath)
				if err != nil {
					return nil, err
				}
				files = append(files, treeFiles...)
			} else {
				var lines int64
				if len(fileMatch.LineMatches()) > 0 {
					lines = int64(len(fileMatch.LineMatches()))
				} else {
					content, err := fileMatch.File().Content(ctx)
					if err != nil {
						return nil, err
					}
					lines = int64(strings.Count(content, "\n"))
				}
				files = append(files, &fileInfo{path: fileMatch.JPath, isDir: fileMatch.File().IsDirectory(), size: lines})
			}
		} else if repo, ok := res.ToRepository(); ok {
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
			inv, err := backend.Repos.GetInventory(ctx, repo.repo, api.CommitID(target))
			if err != nil {
				return nil, err
			}
			{
				// TODO!(sqs): hack adjust for lines
				for _, l := range inv.Languages {
					l.TotalBytes = l.TotalBytes / 31
				}
			}
			repoInventories = append(repoInventories, inv)
		} else if commit, ok := res.ToCommitSearchResult(); ok {
			if commit.raw.Diff == nil {
				continue
			}
			fileDiffs, err := diff.ParseMultiFileDiff([]byte(commit.raw.Diff.Raw))
			if err != nil {
				return nil, err
			}
			for _, fileDiff := range fileDiffs {
				var lines int64
				for _, hunk := range fileDiff.Hunks {
					c := int64(hunk.NewLines - hunk.OrigLines)
					if c < 0 {
						c = c * -1
					}
					lines += c
				}
				files = append(files, &fileInfo{path: fileDiff.NewName, isDir: false, size: lines})
			}
		}
	}

	fileInventory, err := inventory.Get(ctx, files)
	if err != nil {
		return nil, err
	}

	allInventories := append([]*inventory.Inventory{fileInventory}, repoInventories...)
	byLang := map[string]*inventory.Lang{}
	for _, inv := range allInventories {
		for _, lang := range inv.Languages {
			langInv, ok := byLang[lang.Name]
			if ok {
				langInv.TotalBytes += lang.TotalBytes
			} else {
				byLang[lang.Name] = lang
			}
		}
	}

	langStats := make([]*LanguageStatistics, 0, len(byLang))
	for _, langInv := range byLang {
		langStats = append(langStats, &LanguageStatistics{langInv})
	}
	sort.Slice(langStats, func(i, j int) bool {
		return langStats[i].Lang.TotalBytes > langStats[j].Lang.TotalBytes
	})
	return langStats, nil
}

type LanguageStatistics struct{ *inventory.Lang }

func (v *LanguageStatistics) Name() string      { return v.Lang.Name }
func (v *LanguageStatistics) TotalBytes() int32 { return int32(v.Lang.TotalBytes) }
func (v *LanguageStatistics) Type() string      { return v.Lang.Type }
