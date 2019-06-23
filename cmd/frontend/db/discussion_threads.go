package db

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/felixfbecker/stringscore"
	"github.com/karrick/tparse"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/discussions/searchquery"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

// TODO(slimsag:discussions): future: tests for DiscussionThreadsListOptions.TargetRepoID
// TODO(slimsag:discussions): future: tests for DiscussionThreadsListOptions.TargetRepoPath

// discussionThreads provides access to the `discussion_threads*` tables.
//
// For a detailed overview of the schema, see schema.md.
type discussionThreads struct{}

// ErrThreadNotFound is the error returned by Discussions methods to indicate
// that the thread could not be found.
type ErrThreadNotFound struct {
	// ThreadID is the thread that was not found.
	ThreadID int64
}

func (e *ErrThreadNotFound) Error() string {
	return fmt.Sprintf("thread %d not found", e.ThreadID)
}

func (t *discussionThreads) Create(ctx context.Context, newThread *types.DiscussionThread) (*types.DiscussionThread, error) {
	if Mocks.DiscussionThreads.Create != nil {
		return Mocks.DiscussionThreads.Create(ctx, newThread)
	}

	// Validate the input thread.
	if newThread == nil {
		return nil, errors.New("newThread is nil")
	}
	if newThread.ID != 0 {
		return nil, errors.New("newThread.ID must be zero")
	}
	if newThread.ProjectID == 0 {
		return nil, errors.New("newThread.ProjectID must be set")
	}
	if strings.TrimSpace(newThread.Title) == "" {
		return nil, errors.New("newThread.Title must be present (and not whitespace)")
	}
	if len([]rune(newThread.Title)) > 500 {
		return nil, errors.New("newThread.Title too long (must be less than 500 UTF-8 characters)")
	}
	if !newThread.CreatedAt.IsZero() {
		return nil, errors.New("newThread.CreatedAt must not be specified")
	}
	if newThread.ArchivedAt != nil {
		return nil, errors.New("newThread.ArchivedAt must not be specified")
	}
	if !newThread.UpdatedAt.IsZero() {
		return nil, errors.New("newThread.UpdatedAt must not be specified")
	}
	if newThread.DeletedAt != nil {
		return nil, errors.New("newThread.DeletedAt must not be specified")
	}

	// TODO(slimsag:discussions): should be in a transaction

	// First, create the thread itself. Initially it will have no target.
	newThread.CreatedAt = time.Now()
	newThread.UpdatedAt = newThread.CreatedAt
	err := dbconn.Global.QueryRowContext(ctx, `INSERT INTO discussion_threads(
		project_id,
		author_user_id,
		title,
		settings,
		type,
		status,
		created_at,
		updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7,   $8) RETURNING id`,
		newThread.ProjectID,
		newThread.AuthorUserID,
		newThread.Title,
		newThread.Settings,
		newThread.Type,
		newThread.Status,
		newThread.CreatedAt,
		newThread.UpdatedAt,
	).Scan(&newThread.ID)
	if err != nil {
		return nil, errors.Wrap(err, "create thread")
	}
	return newThread, nil
}

func (t *discussionThreads) Get(ctx context.Context, threadID int64) (*types.DiscussionThread, error) {
	if Mocks.DiscussionThreads.Get != nil {
		return Mocks.DiscussionThreads.Get(threadID)
	}

	threads, err := t.List(ctx, &DiscussionThreadsListOptions{
		ThreadIDs: []int64{threadID},
	})
	if err != nil {
		return nil, err
	}
	if len(threads) == 0 {
		return nil, &ErrThreadNotFound{ThreadID: threadID}
	}
	return threads[0], nil
}

type DiscussionThreadsUpdateOptions struct {
	// Title, when non-nil, updates the thread's title.
	Title *string

	// Settings, when non-nil, updates the thread's settings.
	Settings *string

	// Active, when non-nil, specifies whether the check is active or not.
	Active *bool

	// Archive, when non-nil, specifies whether the thread is archived or not.
	Archive *bool

	// Delete, when true, specifies that the thread should be deleted. This
	// operation cannot be undone.
	Delete bool
}

func (t *discussionThreads) Update(ctx context.Context, threadID int64, opts *DiscussionThreadsUpdateOptions) (*types.DiscussionThread, error) {
	if Mocks.DiscussionThreads.Update != nil {
		return Mocks.DiscussionThreads.Update(ctx, threadID, opts)
	}
	if opts == nil {
		return nil, errors.New("options must not be nil")
	}
	now := time.Now()

	// TODO(slimsag:discussions): should be in a transaction

	anyUpdate := false
	if opts.Title != nil {
		anyUpdate = true
		if _, err := dbconn.Global.ExecContext(ctx, "UPDATE discussion_threads SET title=$1 WHERE id=$2 AND deleted_at IS NULL", opts.Title, threadID); err != nil {
			return nil, err
		}
	}
	if opts.Settings != nil {
		anyUpdate = true
		if _, err := dbconn.Global.ExecContext(ctx, "UPDATE discussion_threads SET settings=$1 WHERE id=$2 AND deleted_at IS NULL", opts.Settings, threadID); err != nil {
			return nil, err
		}
	}
	if opts.Active != nil {
		anyUpdate = true
		if _, err := dbconn.Global.ExecContext(ctx, "UPDATE discussion_threads SET status=$1 WHERE id=$2 AND deleted_at IS NULL", *opts.Active, threadID); err != nil {
			return nil, err
		}
	}
	if opts.Archive != nil {
		anyUpdate = true
		var archivedAt *time.Time
		if *opts.Archive {
			archivedAt = &now
		}
		if _, err := dbconn.Global.ExecContext(ctx, "UPDATE discussion_threads SET archived_at=$1 WHERE id=$2 AND deleted_at IS NULL", archivedAt, threadID); err != nil {
			return nil, err
		}
	}
	if opts.Delete {
		anyUpdate = true
		if _, err := dbconn.Global.ExecContext(ctx, "UPDATE discussion_threads SET deleted_at=$1 WHERE id=$2 AND deleted_at IS NULL", now, threadID); err != nil {
			return nil, err
		}

		// Mark all comments in the thread as deleted.
		comments, err := DiscussionComments.List(ctx, &DiscussionCommentsListOptions{
			ThreadID: &threadID,
		})
		if err != nil {
			return nil, err
		}
		for _, comment := range comments {
			_, err := DiscussionComments.Update(ctx, comment.ID, &DiscussionCommentsUpdateOptions{Delete: true, noThreadDelete: true})
			if err != nil {
				return nil, err
			}
		}
	}
	if anyUpdate {
		if _, err := dbconn.Global.ExecContext(ctx, "UPDATE discussion_threads SET updated_at=$1 WHERE id=$2 AND deleted_at IS NULL", now, threadID); err != nil {
			return nil, err
		}
	}
	if opts.Delete {
		return nil, nil
	}
	return t.Get(ctx, threadID)
}

type DiscussionThreadsListOptions struct {
	// LimitOffset specifies SQL LIMIT and OFFSET counts. It may be nil (no limit / offset).
	*LimitOffset

	// IsOpen, when non-nil, specifies that only threads that are open (true) or closed (false)
	// should be returned.
	IsOpen *bool

	// Type, when non-zero, specifies that only threads of the given type should be returned.
	Type types.ThreadType

	// Status, when non-zero, specifies that only threads with the given status should be returned.
	Status types.ThreadStatus

	// TitleQuery, when non-nil, specifies that only threads whose title
	// matches this string should be returned.
	TitleQuery    *string
	NotTitleQuery *string

	// ThreadIDs, when len() > 0, specifies that only the thread with one of
	// these IDs should be returned. See also DiscussionThreads.Get.
	ThreadIDs    []int64
	NotThreadIDs []int64

	// AuthorUserID, when len() > 0, specifies that only threads made by this
	// author should be returned.
	AuthorUserIDs    []int32
	NotAuthorUserIDs []int32

	// TargetRepoID, when non-nil, specifies that only threads that have a repo target and
	// this repo ID should be returned.
	TargetRepoID    *api.RepoID
	NotTargetRepoID *api.RepoID

	// TargetRepoPath, when non-nil, specifies that only threads that have a repo target
	// and this path should be returned.
	TargetRepoPath    *string
	NotTargetRepoPath *string

	// CreatedBefore, when non-nil, specifies that only threads that were
	// created before this time should be returned.
	CreatedBefore *time.Time
	CreatedAfter  *time.Time

	// Whether or not to return results in ascending (oldest first) order. When
	// false, descending (latest first) order is used.
	AscendingOrder bool

	// Reported, when true, specifies that only threads with at least one
	// reported comment should be returned.
	Reported bool
}

// SetFromQuery sets the options based on the search query string.
func (opts *DiscussionThreadsListOptions) SetFromQuery(ctx context.Context, query string) {
	userList := func(value string) (users []*types.User) {
		for _, username := range strings.Fields(value) {
			username = strings.TrimSpace(strings.TrimPrefix(username, "@"))
			user, err := Users.GetByUsername(ctx, username)
			if err != nil {
				continue
			}
			users = append(users, user)
		}
		return
	}
	userIDsList := func(value string) (users []int32) {
		for _, user := range userList(value) {
			users = append(users, user.ID)
		}
		return
	}

	findInvolvedThreadIDs := func(value string) (threadIDs []int64) {
		set := map[int64]struct{}{}
		for _, user := range userList(value) {
			comments, err := DiscussionComments.List(ctx, &DiscussionCommentsListOptions{
				AuthorUserID: &user.ID,
			})
			if err != nil {
				continue
			}
			for _, comment := range comments {
				if _, ok := set[comment.ThreadID]; !ok {
					set[comment.ThreadID] = struct{}{}
					threadIDs = append(threadIDs, comment.ThreadID)
				}
			}
		}
		return
	}

	parseTimeOrDuration := func(value string) *time.Time {
		// Try parsing as RFC3339 / ISO 8601 first.
		t, err := time.Parse(time.RFC3339, value)
		if err == nil {
			return &t
		}

		// Try parsing as a relative duration, e.g. "3d ago", "3h4m", etc.
		value = strings.TrimSuffix(value, " ago")
		t, err = tparse.ParseNow(time.RFC3339, "now-"+value)
		if err != nil {
			return nil
		}
		return &t
	}

	var reported bool
	operators := map[string]func(value string){
		// syntax: "is:open"/"is:closed"/"is:thread"
		"is": func(value string) {
			if t := strings.ToUpper(value); types.IsValidThreadType(t) {
				opts.Type = types.ThreadType(t)
			}
			if s := strings.ToUpper(value); types.IsValidThreadStatus(s) {
				opts.Status = types.ThreadStatus(s)
			}
			if s := strings.ToLower(value); s == "open" || s == "active" {
				opts.Status = types.ThreadStatusOpenActive
			}
		},

		// syntax: `title:"some title"` or "title:sometitle"
		// Primarily exists for the negation mode.
		"title": func(value string) {
			opts.TitleQuery = &value
		},
		"-title": func(value string) {
			opts.NotTitleQuery = &value
		},

		// syntax: "involves:slimsag" or "involves:@slimsag" or "involves:slimsag @jack"
		"involves": func(value string) {
			opts.ThreadIDs = append(opts.ThreadIDs, findInvolvedThreadIDs(value)...)
			if len(opts.ThreadIDs) == 0 {
				opts.ThreadIDs = []int64{-1}
			}
		},
		"-involves": func(value string) {
			opts.NotThreadIDs = append(opts.NotThreadIDs, findInvolvedThreadIDs(value)...)
		},

		// syntax: "author:slimsag" or "author:@slimsag" or `author:"slimsag @jack"`
		"author": func(value string) {
			opts.AuthorUserIDs = userIDsList(value)
			if len(opts.AuthorUserIDs) == 0 {
				opts.AuthorUserIDs = []int32{-1}
			}
		},
		"-author": func(value string) {
			opts.NotAuthorUserIDs = userIDsList(value)
		},

		// syntax: "repo:github.com/gorilla/mux" or "repo:some/repo"
		// TODO(slimsag:discussions): support list syntax here.
		"repo": func(value string) {
			repo, err := Repos.GetByName(ctx, api.RepoName(value))
			if err != nil {
				tmp := api.RepoID(-1)
				opts.TargetRepoID = &tmp
				return
			}
			opts.TargetRepoID = &repo.ID
		},
		"-repo": func(value string) {
			repo, err := Repos.GetByName(ctx, api.RepoName(value))
			if err != nil {
				return
			}
			opts.NotTargetRepoID = &repo.ID
		},

		// syntax: "file:dir/file.go" or "file:something.go"
		// TODO(slimsag:discussions): support list syntax here.
		"file": func(value string) {
			opts.TargetRepoPath = &value
		},
		"-file": func(value string) {
			opts.NotTargetRepoPath = &value
		},

		// syntax: "file:dir/file.go" or "file:something.go"
		"before": func(value string) {
			opts.CreatedBefore = parseTimeOrDuration(value)
		},
		"after": func(value string) {
			opts.CreatedAfter = parseTimeOrDuration(value)
		},

		// syntax: "order:oldest" OR "order:ascending" etc.
		"order": func(value string) {
			value = strings.ToLower(value)
			opts.AscendingOrder = value == "oldest" || value == "oldest-first" || value == "asc" || value == "ascending"
		},

		"reported": func(value string) {
			reported, _ = strconv.ParseBool(value)
		},
	}
	remaining, operations := searchquery.Parse(query)
	for _, operation := range operations {
		operation, value := operation[0], operation[1]
		if handler, ok := operators[operation]; ok {
			handler(value)
			continue
		}
		// Since we don't have an operator for this, consider it part of
		// the remaining search query.
		remaining = strings.Join([]string{remaining, operation + ":" + value}, " ")
	}
	opts.TitleQuery = &remaining

	if reported {
		// Searching only for reported threads.
		if len(opts.ThreadIDs) > 0 {
			// Already have a list of threads we're interested in, e.g. from `involves:slimsag`.
			// Narrow the list down.
			var newThreads []int64
			for _, threadID := range opts.ThreadIDs {
				reportedComments, err := DiscussionComments.Count(ctx, &DiscussionCommentsListOptions{
					ThreadID: &threadID,
					Reported: true,
				})
				if err != nil {
					continue
				}
				if reportedComments == 0 {
					continue
				}
				newThreads = append(newThreads, threadID)
			}
			opts.ThreadIDs = newThreads
			if len(opts.ThreadIDs) == 0 {
				opts.ThreadIDs = []int64{-1}
			}
		} else {
			// We don't have an existing list of threads we're interested in.
			// Compile it now.
			comments, _ := DiscussionComments.List(ctx, &DiscussionCommentsListOptions{
				Reported: true,
			})
			set := map[int64]struct{}{}
			for _, comment := range comments {
				if _, ok := set[comment.ThreadID]; ok {
					continue
				}
				set[comment.ThreadID] = struct{}{}
				opts.ThreadIDs = append(opts.ThreadIDs, comment.ThreadID)
			}
			if len(opts.ThreadIDs) == 0 {
				opts.ThreadIDs = []int64{-1}
			}
		}
	}
}

func (t *discussionThreads) List(ctx context.Context, opts *DiscussionThreadsListOptions) ([]*types.DiscussionThread, error) {
	if Mocks.DiscussionThreads.List != nil {
		return Mocks.DiscussionThreads.List(ctx, opts)
	}
	if opts == nil {
		return nil, errors.New("options must not be nil")
	}
	conds := t.getListSQL(opts)
	order := "DESC"
	if opts.AscendingOrder {
		order = "ASC"
	}
	q := sqlf.Sprintf("WHERE %s ORDER BY id "+order+" %s", sqlf.Join(conds, "AND"), opts.LimitOffset.SQL())

	threads, err := t.getBySQL(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}
	return t.fuzzyFilterThreads(opts, threads), nil
}

func (t *discussionThreads) Count(ctx context.Context, opts *DiscussionThreadsListOptions) (int, error) {
	if Mocks.DiscussionThreads.Count != nil {
		return Mocks.DiscussionThreads.Count(ctx, opts)
	}
	if opts == nil {
		return 0, errors.New("options must not be nil")
	}
	if opts.TitleQuery != nil {
		// TitleQuery requires post-query filtering (we must grab at least the
		// title of the thread). So we take the easy way out here and just
		// actually determine the results to find the count.
		threads, err := t.List(ctx, opts)
		return len(threads), err
	}
	conds := t.getListSQL(opts)
	q := sqlf.Sprintf("WHERE %s", sqlf.Join(conds, "AND"))
	return t.getCountBySQL(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
}

func (t *discussionThreads) fuzzyFilterThreads(opts *DiscussionThreadsListOptions, threads []*types.DiscussionThread) []*types.DiscussionThread {
	if opts.TitleQuery != nil && strings.TrimSpace(*opts.TitleQuery) != "" {
		var (
			scoresByThread  = make(map[*types.DiscussionThread]int, len(threads))
			threadsToRemove []*types.DiscussionThread
		)
		for _, t := range threads {
			score := stringscore.Score(t.Title, *opts.TitleQuery)
			if score > 0 {
				scoresByThread[t] = score
			} else {
				threadsToRemove = append(threadsToRemove, t)
			}
		}
		for _, rm := range threadsToRemove {
			for i, t := range threads {
				if t == rm {
					threads = append(threads[:i], threads[i+1:]...)
					break
				}
			}
		}

		// TODO(slimsag:discussions): future: whether or not to sort based on
		// best match here should be optional.
		sort.Slice(threads, func(i, j int) bool {
			return scoresByThread[threads[i]] > scoresByThread[threads[j]]
		})
	}
	return threads
}

func (*discussionThreads) getListSQL(opts *DiscussionThreadsListOptions) (conds []*sqlf.Query) {
	conds = []*sqlf.Query{sqlf.Sprintf("TRUE")}
	conds = append(conds, sqlf.Sprintf("deleted_at IS NULL"))
	if opts.Status != "" {
		conds = append(conds, sqlf.Sprintf("status = %s", opts.Status))
	}
	if opts.Type != "" {
		conds = append(conds, sqlf.Sprintf("type = %s", opts.Type))
	}
	if opts.TitleQuery != nil && strings.TrimSpace(*opts.TitleQuery) != "" {
		conds = append(conds, sqlf.Sprintf("title ILIKE %v", extraFuzzy(*opts.TitleQuery)))
	}
	if opts.NotTitleQuery != nil && strings.TrimSpace(*opts.NotTitleQuery) != "" {
		// Using extraFuzzy here would exclude too many results, so instead we
		// just do prefix/suffix fuzziness for now.
		conds = append(conds, sqlf.Sprintf("title NOT ILIKE %v", "%"+*opts.NotTitleQuery+"%"))
	}
	if len(opts.ThreadIDs) > 0 {
		conds = append(conds, sqlf.Sprintf("id = ANY(%v)", pq.Array(opts.ThreadIDs)))
	}
	if len(opts.NotThreadIDs) > 0 {
		conds = append(conds, sqlf.Sprintf("id != ANY(%v)", pq.Array(opts.NotThreadIDs)))
	}
	if len(opts.AuthorUserIDs) > 0 {
		conds = append(conds, sqlf.Sprintf("author_user_id = ANY(%v)", pq.Array(opts.AuthorUserIDs)))
	}
	if len(opts.NotAuthorUserIDs) > 0 {
		conds = append(conds, sqlf.Sprintf("author_user_id != ANY(%v)", pq.Array(opts.NotAuthorUserIDs)))
	}
	if opts.CreatedBefore != nil {
		conds = append(conds, sqlf.Sprintf("created_at < %v", *opts.CreatedBefore))
	}
	if opts.CreatedAfter != nil {
		conds = append(conds, sqlf.Sprintf("created_at > %v", *opts.CreatedAfter))
	}

	if opts.TargetRepoID != nil || opts.TargetRepoPath != nil || opts.NotTargetRepoID != nil || opts.NotTargetRepoPath != nil {
		targetRepoConds := []*sqlf.Query{}
		if opts.TargetRepoID != nil {
			targetRepoConds = append(targetRepoConds, sqlf.Sprintf("repo_id = %v", *opts.TargetRepoID))
		}
		if opts.NotTargetRepoID != nil {
			targetRepoConds = append(targetRepoConds, sqlf.Sprintf("repo_id != %v", *opts.NotTargetRepoID))
		}
		if opts.TargetRepoPath != nil {
			if strings.HasSuffix(*opts.TargetRepoPath, "/**") {
				match := strings.TrimSuffix(*opts.TargetRepoPath, "/**") + "%"
				targetRepoConds = append(targetRepoConds, sqlf.Sprintf("path LIKE %v", match))
			} else {
				targetRepoConds = append(targetRepoConds, sqlf.Sprintf("path=%v", *opts.TargetRepoPath))
			}
		}
		if opts.NotTargetRepoPath != nil {
			if strings.HasSuffix(*opts.NotTargetRepoPath, "/**") {
				match := strings.TrimSuffix(*opts.NotTargetRepoPath, "/**") + "%"
				targetRepoConds = append(targetRepoConds, sqlf.Sprintf("path NOT LIKE %v", match))
			} else {
				targetRepoConds = append(targetRepoConds, sqlf.Sprintf("path!=%v", *opts.NotTargetRepoPath))
			}
		}
		conds = append(conds, sqlf.Sprintf("id IN (SELECT thread_id FROM discussion_threads_target_repo WHERE %v)", sqlf.Join(targetRepoConds, "AND")))
	}
	return conds
}

func (*discussionThreads) getCountBySQL(ctx context.Context, query string, args ...interface{}) (int, error) {
	var count int
	rows := dbconn.Global.QueryRowContext(ctx, "SELECT count(id) FROM discussion_threads t "+query, args...)
	err := rows.Scan(&count)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return count, err
}

// getBySQL returns threads matching the SQL query, if any exist.
func (t *discussionThreads) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*types.DiscussionThread, error) {
	rows, err := dbconn.Global.QueryContext(ctx, `
		SELECT
			t.id,
			t.project_id,
			t.author_user_id,
			t.title,
			t.settings,
			t.type,
			t.status,
			t.created_at,
			t.archived_at,
			t.updated_at
		FROM discussion_threads t `+query, args...)
	if err != nil {
		return nil, err
	}

	threads := []*types.DiscussionThread{}
	defer rows.Close()
	for rows.Next() {
		var thread types.DiscussionThread
		err := rows.Scan(
			&thread.ID,
			&thread.ProjectID,
			&thread.AuthorUserID,
			&thread.Title,
			&thread.Settings,
			&thread.Type,
			&thread.Status,
			&thread.CreatedAt,
			&thread.ArchivedAt,
			&thread.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		threads = append(threads, &thread)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return threads, nil
}

// extraFuzzy turns a string like "cat" into "%c%a%t%". It can be used with a
// LIKE query to filter out results that cannot possibly match a fuzzy search
// query. This returns 'extra fuzzy' results, which are usually subsequently
// filtered in Go using github.com/felixfbecker/stringscore.
func extraFuzzy(s string) string {
	if strings.TrimSpace(s) == "" {
		return ""
	}
	input := []rune(s)

	result := make([]rune, 0, 1+(len(input)*2))
	result = append(result, '%')
	for _, r := range input {
		result = append(result, r)
		result = append(result, '%')
	}
	return string(result)
}
