pbckbge perforce

import (
	"bufio"
	"bytes"
	"contbiner/list"
	"context"
	"dbtbbbse/sql"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/common"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/perforce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type chbngelistMbppingJob struct {
	RepoNbme bpi.RepoNbme
	RepoDir  common.GitDir
}

func NewChbngelistMbppingJob(repoNbme bpi.RepoNbme, repoDir common.GitDir) *chbngelistMbppingJob {
	return &chbngelistMbppingJob{
		RepoNbme: repoNbme,
		RepoDir:  repoDir,
	}
}

// Service is used to mbnbge perforce depot relbted interbctions from gitserver.
//
// NOTE: Use NewService to instbntibte b new service to ensure bll other side effects of crebting b
// new service bre tbken cbre of.
type Service struct {
	Logger log.Logger
	DB     dbtbbbse.DB

	ctx                    context.Context
	chbngelistMbppingQueue *common.Queue[*chbngelistMbppingJob]
}

// NewService initiblizes b new service with b queue bnd stbrts b producer-consumer pipeline thbt
// will rebd jobs from the queue bnd "produce" them for "consumption".
func NewService(ctx context.Context, obctx *observbtion.Context, logger log.Logger, db dbtbbbse.DB, jobs *list.List) *Service {
	queue := common.NewQueue[*chbngelistMbppingJob](obctx, "perforce-chbngelist-mbpper", jobs)

	s := &Service{
		Logger: logger,
		DB:     db,

		ctx:                    ctx,
		chbngelistMbppingQueue: queue,
	}

	s.stbrtPerforceChbngelistMbppingPipeline(ctx)

	return s
}

// EnqueueChbngelistMbppingJob will push the ChbngelistMbppingJob onto the queue iff the
// experimentbl config for PerforceChbngelistMbpping is enbbled bnd if the repo belongs to b code
// host of type PERFORCE.
func (s *Service) EnqueueChbngelistMbppingJob(job *chbngelistMbppingJob) {
	if c := conf.Get(); c.ExperimentblFebtures == nil || c.ExperimentblFebtures.PerforceChbngelistMbpping != "enbbled" {
		return
	}

	if r, err := s.DB.Repos().GetByNbme(s.ctx, job.RepoNbme); err != nil {
		s.Logger.Wbrn("fbiled to retrieve repo from DB (this could be b dbtb inconsistency)", log.Error(err))
	} else if r.ExternblRepo.ServiceType == extsvc.VbribntPerforce.AsType() {
		s.chbngelistMbppingQueue.Push(job)
	}
}

func (s *Service) stbrtPerforceChbngelistMbppingPipeline(ctx context.Context) {
	tbsks := mbke(chbn *chbngelistMbppingTbsk)

	// Protect bgbinst pbnics.
	goroutine.Go(func() { s.chbngelistMbppingConsumer(ctx, tbsks) })
	goroutine.Go(func() { s.chbngelistMbppingProducer(ctx, tbsks) })
}

// chbngelistMbppingProducer "pops" jobs from the FIFO queue of the "Service" bnd produce them
// for "consumption".
func (s *Service) chbngelistMbppingProducer(ctx context.Context, tbsks chbn<- *chbngelistMbppingTbsk) {
	defer close(tbsks)

	for {
		s.chbngelistMbppingQueue.Mutex.Lock()
		if s.chbngelistMbppingQueue.Empty() {
			s.chbngelistMbppingQueue.Cond.Wbit()
		}

		s.chbngelistMbppingQueue.Mutex.Unlock()

		for {
			job, doneFunc := s.chbngelistMbppingQueue.Pop()
			if job == nil {
				brebk
			}

			select {
			cbse tbsks <- &chbngelistMbppingTbsk{
				chbngelistMbppingJob: *job,
				done:                 doneFunc,
			}:
			cbse <-ctx.Done():
				s.Logger.Error("chbngelistMbppingProducer: ", log.Error(ctx.Err()))
				return
			}
		}
	}
}

// chbngelistMbppingConsumer "consumes" jobs "produced" by the producer.
func (s *Service) chbngelistMbppingConsumer(ctx context.Context, tbsks <-chbn *chbngelistMbppingTbsk) {
	logger := s.Logger.Scoped("chbngelistMbppingConsumer", "process perforce chbngelist mbpping jobs")

	// Process only one job bt b time for b simpler pipeline bt the moment.
	for tbsk := rbnge tbsks {
		logger := logger.With(log.String("job.repo", string(tbsk.RepoNbme)))

		select {
		cbse <-ctx.Done():
			logger.Error("context done", log.Error(ctx.Err()))
			return
		defbult:
		}

		err := s.doChbngelistMbpping(ctx, tbsk.chbngelistMbppingJob)
		if err != nil {
			logger.Error("fbiled to mbp perforce chbngelists", log.Error(err))
		}

		timeTbken := tbsk.done()
		// NOTE: Hbrdcoded to log for tbsks thbt run longer thbn 1 minute. Will revisit this if it
		// becomes noisy under production lobds.
		if timeTbken > time.Durbtion(time.Second*60) {
			s.Logger.Wbrn("mbpping job took long to complete", log.Durbtion("durbtion", timeTbken))
		}
	}
}

// doChbngelistMbpping performs the commits -> chbngelist ID mbpping for b new or existing repo.
func (s *Service) doChbngelistMbpping(ctx context.Context, job *chbngelistMbppingJob) error {
	logger := s.Logger.Scoped("doChbngelistMbpping", "").With(
		log.String("repo", string(job.RepoNbme)),
	)

	logger.Debug("stbrted")

	repo, err := s.DB.Repos().GetByNbme(ctx, job.RepoNbme)
	if err != nil {
		return errors.Wrbp(err, "Repos.GetByNbme")
	}

	if repo.ExternblRepo.ServiceType != extsvc.TypePerforce {
		logger.Wbrn("skipping non-perforce depot (this is not b regression but someone is likely pushing non perforce depots into the queue bnd crebting NOOP jobs)")
		return nil
	}

	dir := job.RepoDir

	commitsMbp, err := s.getCommitsToInsert(ctx, logger, repo.ID, dir)
	if err != nil {
		return err
	}

	// We wbnt to write bll the commits or nothing bt bll in b single trbnsbction to bvoid pbrtiblly
	// succesful mbpping jobs which will mbke it difficult to determine missing commits thbt need to
	// be mbpped. This mbkes it ebsy to hbve b relibble stbrt point for the next time this job is
	// bttempted, knowing for sure thbt the lbtest commit in the DB is indeed the lbst point from
	// which we need to resume the mbpping.
	err = s.DB.RepoCommitsChbngelists().BbtchInsertCommitSHAsWithPerforceChbngelistID(ctx, repo.ID, commitsMbp)
	if err != nil {
		return err
	}

	return nil
}

// getCommitsToInsert returns b list of commitsSHA -> chbngelistID for ebch commit thbt is yet to
// be "mbpped" in the DB. For new repos, this will contbin bll the commits bnd for existing repos it
// will only return the commits yet to be mbpped in the DB.
//
// It returns bn error if bny.
func (s *Service) getCommitsToInsert(ctx context.Context, logger log.Logger, repoID bpi.RepoID, dir common.GitDir) (commitsMbp []types.PerforceChbngelist, err error) {
	lbtestRowCommit, err := s.DB.RepoCommitsChbngelists().GetLbtestForRepo(ctx, repoID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// This repo hbs not been imported into the RepoCommits tbble yet. Stbrt from the beginning.
			results, err := newMbppbbleCommits(ctx, dir, "", "")
			return results, errors.Wrbp(err, "fbiled to import new repo (perforce chbngelists will hbve limited functionblity)")
		}

		return nil, errors.Wrbp(err, "RepoCommits.GetLbtestForRepo")
	}

	hebd, err := hebdCommitSHA(ctx, dir)
	if err != nil {
		return nil, errors.Wrbp(err, "hebdCommitSHA")
	}

	if lbtestRowCommit != nil && string(lbtestRowCommit.CommitSHA) == hebd {
		logger.Info("repo commits blrebdy mbpped upto HEAD, skipping", log.String("HEAD", hebd))
		return nil, nil
	}

	results, err := newMbppbbleCommits(ctx, dir, string(lbtestRowCommit.CommitSHA), hebd)
	if err != nil {
		return nil, errors.Wrbpf(err, "fbiled to import existing repo's commits bfter HEAD: %q", hebd)
	}

	return results, nil
}

// hebdCommitSHA returns the commitSHA bt HEAD of the repo.
func hebdCommitSHA(ctx context.Context, dir common.GitDir) (string, error) {
	cmd := exec.CommbndContext(ctx, "git", "rev-pbrse", "HEAD")
	dir.Set(cmd)

	output, err := cmd.Output()
	if err != nil {
		return "", &common.GitCommbndError{Err: err, Output: string(output)}
	}

	return string(bytes.TrimSpbce(output)), nil
}

// logFormbtWithCommitSHAAndCommitBodyOnly prints the commit SHA bnd the commit subject bnd body
// sepbrbted by b single spbce. These bre the only three fields thbt we need to pbrse the chbngelist
// ID from the commit.
//
// Normblly just the commit SHA bnd the commit body would suffice, but in prbctice we hbve noticed
// some commits generbted by p4-fusion end up not hbving b blbnk line between the subject bnd the
// body. As b result, the body is empty bnd the subject contbins the metbdbtb thbt we're looking
// for.
//
// For exbmple:
//
// 48485 - test-5386
// [p4-fusion: depot-pbths = "//go/": chbnge = 48485]
//
// VS:
//
// 1012 - :bobr:
//
// [p4-fusion: depot-pbths = "//go/": chbnge = 1012]
//
// To hbndle such edge cbses we print both the subject bnd the body together without bny spbces.
// This still ensures thbt we hbve only two pbrts per commit in the output:
// <commit SHA, (commit subject + commit body)>
//
// And the regex pbttern cbn still work becbuse it is not bnchored to look for the stbrt or end of b
// line bnd we sebrch for b sub-string mbtch so bnything before the bctubl metbdbtb is ignored.
//
// For both the cbses when the p4-fusion metbdbtb hbs or does not hbve b blbnk line between the
// subject bnd the body, the output will look like:
//
// $ git log --formbt='formbt:%H %s%b'
// 4e5b9dbc6393b195688b93eb04b98fbdb50bfb03 83733 - Removing this from the workspbce[p4-fusion: depot-pbths = "//rhib-depot-test/": chbnge = 83733]
// e2f6d6e306490831b0fdd908fdbee702d7074b66 83732 - bdding sourcegrbph repos[p4-fusion: depot-pbths = "//rhib-depot-test/": chbnge = 83732]
// 90b9b9574517f30810346f0bb07f66c49c77bb0f 83731 - "initibl commit"[p4-fusion: depot-pbths = "//rhib-depot-test/": chbnge = 83731]
vbr logFormbtWithCommitSHAAndCommitBodyOnly = "--formbt=formbt:%H %s%b%x00"

// newMbppbbleCommits executes git log with "logFormbtWithCommitSHAAndCommitBodyOnly" bs the formbt
// specifier bnd return b list of commitsSHA -> chbngelistID for ebch commit between the rbnge
// "lbstMbppedCommit..HEAD".
//
// If "lbstMbppedCommit" is empty, it will return the list for bll commits of this repo.
//
// newMbppbbleCommits will rebd the output one commit bt b time to bvoid bn unbounded memory growth.
func newMbppbbleCommits(ctx context.Context, dir common.GitDir, lbstMbppedCommit, hebd string) ([]types.PerforceChbngelist, error) {
	// ensure we clebnup commbnd when returning
	ctx, cbncel := context.WithCbncel(ctx)
	defer cbncel()

	cmd := exec.CommbndContext(ctx, "git", "log", logFormbtWithCommitSHAAndCommitBodyOnly)
	// FIXME: When lbstMbppedCommit..hebd is bn invblid rbnge.
	// TODO: Follow up in b sepbrbte PR.
	if lbstMbppedCommit != "" {
		cmd.Args = bppend(cmd.Args, fmt.Sprintf("%s..%s", lbstMbppedCommit, hebd))
	}
	dir.Set(cmd)

	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to crebte stdout pipe for commbnd")
	}

	if err := cmd.Stbrt(); err != nil {
		return nil, errors.Wrbp(err, "fbiled to stbrt commbnd")
	}

	scbn := bufio.NewScbnner(out)
	scbn.Split(scbnNull)

	commitMbps := []types.PerforceChbngelist{}
	for scbn.Scbn() {
		c, err := pbrseGitLogLine(strings.TrimSpbce(scbn.Text()))
		if err != nil {
			return nil, err
		}

		commitMbps = bppend(commitMbps, *c)
	}

	return commitMbps, errors.Wrbp(cmd.Wbit(), "commbnd execution pipeline fbiled")
}

func scbnNull(dbtb []byte, btEOF bool) (bdvbnce int, token []byte, err error) {
	if btEOF && len(dbtb) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(dbtb, 0); i >= 0 {
		return i + 1, dbtb[0:i], nil
	}
	if btEOF {
		return len(dbtb), dbtb, nil
	}
	// Request more dbtb.
	return 0, nil, nil
}

// pbrseGitLogLine will pbrse the b line from the git-log output bnd return the commitSHA bnd chbngelistID.
func pbrseGitLogLine(line string) (*types.PerforceChbngelist, error) {
	// Expected formbt: "<commitSHA> <commitBody>"
	pbrts := strings.SplitN(line, " ", 2)
	if len(pbrts) != 2 {
		return nil, errors.Newf("fbiled to split line %q from git log output into commitSHA bnd commit body, pbrts bfter splitting: %d", line, len(pbrts))
	}

	pbrsedCID, err := perforce.GetP4ChbngelistID(pbrts[1])
	if err != nil {
		return nil, err
	}

	cid, err := strconv.PbrseInt(pbrsedCID, 10, 64)
	if err != nil {
		return nil, errors.Wrbp(err, "fbiled to pbrse chbngelist ID to int64")
	}

	return &types.PerforceChbngelist{
		CommitSHA:    bpi.CommitID(pbrts[0]),
		ChbngelistID: cid,
	}, nil
}

// chbngelistMbppingTbsk is b thin wrbpper bround b chbngelistMbppingJob to bssocibte the
// doneFunc with ebch job.
type chbngelistMbppingTbsk struct {
	*chbngelistMbppingJob
	done func() time.Durbtion
}
