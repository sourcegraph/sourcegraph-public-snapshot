package resolvers

import (
	"context"
	"io/fs"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func (r *componentResolver) sourceLocationSetResolver(ctx context.Context) (*sourceLocationSetResolver, error) {
	slocs, err := r.sourceLocations(ctx)
	if err != nil {
		return nil, err
	}
	return &sourceLocationSetResolver{
		slocs:            slocs,
		getUsageResolver: r.getUsageResolver,
		db:               r.db,
	}, nil
}

func sourceLocationFromTreeEntry(treeEntry *gql.GitTreeEntryResolver, isPrimary bool) *componentSourceLocationResolver {
	return &componentSourceLocationResolver{
		repo:   treeEntry.Repository(),
		commit: treeEntry.Commit(),
		tree:   treeEntry,

		repoName:  treeEntry.Repository().RepoName(),
		commitID:  api.CommitID(treeEntry.Commit().OID()),
		path:      treeEntry.Path(),
		isPrimary: isPrimary,
	}
}

func sourceLocationSetResolverFromTreeEntry(treeEntry *gql.GitTreeEntryResolver, db database.DB) *sourceLocationSetResolver {
	return &sourceLocationSetResolver{
		slocs: []*componentSourceLocationResolver{sourceLocationFromTreeEntry(treeEntry, true)},
		db:    db,
	}
}

type sourceLocationSetResolver struct {
	slocs []*componentSourceLocationResolver

	getUsageResolver func(context.Context) (gql.ComponentUsageResolver, error)

	db database.DB
}

type fileInfo struct {
	fs.FileInfo
	repo   api.RepoName
	commit api.CommitID
}

func (r *sourceLocationSetResolver) allFiles(ctx context.Context) ([]fileInfo, error) {
	var allFiles []fileInfo
	for _, sloc := range r.slocs {
		// TODO(sqs): doesnt check perms? SECURITY
		entries, err := git.ReadDir(ctx, authz.DefaultSubRepoPermsChecker, sloc.repoName, sloc.commitID, sloc.path, true)
		if err != nil {
			return nil, err
		}
		allFiles = append(allFiles, entriesToFileInfos(entries, sloc.repoName, sloc.commitID)...)
	}
	return allFiles, nil

}

func entriesToFileInfos(entries []fs.FileInfo, repo api.RepoName, commitID api.CommitID) []fileInfo {
	var fileInfos []fileInfo
	for _, e := range entries {
		if !e.Mode().IsRegular() {
			continue // ignore dirs and submodules
		}
		fileInfos = append(fileInfos, fileInfo{
			FileInfo: e,
			repo:     repo,
			commit:   commitID,
		})
	}
	return fileInfos
}
