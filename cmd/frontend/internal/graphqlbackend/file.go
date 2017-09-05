package graphqlbackend

import (
	"context"
	"encoding/json"
	"path"
	"strings"
	"time"

	"github.com/sourcegraph/gosyntect"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/ui2"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
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
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	vcsrepo, err := localstore.RepoVCS.Open(ctx, r.commit.RepoID)
	if err != nil {
		return "", err
	}

	contents, err := vcsrepo.ReadFile(ctx, vcs.CommitID(r.commit.CommitID), r.path)
	if err != nil {
		return "", err
	}

	return string(contents), nil
}

func (r *fileResolver) HighlightedContentHTML(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	vcsrepo, err := localstore.RepoVCS.Open(ctx, r.commit.RepoID)
	if err != nil {
		return "", err
	}

	contents, err := vcsrepo.ReadFile(ctx, vcs.CommitID(r.commit.CommitID), r.path)
	if err != nil {
		return "", err
	}

	resp, err := ui2.SyntectClient.Highlight(ctx, &gosyntect.Query{
		Code:      string(contents),
		Extension: strings.TrimPrefix(path.Ext(r.path), "."),
		Theme:     "Visual Studio Dark", // In the future, we could let the user choose the theme.
	})
	if err != nil {
		return "", err
	}
	return ui2.PreSpansToTable(resp.Data)
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

func (r *fileResolver) BlameRaw(ctx context.Context, args *struct {
	StartLine int32
	EndLine   int32
}) (string, error) {
	vcsrepo, err := localstore.RepoVCS.Open(ctx, r.commit.RepoID)
	if err != nil {
		return "", err
	}

	rawBlame, err := vcsrepo.BlameFileRaw(ctx, r.path, &vcs.BlameOptions{
		NewestCommit: vcs.CommitID(r.commit.CommitID),
		StartLine:    int(args.StartLine),
		EndLine:      int(args.EndLine),
	})
	if err != nil {
		return "", err
	}
	return rawBlame, nil
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
		CommitID:  r.commit.CommitID,
		Language:  args.Language,
		File:      r.path,
		Line:      int(args.Line),
		Character: int(args.Character),
		Limit:     20,
	})
	if err != nil {
		return nil, err
	}

	var referenceResolver []*dependencyReferenceResolver
	var repos []*repositoryResolver
	var repoIDs []int32
	for _, ref := range depRefs.References {
		if ref.RepoID == r.commit.RepoID {
			continue
		}

		repo, err := localstore.Repos.Get(ctx, ref.RepoID)
		if err != nil {
			return nil, err
		}

		repos = append(repos, &repositoryResolver{repo: repo})
		repoIDs = append(repoIDs, repo.ID)

		depData, err := json.Marshal(ref.DepData)
		if err != nil {
			return nil, err
		}

		hints, err := json.Marshal(ref.Hints)
		if err != nil {
			return nil, err
		}

		referenceResolver = append(referenceResolver, &dependencyReferenceResolver{
			dependencyData: string(depData[:]),
			repoID:         ref.RepoID,
			hints:          string(hints)[:],
		})
	}

	loc, err := json.Marshal(depRefs.Location.Location)
	if err != nil {
		return nil, err
	}

	symbol, err := json.Marshal(depRefs.Location.Symbol)
	if err != nil {
		return nil, err
	}

	return &dependencyReferencesResolver{
		dependencyReferenceData: &dependencyReferencesDataResolver{
			references: referenceResolver,
			location: &dependencyLocationResolver{
				location: string(loc[:]),
				symbol:   string(symbol[:]),
			},
		},
		repoData: &repoDataMapResolver{
			repos:   repos,
			repoIDs: repoIDs,
		},
	}, nil
}
