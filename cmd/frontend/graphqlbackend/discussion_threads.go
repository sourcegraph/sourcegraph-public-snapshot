package graphqlbackend

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/discussions"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/discussions/ratelimit"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
)

type discussionsMutationResolver struct {
}

type discussionThreadTargetRepoSelectionInput struct {
	StartLine      int32
	StartCharacter int32
	EndLine        int32
	EndCharacter   int32
	LinesBefore    *[]string
	Lines          *[]string
	LinesAfter     *[]string
}

// discussionsResolveRepository resolves the repository given an ID, name, or
// git clone URL. Only one must be specified, or else this function will panic.
func discussionsResolveRepository(ctx context.Context, id *graphql.ID, name, gitCloneURL *string) (*repositoryResolver, error) {
	switch {
	case id != nil:
		return repositoryByID(ctx, *id)
	case name != nil:
		repo, err := backend.Repos.GetByName(ctx, api.RepoName(*name))
		if err != nil {
			return nil, err
		}
		return repositoryByIDInt32(ctx, repo.ID)
	case gitCloneURL != nil:
		repositoryName, err := cloneURLToRepoName(ctx, *gitCloneURL)
		if err != nil {
			return nil, err
		}
		repo, err := backend.Repos.GetByName(ctx, api.RepoName(repositoryName))
		if err != nil {
			return nil, err
		}
		return repositoryByIDInt32(ctx, repo.ID)
	default:
		panic("invalid state")
	}
}

type discussionThreadTargetRepoInput struct {
	RepositoryID          *graphql.ID
	RepositoryName        *string
	RepositoryGitCloneURL *string
	Path                  *string
	Branch                *string
	Revision              *string
	Selection             *discussionThreadTargetRepoSelectionInput
}

func (d *discussionThreadTargetRepoInput) convert(ctx context.Context) (*types.DiscussionThreadTargetRepo, error) {
	count := 0
	if d.RepositoryID != nil {
		count++
	}
	if d.RepositoryName != nil {
		count++
	}
	if d.RepositoryGitCloneURL != nil {
		count++
	}
	if count != 1 {
		return nil, errors.New("exactly one of repositoryID, repositoryName, or repositoryGitCloneURL must be specified")
	}
	repo, err := discussionsResolveRepository(ctx, d.RepositoryID, d.RepositoryName, d.RepositoryGitCloneURL)
	if err != nil {
		return nil, err
	}
	tr := &types.DiscussionThreadTargetRepo{
		RepoID:   repo.repo.ID,
		Path:     d.Path,
		Branch:   d.Branch,
		Revision: d.Revision,
	}
	if d.Selection != nil {
		tr.StartLine = &d.Selection.StartLine
		tr.EndLine = &d.Selection.EndLine
		tr.StartCharacter = &d.Selection.StartCharacter
		tr.EndCharacter = &d.Selection.EndCharacter

		if d.Selection.Lines == nil {
			// The caller wishes for us to populate the lines using repository
			// data. We do this now.
			if err := d.populateLinesFromRepository(ctx, repo); err != nil {
				return nil, err
			}
		}
		tr.LinesBefore = d.Selection.LinesBefore
		tr.Lines = d.Selection.Lines
		tr.LinesAfter = d.Selection.LinesAfter
	}
	return tr, nil
}

// validate checks the validity of the input and returns an error, if any.
func (d *discussionThreadTargetRepoInput) validate() error {
	if d.Selection != nil {
		// Check that the caller either specified all line fields or didn't specify
		// any at all (specifying some but not others makes no sense, see the
		// schema for details).
		equal := func(a, b, c bool) bool {
			return a != b || b != c
		}
		if ds := d.Selection; equal(ds.LinesBefore != nil, ds.Lines != nil, ds.LinesAfter != nil) {
			return errors.New("DiscussionThreadTargetRepoSelectionInput: linesBefore, lines, and linesAfter must all be null or non-null (not mixed)")
		}
		if d.Selection.Lines == nil {
			if d.Path == nil {
				return errors.New("DiscussionThreadTargetRepoSelectionInput: when lines are null, path field must be specified")
			}
			if d.Branch == nil && d.Revision == nil {
				return errors.New("DiscussionThreadTargetRepoSelectionInput: when lines are null, branch or revision field must be specified")
			}
		}
	}
	return nil
}

// populateLinesFromRepository populates the d.LinesBefore, d.Lines and
// d.LinesAfter fields by pulling the information directly from the repository.
//
// Precondition: d.Selection != nil && d.validate() == nil
func (d *discussionThreadTargetRepoInput) populateLinesFromRepository(ctx context.Context, repo *repositoryResolver) error {
	if d.Selection == nil {
		panic("precondition failed")
	}

	// First we must get the commit resolver with whichever revision is more
	// precise (branches can change revisions).
	var rev string
	if d.Revision != nil {
		rev = *d.Revision
	} else if d.Branch != nil {
		rev = *d.Branch
	} else {
		panic("precondition failed (protected by validation)")
	}
	commit, err := repo.Commit(ctx, &repositoryCommitArgs{Rev: rev})
	if err != nil {
		return err
	}

	// Now we can actually get the file content.
	if d.Path == nil {
		panic("precondition failed (protected by validation)")
	}
	blob, err := commit.Blob(ctx, &struct{ Path string }{Path: *d.Path})
	if err != nil {
		return err
	}
	fileContent, err := blob.Content(ctx)
	if err != nil {
		return err
	}

	// Grab the lines for the selection, populate the struct, and we're finished.
	linesBefore, lines, linesAfter := discussions.LinesForSelection(fileContent, discussions.LineRange{
		StartLine: int(d.Selection.StartLine),
		EndLine:   int(d.Selection.EndLine),
	})
	d.Selection.LinesBefore = &linesBefore
	d.Selection.Lines = &lines
	d.Selection.LinesAfter = &linesAfter
	return nil
}

func (r *discussionsMutationResolver) CreateThread(ctx context.Context, args *struct {
	Input *struct {
		Project  graphql.ID
		Title    *string
		Contents string
		Target   *discussionThreadTargetInput
		Settings *string
		Type     types.ThreadType
		Status   *types.ThreadStatus
	}
}) (*discussionThreadResolver, error) {
	if args.Input.Title == nil {
		// Title defaults to first line of contents.
		title := strings.TrimSpace(strings.SplitN(strings.TrimSpace(args.Input.Contents), "\n", 2)[0])
		args.Input.Title = &title
	}
	// Validate the target input.
	if args.Input.Target != nil {
		if err := args.Input.Target.validate(); err != nil {
			return nil, err
		}
	}

	// 🚨 SECURITY: Only signed in users with a verified email may add comments
	// to a discussion thread.
	//
	// The verified email requirement for public instances is a security
	// measure to prevent spam. For private instances, it is a UX feature
	// (because we would not be able to send the author of this comment email
	// notifications anyway).
	currentUser, err := checkSignedInAndEmailVerified(ctx)
	if err != nil {
		return nil, err
	}
	if currentUser == nil {
		return nil, errors.New("no current user")
	}
	if dc := conf.Get().Discussions; dc != nil && dc.AbuseProtection {
		if mustWait := ratelimit.TimeUntilUserCanCreateThread(ctx, currentUser.user.ID, *args.Input.Title, args.Input.Contents); mustWait != 0 {
			return nil, fmt.Errorf("You are creating threads too quickly. You may create a new one after %v", mustWait.Round(time.Second))
		}
	}

	// Apply default status.
	var status types.ThreadStatus
	switch {
	case args.Input.Status != nil:
		status = *args.Input.Status
	case args.Input.Type == types.ThreadTypeThread || args.Input.Type == types.ThreadTypeChangeset:
		status = types.ThreadStatusOpenActive
	case args.Input.Type == types.ThreadTypeCheck:
		status = types.ThreadStatusInactive
	default:
		return nil, fmt.Errorf("unexpected thread type %q", args.Input.Type)
	}

	// 🚨 SECURITY: Ensure the viewer can view the project (which is a requirement for creating a
	// thread in it).
	project, err := ProjectByID(ctx, args.Input.Project)
	if err != nil {
		return nil, err
	}

	// Create the thread.
	newThread := &types.DiscussionThread{
		AuthorUserID: currentUser.user.ID,
		ProjectID:    project.DBID(),
		Title:        *args.Input.Title,
		Settings:     args.Input.Settings,
		Type:         args.Input.Type,
		Status:       status,
	}
	thread, err := db.DiscussionThreads.Create(ctx, newThread)
	if err != nil {
		return nil, errors.Wrap(err, "DiscussionThreads.Create")
	}

	// Add the target, if any.
	if args.Input.Target != nil {
		target, err := args.Input.Target.validateAndGetTarget(ctx)
		if err != nil {
			return nil, err
		}
		target.ThreadID = thread.ID
		if _, err := db.DiscussionThreads.AddTarget(ctx, target); err != nil {
			return nil, err
		}
	}

	// Create the first comment in the thread.
	newComment := &types.DiscussionComment{
		ThreadID:     newThread.ID,
		AuthorUserID: currentUser.user.ID,
		Contents:     args.Input.Contents,
	}
	_, err = db.DiscussionComments.Create(ctx, newComment)
	if err != nil {
		return nil, errors.Wrap(err, "DiscussionComments.Create")
	}
	discussions.NotifyNewThread(newThread, newComment)
	return &discussionThreadResolver{t: thread}, nil
}

func (r *discussionsMutationResolver) UpdateThread(ctx context.Context, args *struct {
	Input *struct {
		ThreadID graphql.ID
		Title    *string
		Settings *string
		Archive  *bool
		Active   *bool
		Delete   *bool
	}
}) (*discussionThreadResolver, error) {
	// 🚨 SECURITY: Only signed in users may update a discussion thread.
	currentUser, err := CurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	if currentUser == nil {
		return nil, errors.New("no current user")
	}

	var delete bool
	if args.Input.Delete != nil && *args.Input.Delete {
		// 🚨 SECURITY: Only site admins can delete discussion threads.
		if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
			return nil, err
		}
		delete = *args.Input.Delete
	}

	threadID, err := unmarshalDiscussionThreadID(args.Input.ThreadID)
	if err != nil {
		return nil, err
	}
	thread, err := db.DiscussionThreads.Update(ctx, threadID, &db.DiscussionThreadsUpdateOptions{
		Active:   args.Input.Active,
		Archive:  args.Input.Archive,
		Delete:   delete,
		Settings: args.Input.Settings,
		Title:    args.Input.Title,
	})
	if err != nil {
		return nil, errors.Wrap(err, "DiscussionThreads.Update")
	}
	if thread == nil {
		// deleted
		return nil, nil
	}
	return &discussionThreadResolver{t: thread}, nil
}

func (*schemaResolver) Discussions(ctx context.Context) (*discussionsMutationResolver, error) {
	if err := viewerCanUseDiscussions(ctx); err != nil {
		return nil, err
	}
	return &discussionsMutationResolver{}, nil
}

func (*schemaResolver) DiscussionThreads(ctx context.Context, args *struct {
	graphqlutil.ConnectionArgs
	Query                       *string
	ThreadID                    *graphql.ID
	AuthorUserID                *graphql.ID
	TargetRepositoryID          *graphql.ID
	TargetRepositoryName        *string
	TargetRepositoryGitCloneURL *string
	TargetRepositoryPath        *string
}) (*discussionThreadsConnectionResolver, error) {
	if err := viewerCanUseDiscussions(ctx); err != nil {
		return nil, err
	}

	// 🚨 SECURITY: No authentication is required to list discussions. They are
	// public unless the Sourcegraph instance itself (and inherently, the
	// GraphQL API) is private.

	opt := &db.DiscussionThreadsListOptions{
		TargetRepoPath: args.TargetRepositoryPath,
	}
	if args.Query != nil {
		opt.SetFromQuery(ctx, *args.Query)
	}
	args.ConnectionArgs.Set(&opt.LimitOffset)

	if args.ThreadID != nil {
		// BACKCOMPAT DEPRECATED: For backcompat, this value is treated as
		// DiscussionThread#idWithoutKind and is a strictly numeric string like "1234", not a
		// conventional graphql.ID value (i.e., base64("DiscussionThread:1234")).
		dbID, err := strconv.ParseInt(string(*args.ThreadID), 10, 64)
		if err != nil {
			return nil, err
		}
		opt.ThreadIDs = []int64{dbID}
	}
	if args.AuthorUserID != nil {
		authorUserID, err := UnmarshalUserID(*args.AuthorUserID)
		if err != nil {
			return nil, err
		}
		opt.AuthorUserIDs = []int32{authorUserID}
	}

	count := 0
	if args.TargetRepositoryID != nil {
		count++
	}
	if args.TargetRepositoryName != nil {
		count++
	}
	if args.TargetRepositoryGitCloneURL != nil {
		count++
	}
	if count == 1 {
		repo, err := discussionsResolveRepository(ctx, args.TargetRepositoryID, args.TargetRepositoryName, args.TargetRepositoryGitCloneURL)
		if err != nil {
			return nil, err
		}
		opt.TargetRepoID = &repo.repo.ID
	} else if count > 1 {
		return nil, errors.New("only one of targetRepositoryID, targetRepositoryName, or targetRepositoryGitCloneURL can be specified")
	}
	return &discussionThreadsConnectionResolver{opt: opt}, nil
}

func (schemaResolver) DiscussionThread(ctx context.Context, args *struct {
	IDWithoutKind string
}) (*discussionThreadResolver, error) {
	dbID, err := strconv.ParseInt(args.IDWithoutKind, 10, 64)
	if err != nil {
		return nil, err
	}
	return DiscussionThreadByID(ctx, marshalDiscussionThreadID(dbID))
}

type discussionThreadTargetRepoSelectionResolver struct {
	t *types.DiscussionThreadTargetRepo
}

func (r *discussionThreadTargetRepoSelectionResolver) StartLine() int32 { return *r.t.StartLine }
func (r *discussionThreadTargetRepoSelectionResolver) StartCharacter() int32 {
	return *r.t.StartCharacter
}
func (r *discussionThreadTargetRepoSelectionResolver) EndLine() int32        { return *r.t.EndLine }
func (r *discussionThreadTargetRepoSelectionResolver) EndCharacter() int32   { return *r.t.EndCharacter }
func (r *discussionThreadTargetRepoSelectionResolver) LinesBefore() []string { return *r.t.LinesBefore }
func (r *discussionThreadTargetRepoSelectionResolver) Lines() []string       { return *r.t.Lines }
func (r *discussionThreadTargetRepoSelectionResolver) LinesAfter() []string  { return *r.t.LinesAfter }

type discussionThreadTargetRepoResolver struct {
	t *types.DiscussionThreadTargetRepo
}

func (r *discussionThreadTargetRepoResolver) Repository(ctx context.Context) (*repositoryResolver, error) {
	return repositoryByIDInt32(ctx, r.t.RepoID)
}

func (r *discussionThreadTargetRepoResolver) Path() *string { return r.t.Path }

func (r *discussionThreadTargetRepoResolver) Branch(ctx context.Context) (*gitRefResolver, error) {
	return r.branchOrRevision(ctx, r.t.Branch)
}

func (r *discussionThreadTargetRepoResolver) Revision(ctx context.Context) (*gitRefResolver, error) {
	return r.branchOrRevision(ctx, r.t.Revision)
}

func (r *discussionThreadTargetRepoResolver) branchOrRevision(ctx context.Context, rev *string) (*gitRefResolver, error) {
	if rev == nil {
		return nil, nil
	}
	repo, err := repositoryByIDInt32(ctx, r.t.RepoID)
	if err != nil {
		return nil, err
	}
	return &gitRefResolver{repo: repo, name: *rev}, nil
}

func (r *discussionThreadTargetRepoResolver) Selection() *discussionThreadTargetRepoSelectionResolver {
	if !r.t.HasSelection() {
		return nil
	}
	return &discussionThreadTargetRepoSelectionResolver{t: r.t}
}

func (r *discussionThreadTargetRepoResolver) RelativePath(ctx context.Context, args *struct {
	Rev string
}) (*string, error) {
	if r.t.Path == nil {
		return nil, nil
	}
	repo, err := repositoryByIDInt32(ctx, r.t.RepoID)
	if err != nil {
		return nil, err
	}
	if r.t.Revision == nil && r.t.Branch == nil {
		// The thread wasn't created on a specific revision or branch, so we
		// cannot walk the history. Instead, we must assume its location and
		// check in the relative revision.
		commit, err := repo.Commit(ctx, &repositoryCommitArgs{Rev: args.Rev})
		if err != nil {
			return nil, err
		}
		_, err = commit.File(ctx, &struct{ Path string }{Path: *r.t.Path})
		if err != nil {
			// File does not exist in this revision.
			return nil, nil
		}
		return r.t.Path, nil // File exists at that path.
	}

	var rev string
	if r.t.Revision != nil {
		rev = *r.t.Revision
	} else if r.t.Branch != nil {
		rev = *r.t.Branch
	}
	comparison, err := repo.Comparison(ctx, &repositoryComparisonInput{
		Base: &rev,
		Head: &args.Rev,
	})
	if err != nil {
		return nil, err
	}
	currentPath := *r.t.Path
	fileDiffs, err := comparison.FileDiffs(&struct{ First *int32 }{}).Nodes(ctx)
	if err != nil {
		return nil, err
	}
	for _, fileDiff := range fileDiffs {
		oldPath := fileDiff.OldPath()
		newPath := fileDiff.NewPath()

		if oldPath == nil && newPath != nil {
			// newPath was added. We don't need to do anything because this
			// could only indicate the file we're tracking was added.
		} else if oldPath != nil && newPath == nil {
			// oldPath was removed
			if currentPath == *oldPath {
				// The file we are tracking was removed!
				return nil, nil
			}
		} else if oldPath != nil && newPath != nil {
			// oldPath was renamed to newPath
			if currentPath == *oldPath {
				// The file we are tracking was renamed.
				currentPath = *newPath
			}
		}
	}
	return &currentPath, nil
}

type discussionSelectionRangeResolver struct {
	startLine, startCharacter, endLine, endCharacter int32
}

func (r *discussionSelectionRangeResolver) StartLine() int32      { return r.startLine }
func (r *discussionSelectionRangeResolver) StartCharacter() int32 { return r.startCharacter }
func (r *discussionSelectionRangeResolver) EndLine() int32        { return r.endLine }
func (r *discussionSelectionRangeResolver) EndCharacter() int32   { return r.endCharacter }

func discussionSelectionRelativeTo(oldSel *types.DiscussionThreadTargetRepo, newContent string) *discussionSelectionRangeResolver {
	mustFindLines := 4

	search := func(searchForLines string) *discussionSelectionRangeResolver {
		if len(strings.Split(searchForLines, "\n")) < mustFindLines {
			// We do not have enough search lines to find a good match.
			return nil
		}
		matches := strings.Count(newContent, searchForLines)
		switch {
		case matches > 1:
			// The lines we are searching for produced too many matches.
			return nil
		case matches == 1:
			// We found a perfect match.
			idx := strings.Index(newContent, searchForLines)
			startLine := int32(len(strings.Split(newContent[:idx], "\n")))
			return &discussionSelectionRangeResolver{
				startCharacter: *oldSel.StartCharacter,
				endCharacter:   *oldSel.EndCharacter,
				startLine:      startLine,
				endLine:        startLine + int32(len(*oldSel.Lines)),
			}
		default:
			return nil
		}
	}

	// Start removing lines until we find a result (or fail to find one).
	allLines := *oldSel.LinesBefore
	allLines = append(allLines, *oldSel.Lines...)
	allLines = append(allLines, *oldSel.LinesAfter...)
	removeLines := 0
	for {
		if removeLines > len(allLines) {
			return nil
		}
		// Try removing N lines from the top.
		if r := search(strings.Join(allLines[removeLines:], "\n")); r != nil {
			offset := int32(len(*oldSel.LinesBefore) - 1 - removeLines)
			r.startLine += offset
			r.endLine += offset
			return r
		}

		// Try removing N lines from the bottom.
		if r := search(strings.Join(allLines[:len(allLines)-removeLines], "\n")); r != nil {
			offset := int32(len(*oldSel.LinesAfter) - 1 - removeLines)
			r.startLine += offset
			r.endLine += offset
			return r
		}
		removeLines++
	}
}

func (r *discussionThreadTargetRepoResolver) RelativeSelection(ctx context.Context, args *struct {
	Rev string
}) (*discussionSelectionRangeResolver, error) {
	if !r.t.HasSelection() {
		return nil, nil
	}
	path, err := r.RelativePath(ctx, args)
	if err != nil {
		return nil, err
	}
	if path == nil {
		return nil, nil
	}
	repo, err := repositoryByIDInt32(ctx, r.t.RepoID)
	if err != nil {
		return nil, err
	}
	commit, err := repo.Commit(ctx, &repositoryCommitArgs{Rev: args.Rev})
	if err != nil {
		return nil, err
	}
	oldSel := &discussionSelectionRangeResolver{
		startLine:      *r.t.StartLine,
		startCharacter: *r.t.StartCharacter,
		endLine:        *r.t.EndLine,
		endCharacter:   *r.t.EndCharacter,
	}
	if r.t.Revision != nil && *r.t.Revision == string(commit.OID()) {
		return oldSel, nil // nothing to do (requested relative revision is identical to the stored revision)
	}
	if r.t.Branch != nil {
		branchCommit, err := repo.Commit(ctx, &repositoryCommitArgs{Rev: *r.t.Branch})
		if err != nil {
			return nil, err
		}
		if branchCommit.OID() == commit.OID() {
			return oldSel, nil // nothing to do (requested relative revision is identical to the stored branch revision)
		}
	}
	file, err := commit.File(ctx, &struct{ Path string }{Path: *path})
	if err != nil {
		return nil, err
	}
	newContent, err := file.Content(ctx)
	if err != nil {
		return nil, err
	}
	return discussionSelectionRelativeTo(r.t, newContent), nil
}

type discussionThreadTargetResolver struct {
	t *types.DiscussionThreadTargetRepo

	// BACKCOMPAT: See (*discussionThreadResolver).Target for more information.
	unrecognized bool
}

func (r *discussionThreadTargetResolver) ToDiscussionThreadTargetRepo() (*discussionThreadTargetRepoResolver, bool) {
	if r.t == nil {
		return nil, false
	}
	return &discussionThreadTargetRepoResolver{t: r.t}, true
}

func (r *discussionThreadTargetResolver) ToEmptyResponse() (*EmptyResponse, bool) {
	if r.unrecognized {
		return &EmptyResponse{}, true
	}
	return nil, false
}

func marshalDiscussionThreadID(dbID int64) graphql.ID {
	return relay.MarshalID("DiscussionThread", strconv.FormatInt(dbID, 36))
}

func unmarshalDiscussionThreadID(id graphql.ID) (dbID int64, err error) {
	var dbIDStr string
	err = relay.UnmarshalSpec(id, &dbIDStr)
	if err == nil {
		dbID, err = strconv.ParseInt(dbIDStr, 36, 64)
	}
	return
}

// DiscussionThreadByID looks up a DiscussionThread by its GraphQL ID.
func DiscussionThreadByID(ctx context.Context, id graphql.ID) (*discussionThreadResolver, error) {
	dbID, err := unmarshalDiscussionThreadID(id)
	if err != nil {
		return nil, err
	}
	// 🚨 SECURITY: No authentication is required to get a discussion. Discussions are public unless
	// the Sourcegraph instance itself (and inherently, the GraphQL API) is private.
	thread, err := db.DiscussionThreads.Get(ctx, dbID)
	if err != nil {
		return nil, err
	}
	return &discussionThreadResolver{t: thread}, nil
}

// 🚨 SECURITY: When instantiating an discussionThreadResolver value, the
// caller MUST check permissions.
type discussionThreadResolver struct {
	t *types.DiscussionThread
}

func (d *discussionThreadResolver) ID() graphql.ID {
	return marshalDiscussionThreadID(d.t.ID)
}

func (d *discussionThreadResolver) DBID() int64 { return d.t.ID }

func (d *discussionThreadResolver) IDWithoutKind() string {
	return strconv.FormatInt(d.t.ID, 10)
}

func (d *discussionThreadResolver) Project(ctx context.Context) (Project, error) {
	return ProjectByDBID(ctx, d.t.ProjectID)
}

func (d *discussionThreadResolver) Author(ctx context.Context) (*UserResolver, error) {
	return UserByIDInt32(ctx, d.t.AuthorUserID)
}

func (d *discussionThreadResolver) Title() string { return d.t.Title }

func (d *discussionThreadResolver) Targets(ctx context.Context, args *discussionThreadTargetConnectionArgs) *discussionThreadTargetConnectionResolver {
	return &discussionThreadTargetConnectionResolver{threadID: d.t.ID, args: args}
}

func (d *discussionThreadResolver) Target(ctx context.Context) (*discussionThreadTargetResolver, error) {
	// TODO(sqs): This only takes the 1st target. Support multiple targets.
	targets, err := db.DiscussionThreads.ListTargets(ctx, db.DiscussionThreadsListTargetsOptions{ThreadID: d.t.ID})
	if err != nil {
		return nil, err
	}
	if len(targets) > 0 {
		return &discussionThreadTargetResolver{t: targets[0]}, nil
	}

	// BACKCOMPAT: For backcompat with clients expecting this to be non-nullable, return a value
	// that is of a different __typename. The documentation has always required callers to handle
	// unrecognized __typenames gracefully.
	return &discussionThreadTargetResolver{unrecognized: true}, nil
}

func (d *discussionThreadResolver) Settings(ctx context.Context) string {
	if settings := d.t.Settings; settings != nil {
		return *settings
	}
	return "{}"
}

func (d *discussionThreadResolver) Status() types.ThreadStatus {
	return d.t.Status
}

func (d *discussionThreadResolver) Type() types.ThreadType {
	return types.ThreadType(d.t.Type)
}

func (d *discussionThreadResolver) URL(ctx context.Context) string {
	// TODO!(sqs): hardcoded /p/
	return fmt.Sprintf("/p/%d/%s/%s", d.t.ProjectID, strings.ToLower(string(d.t.Type))+"s", d.IDWithoutKind())
}

func (d *discussionThreadResolver) InlineURL(ctx context.Context) (*string, error) {
	url, err := discussions.URLToInlineThread(ctx, d.t)
	if err != nil || url == nil {
		return nil, err
	}
	return strptr(url.String()), nil
}

func (d *discussionThreadResolver) CreatedAt(ctx context.Context) string {
	return d.t.CreatedAt.Format(time.RFC3339)
}

func (d *discussionThreadResolver) UpdatedAt(ctx context.Context) string {
	return d.t.UpdatedAt.Format(time.RFC3339)
}

func (d *discussionThreadResolver) ArchivedAt(ctx context.Context) *string {
	if d.t.ArchivedAt == nil {
		return nil
	}
	return strptr(d.t.ArchivedAt.Format(time.RFC3339))
}

func (d *discussionThreadResolver) Comments(ctx context.Context, args *struct {
	graphqlutil.ConnectionArgs
}) *discussionCommentsConnectionResolver {
	// 🚨 SECURITY: Anyone with access to the thread also has access to its
	// comments. Hence, since we are only accessing the threads comments here
	// (and not other thread's comments) we are covered security-wise here
	// implicitly.

	opt := &db.DiscussionCommentsListOptions{ThreadID: &d.t.ID}
	args.ConnectionArgs.Set(&opt.LimitOffset)
	return &discussionCommentsConnectionResolver{opt: opt}
}

func (d *discussionThreadResolver) ToDiscussionThread() (*discussionThreadResolver, bool) {
	return d, true
}

func (d *discussionThreadResolver) Labels(ctx context.Context, arg *graphqlutil.ConnectionArgs) (LabelConnection, error) {
	return LabelsFor(ctx, d.ID(), arg)
}

// discussionThreadsConnectionResolver resolves a list of discussion comments.
//
// 🚨 SECURITY: When instantiating an discussionThreadsConnectionResolver
// value, the caller MUST check permissions.
type discussionThreadsConnectionResolver struct {
	opt *db.DiscussionThreadsListOptions

	// cache results because they are used by multiple fields
	once     sync.Once
	comments []*types.DiscussionThread
	err      error
}

func (r *discussionThreadsConnectionResolver) compute(ctx context.Context) ([]*types.DiscussionThread, error) {
	r.once.Do(func() {
		opt2 := *r.opt
		if opt2.LimitOffset != nil {
			tmp := *opt2.LimitOffset
			opt2.LimitOffset = &tmp
			opt2.Limit++ // so we can detect if there is a next page
		}

		r.comments, r.err = db.DiscussionThreads.List(ctx, &opt2)
	})
	return r.comments, r.err
}

func (r *discussionThreadsConnectionResolver) Nodes(ctx context.Context) ([]*discussionThreadResolver, error) {
	threads, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	var l []*discussionThreadResolver
	for _, thread := range threads {
		l = append(l, &discussionThreadResolver{t: thread})
	}
	return l, nil
}

func (r *discussionThreadsConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	withoutLimit := *r.opt
	withoutLimit.LimitOffset = nil
	count, err := db.DiscussionThreads.Count(ctx, &withoutLimit)
	return int32(count), err
}

func (r *discussionThreadsConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	threads, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(r.opt.LimitOffset != nil && len(threads) > r.opt.Limit), nil
}

var mockViewerCanUseDiscussions func() error

// viewerCanUseDiscussions returns an error if the user in the context cannot
// use code discussions, e.g. due to the extension not being installed or
// enabled.
func viewerCanUseDiscussions(ctx context.Context) error {
	if mockViewerCanUseDiscussions != nil {
		return mockViewerCanUseDiscussions()
	}

	merged, err := viewerFinalSettings(ctx)
	if err != nil {
		return err
	}
	var settings schema.Settings
	if err := jsonc.Unmarshal(merged.Contents(), &settings); err != nil {
		return err
	}
	enabled, ok := settings.Extensions["sourcegraph/code-discussions"]
	if !ok {
		return errors.New("Sourcegraph Code Discussions extension must be added for the active user to use this API")
	}
	if !enabled {
		return errors.New("Sourcegraph Code Discussions extension must be enabled for the active user to use this API")
	}
	return nil
}
