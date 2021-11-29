package resolvers

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type blameAuthor struct {
	Name, Email    string
	LineCount      int
	LastCommit     api.CommitID
	LastCommitDate time.Time
}

// TODO(sqs): the "reduce" step is duplicated in this getBlameAuthors func body and above in the
// Authors method, maybe make this func return raw-er data to avoid the duplication?
func getBlameAuthors(ctx context.Context, repoName api.RepoName, path string, opt git.BlameOptions) (authorsByEmail map[string]*blameAuthor, totalLineCount int, err error) {
	// TODO(sqs): SECURITY does this check perms?
	hunks, err := git.BlameFile(ctx, repoName, path, &opt)
	if err != nil {
		return nil, 0, err
	}

	// TODO(sqs): normalize email (eg case-insensitive?)
	authorsByEmail = map[string]*blameAuthor{}
	for _, hunk := range hunks {
		a := authorsByEmail[hunk.Author.Email]
		if a == nil {
			a = &blameAuthor{
				Name:  hunk.Author.Name,
				Email: hunk.Author.Email,
			}
			authorsByEmail[hunk.Author.Email] = a
		}

		lineCount := hunk.EndLine - hunk.StartLine
		totalLineCount += lineCount
		a.LineCount += lineCount

		if hunk.Author.Date.After(a.LastCommitDate) {
			a.Name = hunk.Author.Name // use latest name in case it changed over time
			a.LastCommit = hunk.CommitID
			a.LastCommitDate = hunk.Author.Date
		}
	}

	return authorsByEmail, totalLineCount, nil
}
