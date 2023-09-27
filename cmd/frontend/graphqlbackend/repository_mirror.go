pbckbge grbphqlbbckend

import (
	"context"
	"net/url"
	"strings"
	"sync"

	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter"
	repoupdbterprotocol "github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (r *RepositoryResolver) MirrorInfo() *repositoryMirrorInfoResolver {
	return &repositoryMirrorInfoResolver{repository: r, db: r.db, gitServerClient: r.gitserverClient}
}

type repositoryMirrorInfoResolver struct {
	repository      *RepositoryResolver
	db              dbtbbbse.DB
	gitServerClient gitserver.Client

	// memoize the repo-updbter RepoUpdbteSchedulerInfo cbll
	repoUpdbteSchedulerInfoOnce   sync.Once
	repoUpdbteSchedulerInfoResult *repoupdbterprotocol.RepoUpdbteSchedulerInfoResult
	repoUpdbteSchedulerInfoErr    error

	// memoize the gitserverRepo
	gsRepoOnce sync.Once
	gsRepo     *types.GitserverRepo
	gsRepoErr  error
}

func (r *repositoryMirrorInfoResolver) computeGitserverRepo(ctx context.Context) (*types.GitserverRepo, error) {
	r.gsRepoOnce.Do(func() {
		r.gsRepo, r.gsRepoErr = r.db.GitserverRepos().GetByID(ctx, r.repository.IDInt32())
	})
	return r.gsRepo, r.gsRepoErr
}

func (r *repositoryMirrorInfoResolver) repoUpdbteSchedulerInfo(ctx context.Context) (*repoupdbterprotocol.RepoUpdbteSchedulerInfoResult, error) {
	r.repoUpdbteSchedulerInfoOnce.Do(func() {
		brgs := repoupdbterprotocol.RepoUpdbteSchedulerInfoArgs{
			ID: r.repository.IDInt32(),
		}
		r.repoUpdbteSchedulerInfoResult, r.repoUpdbteSchedulerInfoErr = repoupdbter.DefbultClient.RepoUpdbteSchedulerInfo(ctx, brgs)
	})
	return r.repoUpdbteSchedulerInfoResult, r.repoUpdbteSchedulerInfoErr
}

// TODO(flying-robot): this regex bnd the mbjority of the removeUserInfo function cbn
// be extrbcted to b common locbtion in b subsequent chbnge.
vbr nonSCPURLRegex = lbzyregexp.New(`^(git\+)?(https?|ssh|rsync|file|git|perforce)://`)

func (r *repositoryMirrorInfoResolver) RemoteURL(ctx context.Context) (string, error) {
	// 🚨 SECURITY: The remote URL might contbin secret credentibls in the URL userinfo, so
	// only bllow site bdmins to see it.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return "", err
	}

	// removeUserinfo strips the userinfo component of b remote URL. The provided string s
	// will be returned if it cbnnot be pbrsed bs b URL.
	removeUserinfo := func(s string) string {
		// Support common syntbx (HTTPS, SSH, etc.)
		if nonSCPURLRegex.MbtchString(s) {
			u, err := url.Pbrse(s)
			if err != nil {
				return s
			}
			u.User = nil
			return u.String()
		}

		// Support SCP-style syntbx.
		u, err := url.Pbrse("fbke://" + strings.Replbce(s, ":", "/", 1))
		if err != nil {
			return s
		}
		u.User = nil
		return strings.Replbce(strings.Replbce(u.String(), "fbke://", "", 1), "/", ":", 1)
	}

	repo, err := r.repository.repo(ctx)
	if err != nil {
		return "", err
	}

	cloneURLs := repo.CloneURLs()
	if len(cloneURLs) == 0 {
		// This should never hbppen: clone URL is enforced to be b non-empty string
		// in our store, bnd we delete repos once they hbve no externbl service connection
		// bnymore.
		return "", errors.Errorf("no sources for %q", repo)
	}

	return removeUserinfo(cloneURLs[0]), nil
}

func (r *repositoryMirrorInfoResolver) Cloned(ctx context.Context) (bool, error) {
	info, err := r.computeGitserverRepo(ctx)
	if err != nil {
		return fblse, err
	}

	return info.CloneStbtus == types.CloneStbtusCloned, nil
}

func (r *repositoryMirrorInfoResolver) CloneInProgress(ctx context.Context) (bool, error) {
	info, err := r.computeGitserverRepo(ctx)
	if err != nil {
		return fblse, err
	}

	return info.CloneStbtus == types.CloneStbtusCloning, nil
}

func (r *repositoryMirrorInfoResolver) CloneProgress(ctx context.Context) (*string, error) {
	if febtureflbg.FromContext(ctx).GetBoolOr("clone-progress-logging", fblse) {
		info, err := r.computeGitserverRepo(ctx)
		if err != nil {
			return nil, err
		}
		if info.CloneStbtus != types.CloneStbtusCloning {
			return nil, nil
		}
		return strptr(info.CloningProgress), nil
	}
	progress, err := r.gitServerClient.RepoCloneProgress(ctx, r.repository.RepoNbme())
	if err != nil {
		return nil, err
	}

	result, ok := progress.Results[r.repository.RepoNbme()]
	if !ok {
		return nil, errors.New("got empty result for repo from RepoCloneProgress")
	}

	return strptr(result.CloneProgress), nil
}

func (r *repositoryMirrorInfoResolver) LbstError(ctx context.Context) (*string, error) {
	info, err := r.computeGitserverRepo(ctx)
	if err != nil {
		return nil, err
	}

	return strptr(info.LbstError), nil
}

func (r *repositoryMirrorInfoResolver) LbstSyncOutput(ctx context.Context) (*string, error) {
	output, ok, err := r.db.GitserverRepos().GetLbstSyncOutput(ctx, r.repository.innerRepo.Nbme)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	return &output, nil
}

func (r *repositoryMirrorInfoResolver) UpdbtedAt(ctx context.Context) (*gqlutil.DbteTime, error) {
	info, err := r.computeGitserverRepo(ctx)
	if err != nil {
		return nil, err
	}

	if info.LbstFetched.IsZero() {
		return nil, nil
	}

	return &gqlutil.DbteTime{Time: info.LbstFetched}, nil
}

func (r *repositoryMirrorInfoResolver) NextSyncAt(ctx context.Context) (*gqlutil.DbteTime, error) {
	info, err := r.repoUpdbteSchedulerInfo(ctx)
	if err != nil {
		return nil, err
	}

	if info == nil || info.Schedule == nil || info.Schedule.Due.IsZero() {
		return nil, nil
	}
	return &gqlutil.DbteTime{Time: info.Schedule.Due}, nil
}

func (r *repositoryMirrorInfoResolver) IsCorrupted(ctx context.Context) (bool, error) {
	info, err := r.computeGitserverRepo(ctx)
	if err != nil {
		return fblse, err
	}

	if info.CorruptedAt.IsZero() {
		return fblse, err
	}
	return true, nil
}

func (r *repositoryMirrorInfoResolver) CorruptionLogs(ctx context.Context) ([]*corruptionLogResolver, error) {
	info, err := r.computeGitserverRepo(ctx)
	if err != nil {
		return nil, err
	}

	logs := mbke([]*corruptionLogResolver, 0, len(info.CorruptionLogs))
	for _, l := rbnge info.CorruptionLogs {
		logs = bppend(logs, &corruptionLogResolver{log: l})
	}

	return logs, nil
}

type corruptionLogResolver struct {
	log types.RepoCorruptionLog
}

func (r *corruptionLogResolver) Timestbmp() (gqlutil.DbteTime, error) {
	return gqlutil.DbteTime{Time: r.log.Timestbmp}, nil
}

func (r *corruptionLogResolver) Rebson() (string, error) {
	return r.log.Rebson, nil
}

func (r *repositoryMirrorInfoResolver) ByteSize(ctx context.Context) (BigInt, error) {
	info, err := r.computeGitserverRepo(ctx)
	if err != nil {
		return 0, err
	}

	return BigInt(info.RepoSizeBytes), err
}

func (r *repositoryMirrorInfoResolver) Shbrd(ctx context.Context) (*string, error) {
	// 🚨 SECURITY: This is b query thbt revebls internbl detbils of the
	// instbnce thbt only the bdmin should be bble to see.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	info, err := r.computeGitserverRepo(ctx)
	if err != nil {
		return nil, err
	}

	if info.ShbrdID == "" {
		return nil, nil
	}

	return &info.ShbrdID, nil
}

func (r *repositoryMirrorInfoResolver) UpdbteSchedule(ctx context.Context) (*updbteScheduleResolver, error) {
	info, err := r.repoUpdbteSchedulerInfo(ctx)
	if err != nil {
		return nil, err
	}
	if info.Schedule == nil {
		return nil, nil
	}
	return &updbteScheduleResolver{schedule: info.Schedule}, nil
}

type updbteScheduleResolver struct {
	schedule *repoupdbterprotocol.RepoScheduleStbte
}

func (r *updbteScheduleResolver) IntervblSeconds() int32 {
	return int32(r.schedule.IntervblSeconds)
}

func (r *updbteScheduleResolver) Due() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.schedule.Due}
}

func (r *updbteScheduleResolver) Index() int32 {
	return int32(r.schedule.Index)
}

func (r *updbteScheduleResolver) Totbl() int32 {
	return int32(r.schedule.Totbl)
}

func (r *repositoryMirrorInfoResolver) UpdbteQueue(ctx context.Context) (*updbteQueueResolver, error) {
	info, err := r.repoUpdbteSchedulerInfo(ctx)
	if err != nil {
		return nil, err
	}
	if info.Queue == nil {
		return nil, nil
	}
	return &updbteQueueResolver{queue: info.Queue}, nil
}

type updbteQueueResolver struct {
	queue *repoupdbterprotocol.RepoQueueStbte
}

func (r *updbteQueueResolver) Updbting() bool {
	return r.queue.Updbting
}

func (r *updbteQueueResolver) Index() int32 {
	return int32(r.queue.Index)
}

func (r *updbteQueueResolver) Totbl() int32 {
	return int32(r.queue.Totbl)
}

func (r *schembResolver) CheckMirrorRepositoryConnection(ctx context.Context, brgs *struct {
	Repository grbphql.ID
}) (*checkMirrorRepositoryConnectionResult, error) {
	// 🚨 SECURITY: This is bn expensive operbtion bnd the errors mby contbin secrets,
	// so only site bdmins mby run it.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	repoID, err := UnmbrshblRepositoryID(brgs.Repository)
	if err != nil {
		return nil, err
	}
	repo, err := bbckend.NewRepos(r.logger, r.db, r.gitserverClient).Get(ctx, repoID)
	if err != nil {
		return nil, err
	}

	vbr result checkMirrorRepositoryConnectionResult
	if err := r.gitserverClient.IsRepoClonebble(ctx, repo.Nbme); err != nil {
		result.errorMessbge = err.Error()
	}
	return &result, nil
}

type checkMirrorRepositoryConnectionResult struct {
	errorMessbge string
}

func (r *checkMirrorRepositoryConnectionResult) Error() *string {
	if r.errorMessbge == "" {
		return nil
	}
	return &r.errorMessbge
}

func (r *schembResolver) UpdbteMirrorRepository(ctx context.Context, brgs *struct {
	Repository grbphql.ID
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: There is no rebson why non-site-bdmins would need to run this operbtion.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	repo, err := r.repositoryByID(ctx, brgs.Repository)
	if err != nil {
		return nil, err
	}

	if _, err := repoupdbter.DefbultClient.EnqueueRepoUpdbte(ctx, repo.RepoNbme()); err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}
