package resolvers

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
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
	patches, _, err := store.ListPatches(ctx, opts)
	if err != nil {
		return nil, err
	}

	repoIDs := make([]api.RepoID, 0, len(patches))
	for _, p := range patches {
		repoIDs = append(repoIDs, p.RepoID)
	}

	// 🚨 SECURITY: We use db.Repos.GetByIDs to filter out repositories the
	// user doesn't have access to.
	accessibleRepos, err := db.Repos.GetByIDs(ctx, repoIDs...)
	if err != nil {
		return nil, err
	}

	accessibleRepoIDs := make(map[api.RepoID]struct{}, len(accessibleRepos))
	for _, r := range accessibleRepos {
		accessibleRepoIDs[r.ID] = struct{}{}
	}

	total := &graphqlbackend.DiffStat{}
	for _, p := range patches {
		// 🚨 SECURITY: We filter out the patches that belong to repositories the
		// user does NOT have access to.
		if _, ok := accessibleRepoIDs[p.RepoID]; !ok {
			continue
		}

		s, ok := p.DiffStat()
		if !ok {
			return nil, fmt.Errorf("patch %d has no diff stat", p.ID)
		}

		total.AddStat(s)
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
	patches                []*campaigns.Patch
	reposByID              map[api.RepoID]*types.Repo
	changesetJobsByPatchID map[int64]*campaigns.ChangesetJob
	next                   int64
	err                    error
}

func (r *patchesConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.PatchInterfaceResolver, error) {
	patches, reposByID, changesetJobsByPatchID, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := make([]graphqlbackend.PatchInterfaceResolver, 0, len(patches))
	for _, j := range patches {
		repo, ok := reposByID[j.RepoID]
		if !ok {
			// If it's not in reposByID the repository was either deleted or
			// filtered out by the authz-filter.
			// Use a hiddenPatchResolver.
			resolvers = append(resolvers, &hiddenPatchResolver{patch: j})
			continue
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

func (r *patchesConnectionResolver) compute(ctx context.Context) ([]*campaigns.Patch, map[api.RepoID]*types.Repo, map[int64]*campaigns.ChangesetJob, int64, error) {
	r.once.Do(func() {
		r.patches, r.next, r.err = r.store.ListPatches(ctx, r.opts)
		if r.err != nil {
			return
		}

		repoIDs := make([]api.RepoID, len(r.patches))
		for i, j := range r.patches {
			repoIDs[i] = j.RepoID
		}

		// 🚨 SECURITY: db.Repos.GetByIDs uses the authzFilter under the hood and
		// filters out repositories that the user doesn't have access to.
		rs, err := db.Repos.GetByIDs(ctx, repoIDs...)
		if err != nil {
			r.err = err
			return
		}

		r.reposByID = make(map[api.RepoID]*types.Repo, len(rs))
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
	return r.patches, r.reposByID, r.changesetJobsByPatchID, r.next, r.err
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
	preloadedRepo *types.Repo

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

func (r *patchResolver) ToPatch() (graphqlbackend.PatchResolver, bool) {
	return r, true
}

func (r *patchResolver) ToHiddenPatch() (graphqlbackend.HiddenPatchResolver, bool) {
	return nil, false
}

func (r *patchResolver) computeRepoCommit(ctx context.Context) (*graphqlbackend.RepositoryResolver, *graphqlbackend.GitCommitResolver, error) {
	r.once.Do(func() {
		if r.preloadedRepo != nil {
			r.repo = graphqlbackend.NewRepositoryResolver(r.preloadedRepo)
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

func (r *patchResolver) Diff() graphqlbackend.PatchResolver {
	return r
}

func (r *patchResolver) FileDiffs(ctx context.Context, args *graphqlbackend.FileDiffsConnectionArgs) (graphqlbackend.FileDiffConnection, error) {
	_, commit, err := r.computeRepoCommit(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlbackend.NewFileDiffConnectionResolver(commit, commit, args, fileDiffConnectionCompute(r.patch), previewNewFile), nil
}

func fileDiffConnectionCompute(patch *campaigns.Patch) func(ctx context.Context, args *graphqlbackend.FileDiffsConnectionArgs) ([]*diff.FileDiff, int32, bool, error) {
	var (
		once        sync.Once
		fileDiffs   []*diff.FileDiff
		afterIdx    int32
		hasNextPage bool
		err         error
	)
	return func(ctx context.Context, args *graphqlbackend.FileDiffsConnectionArgs) ([]*diff.FileDiff, int32, bool, error) {
		once.Do(func() {
			if args.After != nil {
				parsedIdx, err := strconv.ParseInt(*args.After, 0, 32)
				if err != nil {
					return
				}
				if parsedIdx < 0 {
					parsedIdx = 0
				}
				afterIdx = int32(parsedIdx)
			}
			totalAmount := afterIdx
			if args.First != nil {
				totalAmount += *args.First
			}

			dr := diff.NewMultiFileDiffReader(strings.NewReader(patch.Diff))
			for {
				var fileDiff *diff.FileDiff
				fileDiff, err = dr.ReadFile()
				if err == io.EOF {
					err = nil
					break
				}
				if err != nil {
					return
				}
				fileDiffs = append(fileDiffs, fileDiff)
				if len(fileDiffs) == int(totalAmount) {
					// Check for hasNextPage.
					_, err = dr.ReadFile()
					if err != nil && err != io.EOF {
						return
					}
					if err == io.EOF {
						err = nil
					} else {
						hasNextPage = true
					}
					break
				}
			}
		})
		return fileDiffs, afterIdx, hasNextPage, err
	}
}

func previewNewFile(r *graphqlbackend.FileDiffResolver) graphqlbackend.FileResolver {
	fileStat := graphqlbackend.CreateFileInfo(r.FileDiff.NewName, false)
	return graphqlbackend.NewVirtualFileResolver(fileStat, fileDiffVirtualFileContent(r))
}

func fileDiffVirtualFileContent(r *graphqlbackend.FileDiffResolver) graphqlbackend.FileContentFunc {
	var (
		once       sync.Once
		newContent string
		err        error
	)
	return func(ctx context.Context) (string, error) {
		once.Do(func() {
			var oldContent string
			if oldFile := r.OldFile(); oldFile != nil {
				var err error
				oldContent, err = r.OldFile().Content(ctx)
				if err != nil {
					return
				}
			}
			newContent = applyPatch(oldContent, r.FileDiff)
		})
		return newContent, err
	}
}

func applyPatch(fileContent string, fileDiff *diff.FileDiff) string {
	contentLines := strings.Split(fileContent, "\n")
	newContentLines := make([]string, 0)
	var lastLine int32 = 1
	// Assumes the hunks are sorted by ascending lines.
	for _, hunk := range fileDiff.Hunks {
		// Detect holes.
		if hunk.OrigStartLine != 0 && hunk.OrigStartLine != lastLine {
			originalLines := contentLines[lastLine-1 : hunk.OrigStartLine-1]
			newContentLines = append(newContentLines, originalLines...)
			lastLine += int32(len(originalLines))
		}
		hunkLines := strings.Split(string(hunk.Body), "\n")
		for _, line := range hunkLines {
			switch {
			case line == "":
				// Skip
			case strings.HasPrefix(line, "-"):
				lastLine++
			case strings.HasPrefix(line, "+"):
				newContentLines = append(newContentLines, line[1:])
			default:
				newContentLines = append(newContentLines, contentLines[lastLine-1])
				lastLine++
			}
		}
	}
	// Append remaining lines from original file.
	if origLines := int32(len(contentLines)); origLines > 0 && origLines != lastLine {
		newContentLines = append(newContentLines, contentLines[lastLine-1:]...)
	}
	return strings.Join(newContentLines, "\n")
}

type hiddenPatchResolver struct {
	patch *campaigns.Patch
}

func (r *hiddenPatchResolver) ToPatch() (graphqlbackend.PatchResolver, bool) {
	return nil, false
}

func (r *hiddenPatchResolver) ToHiddenPatch() (graphqlbackend.HiddenPatchResolver, bool) {
	return r, true
}

func (r *hiddenPatchResolver) ID() graphql.ID {
	return marshalPatchID(r.patch.ID)
}
