package localstore

import (
	"bytes"
	"errors"
	"log"
	"regexp"
	regexpsyntax "regexp/syntax"
	"strings"
	"time"
	"unicode"

	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf/feature"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/github"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
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

	repos, err := s.getBySQL(ctx, sqlf.Sprintf("WHERE id=%d LIMIT 1", id))
	if err != nil {
		return nil, err
	}

	if len(repos) == 0 {
		return nil, ErrRepoNotFound
	}
	repo := repos[0]

	if !feature.Features.Sep20Auth {
		// 🚨 SECURITY: access control check here 🚨
		if repo.Private && !verifyUserHasRepoURIAccess(ctx, repo.URI) {
			return nil, ErrRepoNotFound
		}
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

		// Auto-add repository if possible
		if strings.HasPrefix(strings.ToLower(uri), "github.com/") {
			if ghRepo, err := s.addFromGitHubAPI(ctx, uri); err == nil {
				return ghRepo, nil
			}
			return nil, ErrRepoNotFound
		}
		cloneable, err := gitserver.DefaultClient.IsRepoCloneable(ctx, uri)
		if err != nil {
			return nil, err
		}
		if !cloneable {
			return nil, ErrRepoNotFound
		}
		if err := s.TryInsertNew(ctx, uri, "", false, false); err != nil {
			return nil, err
		}
		return s.getByURI(ctx, uri)
	}

	return repo, nil
}

func (s *repos) addFromGitHubAPI(ctx context.Context, uri string) (*sourcegraph.Repo, error) {
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

func (s *repos) getByURI(ctx context.Context, uri string) (*sourcegraph.Repo, error) {
	repos, err := s.getBySQL(ctx, sqlf.Sprintf("WHERE uri=%s LIMIT 1", uri))
	if err != nil {
		return nil, err
	}

	if len(repos) == 0 {
		return nil, ErrRepoNotFound
	}
	repo := repos[0]

	if !feature.Features.Sep20Auth {
		// 🚨 SECURITY: access control check here 🚨
		if repo.Private && !verifyUserHasRepoURIAccess(ctx, repo.URI) {
			return nil, ErrRepoNotFound
		}
	}

	return repo, nil
}

func (s *repos) getBySQL(ctx context.Context, querySuffix *sqlf.Query) ([]*sourcegraph.Repo, error) {
	q := sqlf.Sprintf("SELECT id, uri, description, homepage_url, default_branch, language, blocked, fork, private, indexed_revision, created_at, updated_at, pushed_at, freeze_indexed_revision FROM repo %s", querySuffix)
	rows, err := globalDB.Query(q.Query(sqlf.PostgresBindVar), q.Args()...)
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

// RepoListOp specifies the options for listing repositories.
//
// Query and IncludePatterns/ExcludePatterns may not be used together.
type RepoListOp struct {
	// Query specifies a search query for repositories. If specified, then the Sort and
	// Direction options are ignored
	Query string

	// IncludePatterns is a list of regular expressions, all of which must match all
	// repositories returned in the list.
	IncludePatterns []string

	// ExcludePattern is a regular expression that must not match any repository
	// returned in the list.
	ExcludePattern string

	sourcegraph.ListOptions
}

// makeFuzzyLikeRepoQuery turns a string of "foo/bar" into "%foo%/%bar%".
// Anything that is not a letter or digit is turned turned surrounded by %.
// Except for space, which is just turned into %.
func makeFuzzyLikeRepoQuery(q string) string {
	var last rune
	var b bytes.Buffer
	b.Grow(len(q) + 4) // most queries will add around 4 wildcards (prefix, postfix and around separator)
	writeRune := func(r rune) {
		if r == '%' && last == '%' {
			return
		}
		last = r
		b.WriteRune(r)
	}
	writeEscaped := func(r rune) {
		if last != '%' {
			b.WriteRune('%')
		}
		b.WriteRune('\\')
		b.WriteRune(r)
		b.WriteRune('%')
		last = '%'
	}

	writeRune('%') // prefix
	for _, r := range q {
		switch r {
		case ' ':
			// Ignore space, since repo URI can't contain it. Just add a wildcard
			writeRune('%')
		case '\\':
			writeEscaped(r)
		case '%':
			writeEscaped(r)
		case '_':
			writeEscaped(r)
		default:
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				writeRune(r)
			} else {
				writeRune('%')
				writeRune(r)
				writeRune('%')
			}
		}
	}
	writeRune('%') // postfix

	return b.String()
}

// List lists repositories in the Sourcegraph repository
//
// This will not return any repositories from external services that are not present in the Sourcegraph repository.
// The result list is unsorted and has a fixed maximum limit of 1000 items.
// Matching is done with fuzzy matching, i.e. "query" will match any repo URI that matches the regexp `q.*u.*e.*r.*y`
func (s *repos) List(ctx context.Context, opt *RepoListOp) ([]*sourcegraph.Repo, error) {
	if Mocks.Repos.List != nil {
		return Mocks.Repos.List(ctx, opt)
	}

	if opt == nil {
		opt = &RepoListOp{}
	}

	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if opt.Query != "" && (len(opt.IncludePatterns) > 0 || opt.ExcludePattern != "") {
		return nil, errors.New("Repos.List: Query and IncludePatterns/ExcludePattern options are mutually exclusive")
	}
	if opt.Query != "" {
		conds = append(conds, sqlf.Sprintf("lower(uri) LIKE %s", makeFuzzyLikeRepoQuery(strings.ToLower(opt.Query))))
	}
	for _, includePattern := range opt.IncludePatterns {
		exact, like, pattern, err := parseIncludePattern(includePattern)
		if err != nil {
			return nil, err
		}
		if exact != nil {
			if len(exact) == 0 || (len(exact) == 1 && exact[0] == "") {
				conds = append(conds, sqlf.Sprintf("FALSE"))
			} else {
				items := []*sqlf.Query{}
				for _, v := range exact {
					items = append(items, sqlf.Sprintf("%s", v))
				}
				conds = append(conds, sqlf.Sprintf("uri IN (%s)", sqlf.Join(items, ",")))
			}
		}
		if like != nil && len(like) > 0 {
			var likeConds []*sqlf.Query
			for _, v := range like {
				likeConds = append(likeConds, sqlf.Sprintf(`lower(uri) LIKE %s`, v))
			}
			conds = append(conds, sqlf.Join(likeConds, " OR "))
		}
		if pattern != "" {
			conds = append(conds, sqlf.Sprintf("lower(uri) ~* %s", pattern))
		}
	}
	if opt.ExcludePattern != "" {
		conds = append(conds, sqlf.Sprintf("lower(uri) !~* %s", opt.ExcludePattern))
	}

	// fetch matching repos unordered
	rawRepos, err := s.getBySQL(ctx, sqlf.Sprintf("WHERE %s LIMIT 1000", sqlf.Join(conds, "AND")))

	if err != nil {
		return nil, err
	}

	var repos []*sourcegraph.Repo
	if !feature.Features.Sep20Auth {
		// 🚨 SECURITY: It is very important that the input list of repos (rawRepos) 🚨
		// comes directly from the DB as verifyUserHasReadAccessAll relies directly
		// on the accuracy of the Repo.Private field.
		repos, err = verifyUserHasReadAccessAll(ctx, "Repos.List", rawRepos)
		if err != nil {
			return nil, err
		}
	} else {
		repos = rawRepos
	}

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

// parseIncludePattern either (1) parses the pattern into a list of exact possible
// string values and LIKE patterns if such a list can be determined from the pattern,
// and (2) returns the original regexp if those patterns are not equivalent to the
// regexp.
//
// It allows Repos.List to optimize for the common case where a pattern like
// `(^github.com/foo/bar$)|(^github.com/baz/qux$)` is provided. In that case,
// it's faster to query for "WHERE uri IN (...)" the two possible exact values
// (because it can use an index) instead of using a "WHERE uri ~*" regexp condition
// (which generally can't use an index).
//
// This optimization is necessary for good performance when there are many repos
// in the database. With this optimization, specifying a "repogroup:" in the query
// will be fast (even if there are many repos) because the query can be constrained
// efficiently to only the repos in the group.
func parseIncludePattern(pattern string) (exact, like []string, regexp string, err error) {
	re, err := regexpsyntax.Parse(pattern, regexpsyntax.OneLine)
	if err != nil {
		return nil, nil, "", err
	}
	exact, contains, prefix, suffix, err := allMatchingStrings(re.Simplify())
	if err != nil {
		return nil, nil, "", err
	}
	for _, v := range contains {
		like = append(like, "%"+v+"%")
	}
	for _, v := range prefix {
		like = append(like, v+"%")
	}
	for _, v := range suffix {
		like = append(like, "%"+v)
	}
	if exact != nil || like != nil {
		return exact, like, "", nil
	}
	return nil, nil, pattern, nil
}

// allMatchingStrings returns a complete list of the strings that re
// matches, if it's possible to determine the list.
func allMatchingStrings(re *regexpsyntax.Regexp) (exact, contains, prefix, suffix []string, err error) {
	prog, err := regexpsyntax.Compile(re)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	switch {
	case re.Op == regexpsyntax.OpEmptyMatch:
		return []string{""}, nil, nil, nil, nil
	case re.Op == regexpsyntax.OpLiteral:
		prefix, complete := prog.Prefix()
		if complete {
			return nil, []string{prefix}, nil, nil, nil
		}
		return nil, nil, nil, nil, nil

	case re.Op == regexpsyntax.OpCharClass:
		// Only handle simple case of one range.
		if len(re.Rune) == 2 {
			len := int(re.Rune[1] - re.Rune[0] + 1)
			if len > 26 {
				// Avoid large character ranges (which could blow up the number
				// of possible matches).
				return nil, nil, nil, nil, nil
			}
			chars := make([]string, len)
			for r := re.Rune[0]; r <= re.Rune[1]; r++ {
				chars[r-re.Rune[0]] = string(r)
			}
			return nil, chars, nil, nil, nil
		}
		return nil, nil, nil, nil, nil

	case re.Op == regexpsyntax.OpBeginText:
		return nil, nil, []string{""}, nil, nil

	case re.Op == regexpsyntax.OpEndText:
		return nil, nil, nil, []string{""}, nil

	case re.Op == regexpsyntax.OpCapture:
		return allMatchingStrings(re.Sub0[0])

	case re.Op == regexpsyntax.OpConcat:
		var begin, end bool
		for i, sub := range re.Sub {
			if sub.Op == regexpsyntax.OpBeginText && i == 0 {
				begin = true
				continue
			}
			if sub.Op == regexpsyntax.OpEndText && i == len(re.Sub)-1 {
				end = true
				continue
			}
			subexact, subcontains, subprefix, subsuffix, err := allMatchingStrings(sub)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			if subexact == nil && subcontains == nil && subprefix == nil && subsuffix == nil {
				return nil, nil, nil, nil, nil
			}

			if subexact == nil {
				subexact = subcontains
			}
			if exact == nil {
				exact = subexact
			} else {
				size := len(exact) * len(subexact)
				if len(subexact) > 4 || size > 30 {
					// Avoid blowup in number of possible matches.
					return nil, nil, nil, nil, nil
				}
				combined := make([]string, 0, size)
				for _, match := range exact {
					for _, submatch := range subexact {
						combined = append(combined, match+submatch)
					}
				}
				exact = combined
			}
		}
		if exact == nil {
			exact = []string{""}
		}
		if begin && end {
			return exact, nil, nil, nil, nil
		} else if begin {
			return nil, nil, exact, nil, nil
		} else if end {
			return nil, nil, nil, exact, nil
		}
		return nil, exact, nil, nil, nil

	case re.Op == regexpsyntax.OpAlternate:
		for _, sub := range re.Sub {
			subexact, subcontains, subprefix, subsuffix, err := allMatchingStrings(sub)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			exact = append(exact, subexact...)
			contains = append(contains, subcontains...)
			prefix = append(prefix, subprefix...)
			suffix = append(suffix, subsuffix...)
		}
		return exact, contains, prefix, suffix, nil
	}

	return nil, nil, nil, nil, nil
}

func (s *repos) Delete(ctx context.Context, repo int32) error {
	if Mocks.Repos.Delete != nil {
		return Mocks.Repos.Delete(ctx, repo)
	}

	q := sqlf.Sprintf("DELETE FROM REPO WHERE id=%d", repo)
	_, err := globalDB.Exec(q.Query(sqlf.PostgresBindVar), q.Args()...)
	return err
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

	var updates []*sqlf.Query
	if ghrepo.Description != repo.Description {
		updates = append(updates, sqlf.Sprintf("description=%s", ghrepo.Description))
	}
	if ghrepo.HomepageURL != repo.HomepageURL {
		updates = append(updates, sqlf.Sprintf("homepage_url=%s", ghrepo.HomepageURL))
	}
	if ghrepo.DefaultBranch != repo.DefaultBranch {
		updates = append(updates, sqlf.Sprintf("default_branch=%s", ghrepo.DefaultBranch))
	}
	if ghrepo.Private != repo.Private {
		updates = append(updates, sqlf.Sprintf("private=%v", ghrepo.Private))
	}

	if !timestampEqual(repo.UpdatedAt, ghrepo.UpdatedAt) {
		updates = append(updates, sqlf.Sprintf("updated_at=%s", ghrepo.UpdatedAt))
	}
	if !timestampEqual(repo.PushedAt, ghrepo.PushedAt) {
		updates = append(updates, sqlf.Sprintf("pushed_at=%s", ghrepo.PushedAt))
	}

	if len(updates) > 0 {
		q := sqlf.Sprintf("UPDATE repo SET %s WHERE id=%d", sqlf.Join(updates, ","), repo.ID)
		if _, err := globalDB.Exec(q.Query(sqlf.PostgresBindVar), q.Args()...); err != nil {
			return err
		}
	}

	return nil
}

func (s *repos) UpdateLanguage(ctx context.Context, repoID int32, language string) error {
	_, err := globalDB.Exec("UPDATE repo SET language=$1 WHERE id=$2", language, repoID)
	return err
}

func (s *repos) UpdateIndexedRevision(ctx context.Context, repoID int32, rev string) error {
	_, err := globalDB.Exec("UPDATE repo SET indexed_revision=$1 WHERE id=$2", rev, repoID)
	return err
}

// TryInsertNew attempts to insert the repository rp into the db. It returns no error if a repo
// with the given uri already exists.
func (s *repos) TryInsertNew(ctx context.Context, uri string, description string, fork bool, private bool) error {
	_, err := globalDB.Exec("INSERT INTO repo (uri, description, fork, private, created_at, vcs, default_branch, homepage_url, language, blocked) VALUES ($1, $2, $3, $4, $5, '', '', '', '', false)", uri, description, fork, private, time.Now()) // FIXME: bad DB schema: nullable columns
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
