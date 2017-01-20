package graphqlbackend

import (
	"context"
	"encoding/json"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
)

type fileResolver struct {
	commit commitSpec
	name   string
	path   string
}

func (r *fileResolver) Name() string {
	return r.name
}

func (r *fileResolver) Content(ctx context.Context) (string, error) {
	file, err := backend.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{
		Entry: sourcegraph.TreeEntrySpec{
			RepoRev: sourcegraph.RepoRevSpec{
				Repo:     r.commit.RepoID,
				CommitID: r.commit.CommitID,
			},
			Path: r.path,
		},
	})
	if err != nil {
		return "", err
	}
	return string(file.Contents), nil
}

func (r *fileResolver) Commits(ctx context.Context) ([]*commitInfoResolver, error) {
	vcsrepo, err := localstore.RepoVCS.Open(ctx, r.commit.RepoID)
	if err != nil {
		return nil, err
	}

	commits, _, err := vcsrepo.Commits(ctx, vcs.CommitsOptions{
		Head: vcs.CommitID(r.commit.DefaultBranch),
		N:    20,
		Path: r.path,
	})
	if err != nil {
		return nil, err
	}
	commitsResolver := make([]*commitInfoResolver, len(commits))
	for i, commit := range commits {
		commitsResolver[i] = &commitInfoResolver{
			rev: string(commit.ID),
			author: &signatureResolver{
				person: &personResolver{
					name:         commit.Author.Name,
					email:        commit.Author.Email,
					gravatarHash: commit.Author.Email,
				},
				date: commit.Author.Date.String(),
			},
			committer: &signatureResolver{
				person: &personResolver{
					name:         commit.Author.Name,
					email:        commit.Author.Email,
					gravatarHash: commit.Author.Email,
				},
				date: commit.Committer.Date.String(),
			},
			message: commit.Message,
		}
	}

	return commitsResolver, nil
}

func (r *fileResolver) Blame(ctx context.Context,
	args *struct {
		StartLine int32
		EndLine   int32
	}) ([]*hunkResolver, error) {

	vcsrepo, err := localstore.RepoVCS.Open(ctx, r.commit.RepoID)
	if err != nil {
		return nil, err
	}

	hunks, err := vcsrepo.BlameFile(ctx, r.path, &vcs.BlameOptions{
		NewestCommit: vcs.CommitID(r.commit.CommitID),
		StartLine:    int(args.StartLine),
		EndLine:      int(args.EndLine),
	})
	if err != nil {
		return nil, err
	}

	var hunksResolver []*hunkResolver
	for _, hunk := range hunks {
		hunksResolver = append(hunksResolver, &hunkResolver{
			hunk: hunk,
		})
	}

	return hunksResolver, nil
}

func (r *fileResolver) DependencyReferences(ctx context.Context, args *struct {
	Language  string
	Line      int32
	Character int32
}) (*dependencyReferencesResolver, error) {
	depRefs, err := backend.Defs.DependencyReferences(ctx, sourcegraph.DependencyReferencesOptions{
		RepoID:    r.commit.RepoID,
		Language:  args.Language,
		File:      r.path,
		Line:      int(args.Line),
		Character: int(args.Character),
		Limit:     20,
	})
	if err != nil {
		return nil, err
	}

	refMap := make(map[int32]interface{}, len(depRefs.References))
	for _, ref := range depRefs.References {
		repo, err := localstore.Repos.Get(ctx, ref.RepoID)
		if err != nil {
			return nil, err
		}
		refMap[ref.RepoID] = repo
	}

	slcB, _ := json.Marshal(struct {
		Data     *sourcegraph.DependencyReferences
		RepoData map[int32]interface{}
	}{
		Data:     depRefs,
		RepoData: refMap,
	})

	return &dependencyReferencesResolver{
		data: string(slcB),
	}, nil
}
