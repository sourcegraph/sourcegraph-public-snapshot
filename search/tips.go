package search

import (
	"fmt"
	"math"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/svc"
)

// Tips produces a list of helpful tips (warnings about common issues,
// such as a repo having no builds) for the query.
//
// The tips are returned as TokenErrors since that type allows us to
// associate a message with a token and perform typesafe JSON
// de/serialization.
//
// If cancel is true, the query should not be planned and executed
// until the user fixes it (using the guidance provided by the tips).
func Tips(ctx context.Context, resolved []sourcegraph.Token) (cancel bool, tips []sourcegraph.TokenError, err error) {
	addTip := func(i int, tok sourcegraph.Token, msg string) {
		tips = append(tips, sourcegraph.TokenError{
			Index:   int32(i + 1),
			Token:   tp(tok),
			Message: msg,
		})
	}

	// First pass.
	var (
		repoToken  sourcegraph.RepoToken
		repoTokenI = math.MaxInt32
		revToken   sourcegraph.RevToken
		userToken  sourcegraph.UserToken
		fileToken  sourcegraph.FileToken
		unitToken  sourcegraph.UnitToken
	)
	for i, tok := range resolved {
		switch tok := tok.(type) {
		case sourcegraph.RepoToken:
			if repoToken.URI != "" {
				addTip(i, tok, fmt.Sprintf("Searching more than 1 repository at a time isn't currently supported. You already specified the repository %q. Remove one of the repositories and try again.", repoToken.URI))
				cancel = true
				return
			}

			repoToken = tok
			repoTokenI = i

		case sourcegraph.RevToken:
			if i < repoTokenI {
				addTip(i, tok, "You must specify a repository to search (e.g., 'github.com/user/repo') before specifying a revision.")
				cancel = true
				return
			}

			if revToken.Rev != "" {
				addTip(i, tok, fmt.Sprintf("Searching more than 1 revision at a time isn't currently supported. You already specified the revision %q. Remove one of the revisions and try again.", revToken.Rev))
				cancel = true
				return
			}

			revToken = tok

		case sourcegraph.UnitToken:
			if i < repoTokenI {
				addTip(i, tok, "You must specify a repository to search (e.g., 'github.com/user/repo') before specifying a source unit (package).")
				cancel = true
				return
			}

			if unitToken.Name != "" || unitToken.UnitType != "" {
				addTip(i, tok, fmt.Sprintf("Searching in more than 1 source unit (package) at a time isn't currently supported. You already specified %q. Remove one of them and try again.", unitToken.Name))
				cancel = true
				return
			}

			unitToken = tok

		case sourcegraph.FileToken:
			if i < repoTokenI {
				addTip(i, tok, "You must specify a repository to search (e.g., 'github.com/user/repo') before specifying a file.")
				cancel = true
				return
			}

			if fileToken.Path != "" {
				addTip(i, tok, fmt.Sprintf("Searching in more than 1 file at a time isn't currently supported. You already specified %q. Remove one of the files and try again.", fileToken.Path))
				cancel = true
				return
			}

			fileToken = tok

		case sourcegraph.UserToken:
			if userToken.Login != "" {
				addTip(i, tok, fmt.Sprintf("Searching more than 1 user's or org's repositories at a time isn't currently supported. You already specified %q. Remove one of the users or orgs and try again.", userToken.Login))
				cancel = true
				return
			}

			userToken = tok
		}
	}

	for i, tok := range resolved {
		switch tok := tok.(type) {
		case sourcegraph.RepoToken:
			if tok.URI != "" && tok.Repo != nil {
				repoRevSpec := sourcegraph.RepoRevSpec{RepoSpec: sourcegraph.RepoSpec{URI: tok.URI}}
				if revToken.Rev == "" {
					repoRevSpec.Rev = repoToken.Repo.DefaultBranch
				} else {
					repoRevSpec.Rev = revToken.Rev
					if revToken.Commit != nil {
						repoRevSpec.CommitID = string(revToken.Commit.ID)
					}
				}

				buildInfo, err := svc.Builds(ctx).GetRepoBuildInfo(ctx, &sourcegraph.BuildsGetRepoBuildInfoOp{Repo: repoRevSpec})
				if err != nil {
					addTip(i, tok, fmt.Sprintf("No build found for revision %q in %q.", repoRevSpec.Rev, repoRevSpec.URI))
					continue
				}
				if buildInfo.Exact == nil {
					addTip(i, tok, fmt.Sprintf("Latest build is %d commits behind (%s).", buildInfo.CommitsBehind, buildInfo.LastSuccessfulCommit.ID))
					continue
				}
			}
		}
	}

	return cancel, tips, nil
}
