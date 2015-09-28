package search

import (
	"log"
	"strings"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/fed/discover"
	"src.sourcegraph.com/sourcegraph/store"
	"src.sourcegraph.com/sourcegraph/svc"
)

// Resolve resolves a query by ensuring that all of the tokens refer
// to existing repositories/users/defs/etc. For example, a query like
// "a/b" where no such repo "a/b" exists would cause a resolution
// error.
//
// If resolving a token fails, an error is added to the resolveErrors
// return value. If an unexpected error occurs, it is returned as
// err. Callers should check the value of resolveErrors and handle
// errors appropriately (e.g., by displaying them to the user).
func Resolve(ctx context.Context, q []sourcegraph.Token) (resolved []sourcegraph.Token, resolveErrors []sourcegraph.TokenError, err error) {
	if len(q) == 0 {
		return nil, nil, nil
	}

	// Scope (filled in as it is determined).
	var repoRevSpec sourcegraph.RepoRevSpec

	hasUnitToken := false

	var rserrs []sourcegraph.TokenError
	for i, tok := range q {
		addResolveError := func(msg string) {
			rserrs = append(rserrs, sourcegraph.TokenError{
				Index:   int32(i + 1),
				Token:   tp(tok),
				Message: msg,
			})
		}

		switch tok := tok.(type) {
		case sourcegraph.RepoToken:
			if tok.URI == "" {
				continue
			}

			tryURIs := getTryURIs(tok.URI)
			var repo *sourcegraph.Repo
			var err error
			for _, uri := range tryURIs {
				repo, err = svc.Repos(ctx).Get(ctx, &sourcegraph.RepoSpec{URI: uri})
				if repo != nil {
					tok.URI = repo.URI
					break
				}
			}
			if repo == nil || err != nil {
				// TODO(sqs): check that this is a repo-not-found
				// error in a way that is compatible with the errors
				// returned by both the backend svc DB impl of
				// ReposService and the HTTP API client. currently
				// this assumes that any error is a 404.
				addResolveError("Repository not found: " + tok.URI)
				return resolved, rserrs, nil
			}

			tok.Repo = repo
			resolved = append(resolved, tok)

			if repo != nil {
				repoRevSpec.URI = repo.URI
			}

		case sourcegraph.RevToken:
			if repoRevSpec.URI != "" {
				repoRevSpec.Rev = tok.Rev
				commit, err := svc.Repos(ctx).GetCommit(ctx, &repoRevSpec)
				if err != nil {
					addResolveError("No such revision found.")
				}
				tok.Commit = commit
			}

			resolved = append(resolved, tok)

			if tok.Commit != nil {
				repoRevSpec.CommitID = string(tok.Commit.ID)
			}

		case sourcegraph.UnitToken:
			if tok.Name != "" || tok.UnitType != "" {
				if repoRevSpec.URI != "" {
					if repoRevSpec.Rev == "" {
						// Look up repo's default branch.
						repo, err := svc.Repos(ctx).Get(ctx, &repoRevSpec.RepoSpec)
						if err != nil {
							addResolveError("Couldn't find repository to get default branch for file query.")
						}
						if repo != nil {
							repoRevSpec.Rev = repo.DefaultBranch
						}
					}
					// Actually resolve since Units.Get requires us to do so.
					if repoRevSpec.CommitID == "" {
						buildInfo, err := svc.Builds(ctx).GetRepoBuildInfo(ctx, &sourcegraph.BuildsGetRepoBuildInfoOp{Repo: repoRevSpec})
						if err != nil {
							addResolveError("Couldn't resolve commit build for source unit query.")
						}
						if buildInfo != nil && buildInfo.LastSuccessful != nil {
							repoRevSpec.CommitID = buildInfo.LastSuccessful.CommitID
						}
					}

					unit, err := svc.Units(ctx).Get(ctx, &sourcegraph.UnitSpec{RepoRevSpec: repoRevSpec, Unit: tok.Name, UnitType: tok.UnitType})
					if err != nil {
						log.Println("Units.Get: ", err)
						addResolveError("Couldn't find source unit.")
					}
					if unit != nil {
						tok.Unit = unit
					}
				}
			}

			hasUnitToken = true
			resolved = append(resolved, tok)

		case sourcegraph.FileToken:
			if repoRevSpec.URI != "" && repoRevSpec.Rev == "" {
				// Look up repo's default branch.
				repo, err := svc.Repos(ctx).Get(ctx, &repoRevSpec.RepoSpec)
				if err != nil {
					addResolveError("Couldn't find repository to get default branch for file query.")
				}
				if repo != nil {
					repoRevSpec.Rev = repo.DefaultBranch
				}
			}
			if repoRevSpec.Rev == "" {
				resolved = append(resolved, tok)
				continue
			}

			// Actually resolve since RepoTree.getFromVCS requires us to do so.
			if repoRevSpec.CommitID == "" {
				commit, err := svc.Repos(ctx).GetCommit(ctx, &repoRevSpec)
				if err != nil {
					addResolveError("Couldn't resolve commit for file query.")
				}
				if commit != nil {
					repoRevSpec.CommitID = string(commit.ID)
				}
			}

			if repoRevSpec.CommitID != "" {
				entrySpec := sourcegraph.TreeEntrySpec{
					RepoRev: repoRevSpec,
					Path:    tok.Path,
				}
				entry, err := svc.RepoTree(ctx).Get(ctx, &sourcegraph.RepoTreeGetOp{Entry: entrySpec})
				if err != nil {
					addResolveError("Couldn't find file.")
				}
				if entry != nil {
					tok.Entry = entry.TreeEntry
				}
			}

			resolved = append(resolved, tok)

		case sourcegraph.UserToken:
			if tok.Login == "" {
				continue
			}

			user, err := svc.Users(ctx).Get(ctx, &sourcegraph.UserSpec{Login: tok.Login})
			if err != nil {
				// TODO(sqs): check that this is a user-not-found
				// error in a way that is compatible with the errors
				// returned by both the backend svc DB impl of
				// UsersService and the HTTP API client. currently
				// this assumes that any error is a 404.
				addResolveError("User/org not found: " + tok.Login)
			}

			tok.User = user
			resolved = append(resolved, tok)

		case sourcegraph.AnyToken:
			// Prefer to resolve the first token to a repo, if there
			// is an exact match. This is required for search to work
			// on single-path-component repos (otherwise the first
			// token would be ambiguous).
			if i == 0 {
				repo, err := svc.Repos(ctx).Get(ctx, &sourcegraph.RepoSpec{URI: string(tok)})
				if err != nil && grpc.Code(err) != codes.NotFound && !discover.IsNotFound(err) && !store.IsRepoNotFound(err) {
					return nil, nil, err
				} else if err == nil {
					resolved = append(resolved, sourcegraph.RepoToken{URI: repo.URI, Repo: repo})
					continue
				}
			}

			// Don't resolve to defs; leave those as unresolved sourcegraph.Term
			// queries. And don't resolve units if we already have a unit.
			tc := TokenCompletionConfig{DontResolveUnits: hasUnitToken, DontResolveDefs: true, MaxPerType: 1}
			comps, err := CompleteToken(ctx, tok, resolved, tc)
			if err != nil {
				return nil, nil, err
			}
			if len(comps) != 0 {
				resolved = append(resolved, comps[0])
			} else {
				resolved = append(resolved, sourcegraph.Term(tok.Token()))
			}

		default:
			resolved = append(resolved, sourcegraph.Term(tok.Token()))
		}
	}

	return resolved, rserrs, nil
}

var repoURIPrefixes = []string{
	"sourcegraph.com/",
}

// getTryURIs returns all of the repo URIs that should be tried when
// fetching uri.
func getTryURIs(uri string) []string {
	if parts := strings.Split(uri, "/"); len(parts) == 2 && !strings.Contains(parts[0], ".") {
		uris := make([]string, len(repoURIPrefixes)+1)
		uris[0] = uri // try the URI itself (perhaps it's a local repo)
		for i, p := range repoURIPrefixes {
			uris[i+1] = p + uri
		}
		return uris
	}
	return []string{uri}
}
