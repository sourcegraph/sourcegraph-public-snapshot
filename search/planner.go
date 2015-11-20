package search

import (
	"fmt"
	pathpkg "path"
	"strings"

	"golang.org/x/net/context"

	"sourcegraph.com/sourcegraph/go-vcs/vcs"
	"sourcegraph.com/sourcegraph/vcsstore/vcsclient"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/svc"
)

// NewPlan creates a primary query plan for the given query (which has
// been tokenized using Tokenize).
func NewPlan(ctx context.Context, tokens []sourcegraph.Token) (*sourcegraph.Plan, error) {
	p := &sourcegraph.Plan{}
	var terms []string
	for _, tok := range tokens {
		switch tok := tok.(type) {
		case sourcegraph.Term:
			terms = append(terms, tok.UnquotedToken())

		case sourcegraph.RepoToken:
			if p.Repos == nil {
				p.Repos = &sourcegraph.RepoListOptions{}
			}
			p.Repos.URIs = append(p.Repos.URIs, tok.URI)

			if p.Defs == nil {
				p.Defs = &sourcegraph.DefListOptions{}
			}
			p.Defs.RepoRevs = append(p.Defs.RepoRevs, tok.URI)

			p.TreeRepoRevs = append(p.TreeRepoRevs, tok.URI)

		case sourcegraph.RevToken:
			// TODO(sqs): check and resolve the rev too?

			appendRev := func(repoRevs []string) {
				for i := range repoRevs {
					repoRevs[i] += "@" + tok.Rev
				}
			}

			if p.Defs == nil {
				p.Defs = &sourcegraph.DefListOptions{}
			}
			appendRev(p.Defs.RepoRevs)

			appendRev(p.TreeRepoRevs)

		case sourcegraph.UnitToken:
			if p.Defs == nil {
				p.Defs = &sourcegraph.DefListOptions{}
			}
			if tok.Name != "" {
				p.Defs.Unit = tok.Name
			}
			if tok.UnitType != "" {
				p.Defs.UnitType = tok.UnitType
			}

		case sourcegraph.FileToken:
			// TODO(sqs): check and resolve the rev too?
			if p.Defs == nil {
				p.Defs = &sourcegraph.DefListOptions{}
			}
			path := pathpkg.Clean(tok.Path)
			if tok.Entry != nil && tok.Entry.Type == vcsclient.DirEntry {
				p.Defs.FilePathPrefix = path
			} else {
				p.Defs.File = path
			}

		case sourcegraph.UserToken:
			if tok.Login != "" {
				if p.Repos == nil {
					p.Repos = &sourcegraph.RepoListOptions{}
				}
				if p.Repos.Owner != "" {
					return nil, &PlanError{Token: tok, Reason: "only one repo owner (repo:@LOGIN) may be specified"}
				}
				p.Repos.Owner = tok.Login
			}

			// Expand the user's repos list so we can create a def
			// query for all defs in any of those repos.
			if tok.Login != "" {
				repos, err := svc.Repos(ctx).List(ctx, &sourcegraph.RepoListOptions{
					Owner:       tok.Login,
					BuiltOnly:   true,
					Sort:        "pushed",
					Direction:   "desc",
					ListOptions: sourcegraph.ListOptions{PerPage: 20},
				})
				if err != nil {
					return nil, err
				}

				// TODO(sqs): produce a resolution warning/error if
				// this user has no repos that are built.

				if p.Defs == nil {
					p.Defs = &sourcegraph.DefListOptions{}
				}
				for _, repo := range repos.Repos {
					repoRev := repo.URI + "@" + repo.DefaultBranch
					p.Defs.RepoRevs = append(p.Defs.RepoRevs, repoRev)
					p.TreeRepoRevs = append(p.TreeRepoRevs, repoRev)
				}
			}

			if tok.Login != "" {
				if p.Users == nil {
					p.Users = &sourcegraph.UsersListOptions{}
				}
				p.Users.Query = tok.Login
			}

		default:
			panic(fmt.Sprintf("unrecognized token type %T", tok))
		}
	}

	if queryStr := strings.TrimSpace(strings.Join(terms, " ")); queryStr != "" {
		if p.Repos == nil && p.Defs == nil && p.Users == nil && p.Tree == nil {
			// If no other constraints are set, apply query to repos and users.
			p.Defs = &sourcegraph.DefListOptions{Query: queryStr}
			p.Repos = &sourcegraph.RepoListOptions{Query: queryStr}
			p.Users = &sourcegraph.UsersListOptions{Query: queryStr}
		} else {
			// Otherwise, add the query to existing options.
			if p.Repos != nil {
				p.Repos.Query += " " + queryStr
			}
			if p.Users != nil {
				p.Users.Query += " " + queryStr
			}
			if p.Defs != nil {
				p.Defs.Query += " " + queryStr
			}
			p.Tree = &sourcegraph.RepoTreeSearchOptions{
				SearchOptions: vcs.SearchOptions{
					Query:     queryStr,
					QueryType: vcs.FixedQuery,
					N:         10,
				},
			}
		}
	}

	// Trim whitespace and set other defaults.
	if p.Repos != nil {
		p.Repos.Query = strings.TrimSpace(p.Repos.Query)
		p.Repos.NoFork = true
		p.Repos.Sort = "updated"
		p.Repos.Direction = "desc"
	}
	if p.Users != nil {
		p.Users.Query = strings.TrimSpace(p.Users.Query)
	}
	if p.Defs != nil {
		p.Defs.Query = strings.TrimSpace(p.Defs.Query)
		p.Defs.Nonlocal = true
		if len(p.Defs.RepoRevs) > 1 {
			p.Defs.Exported = true
			p.Defs.IncludeTest = false
		}
	}

	if p.Tree == nil {
		p.TreeRepoRevs = nil
	}

	return p, nil
}

// A PlanError occurs during query planning when an invalid query is specified.
type PlanError struct {
	Token  sourcegraph.Token
	Reason string
}

func (e *PlanError) Error() string {
	return fmt.Sprintf("query: plan failed at %q: %s", e.Token, e.Reason)
}
