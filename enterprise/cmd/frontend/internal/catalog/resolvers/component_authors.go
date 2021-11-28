package resolvers

import (
	"context"
	"sort"
	"time"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func (r *catalogComponentResolver) Authors(ctx context.Context) (*[]gql.CatalogComponentAuthorEdgeResolver, error) {
	entries, err := git.ReadDir(ctx, api.RepoName(r.sourceRepo), api.CommitID(r.sourceCommit), r.sourcePath, true)
	if err != nil {
		return nil, err
	}

	// TODO(sqs): HACK make this go faster in local dev
	if max := 7; len(entries) > max {
		entries = entries[:max]
	}

	all := map[string]*blameAuthor{}
	var totalLineCount int
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		authorsByEmail, lineCount, err := getBlameAuthors(ctx, api.RepoName(r.sourceRepo), api.CommitID(r.sourceCommit), e.Name())
		if err != nil {
			return nil, err
		}

		totalLineCount += lineCount
		for email, a := range authorsByEmail {
			ca := all[email]
			if ca == nil {
				all[email] = a
			} else {
				ca.lineCount += a.lineCount
				if a.lastCommitDate.After(ca.lastCommitDate) {
					ca.name = a.name // use latest name in case it changed over time
					ca.lastCommit = a.lastCommit
					ca.lastCommitDate = a.lastCommitDate
				}
			}
		}
	}

	edges := make([]gql.CatalogComponentAuthorEdgeResolver, 0, len(all))
	for _, a := range all {
		edges = append(edges, &catalogComponentAuthorEdgeResolver{
			db:             r.db,
			component:      r,
			blameAuthor:    a,
			totalLineCount: totalLineCount,
		})
	}

	sort.Slice(edges, func(i, j int) bool {
		return edges[i].AuthoredLineCount() > edges[j].AuthoredLineCount()
	})

	return &edges, nil
}

type blameAuthor struct {
	name, email    string
	lineCount      int
	lastCommit     api.CommitID
	lastCommitDate time.Time
}

// TODO(sqs): the "reduce" step is duplicated in this getBlameAuthors func body and above in the
// Authors method, maybe make this func return raw-er data to avoid the duplication?
func getBlameAuthors(ctx context.Context, repoName api.RepoName, commit api.CommitID, path string) (authorsByEmail map[string]*blameAuthor, totalLineCount int, err error) {
	// TODO(sqs): SECURITY does this check perms?
	hunks, err := git.BlameFile(ctx, repoName, path, &git.BlameOptions{NewestCommit: commit})
	if err != nil {
		return nil, 0, err
	}

	// TODO(sqs): normalize email (eg case-insensitive?)
	authorsByEmail = map[string]*blameAuthor{}
	for _, hunk := range hunks {
		a := authorsByEmail[hunk.Author.Email]
		if a == nil {
			a = &blameAuthor{
				name:  hunk.Author.Name,
				email: hunk.Author.Email,
			}
			authorsByEmail[hunk.Author.Email] = a
		}

		lineCount := hunk.EndLine - hunk.StartLine
		totalLineCount += lineCount
		a.lineCount += lineCount

		if hunk.Author.Date.After(a.lastCommitDate) {
			a.name = hunk.Author.Name // use latest name in case it changed over time
			a.lastCommit = hunk.CommitID
			a.lastCommitDate = hunk.Author.Date
		}
	}

	return authorsByEmail, totalLineCount, nil
}

type catalogComponentAuthorEdgeResolver struct {
	db        database.DB
	component *catalogComponentResolver
	*blameAuthor
	totalLineCount int
}

func (r *catalogComponentAuthorEdgeResolver) Component() gql.CatalogComponentResolver {
	return r.component
}

func (r *catalogComponentAuthorEdgeResolver) Person() *gql.PersonResolver {
	return gql.NewPersonResolver(r.db, r.name, r.email, true)
}

func (r *catalogComponentAuthorEdgeResolver) AuthoredLineCount() int32 { return int32(r.lineCount) }

func (r *catalogComponentAuthorEdgeResolver) AuthoredLineProportion() float64 {
	return float64(r.lineCount) / float64(r.totalLineCount)
}

func (r *catalogComponentAuthorEdgeResolver) LastCommit(ctx context.Context) (*gql.GitCommitResolver, error) {
	repoResolver, err := r.component.sourceRepoResolver(ctx)
	if err != nil {
		return nil, err
	}

	return gql.NewGitCommitResolver(r.db, repoResolver, r.lastCommit, nil), nil
}
