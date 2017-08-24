package localstore

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"

	"context"

	"github.com/lib/pq"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/github"
)

var autoRepoWhitelist []*regexp.Regexp

func init() {
	for _, pattern := range strings.Fields(env.Get("AUTO_REPO_WHITELIST", ".+", "whitelist of repositories that will be automatically added to the DB when opened (space-separated list of lower-case regular expressions)")) {
		expr, err := regexp.Compile("^" + pattern + "$")
		if err != nil {
			log.Fatalf("invalid regular expression %q in AUTO_REPO_WHITELIST: %s", pattern, err)
		}
		autoRepoWhitelist = append(autoRepoWhitelist, expr)
	}
}

// repos is a DB-backed implementation of the Repos
type repos struct{}

// Get returns metadata for the request repository ID. It fetches data
// only from the database and NOT from any external sources. If the
// caller is concerned the copy of the data in the database might be
// stale, the caller is responsible for fetching data from any
// external services.
func (s *repos) Get(ctx context.Context, id int32) (*sourcegraph.Repo, error) {
	if Mocks.Repos.Get != nil {
		return Mocks.Repos.Get(ctx, id)
	}

	repos, err := s.getBySQL(ctx, "WHERE id=$1 LIMIT 1", id)
	if err != nil {
		return nil, err
	}

	if len(repos) == 0 {
		return nil, ErrRepoNotFound
	}
	repo := repos[0]

	// 🚨 SECURITY: access control check here 🚨
	if repo.Private && !verifyUserHasRepoURIAccess(ctx, repo.URI) {
		return nil, ErrRepoNotFound
	}
	return repo, nil
}

// GetByURI returns metadata for the request repository URI. See the
// documentation for repos.Get for the contract on the freshness of
// the data returned.
//
// If the repository doesn't already exist in the db, this method will
// add it to the db if the repo exists and start cloning, but will
// not wait for cloning to finish before returning.
//
// If the repository already exists in the db, that information is returned
// and no effort is made to detect if the repo is cloned or cloning.
func (s *repos) GetByURI(ctx context.Context, uri string) (*sourcegraph.Repo, error) {
	if Mocks.Repos.GetByURI != nil {
		return Mocks.Repos.GetByURI(ctx, uri)
	}

	repo, err := s.getByURI(ctx, uri)
	if err != nil {
		whitelisted := false
		for _, expr := range autoRepoWhitelist {
			if expr.MatchString(strings.ToLower(uri)) {
				whitelisted = true
				break
			}
		}
		if !whitelisted {
			return nil, err
		}

		if strings.HasPrefix(strings.ToLower(uri), "github.com/") {
			// Repo does not exist in DB, create new entry.
			ctx = context.WithValue(ctx, github.GitHubTrackingContextKey, "Repos.GetByURI")
			ghRepo, err := github.GetRepo(ctx, uri)
			if err != nil {
				return nil, err
			}
			if ghRepo.URI != uri {
				// not canonical name (the GitHub api will redirect from the old name to
				// the results for the new name if the repo got renamed on GitHub)
				if repo, err := s.getByURI(ctx, ghRepo.URI); err == nil {
					return repo, nil
				}
			}

			if err := s.TryInsertNew(ctx, ghRepo.URI, ghRepo.Description, ghRepo.Fork, ghRepo.Private); err != nil {
				return nil, err
			}

			return s.getByURI(ctx, ghRepo.URI)
		}

		return nil, err
	}

	return repo, nil
}

func (s *repos) getByURI(ctx context.Context, uri string) (*sourcegraph.Repo, error) {
	repos, err := s.getBySQL(ctx, "WHERE uri=$1 LIMIT 1", uri)
	if err != nil {
		return nil, err
	}

	if len(repos) == 0 {
		return nil, ErrRepoNotFound
	}
	repo := repos[0]

	// 🚨 SECURITY: access control check here 🚨
	if repo.Private && !verifyUserHasRepoURIAccess(ctx, repo.URI) {
		return nil, ErrRepoNotFound
	}

	return repo, nil
}

func (s *repos) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*sourcegraph.Repo, error) {
	rows, err := appDBH(ctx).Query("SELECT id, uri, description, homepage_url, default_branch, language, blocked, fork, private, indexed_revision, created_at, updated_at, pushed_at, freeze_indexed_revision FROM repo "+query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []*sourcegraph.Repo
	for rows.Next() {
		var repo sourcegraph.Repo
		var freezeIndexedRevision *bool

		if err := rows.Scan(
			&repo.ID,
			&repo.URI,
			&repo.Description,
			&repo.HomepageURL,
			&repo.DefaultBranch,
			&repo.Language,
			&repo.Blocked,
			&repo.Fork,
			&repo.Private,
			&repo.IndexedRevision,
			&repo.CreatedAt,
			&repo.UpdatedAt,
			&repo.PushedAt,
			&freezeIndexedRevision,
		); err != nil {
			return nil, err
		}

		repo.FreezeIndexedRevision = freezeIndexedRevision != nil && *freezeIndexedRevision // FIXME: bad DB schema: nullable boolean

		repos = append(repos, &repo)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return repos, nil
}

type RepoListOp struct {
	// Query specifies a search query for repositories. If specified, then the Sort and
	// Direction options are ignored
	Query string

	sourcegraph.ListOptions
}

// List repositories in the Sourcegraph repository  Note:
// this will not return any repositories from external services
// that are not present in the Sourcegraph repository
func (s *repos) List(ctx context.Context, opt *RepoListOp) ([]*sourcegraph.Repo, error) {
	if Mocks.Repos.List != nil {
		return Mocks.Repos.List(ctx, opt)
	}

	if opt == nil {
		opt = &RepoListOp{}
	}

	var args []interface{}
	arg := func(a interface{}) string {
		v := fmt.Sprintf("$%d", len(args)+1)
		args = append(args, a)
		return v
	}

	terms := strings.Fields(opt.Query)
	if len(terms) > 10 {
		terms = terms[:10]
	}

	conds := []string{"TRUE"}
	for _, term := range terms {
		term = strings.ToLower(term)
		term = strings.Replace(term, `\`, `\\`, -1)
		term = strings.Replace(term, "%", `\%`, -1)
		term = strings.Replace(term, "_", `\_`, -1)
		conds = append(conds, "lower(uri) LIKE "+arg("%"+term+"%"))
	}

	// fetch matching repos unordered
	rawRepos, err := s.getBySQL(ctx, "WHERE "+strings.Join(conds, " AND ")+" LIMIT 1000", args...)

	// 🚨 SECURITY: It is very important that the input list of repos (rawRepos) 🚨
	// comes directly from the DB as verifyUserHasReadAccessAll relies directly
	// on the accuracy of the Repo.Private field.
	repos, err := verifyUserHasReadAccessAll(ctx, "Repos.List", rawRepos)
	if err != nil {
		return nil, err
	}

	// sort by position of search terms
	sort.Slice(repos, func(i, j int) bool {
		uri1 := strings.ToLower(repos[i].URI)
		uri2 := strings.ToLower(repos[j].URI)
		for _, term := range terms {
			term = strings.ToLower(term)
			pos1 := strings.Index(uri1, term)
			pos2 := strings.Index(uri2, term)
			if pos1 < pos2 {
				return true
			}
			if pos2 < pos1 {
				return false
			}
		}
		return uri1 < uri2
	})

	// pagination
	if opt.Page > 0 {
		start := (opt.Page - 1) * opt.PerPage
		if int(start) >= len(repos) {
			return nil, nil
		}
		repos = repos[start:]
		if len(repos) > int(opt.PerPage) {
			repos = repos[:opt.PerPage]
		}
	}

	return repos, nil
}

// UpdateRepoFieldsFromRemote updates the DB from the remote (e.g., GitHub).
func (s *repos) UpdateRepoFieldsFromRemote(ctx context.Context, repoID int32) error {
	repo, err := s.Get(ctx, repoID)
	if err != nil {
		return err
	}

	if strings.HasPrefix(strings.ToLower(repo.URI), "github.com/") {
		return s.updateRepoFieldsFromGitHub(ctx, repo)
	}
	return nil
}

func (s *repos) updateRepoFieldsFromGitHub(ctx context.Context, repo *sourcegraph.Repo) error {
	// Fetch latest metadata from GitHub
	ghrepo, err := github.GetRepo(ctx, repo.URI)
	if err != nil {
		return err
	}

	if ghrepo.Description != repo.Description {
		if _, err := appDBH(ctx).Exec("UPDATE repo SET description=$1 WHERE id=$2", ghrepo.Description, repo.ID); err != nil {
			return err
		}
	}
	if ghrepo.HomepageURL != repo.HomepageURL {
		if _, err := appDBH(ctx).Exec("UPDATE repo SET homepage_url=$1 WHERE id=$2", ghrepo.HomepageURL, repo.ID); err != nil {
			return err
		}
	}
	if ghrepo.DefaultBranch != repo.DefaultBranch {
		if _, err := appDBH(ctx).Exec("UPDATE repo SET default_branch=$1 WHERE id=$2", ghrepo.DefaultBranch, repo.ID); err != nil {
			return err
		}
	}
	if ghrepo.Private != repo.Private {
		if _, err := appDBH(ctx).Exec("UPDATE repo SET private=$1 WHERE id=$2", ghrepo.Private, repo.ID); err != nil {
			return err
		}
	}

	if !timestampEqual(repo.UpdatedAt, ghrepo.UpdatedAt) {
		if _, err := appDBH(ctx).Exec("UPDATE repo SET updated_at=$1 WHERE id=$2", ghrepo.UpdatedAt, repo.ID); err != nil {
			return err
		}
	}
	if !timestampEqual(repo.PushedAt, ghrepo.PushedAt) {
		if _, err := appDBH(ctx).Exec("UPDATE repo SET pushed_at=$1 WHERE id=$2", ghrepo.PushedAt, repo.ID); err != nil {
			return err
		}
	}

	return nil
}

func (s *repos) UpdateLanguage(ctx context.Context, repoID int32, language string) error {
	_, err := appDBH(ctx).Exec("UPDATE repo SET language=$1 WHERE id=$2", language, repoID)
	return err
}

func (s *repos) UpdateIndexedRevision(ctx context.Context, repoID int32, rev string) error {
	_, err := appDBH(ctx).Exec("UPDATE repo SET indexed_revision=$1 WHERE id=$2", rev, repoID)
	return err
}

// TryInsertNew attempts to insert the repository rp into the db. It returns no error if a repo
// with the given uri already exists.
func (s *repos) TryInsertNew(ctx context.Context, uri string, description string, fork bool, private bool) error {
	_, err := appDBH(ctx).Exec("INSERT INTO repo (uri, description, fork, private, created_at, vcs, default_branch, homepage_url, language, blocked) VALUES ($1, $2, $3, $4, $5, '', '', '', '', false)", uri, description, fork, private, time.Now()) // FIXME: bad DB schema: nullable columns
	if err != nil {
		if isPQErrorUniqueViolation(err) {
			if c := err.(*pq.Error).Constraint; c == "repo_uri_unique" {
				return nil // repo with given uri already exists
			}
		}
		return err
	}
	return nil
}

func timestampEqual(a, b *time.Time) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(*b)
}
