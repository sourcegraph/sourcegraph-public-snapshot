package resolvers

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
)

const patchSetIDKind = "PatchSet"

func marshalPatchSetID(id int64) graphql.ID {
	return relay.MarshalID(patchSetIDKind, id)
}

func unmarshalPatchSetID(id graphql.ID) (patchSetID int64, err error) {
	err = relay.UnmarshalSpec(id, &patchSetID)
	return
}

const patchIDKind = "Patch"

func marshalPatchID(id int64) graphql.ID {
	return relay.MarshalID(patchIDKind, id)
}

func unmarshalPatchID(id graphql.ID) (cid int64, err error) {
	err = relay.UnmarshalSpec(id, &cid)
	return
}

var _ graphqlbackend.PatchSetResolver = &patchSetResolver{}

type patchSetResolver struct {
	store    *ee.Store
	patchSet *campaigns.PatchSet
}

func (r *patchSetResolver) ID() graphql.ID {
	return marshalPatchSetID(r.patchSet.ID)
}

func (r *patchSetResolver) Patches(
	ctx context.Context,
	args *graphqlutil.ConnectionArgs,
) graphqlbackend.PatchConnectionResolver {
	return &patchesConnectionResolver{
		store: r.store,
		opts: ee.ListPatchesOpts{
			PatchSetID:   r.patchSet.ID,
			Limit:        int(args.GetFirst()),
			OnlyWithDiff: true,
		},
	}
}

func (r *patchSetResolver) DiffStat(ctx context.Context) (*graphqlbackend.DiffStat, error) {
	return patchSetDiffStat(ctx, r.store, ee.ListPatchesOpts{
		PatchSetID:   r.patchSet.ID,
		Limit:        -1, // Fetch all patches in a patch set
		OnlyWithDiff: true,
	})
}

func patchSetDiffStat(ctx context.Context, store *ee.Store, opts ee.ListPatchesOpts) (*graphqlbackend.DiffStat, error) {
	patchesConnection := &patchesConnectionResolver{store: store, opts: opts}

	patches, err := patchesConnection.Nodes(ctx)
	if err != nil {
		return nil, err
	}

	total := &graphqlbackend.DiffStat{}
	for _, p := range patches {
		fileDiffs, err := p.FileDiffs(ctx, &graphqlbackend.FileDiffsConnectionArgs{})
		if err != nil {
			return nil, err
		}

		s, err := fileDiffs.DiffStat(ctx)
		if err != nil {
			return nil, err
		}

		total.AddDiffStat(s)
	}

	return total, nil
}

func (r *patchSetResolver) PreviewURL() string {
	u := globals.ExternalURL().ResolveReference(&url.URL{Path: "/campaigns/new"})
	q := url.Values{}
	q.Set("patchSet", string(r.ID()))
	u.RawQuery = q.Encode()
	return u.String()
}

type patchesConnectionResolver struct {
	store *ee.Store
	opts  ee.ListPatchesOpts

	// cache results because they are used by multiple fields
	once                   sync.Once
	jobs                   []*campaigns.Patch
	reposByID              map[api.RepoID]*repos.Repo
	changesetJobsByPatchID map[int64]*campaigns.ChangesetJob
	next                   int64
	err                    error
}

func (r *patchesConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.PatchResolver, error) {
	jobs, reposByID, changesetJobsByPatchID, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := make([]graphqlbackend.PatchResolver, 0, len(jobs))
	for _, j := range jobs {
		repo, ok := reposByID[j.RepoID]
		if !ok {
			return nil, fmt.Errorf("failed to load repo %d", j.RepoID)
		}

		resolver := &patchResolver{
			store:         r.store,
			patch:         j,
			preloadedRepo: repo,
			// We set this to true, because we tried to preload the
			// changestJob, but maybe we couldn't find one.
			attemptedPreloadChangesetJob: true,
		}

		changesetJob, ok := changesetJobsByPatchID[j.ID]
		if ok {
			resolver.preloadedChangesetJob = changesetJob
		}

		resolvers = append(resolvers, resolver)
	}
	return resolvers, nil
}

func (r *patchesConnectionResolver) compute(ctx context.Context) ([]*campaigns.Patch, map[api.RepoID]*repos.Repo, map[int64]*campaigns.ChangesetJob, int64, error) {
	r.once.Do(func() {
		r.jobs, r.next, r.err = r.store.ListPatches(ctx, r.opts)
		if r.err != nil {
			return
		}

		reposStore := repos.NewDBStore(r.store.DB(), sql.TxOptions{})
		repoIDs := make([]api.RepoID, len(r.jobs))
		for i, j := range r.jobs {
			repoIDs[i] = j.RepoID
		}

		rs, err := reposStore.ListRepos(ctx, repos.StoreListReposArgs{IDs: repoIDs})
		if err != nil {
			r.err = err
			return
		}

		r.reposByID = make(map[api.RepoID]*repos.Repo, len(rs))
		for _, repo := range rs {
			r.reposByID[repo.ID] = repo
		}

		cs, _, err := r.store.ListChangesetJobs(ctx, ee.ListChangesetJobsOpts{
			PatchSetID: r.opts.PatchSetID,
			Limit:      -1,
		})
		if err != nil {
			r.err = err
			return
		}
		r.changesetJobsByPatchID = make(map[int64]*campaigns.ChangesetJob, len(cs))
		for _, c := range cs {
			r.changesetJobsByPatchID[c.PatchID] = c
		}
	})
	return r.jobs, r.reposByID, r.changesetJobsByPatchID, r.next, r.err
}

func (r *patchesConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	opts := ee.CountPatchesOpts{
		PatchSetID:                r.opts.PatchSetID,
		OnlyWithDiff:              r.opts.OnlyWithDiff,
		OnlyUnpublishedInCampaign: r.opts.OnlyUnpublishedInCampaign,
	}
	count, err := r.store.CountPatches(ctx, opts)
	return int32(count), err
}

func (r *patchesConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, _, _, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(next != 0), nil
}

type patchResolver struct {
	store *ee.Store

	patch         *campaigns.Patch
	preloadedRepo *repos.Repo

	// Set if we tried to preload the changesetjob
	attemptedPreloadChangesetJob bool
	// This is only set if we tried to preload and found a ChangesetJob. If we
	// tried preloading, but couldn't find anything, it's nil.
	preloadedChangesetJob *campaigns.ChangesetJob

	// cache repo because it's called more than one time
	once   sync.Once
	err    error
	repo   *graphqlbackend.RepositoryResolver
	commit *graphqlbackend.GitCommitResolver
}

func (r *patchResolver) computeRepoCommit(ctx context.Context) (*graphqlbackend.RepositoryResolver, *graphqlbackend.GitCommitResolver, error) {
	r.once.Do(func() {
		if r.preloadedRepo != nil {
			r.repo = newRepositoryResolver(r.preloadedRepo)
		} else {
			r.repo, r.err = graphqlbackend.RepositoryByIDInt32(ctx, r.patch.RepoID)
			if r.err != nil {
				return
			}
		}
		args := &graphqlbackend.RepositoryCommitArgs{Rev: string(r.patch.Rev)}
		r.commit, r.err = r.repo.Commit(ctx, args)
	})
	return r.repo, r.commit, r.err
}

func (r *patchResolver) ID() graphql.ID {
	return marshalPatchID(r.patch.ID)
}

func (r *patchResolver) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	repo, _, err := r.computeRepoCommit(ctx)
	return repo, err
}

func (r *patchResolver) BaseRepository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	return r.Repository(ctx)
}

func (r *patchResolver) Diff() graphqlbackend.PatchResolver {
	return r
}

func (r *patchResolver) FileDiffs(ctx context.Context, args *graphqlbackend.FileDiffsConnectionArgs) (graphqlbackend.PreviewFileDiffConnection, error) {
	_, commit, err := r.computeRepoCommit(ctx)
	if err != nil {
		return nil, err
	}
	return &previewFileDiffConnectionResolver{
		patch:  r.patch,
		commit: commit,
		first:  args.First,
		after:  args.After,
	}, nil
}

func (r *patchResolver) PublicationEnqueued(ctx context.Context) (bool, error) {
	// We tried to preload a ChangesetJob for this Patch
	if r.attemptedPreloadChangesetJob {
		if r.preloadedChangesetJob == nil {
			return false, nil
		}
		return r.preloadedChangesetJob.FinishedAt.IsZero(), nil
	}

	cj, err := r.store.GetChangesetJob(ctx, ee.GetChangesetJobOpts{PatchID: r.patch.ID})
	if err != nil && err != ee.ErrNoResults {
		return false, err
	}
	if err == ee.ErrNoResults {
		return false, nil
	}

	// FinishedAt is always set once the ChangesetJob is finished, even if it
	// failed. If it's zero, we're still executing the job. If not, we're
	// done and the "publication" is not "enqueued" anymore.
	return cj.FinishedAt.IsZero(), nil
}

type previewFileDiffConnectionResolver struct {
	patch  *campaigns.Patch
	commit *graphqlbackend.GitCommitResolver
	first  *int32
	after  *string

	// cache result because it is used by multiple fields
	once        sync.Once
	fileDiffs   []*diff.FileDiff
	afterIdx    int32
	hasNextPage bool
	err         error
}

func (r *previewFileDiffConnectionResolver) compute(ctx context.Context) ([]*diff.FileDiff, int32, error) {
	r.once.Do(func() {
		r.fileDiffs, r.err = diff.ParseMultiFileDiff([]byte(r.patch.Diff))
		if r.err != nil {
			return
		}
		if r.after != nil {
			parsedIdx, err := strconv.ParseInt(*r.after, 0, 32)
			if err != nil {
				r.err = err
				return
			}
			if parsedIdx < 0 {
				parsedIdx = 0
			}
			r.afterIdx = int32(parsedIdx)
		}
		if r.first != nil && len(r.fileDiffs) > int(*r.first+r.afterIdx) {
			r.hasNextPage = true
		}
	})
	return r.fileDiffs, r.afterIdx, r.err
}

func (r *previewFileDiffConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.PreviewFileDiff, error) {
	fileDiffs, afterIdx, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if r.first != nil && int(*r.first+afterIdx) <= len(fileDiffs) {
		fileDiffs = fileDiffs[afterIdx:(*r.first + afterIdx)]
	} else if int(afterIdx) <= len(fileDiffs) {
		fileDiffs = fileDiffs[afterIdx:]
	} else {
		fileDiffs = []*diff.FileDiff{}
	}

	resolvers := make([]graphqlbackend.PreviewFileDiff, len(fileDiffs))
	for i, fileDiff := range fileDiffs {
		resolvers[i] = &previewFileDiffResolver{
			fileDiff: fileDiff,
			commit:   r.commit,
		}
	}
	return resolvers, nil
}

func (r *previewFileDiffConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	fileDiffs, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if r.first == nil || (len(fileDiffs) > int(*r.first)) {
		n := int32(len(fileDiffs))
		return &n, nil
	}
	// This is taken from fileDiffConnectionResolver.TotalCount
	return nil, nil
}

func (r *previewFileDiffConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, afterIdx, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if !r.hasNextPage {
		return graphqlutil.HasNextPage(r.hasNextPage), nil
	}
	next := int32(afterIdx)
	if r.first != nil {
		next += *r.first
	}
	return graphqlutil.NextPageCursor(strconv.Itoa(int(next))), nil
}

func (r *previewFileDiffConnectionResolver) DiffStat(ctx context.Context) (*graphqlbackend.DiffStat, error) {
	fileDiffs, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	stat := &graphqlbackend.DiffStat{}
	for _, fileDiff := range fileDiffs {
		s := fileDiff.Stat()
		stat.AddStat(s)
	}
	return stat, nil
}

func (r *previewFileDiffConnectionResolver) RawDiff(ctx context.Context) (string, error) {
	fileDiffs, _, err := r.compute(ctx)
	if err != nil {
		return "", err
	}
	b, err := diff.PrintMultiFileDiff(fileDiffs)
	return string(b), err
}

type previewFileDiffResolver struct {
	fileDiff *diff.FileDiff
	commit   *graphqlbackend.GitCommitResolver
}

func (r *previewFileDiffResolver) OldPath() *string { return diffPathOrNull(r.fileDiff.OrigName) }
func (r *previewFileDiffResolver) NewPath() *string { return diffPathOrNull(r.fileDiff.NewName) }

func (r *previewFileDiffResolver) Hunks() []*graphqlbackend.DiffHunk {
	highlighter := &previewFileDiffHighlighter{previewFileDiffResolver: r}
	hunks := make([]*graphqlbackend.DiffHunk, len(r.fileDiff.Hunks))
	for i, hunk := range r.fileDiff.Hunks {
		hunks[i] = graphqlbackend.NewDiffHunk(hunk, highlighter)
	}
	return hunks
}

func (r *previewFileDiffResolver) Stat() *graphqlbackend.DiffStat {
	stat := r.fileDiff.Stat()
	return graphqlbackend.NewDiffStat(stat)
}

func (r *previewFileDiffResolver) OldFile() *graphqlbackend.GitTreeEntryResolver {
	if diffPathOrNull(r.fileDiff.OrigName) == nil {
		return nil
	}
	fileStat := graphqlbackend.CreateFileInfo(r.fileDiff.OrigName, false)
	return graphqlbackend.NewGitTreeEntryResolver(r.commit, fileStat)
}

func (r *previewFileDiffResolver) InternalID() string {
	b := sha256.Sum256([]byte(fmt.Sprintf("%d:%s:%s", len(r.fileDiff.OrigName), r.fileDiff.OrigName, r.fileDiff.NewName)))
	return hex.EncodeToString(b[:])[:32]
}

func diffPathOrNull(path string) *string {
	if path == "/dev/null" || path == "" {
		return nil
	}
	return &path
}
