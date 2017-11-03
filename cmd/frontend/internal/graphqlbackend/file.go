package graphqlbackend

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"path"
	"strings"
	"time"
	"unicode/utf8"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/highlight"
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

func (r *fileResolver) IsDirectory(ctx context.Context) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	vcsrepo, err := localstore.RepoVCS.Open(ctx, r.commit.RepoID)
	if err != nil {
		return false, err
	}

	stat, err := vcsrepo.Stat(ctx, vcs.CommitID(r.commit.CommitID), r.path)
	if err != nil {
		return false, err
	}

	return stat.IsDir(), nil
}

func (r *fileResolver) Repository(ctx context.Context) (*repositoryResolver, error) {
	repo, err := localstore.Repos.Get(ctx, r.commit.RepoID)
	if err != nil {
		return nil, err
	}
	return &repositoryResolver{repo: repo}, nil
}

func (r *fileResolver) Binary(ctx context.Context) (bool, error) {
	content, err := r.Content(ctx)
	if err != nil {
		return false, err
	}
	return r.isBinary([]byte(content)), nil
}

// isBinary is a helper to tell if the content of a file is binary or not. It
// is used instead of utf8.Valid in case we ever need to add e.g. extension
// specific checks in addition to checking if the content is valid utf8.
func (r *fileResolver) isBinary(content []byte) bool {
	return !utf8.Valid(content)
}

type highlightedFileResolver struct {
	aborted bool
	html    string
}

func (h *highlightedFileResolver) Aborted() bool { return h.aborted }
func (h *highlightedFileResolver) HTML() string  { return h.html }

func (r *fileResolver) Highlight(ctx context.Context, args *struct {
	DisableTimeout bool
}) (*highlightedFileResolver, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	vcsrepo, err := localstore.RepoVCS.Open(ctx, r.commit.RepoID)
	if err != nil {
		return nil, err
	}

	code, err := vcsrepo.ReadFile(ctx, vcs.CommitID(r.commit.CommitID), r.path)
	if err != nil {
		return nil, err
	}

	// Never pass binary files to the syntax highlighter.
	if r.isBinary(code) {
		return nil, errors.New("cannot render binary file")
	}

	// Highlight the code.
	var (
		html   template.HTML
		result = &highlightedFileResolver{}
	)
	html, result.aborted, err = highlight.Code(ctx, string(code), strings.TrimPrefix(path.Ext(r.path), "."), args.DisableTimeout)
	if err != nil {
		return nil, err
	}
	result.html = string(html)
	return result, nil
}

func (r *fileResolver) Commit(ctx context.Context) (*commitResolver, error) {
	repo, err := localstore.Repos.Get(ctx, r.commit.RepoID)
	if err != nil {
		return nil, err
	}
	return &commitResolver{
		commit: r.commit,
		repo:   *repo,
	}, nil
}

func (r *fileResolver) LastCommit(ctx context.Context) (*commitInfoResolver, error) {
	commits, err := r.commits(ctx, 1)
	if err != nil {
		return nil, err
	}
	return commits[0], nil
}

func (r *fileResolver) Commits(ctx context.Context) ([]*commitInfoResolver, error) {
	return r.commits(ctx, 20)
}

func (r *fileResolver) commits(ctx context.Context, limit uint) ([]*commitInfoResolver, error) {
	vcsrepo, err := localstore.RepoVCS.Open(ctx, r.commit.RepoID)
	if err != nil {
		return nil, err
	}

	commits, _, err := vcsrepo.Commits(ctx, vcs.CommitsOptions{
		Head:    vcs.CommitID(r.commit.DefaultBranch),
		N:       limit,
		Path:    r.path,
		NoTotal: true,
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
					name:  commit.Author.Name,
					email: commit.Author.Email,
				},
				date: commit.Author.Date.String(),
			},
			committer: &signatureResolver{
				person: &personResolver{
					name:  commit.Author.Name,
					email: commit.Author.Email,
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
		Limit:     500,
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
