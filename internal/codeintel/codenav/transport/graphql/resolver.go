package graphql

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/dgraph-io/ristretto"

	"github.com/sourcegraph/go-diff/diff"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	gitserverInternal "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Resolver interface {
	SetUploadsDataLoader(uploads []dbstore.Dump)
	SetLocalGitTreeTranslator(client gitserver.Client, repo *types.Repo, commit, path string) error
	SetLocalCommitCache(client shared.GitserverClient)
	SetMaximumIndexesPerMonikerSearch(maxNumber int)
	SetAuthChecker(authChecker authz.SubRepoPermissionChecker)

	Hover(ctx context.Context, args shared.RequestArgs) (_ string, _ shared.Range, _ bool, err error)
	Definitions(ctx context.Context, args shared.RequestArgs) (_ []shared.UploadLocation, err error)
	References(ctx context.Context, args shared.RequestArgs) (_ []shared.UploadLocation, _ string, err error)
	Implementations(ctx context.Context, args shared.RequestArgs) (_ []shared.UploadLocation, _ string, err error)
	Diagnostics(ctx context.Context, args shared.RequestArgs) (diagnosticsAtUploads []shared.DiagnosticAtUpload, _ int, err error)
	Stencil(ctx context.Context, args shared.RequestArgs) (adjustedRanges []shared.Range, err error)
	Ranges(ctx context.Context, args shared.RequestArgs, startLine, endLine int) (adjustedRanges []shared.AdjustedCodeIntelligenceRange, err error)
}

type resolver struct {
	svc               Service
	requestArgs       *requestArgs
	GitTreeTranslator GitTreeTranslator

	authChecker authz.SubRepoPermissionChecker

	requestState codenav.RequestState

	// maximumIndexesPerMonikerSearch configures the maximum number of reference upload identifiers
	// that can be passed to a single moniker search query. Previously this limit was meant to keep
	// the number of SQLite files we'd have to open within a single call relatively low. Since we've
	// migrated to Postgres this limit is not a concern. Now we only want to limit these values
	// based on the number of elements we can pass to an IN () clause in the codeintel-db, as well
	// as the size required to encode them in a user-facing pagination cursor.
	maximumIndexesPerMonikerSearch int

	// Local Request Caches
	dataLoader    *UploadsDataLoader
	hunkCacheSize int
	commitCache   CommitCache

	// Metrics
	operations *operations
}

func New(svc Service, hunkCacheSize int, observationContext *observation.Context) *resolver {
	return &resolver{
		svc:           svc,
		operations:    newOperations(observationContext),
		dataLoader:    NewUploadsDataLoader(),
		hunkCacheSize: hunkCacheSize,
	}
}

func (r *resolver) SetRequestState(
	uploads []dbstore.Dump,
	authChecker authz.SubRepoPermissionChecker,
	client gitserver.Client, repo *types.Repo, commit, path string,
	gitclient shared.GitserverClient,
	maxIndexes int,
) {
	requestState := codenav.RequestState{}
	requestState.SetUploadsDataLoader(uploads)
	requestState.SetAuthChecker(authChecker)
	requestState.SetLocalGitTreeTranslator(client, repo, commit, path, r.hunkCacheSize)
	requestState.SetLocalCommitCache(gitclient)
	requestState.SetMaximumIndexesPerMonikerSearch(maxIndexes)

	r.requestState = requestState
}

func (r *resolver) SetAuthChecker(authChecker authz.SubRepoPermissionChecker) {
	r.authChecker = authChecker
}

func (r *resolver) SetUploadsDataLoader(uploads []dbstore.Dump) {
	for _, upload := range uploads {
		r.dataLoader.AddUpload(upload)
	}
}

func (r *resolver) SetLocalGitTreeTranslator(client gitserver.Client, repo *types.Repo, commit, path string) error {
	hunkCache, err := NewHunkCache(r.hunkCacheSize)
	if err != nil {
		return err
	}

	args := &requestArgs{
		repo:   repo,
		commit: commit,
		path:   path,
	}

	r.requestArgs = args
	r.GitTreeTranslator = NewGitTreeTranslator(client, args, hunkCache)

	return nil
}

func (r *resolver) SetLocalCommitCache(client shared.GitserverClient) {
	r.commitCache = newCommitCache(client)
}

func (r *resolver) SetMaximumIndexesPerMonikerSearch(maxNumber int) {
	r.maximumIndexesPerMonikerSearch = maxNumber
}

func (r *resolver) Symbol(ctx context.Context, args struct{}) (_ any, err error) {
	ctx, _, endObservation := r.operations.symbol.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	// To be implemented in https://github.com/sourcegraph/sourcegraph/issues/33374
	_, _ = ctx, args
	return nil, errors.New("unimplemented: Symbol")
}

// DELETE

// HunkCache is a LRU cache that holds git diff hunks.
type HunkCache interface {
	// Get returns the value (if any) and a boolean representing whether the value was
	// found or not.
	Get(key any) (any, bool)

	// Set attempts to add the key-value item to the cache with the given cost. If it
	// returns false, then the value as dropped and the item isn't added to the cache.
	Set(key, value any, cost int64) bool
}

// NewHunkCache creates a data cache instance with the given maximum capacity.
func NewHunkCache(size int) (HunkCache, error) {
	return ristretto.NewCache(&ristretto.Config{
		NumCounters: int64(size) * 10,
		MaxCost:     int64(size),
		BufferItems: 64,
	})
}

// GitTreeTranslator translates a position within a git tree at a source commit into the
// equivalent position in a target commit. The git tree translator instance carries
// along with it the source commit.
type GitTreeTranslator interface {
	// GetTargetCommitPathFromSourcePath translates the given path from the source commit into the given target
	// commit. If revese is true, then the source and target commits are swapped.
	GetTargetCommitPathFromSourcePath(ctx context.Context, commit, path string, reverse bool) (string, bool, error)
	// AdjustPath

	// GetTargetCommitPositionFromSourcePosition translates the given position from the source commit into the given
	// target commit. The target commit's path and position are returned, along with a boolean flag
	// indicating that the translation was successful. If revese is true, then the source and
	// target commits are swapped.
	GetTargetCommitPositionFromSourcePosition(ctx context.Context, commit string, px shared.Position, reverse bool) (string, shared.Position, bool, error)
	// AdjustPosition

	// GetTargetCommitRangeFromSourceRange translates the given range from the source commit into the given target
	// commit. The target commit's path and range are returned, along with a boolean flag indicating
	// that the translation was successful. If revese is true, then the source and target commits
	// are swapped.
	GetTargetCommitRangeFromSourceRange(ctx context.Context, commit, path string, rx shared.Range, reverse bool) (string, shared.Range, bool, error)
}

type gitTreeTranslator struct {
	client           gitserver.Client
	localRequestArgs *requestArgs
	hunkCache        HunkCache
}

// NewGitTreeTranslator creates a new GitTreeTranslator with the given repository and source commit.
func NewGitTreeTranslator(client gitserver.Client, args *requestArgs, hunkCache HunkCache) GitTreeTranslator {
	return &gitTreeTranslator{
		client:           client,
		hunkCache:        hunkCache,
		localRequestArgs: args,
	}
}

// GetTargetCommitPathFromSourcePath translates the given path from the source commit into the given target
// commit. If revese is true, then the source and target commits are swapped.
func (g *gitTreeTranslator) GetTargetCommitPathFromSourcePath(ctx context.Context, commit, path string, reverse bool) (string, bool, error) {
	return path, true, nil
}

// GetTargetCommitPositionFromSourcePosition translates the given position from the source commit into the given
// target commit. The target commit path and position are returned, along with a boolean flag
// indicating that the translation was successful. If revese is true, then the source and
// target commits are swapped.
// TODO: No todo just letting me know that I updated path just on this one. Need to do it like that.
func (g *gitTreeTranslator) GetTargetCommitPositionFromSourcePosition(ctx context.Context, commit string, px shared.Position, reverse bool) (string, shared.Position, bool, error) {
	hunks, err := g.readCachedHunks(ctx, g.localRequestArgs.repo, g.localRequestArgs.commit, commit, g.localRequestArgs.path, reverse)
	if err != nil {
		return "", shared.Position{}, false, err
	}

	commitPosition, ok := translatePosition(hunks, px)
	return g.localRequestArgs.path, commitPosition, ok, nil
}

// GetTargetCommitRangeFromSourceRange translates the given range from the source commit into the given target
// commit. The target commit path and range are returned, along with a boolean flag indicating
// that the translation was successful. If revese is true, then the source and target commits
// are swapped.
func (g *gitTreeTranslator) GetTargetCommitRangeFromSourceRange(ctx context.Context, commit, path string, rx shared.Range, reverse bool) (string, shared.Range, bool, error) {
	hunks, err := g.readCachedHunks(ctx, g.localRequestArgs.repo, g.localRequestArgs.commit, commit, path, reverse)
	if err != nil {
		return "", shared.Range{}, false, err
	}

	commitRange, ok := translateRange(hunks, rx)
	return path, commitRange, ok, nil
}

// readCachedHunks returns a position-ordered slice of changes (additions or deletions) of
// the given path between the given source and target commits. If reverse is true, then the
// source and target commits are swapped. If the git tree translator has a hunk cache, it
// will read from it before attempting to contact a remote server, and populate the cache
// with new results
func (g *gitTreeTranslator) readCachedHunks(ctx context.Context, repo *types.Repo, sourceCommit, targetCommit, path string, reverse bool) ([]*diff.Hunk, error) {
	if sourceCommit == targetCommit {
		return nil, nil
	}
	if reverse {
		sourceCommit, targetCommit = targetCommit, sourceCommit
	}

	if g.hunkCache == nil {
		return g.readHunks(ctx, repo, sourceCommit, targetCommit, path)
	}

	key := makeKey(strconv.FormatInt(int64(repo.ID), 10), sourceCommit, targetCommit, path)
	if hunks, ok := g.hunkCache.Get(key); ok {
		if hunks == nil {
			return nil, nil
		}

		return hunks.([]*diff.Hunk), nil
	}

	hunks, err := g.readHunks(ctx, repo, sourceCommit, targetCommit, path)
	if err != nil {
		return nil, err
	}

	g.hunkCache.Set(key, hunks, int64(len(hunks)))

	return hunks, nil
}

// readHunks returns a position-ordered slice of changes (additions or deletions) of
// the given path between the given source and target commits.
func (g *gitTreeTranslator) readHunks(ctx context.Context, repo *types.Repo, sourceCommit, targetCommit, path string) ([]*diff.Hunk, error) {
	return g.client.DiffPath(ctx, authz.DefaultSubRepoPermsChecker, repo.Name, sourceCommit, targetCommit, path)
}

// findHunk returns the last thunk that does not begin after the given line.
func findHunk(hunks []*diff.Hunk, line int) *diff.Hunk {
	i := 0
	for i < len(hunks) && int(hunks[i].OrigStartLine) <= line {
		i++
	}

	if i == 0 {
		return nil
	}
	return hunks[i-1]
}

// translateRange translates the given range by calling translatePosition on both of the range's
// endpoints. This function returns a boolean flag indicating that the translation was
// successful (which occurs when both endpoints of the range can be translated).
func translateRange(hunks []*diff.Hunk, r shared.Range) (shared.Range, bool) {
	start, ok := translatePosition(hunks, r.Start)
	if !ok {
		return shared.Range{}, false
	}

	end, ok := translatePosition(hunks, r.End)
	if !ok {
		return shared.Range{}, false
	}

	return shared.Range{Start: start, End: end}, true
}

// translatePosition translates the given position by setting the line number based on the
// number of additions and deletions that occur before that line. This function returns a
// boolean flag indicating that the translation is successful. A translation fails when the
// line indicated by the position has been edited.
func translatePosition(hunks []*diff.Hunk, pos shared.Position) (shared.Position, bool) {
	line, ok := translateLineNumbers(hunks, pos.Line)
	if !ok {
		return shared.Position{}, false
	}

	return shared.Position{Line: line, Character: pos.Character}, true
}

// translateLineNumbers translates the given line number based on the number of additions and deletions
// that occur before that line. This function returns a boolean flag indicating that the
// translation is successful. A translation fails when the given line has been edited.
func translateLineNumbers(hunks []*diff.Hunk, line int) (int, bool) {
	// Translate from bundle/lsp zero-index to git diff one-index
	line = line + 1

	hunk := findHunk(hunks, line)
	if hunk == nil {
		// Trivial case, no changes before this line
		return line - 1, true
	}

	// If the hunk ends before this line, we can simply set the line offset by the
	// relative difference between the line offsets in each file after this hunk.
	if line >= int(hunk.OrigStartLine+hunk.OrigLines) {
		endOfSourceHunk := int(hunk.OrigStartLine + hunk.OrigLines)
		endOfTargetHunk := int(hunk.NewStartLine + hunk.NewLines)
		targetCommitLineNumber := line + (endOfTargetHunk - endOfSourceHunk)

		// Translate from git diff one-index to bundle/lsp zero-index
		return targetCommitLineNumber - 1, true
	}

	// These offsets start at the beginning of the hunk's delta. The following loop will
	// process the delta line-by-line. For each line that exists the source (orig) or
	// target (new) file, the corresponding offset will be bumped. The values of these
	// offsets once we hit our target line will determine the relative offset between
	// the two files.
	sourceOffset := int(hunk.OrigStartLine)
	targetOffset := int(hunk.NewStartLine)

	for _, deltaLine := range strings.Split(string(hunk.Body), "\n") {
		isAdded := strings.HasPrefix(deltaLine, "+")
		isRemoved := strings.HasPrefix(deltaLine, "-")

		// A line exists in the source file if it wasn't added in the delta. We set
		// this before the next condition so that our comparison with our target line
		// is correct.
		if !isAdded {
			sourceOffset++
		}

		// Hit our target line
		if sourceOffset-1 == line {
			// This particular line was (1) edited; (2) removed, or (3) added.
			// If it was removed, there is nothing to point to in the target file.
			// If it was added, then we don't have any index information for it in
			// our source file. In any case, we won't have a precise translation.
			if isAdded || isRemoved {
				return 0, false
			}

			// Translate from git diff one-index to bundle/lsp zero-index
			return targetOffset - 1, true
		}

		// A line exists in the target file if it wasn't deleted in the delta. We set
		// this after the previous condition so we don't have to re-set the target offset
		// within the exit conditions (this adjustment is only necessary for future iterations).
		if !isRemoved {
			targetOffset++
		}
	}

	// This should never happen unless the git diff content is malformed. We know
	// the target line occurs within the hunk, but iteration of the hunk's body did
	// not contain enough lines attributed to the original file.
	panic("Malformed hunk body")
}

func makeKey(parts ...string) string {
	return strings.Join(parts, ":")
}

type CommitCache interface {
	AreCommitsResolvable(ctx context.Context, commits []gitserverInternal.RepositoryCommit) ([]bool, error)
}

type commitCache struct {
	gitserverClient shared.GitserverClient
	mutex           sync.RWMutex
	cache           map[int]map[string]bool
}

func newCommitCache(client shared.GitserverClient) CommitCache {
	return &commitCache{
		gitserverClient: client,
		cache:           map[int]map[string]bool{},
	}
}

// set marks the given repository and commit as valid and resolvable by gitserver.
func (c *commitCache) set(repositoryID int, commit string) {
	c.setInternal(repositoryID, commit, true)
}

// AreCommitsResolvable determines if the given commits are resolvable for the given repositories.
// If we do not know the answer from a previous call to set or AreCommitsResolvable, we ask gitserver
// to resolve the remaining commits and store the results for subsequent calls. This method
// returns a slice of the same size as the input slice, true indicating that the commit at
// the symmetric index exists.
func (c *commitCache) AreCommitsResolvable(ctx context.Context, commits []gitserverInternal.RepositoryCommit) ([]bool, error) {
	exists := make([]bool, len(commits))
	rcIndexMap := make([]int, 0, len(commits))
	rcs := make([]gitserverInternal.RepositoryCommit, 0, len(commits))

	for i, rc := range commits {
		if e, ok := c.getInternal(rc.RepositoryID, rc.Commit); ok {
			exists[i] = e
		} else {
			rcIndexMap = append(rcIndexMap, i)
			rcs = append(rcs, gitserverInternal.RepositoryCommit{
				RepositoryID: rc.RepositoryID,
				Commit:       rc.Commit,
			})
		}
	}

	// if there are no repository commits to fetch, we're done
	if len(rcs) == 0 {
		return exists, nil
	}

	// Perform heavy work outside of critical section
	e, err := c.gitserverClient.CommitsExist(ctx, rcs)
	if err != nil {
		return nil, errors.Wrap(err, "gitserverClient.CommitsExist")
	}
	if len(e) != len(rcs) {
		panic(strings.Join([]string{
			fmt.Sprintf("Expected slice returned from CommitsExist to have len %d, but has len %d.", len(rcs), len(e)),
			"If this panic occurred dcuring a test, your test is missing a mock definition for CommitsExist.",
			"If this is occurred during runtime, please file a bug.",
		}, " "))
	}

	for i, rc := range rcs {
		exists[rcIndexMap[i]] = e[i]
		c.setInternal(rc.RepositoryID, rc.Commit, e[i])
	}

	return exists, nil
}

func (c *commitCache) getInternal(repositoryID int, commit string) (bool, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if repositoryMap, ok := c.cache[repositoryID]; ok {
		if exists, ok := repositoryMap[commit]; ok {
			return exists, true
		}
	}

	return false, false
}

func (c *commitCache) setInternal(repositoryID int, commit string, exists bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.cache[repositoryID]; !ok {
		c.cache[repositoryID] = map[string]bool{}
	}

	c.cache[repositoryID][commit] = exists
}
