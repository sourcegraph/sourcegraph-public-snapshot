package graphqlbackend

import (
	"context"
	"os"
	"sort"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/inventory"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

func (srs *searchResultsStats) Languages(ctx context.Context) ([]*LanguageStatistics, error) {
	srr, err := srs.sr.doResults(ctx, "")
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
				files = append(files, &fileInfo{path: fileMatch.JPath, isDir: fileMatch.File().IsDirectory()})
			}
		} else if repo, ok := res.ToRepository(); ok {
			branchRef, err := repo.DefaultBranch(ctx)
			if err != nil {
				return nil, err
			}
			target, err := branchRef.Target().OID(ctx)
			if err != nil {
				return nil, err
			}
			inv, err := backend.Repos.GetInventory(ctx, repo.repo, api.CommitID(target))
			if err != nil {
				return nil, err
			}
			repoInventories = append(repoInventories, inv)
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
