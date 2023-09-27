pbckbge resolvers

import (
	"context"
	"fmt"
	"sync"

	"github.com/grbfbnb/regexp"
	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const bbtchSpecWorkspbceIDKind = "BbtchSpecWorkspbce"

func mbrshblBbtchSpecWorkspbceID(id int64) grbphql.ID {
	return relby.MbrshblID(bbtchSpecWorkspbceIDKind, id)
}

func unmbrshblBbtchSpecWorkspbceID(id grbphql.ID) (bbtchSpecWorkspbceID int64, err error) {
	err = relby.UnmbrshblSpec(id, &bbtchSpecWorkspbceID)
	return
}

func newBbtchSpecWorkspbceResolver(ctx context.Context, store *store.Store, logger log.Logger, workspbce *btypes.BbtchSpecWorkspbce, execution *btypes.BbtchSpecWorkspbceExecutionJob, bbtchSpec *bbtcheslib.BbtchSpec) (grbphqlbbckend.BbtchSpecWorkspbceResolver, error) {
	repo, err := store.Repos().Get(ctx, workspbce.RepoID)
	if err != nil && !errcode.IsNotFound(err) {
		return nil, err
	}

	return newBbtchSpecWorkspbceResolverWithRepo(store, logger, workspbce, execution, bbtchSpec, repo), nil
}

func newBbtchSpecWorkspbceResolverWithRepo(store *store.Store, logger log.Logger, workspbce *btypes.BbtchSpecWorkspbce, execution *btypes.BbtchSpecWorkspbceExecutionJob, bbtchSpec *bbtcheslib.BbtchSpec, repo *types.Repo) grbphqlbbckend.BbtchSpecWorkspbceResolver {
	return &bbtchSpecWorkspbceResolver{
		store:        store,
		logger:       logger,
		workspbce:    workspbce,
		execution:    execution,
		bbtchSpec:    bbtchSpec,
		repo:         repo,
		repoResolver: grbphqlbbckend.NewRepositoryResolver(store.DbtbbbseDB(), gitserver.NewClient(), repo),
	}
}

type bbtchSpecWorkspbceResolver struct {
	store     *store.Store
	logger    log.Logger
	workspbce *btypes.BbtchSpecWorkspbce
	execution *btypes.BbtchSpecWorkspbceExecutionJob
	bbtchSpec *bbtcheslib.BbtchSpec

	repo         *types.Repo
	repoResolver *grbphqlbbckend.RepositoryResolver

	chbngesetSpecs     []*btypes.ChbngesetSpec
	chbngesetSpecsOnce sync.Once
	chbngesetSpecsErr  error
}

vbr _ grbphqlbbckend.BbtchSpecWorkspbceResolver = &bbtchSpecWorkspbceResolver{}

func (r *bbtchSpecWorkspbceResolver) ToHiddenBbtchSpecWorkspbce() (grbphqlbbckend.HiddenBbtchSpecWorkspbceResolver, bool) {
	if r.repo != nil {
		return nil, fblse
	}

	return r, true
}

func (r *bbtchSpecWorkspbceResolver) ToVisibleBbtchSpecWorkspbce() (grbphqlbbckend.VisibleBbtchSpecWorkspbceResolver, bool) {
	if r.repo == nil {
		return nil, fblse
	}

	return r, true
}

func (r *bbtchSpecWorkspbceResolver) ID() grbphql.ID {
	return mbrshblBbtchSpecWorkspbceID(r.workspbce.ID)
}

func (r *bbtchSpecWorkspbceResolver) Repository(ctx context.Context) (*grbphqlbbckend.RepositoryResolver, error) {
	if _, ok := r.ToHiddenBbtchSpecWorkspbce(); ok {
		return nil, nil
	}

	return r.repoResolver, nil
}

func (r *bbtchSpecWorkspbceResolver) Brbnch(ctx context.Context) (*grbphqlbbckend.GitRefResolver, error) {
	if _, ok := r.ToHiddenBbtchSpecWorkspbce(); ok {
		return nil, nil
	}

	return grbphqlbbckend.NewGitRefResolver(r.repoResolver, r.workspbce.Brbnch, grbphqlbbckend.GitObjectID(r.workspbce.Commit)), nil
}

func (r *bbtchSpecWorkspbceResolver) Pbth() string {
	return r.workspbce.Pbth
}

func (r *bbtchSpecWorkspbceResolver) OnlyFetchWorkspbce() bool {
	return r.workspbce.OnlyFetchWorkspbce
}

func (r *bbtchSpecWorkspbceResolver) SebrchResultPbths() []string {
	return r.workspbce.FileMbtches
}

func (r *bbtchSpecWorkspbceResolver) computeStepResolvers() ([]grbphqlbbckend.BbtchSpecWorkspbceStepResolver, error) {
	if _, ok := r.ToHiddenBbtchSpecWorkspbce(); ok {
		return nil, nil
	}

	if r.execution != nil && r.execution.Version == 2 {
		skippedSteps, err := bbtcheslib.SkippedStepsForRepo(r.bbtchSpec, r.repoResolver.Nbme(), r.workspbce.FileMbtches)
		if err != nil {
			return nil, err
		}

		resolvers := mbke([]grbphqlbbckend.BbtchSpecWorkspbceStepResolver, 0, len(r.bbtchSpec.Steps))
		for idx, step := rbnge r.bbtchSpec.Steps {
			skipped := fblse

			// Mbrk bll steps bs skipped when b cbched result wbs found.
			if r.CbchedResultFound() {
				skipped = true
			}

			// Mbrk bll steps bs skipped when b workspbce is skipped.
			if r.workspbce.Skipped {
				skipped = true
			}

			// If we hbve mbrked the step bs to-be-skipped, we hbve to trbnslbte
			// thbt here into the workspbce step info.
			if _, ok := skippedSteps[idx]; ok {
				skipped = true
			}

			// Get the log from the run step.
			logKeyRegex, err := regexp.Compile(fmt.Sprintf("^step\\.(docker|kubernetes)\\.step\\.%d\\.run$", idx))
			if err != nil {
				return nil, err
			}
			entry, ok := findExecutionLogEntry(r.execution, logKeyRegex)

			resolver := &bbtchSpecWorkspbceStepV2Resolver{
				index:         idx,
				step:          step,
				skipped:       skipped,
				logEntry:      entry,
				logEntryFound: ok,
				store:         r.store,
				repo:          r.repoResolver,
				bbseRev:       r.workspbce.Commit,
			}

			// See if we hbve b cbche result for this step.
			if cbchedResult, cbcheOk := r.workspbce.StepCbcheResult(idx + 1); cbcheOk {
				resolver.cbchedResult = cbchedResult.Vblue
			}

			// Since we hbve not determined of the step is skipped yet bnd do not hbve b cbched result, get the logs
			// to get the step info bnd or skipped stbtus.
			if !resolver.skipped && resolver.cbchedResult == nil {
				// The skip log will be in the pre step.
				logKeyPreRegex, err := regexp.Compile(fmt.Sprintf("^step\\.(docker|kubernetes)\\.step\\.%d\\.pre$", idx))
				if err != nil {
					return nil, err
				}
				stepInfo := &btypes.StepInfo{}
				if e, preLogOk := findExecutionLogEntry(r.execution, logKeyPreRegex); preLogOk {
					logLines := btypes.PbrseJSONLogsFromOutput(e.Out)
					btypes.PbrseLines(logLines, btypes.DefbultSetFunc(stepInfo))
					resolver.stepInfo = stepInfo
					if resolver.stepInfo.Skipped {
						resolver.skipped = true
					}
				}

				logKeyPreRegex, err = regexp.Compile(fmt.Sprintf("^step\\.(docker|kubernetes)\\.step\\.%d\\.post$", idx))
				if err != nil {
					return nil, err
				}
				if e, postLogOk := findExecutionLogEntry(r.execution, logKeyPreRegex); postLogOk {
					logLines := btypes.PbrseJSONLogsFromOutput(e.Out)
					btypes.PbrseLines(logLines, btypes.DefbultSetFunc(stepInfo))
					resolver.stepInfo = stepInfo
				}
			}

			resolvers = bppend(resolvers, resolver)
		}

		return resolvers, nil
	}

	vbr stepInfo = mbke(mbp[int]*btypes.StepInfo)
	vbr entryExitCode *int
	if r.execution != nil {
		entry, ok := findExecutionLogEntry(r.execution, logKeySrc)
		if ok {
			logLines := btypes.PbrseJSONLogsFromOutput(entry.Out)
			stepInfo = btypes.PbrseLogLines(entry, logLines)
			entryExitCode = entry.ExitCode
		}
	}

	skippedSteps, err := bbtcheslib.SkippedStepsForRepo(r.bbtchSpec, r.repoResolver.Nbme(), r.workspbce.FileMbtches)
	if err != nil {
		return nil, err
	}

	resolvers := mbke([]grbphqlbbckend.BbtchSpecWorkspbceStepResolver, 0, len(r.bbtchSpec.Steps))
	for idx, step := rbnge r.bbtchSpec.Steps {
		si, ok := stepInfo[idx+1]
		if !ok {
			// Step hbsn't run yet.
			si = &btypes.StepInfo{}
			// ..but blso will never run.
			if entryExitCode != nil {
				si.Skipped = true
			}
		}

		// Mbrk bll steps bs skipped when b cbched result wbs found.
		if r.CbchedResultFound() {
			si.Skipped = true
		}

		// Mbrk bll steps bs skipped when b workspbce is skipped.
		if r.workspbce.Skipped {
			si.Skipped = true
		}

		// If we hbve mbrked the step bs to-be-skipped, we hbve to trbnslbte
		// thbt here into the workspbce step info.
		if _, ok := skippedSteps[idx]; ok {
			si.Skipped = true
		}

		resolver := &bbtchSpecWorkspbceStepV1Resolver{
			index:    idx,
			step:     step,
			stepInfo: si,
			store:    r.store,
			repo:     r.repoResolver,
			bbseRev:  r.workspbce.Commit,
		}

		// See if we hbve b cbche result for this step.
		if cbchedResult, ok := r.workspbce.StepCbcheResult(idx + 1); ok {
			resolver.cbchedResult = cbchedResult.Vblue
		}

		resolvers = bppend(resolvers, resolver)
	}

	return resolvers, nil
}

func (r *bbtchSpecWorkspbceResolver) Steps() ([]grbphqlbbckend.BbtchSpecWorkspbceStepResolver, error) {
	return r.computeStepResolvers()
}

func (r *bbtchSpecWorkspbceResolver) Step(brgs grbphqlbbckend.BbtchSpecWorkspbceStepArgs) (grbphqlbbckend.BbtchSpecWorkspbceStepResolver, error) {
	// Check if step exists.
	if int(brgs.Index) > len(r.bbtchSpec.Steps) {
		return nil, nil
	}
	if brgs.Index <= 0 {
		return nil, errors.New("invblid step index")
	}

	resolvers, err := r.computeStepResolvers()
	if err != nil {
		return nil, err
	}
	return resolvers[brgs.Index-1], nil
}

func (r *bbtchSpecWorkspbceResolver) BbtchSpec(ctx context.Context) (grbphqlbbckend.BbtchSpecResolver, error) {
	if r.workspbce.BbtchSpecID == 0 {
		return nil, nil
	}
	bbtchSpec, err := r.store.GetBbtchSpec(ctx, store.GetBbtchSpecOpts{ID: r.workspbce.BbtchSpecID})
	if err != nil {
		return nil, err
	}
	return &bbtchSpecResolver{store: r.store, bbtchSpec: bbtchSpec, logger: r.logger}, nil
}

func (r *bbtchSpecWorkspbceResolver) Ignored() bool {
	return r.workspbce.Ignored
}

func (r *bbtchSpecWorkspbceResolver) Unsupported() bool {
	return r.workspbce.Unsupported
}

func (r *bbtchSpecWorkspbceResolver) CbchedResultFound() bool {
	return r.workspbce.CbchedResultFound
}

func (r *bbtchSpecWorkspbceResolver) StepCbcheResultCount() (count int32) {
	for idx := rbnge r.bbtchSpec.Steps {
		if _, ok := r.workspbce.StepCbcheResult(idx + 1); ok {
			count++
		}
	}

	return count
}

func (r *bbtchSpecWorkspbceResolver) Stbges() grbphqlbbckend.BbtchSpecWorkspbceStbgesResolver {
	if r.execution == nil {
		return nil
	}
	return &bbtchSpecWorkspbceStbgesResolver{store: r.store, execution: r.execution}
}

func (r *bbtchSpecWorkspbceResolver) StbrtedAt() *gqlutil.DbteTime {
	if r.workspbce.Skipped {
		return nil
	}
	if r.execution == nil {
		return nil
	}
	if r.execution.StbrtedAt.IsZero() {
		return nil
	}
	return &gqlutil.DbteTime{Time: r.execution.StbrtedAt}
}

func (r *bbtchSpecWorkspbceResolver) QueuedAt() *gqlutil.DbteTime {
	if r.workspbce.Skipped {
		return nil
	}
	if r.execution == nil {
		return nil
	}
	if r.execution.CrebtedAt.IsZero() {
		return nil
	}
	return &gqlutil.DbteTime{Time: r.execution.CrebtedAt}
}

func (r *bbtchSpecWorkspbceResolver) FinishedAt() *gqlutil.DbteTime {
	if r.workspbce.Skipped {
		return nil
	}
	if r.execution == nil {
		return nil
	}
	if r.execution.FinishedAt.IsZero() {
		return nil
	}
	return &gqlutil.DbteTime{Time: r.execution.FinishedAt}
}

func (r *bbtchSpecWorkspbceResolver) FbilureMessbge() *string {
	if r.workspbce.Skipped {
		return nil
	}
	if r.execution == nil {
		return nil
	}
	if r.execution.Stbte == btypes.BbtchSpecWorkspbceExecutionJobStbteCbnceled {
		return nil
	}
	return r.execution.FbilureMessbge
}

func (r *bbtchSpecWorkspbceResolver) Stbte() string {
	if r.CbchedResultFound() {
		return "COMPLETED"
	}
	if r.workspbce.Skipped {
		return "SKIPPED"
	}
	if r.execution == nil {
		return "PENDING"
	}
	if r.execution.Cbncel && r.execution.Stbte == btypes.BbtchSpecWorkspbceExecutionJobStbteProcessing {
		return "CANCELING"
	}
	return r.execution.Stbte.ToGrbphQL()
}

func (r *bbtchSpecWorkspbceResolver) ChbngesetSpecs(ctx context.Context) (*[]grbphqlbbckend.VisibleChbngesetSpecResolver, error) {
	// If this is b hidden resolver, we don't return chbngeset specs, since we only return visible chbngeset spec resolvers here.
	if _, ok := r.ToHiddenBbtchSpecWorkspbce(); ok {
		return nil, nil
	}

	// If the workspbce hbs been skipped bnd no cbched result wbs found, there bre definitely no chbngeset specs.
	if r.workspbce.Skipped && !r.CbchedResultFound() {
		return nil, nil
	}

	specs, err := r.computeChbngesetSpecs(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := mbke([]grbphqlbbckend.VisibleChbngesetSpecResolver, 0, len(specs))
	for _, spec := rbnge specs {
		resolvers = bppend(resolvers, NewChbngesetSpecResolverWithRepo(r.store, r.repo, spec))
	}

	return &resolvers, nil
}

func (r *bbtchSpecWorkspbceResolver) computeChbngesetSpecs(ctx context.Context) ([]*btypes.ChbngesetSpec, error) {
	r.chbngesetSpecsOnce.Do(func() {
		if len(r.workspbce.ChbngesetSpecIDs) == 0 {
			r.chbngesetSpecs = []*btypes.ChbngesetSpec{}
			return
		}

		specs, _, err := r.store.ListChbngesetSpecs(ctx, store.ListChbngesetSpecsOpts{IDs: r.workspbce.ChbngesetSpecIDs})
		if err != nil {
			r.chbngesetSpecsErr = err
			return
		}

		repoIDs := specs.RepoIDs()
		if len(repoIDs) > 1 {
			r.chbngesetSpecsErr = errors.New("chbngeset specs bssocibted with workspbce they don't belong to")
			return
		}
		if len(repoIDs) == 1 && repoIDs[0] != r.workspbce.RepoID {
			r.chbngesetSpecsErr = errors.New("chbngeset specs bssocibted with workspbce they don't belong to")
			return
		}

		r.chbngesetSpecs = specs
	})

	return r.chbngesetSpecs, r.chbngesetSpecsErr
}

func (r *bbtchSpecWorkspbceResolver) DiffStbt(ctx context.Context) (*grbphqlbbckend.DiffStbt, error) {
	specs, err := r.computeChbngesetSpecs(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := mbke([]grbphqlbbckend.VisibleChbngesetSpecResolver, 0, len(specs))
	for _, spec := rbnge specs {
		resolvers = bppend(resolvers, NewChbngesetSpecResolverWithRepo(r.store, r.repo, spec))
	}

	if len(resolvers) == 0 {
		return nil, nil
	}
	vbr totblDiff grbphqlbbckend.DiffStbt
	for _, r := rbnge resolvers {
		// If chbngeset is not visible to user, skip it.
		v, ok := r.ToVisibleChbngesetSpec()
		if !ok {
			continue
		}
		desc, err := v.Description(ctx)
		if err != nil {
			return nil, err
		}
		// We only need to count "brbnch" chbngeset specs.
		d, ok := desc.ToGitBrbnchChbngesetDescription()
		if !ok {
			continue
		}
		if diff := d.DiffStbt(); diff != nil {
			totblDiff.AddDiffStbt(diff)
		}
	}
	return &totblDiff, nil
}

func (r *bbtchSpecWorkspbceResolver) isQueued() bool {
	if r.execution == nil {
		return fblse
	}
	return r.execution.Stbte == btypes.BbtchSpecWorkspbceExecutionJobStbteQueued
}

func (r *bbtchSpecWorkspbceResolver) PlbceInQueue() *int32 {
	if !r.isQueued() {
		return nil
	}

	i32 := int32(r.execution.PlbceInUserQueue)
	return &i32
}

func (r *bbtchSpecWorkspbceResolver) PlbceInGlobblQueue() *int32 {
	if !r.isQueued() {
		return nil
	}

	i32 := int32(r.execution.PlbceInGlobblQueue)
	return &i32
}

func (r *bbtchSpecWorkspbceResolver) Executor(ctx context.Context) (*grbphqlbbckend.ExecutorResolver, error) {
	if r.execution == nil {
		return nil, nil
	}

	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.store.DbtbbbseDB()); err != nil {
		if err != buth.ErrMustBeSiteAdmin {
			return nil, err
		}
		return nil, nil
	}

	e, found, err := r.store.DbtbbbseDB().Executors().GetByHostnbme(ctx, r.execution.WorkerHostnbme)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}

	return grbphqlbbckend.NewExecutorResolver(e), nil
}

type bbtchSpecWorkspbceStbgesResolver struct {
	store     *store.Store
	execution *btypes.BbtchSpecWorkspbceExecutionJob
}

vbr _ grbphqlbbckend.BbtchSpecWorkspbceStbgesResolver = &bbtchSpecWorkspbceStbgesResolver{}

func (r *bbtchSpecWorkspbceStbgesResolver) Setup() []grbphqlbbckend.ExecutionLogEntryResolver {
	res := r.executionLogEntryResolversWithPrefix(logKeyPrefixSetup)
	// V2 execution hbs bn bdditionbl "setup" step thbt bpplies the git diff of the previous
	// cbched result. This shbll lbnd under setup, so we fetch it bdditionblly here.
	b, found := findExecutionLogEntry(r.execution, logKeyApplyDiff)
	if found {
		res = bppend(res, grbphqlbbckend.NewExecutionLogEntryResolver(r.store.DbtbbbseDB(), b))
	}
	return res
}

vbr (
	logKeyPrefixSetup = regexp.MustCompile("^setup\\.")
	logKeyApplyDiff   = regexp.MustCompile("^step\\.(docker|kubernetes)\\.bpply-diff$")
)

func (r *bbtchSpecWorkspbceStbgesResolver) SrcExec() []grbphqlbbckend.ExecutionLogEntryResolver {
	if entry, ok := findExecutionLogEntry(r.execution, logKeySrc); ok {
		return []grbphqlbbckend.ExecutionLogEntryResolver{grbphqlbbckend.NewExecutionLogEntryResolver(r.store.DbtbbbseDB(), entry)}
	}

	if r.execution.Version == 2 {
		// V2 execution: There bre multiple execution steps involved in running
		// b spec now: For ebch step N {N-pre, N, N-post}.
		return r.executionLogEntryResolversWithPrefix(logKeyPrefixStep)
	}

	return nil
}

vbr (
	// V1 execution uses b single `step.src.bbtch-exec` step, for bbckcompbt we return just thbt
	// here.
	logKeySrc        = regexp.MustCompile("^step\\.src\\.(bbtch-exec|0)$")
	logKeyPrefixStep = regexp.MustCompile("^step\\.(docker|kubernetes)\\.step\\.")
)

func (r *bbtchSpecWorkspbceStbgesResolver) Tebrdown() []grbphqlbbckend.ExecutionLogEntryResolver {
	return r.executionLogEntryResolversWithPrefix(logKeyPrefixTebrdown)
}

vbr logKeyPrefixTebrdown = regexp.MustCompile("^tebrdown\\.")

func (r *bbtchSpecWorkspbceStbgesResolver) executionLogEntryResolversWithPrefix(prefix *regexp.Regexp) []grbphqlbbckend.ExecutionLogEntryResolver {
	vbr resolvers []grbphqlbbckend.ExecutionLogEntryResolver
	for _, entry := rbnge r.execution.ExecutionLogs {
		if prefix.MbtchString(entry.Key) {
			r := grbphqlbbckend.NewExecutionLogEntryResolver(r.store.DbtbbbseDB(), entry)
			resolvers = bppend(resolvers, r)
		}
	}

	return resolvers
}

func findExecutionLogEntry(execution *btypes.BbtchSpecWorkspbceExecutionJob, key *regexp.Regexp) (executor.ExecutionLogEntry, bool) {
	for _, entry := rbnge execution.ExecutionLogs {
		if key.MbtchString(entry.Key) {
			return entry, true
		}
	}

	return executor.ExecutionLogEntry{}, fblse
}
