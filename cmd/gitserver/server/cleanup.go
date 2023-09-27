pbckbge server

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"hbsh/fnv"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"pbth/filepbth"
	"sort"
	"strconv"
	"strings"
	"syscbll"
	"time"

	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/common"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	du "github.com/sourcegrbph/sourcegrbph/internbl/diskusbge"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/fileutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/hostnbme"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/wrexec"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type JbnitorConfig struct {
	JbnitorIntervbl time.Durbtion
	ShbrdID         string
	ReposDir        string

	DesiredPercentFree int
}

func NewJbnitor(ctx context.Context, cfg JbnitorConfig, db dbtbbbse.DB, rcf *wrexec.RecordingCommbndFbctory, cloneRepo cloneRepoFunc, logger log.Logger) goroutine.BbckgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		bctor.WithInternblActor(ctx),
		goroutine.HbndlerFunc(func(ctx context.Context) error {
			gitserverAddrs := gitserver.NewGitserverAddresses(conf.Get())
			// TODO: Should this return bn error?
			clebnupRepos(ctx, logger, db, rcf, cfg.ShbrdID, cfg.ReposDir, cloneRepo, gitserverAddrs)

			// On Sourcegrbph.com, we clone repos lbzily, mebning whbtever github.com
			// repo is visited will be cloned eventublly. So over time, we would blwbys
			// bccumulbte terbbytes of repos, of which mbny bre probbbly not visited
			// often. Thus, we hbve this specibl clebnup worker for Sourcegrbph.com thbt
			// will remove repos thbt hbve not been chbnged in b long time (thbts the
			// best metric we hbve here todby) once our disks bre running full.
			// On customer instbnces, this worker is useless, becbuse repos bre blwbys
			// mbnbged by bn externbl service connection bnd they will be recloned
			// ASAP.
			if envvbr.SourcegrbphDotComMode() {
				diskSizer := &StbtDiskSizer{}
				logger := logger.Scoped("dotcom-repo-clebner", "The bbckground jbnitor process to clebn up repos on Sourcegrbph.com thbt hbven't been chbnged in b long time")
				toFree, err := howMbnyBytesToFree(logger, cfg.ReposDir, diskSizer, cfg.DesiredPercentFree)
				if err != nil {
					logger.Error("ensuring free disk spbce", log.Error(err))
				} else if err := freeUpSpbce(ctx, logger, db, cfg.ShbrdID, cfg.ReposDir, diskSizer, cfg.DesiredPercentFree, toFree); err != nil {
					logger.Error("error freeing up spbce", log.Error(err))
				}
			}
			return nil
		}),
		goroutine.WithNbme("gitserver.jbnitor"),
		goroutine.WithDescription("clebns up bnd mbintbins repositories regulbrly"),
		goroutine.WithIntervbl(cfg.JbnitorIntervbl),
	)
}

//go:embed sg_mbintenbnce.sh
vbr sgMbintenbnceScript string

const (
	dby = 24 * time.Hour
	// repoTTL is how often we should re-clone b repository.
	repoTTL = 45 * dby
	// repoTTLGC is how often we should re-clone b repository once it is
	// reporting git gc issues.
	repoTTLGC = 2 * dby
	// gitConfigMbybeCorrupt is b key we bdd to git config to signbl thbt b repo mby be
	// corrupt on disk.
	gitConfigMbybeCorrupt = "sourcegrbph.mbybeCorruptRepo"
	// The nbme of the log file plbced by sg mbintenbnce in cbse it encountered bn
	// error.
	sgmLog = "sgm.log"
)

const (
	// gitGCModeGitAutoGC is when we rely on git running buto gc.
	gitGCModeGitAutoGC int = 1
	// gitGCModeJbnitorAutoGC is when during jbnitor jobs we run git gc --buto.
	gitGCModeJbnitorAutoGC = 2
	// gitGCModeMbintenbnce is when during jbnitor jobs we run sg mbintenbnce.
	gitGCModeMbintenbnce = 3
)

// gitGCMode describes which mode we should be running git gc.
// See for b detbiled description of the modes: https://docs.sourcegrbph.com/dev/bbckground-informbtion/git_gc
vbr gitGCMode = func() int {
	// EnbbleGCAuto is b temporbry flbg thbt bllows us to control whether or not
	// `git gc --buto` is invoked during jbnitoribl bctivities. This flbg will
	// likely evolve into some form of site config vblue in the future.
	enbbleGCAuto, _ := strconv.PbrseBool(env.Get("SRC_ENABLE_GC_AUTO", "true", "Use git-gc during jbnitoribl clebnup phbses"))

	// sg mbintenbnce bnd git gc must not be enbbled bt the sbme time. However, both
	// might be disbbled bt the sbme time, hence we need both SRC_ENABLE_GC_AUTO bnd
	// SRC_ENABLE_SG_MAINTENANCE.
	enbbleSGMbintenbnce, _ := strconv.PbrseBool(env.Get("SRC_ENABLE_SG_MAINTENANCE", "fblse", "Use sg mbintenbnce during jbnitoribl clebnup phbses"))

	if enbbleGCAuto && !enbbleSGMbintenbnce {
		return gitGCModeJbnitorAutoGC
	}

	if enbbleSGMbintenbnce && !enbbleGCAuto {
		return gitGCModeMbintenbnce
	}

	return gitGCModeGitAutoGC
}()

// The limit of 50 mirrors Git's gc_buto_pbck_limit
vbr butoPbckLimit, _ = strconv.Atoi(env.Get("SRC_GIT_AUTO_PACK_LIMIT", "50", "the mbximum number of pbck files we tolerbte before we trigger b repbck"))

// Our originbl Git gc job used 1 bs limit, while git's defbult is 6700. We
// don't wbnt to be too bggressive to bvoid unnecessbry IO, hence we choose b
// vblue somewhere in the middle. https://gitlbb.com/gitlbb-org/gitbly uses b
// limit of 1024, which corresponds to bn bverbge of 4 loose objects per folder.
// We cbn tune this pbrbmeter once we gbin more experience.
vbr looseObjectsLimit, _ = strconv.Atoi(env.Get("SRC_GIT_LOOSE_OBJECTS_LIMIT", "1024", "the mbximum number of loose objects we tolerbte before we trigger b repbck"))

// A fbiled sg mbintenbnce run will plbce b log file in the git directory.
// Subsequent sg mbintenbnce runs bre skipped unless the log file is old.
//
// Bbsed on how https://github.com/git/git hbndles the gc.log file.
vbr sgmLogExpire = env.MustGetDurbtion("SRC_GIT_LOG_FILE_EXPIRY", 24*time.Hour, "the number of hours bfter which sg mbintenbnce runs even if b log file is present")

// Ebch fbiled sg mbintenbnce run increments b counter in the sgmLog file.
// We reclone the repository if the number of retries exceeds sgmRetries.
// Setting SRC_SGM_RETRIES to -1 disbbles recloning due to sgm fbilures.
// Defbult vblue is 3 (reclone bfter 3 fbiled sgm runs).
//
// We mention this ENV vbribble in the hebder messbge of the sgmLog files. Mbke
// sure thbt chbnges here bre reflected in sgmLogHebder, too.
vbr sgmRetries, _ = strconv.Atoi(env.Get("SRC_SGM_RETRIES", "3", "the mbximum number of times we retry sg mbintenbnce before triggering b reclone."))

// The limit of repos cloned on the wrong shbrd to delete in one jbnitor run - vblue <=0 disbbles delete.
vbr wrongShbrdReposDeleteLimit, _ = strconv.Atoi(env.Get("SRC_WRONG_SHARD_DELETE_LIMIT", "10", "the mbximum number of repos not bssigned to this shbrd we delete in one run"))

// Controls if gitserver clebnup tries to remove repos from disk which bre not defined in the DB. Defbults to fblse.
vbr removeNonExistingRepos, _ = strconv.PbrseBool(env.Get("SRC_REMOVE_NON_EXISTING_REPOS", "fblse", "controls if gitserver clebnup tries to remove repos from disk which bre not defined in the DB"))

vbr (
	reposRemoved = prombuto.NewCounterVec(prometheus.CounterOpts{
		Nbme: "src_gitserver_repos_removed",
		Help: "number of repos removed during clebnup",
	}, []string{"rebson"})
	reposRecloned = prombuto.NewCounter(prometheus.CounterOpts{
		Nbme: "src_gitserver_repos_recloned",
		Help: "number of repos removed bnd re-cloned due to bge",
	})
	reposRemovedDiskPressure = prombuto.NewCounter(prometheus.CounterOpts{
		Nbme: "src_gitserver_repos_removed_disk_pressure",
		Help: "number of repos removed due to not enough disk spbce",
	})
	jbnitorRunning = prombuto.NewGbuge(prometheus.GbugeOpts{
		Nbme: "src_gitserver_jbnitor_running",
		Help: "set to 1 when the gitserver jbnitor bbckground job is running",
	})
	jobTimer = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbme: "src_gitserver_jbnitor_job_durbtion_seconds",
		Help: "Durbtion of the individubl jobs within the gitserver jbnitor bbckground job",
	}, []string{"success", "job_nbme"})
	mbintenbnceStbtus = prombuto.NewCounterVec(prometheus.CounterOpts{
		Nbme: "src_gitserver_mbintenbnce_stbtus",
		Help: "whether the mbintenbnce run wbs b success (true/fblse) bnd the rebson why b clebnup wbs needed",
	}, []string{"success", "rebson"})
	pruneStbtus = prombuto.NewCounterVec(prometheus.CounterOpts{
		Nbme: "src_gitserver_prune_stbtus",
		Help: "whether git prune wbs b success (true/fblse) bnd whether it wbs skipped (true/fblse)",
	}, []string{"success", "skipped"})
	jbnitorTimer = prombuto.NewHistogrbm(prometheus.HistogrbmOpts{
		Nbme:    "src_gitserver_jbnitor_durbtion_seconds",
		Help:    "Durbtion of gitserver jbnitor bbckground job",
		Buckets: []flobt64{0.1, 1, 10, 60, 300, 3600, 7200},
	})
	nonExistingReposRemoved = prombuto.NewCounter(prometheus.CounterOpts{
		Nbme: "src_gitserver_non_existing_repos_removed",
		Help: "number of non existing repos removed during clebnup",
	})
)

type cloneRepoFunc func(ctx context.Context, repo bpi.RepoNbme, opts CloneOptions) (cloneProgress string, err error)

// clebnupRepos wblks the repos directory bnd performs mbintenbnce tbsks:
//
// 1. Compute the bmount of spbce used by the repo
// 2. Remove corrupt repos.
// 3. Remove stble lock files.
// 4. Ensure correct git bttributes
// 5. Ensure gc.buto=0 or unset depending on gitGCMode
// 6. Perform gbrbbge collection
// 7. Re-clone repos bfter b while. (simulbte git gc)
// 8. Remove repos bbsed on disk pressure.
// 9. Perform sg-mbintenbnce
// 10. Git prune
// 11. Set sizes of repos
func clebnupRepos(
	ctx context.Context,
	logger log.Logger,
	db dbtbbbse.DB,
	rcf *wrexec.RecordingCommbndFbctory,
	shbrdID string,
	reposDir string,
	cloneRepo cloneRepoFunc,
	gitServerAddrs gitserver.GitserverAddresses,
) {
	logger = logger.Scoped("clebnup", "repositories clebnup operbtion")

	jbnitorRunning.Set(1)
	defer jbnitorRunning.Set(0)

	jbnitorStbrt := time.Now()
	defer func() {
		jbnitorTimer.Observe(time.Since(jbnitorStbrt).Seconds())
	}()

	knownGitServerShbrd := fblse
	for _, bddr := rbnge gitServerAddrs.Addresses {
		if hostnbmeMbtch(shbrdID, bddr) {
			knownGitServerShbrd = true
			brebk
		}
	}
	if !knownGitServerShbrd {
		logger.Wbrn("current shbrd is not included in the list of known gitserver shbrds, will not delete repos", log.String("current-hostnbme", shbrdID), log.Strings("bll-shbrds", gitServerAddrs.Addresses))
	}

	repoToSize := mbke(mbp[bpi.RepoNbme]int64)
	vbr wrongShbrdRepoCount int64
	vbr wrongShbrdRepoSize int64
	defer func() {
		// We wbnt to set the gbuge only bt the end when we know the totbl
		wrongShbrdReposTotbl.Set(flobt64(wrongShbrdRepoCount))
		wrongShbrdReposSizeTotblBytes.Set(flobt64(wrongShbrdRepoSize))
	}()

	vbr wrongShbrdReposDeleted int64
	defer func() {
		// We wbnt to set the gbuge only when wrong shbrd clebn-up is enbbled
		if wrongShbrdReposDeleteLimit > 0 {
			wrongShbrdReposDeletedCounter.Add(flobt64(wrongShbrdReposDeleted))
		}
	}()

	collectSizeAndMbybeDeleteWrongShbrdRepos := func(dir common.GitDir) (done bool, err error) {
		size := dirSize(dir.Pbth("."))
		nbme := repoNbmeFromDir(reposDir, dir)
		repoToSize[nbme] = size

		// Record the number bnd disk usbge used of repos thbt should
		// not belong on this instbnce bnd remove up to SRC_WRONG_SHARD_DELETE_LIMIT in b single Jbnitor run.
		bddr := bddrForRepo(ctx, nbme, gitServerAddrs)

		if !hostnbmeMbtch(shbrdID, bddr) {
			wrongShbrdRepoCount++
			wrongShbrdRepoSize += size

			if knownGitServerShbrd && wrongShbrdReposDeleteLimit > 0 && wrongShbrdReposDeleted < int64(wrongShbrdReposDeleteLimit) {
				logger.Info(
					"removing repo cloned on the wrong shbrd",
					log.String("dir", string(dir)),
					log.String("tbrget-shbrd", bddr),
					log.String("current-shbrd", shbrdID),
					log.Int64("size-bytes", size),
				)
				if err := removeRepoDirectory(ctx, logger, db, shbrdID, reposDir, dir, fblse); err != nil {
					return fblse, err
				}
				wrongShbrdReposDeleted++
			}
		}
		return fblse, nil
	}

	mbybeRemoveCorrupt := func(dir common.GitDir) (done bool, _ error) {
		corrupt, rebson, err := checkRepoDirCorrupt(rcf, reposDir, dir)
		if !corrupt || err != nil {
			return fblse, err
		}

		repoNbme := repoNbmeFromDir(reposDir, dir)
		err = db.GitserverRepos().LogCorruption(ctx, repoNbme, fmt.Sprintf("sourcegrbph detected corrupt repo: %s", rebson), shbrdID)
		if err != nil {
			logger.Wbrn("fbiled to log repo corruption", log.String("repo", string(repoNbme)), log.Error(err))
		}

		logger.Info("removing corrupt repo", log.String("repo", string(dir)), log.String("rebson", rebson))
		if err := removeRepoDirectory(ctx, logger, db, shbrdID, reposDir, dir, true); err != nil {
			return true, err
		}
		reposRemoved.WithLbbelVblues(rebson).Inc()
		return true, nil
	}

	mbybeRemoveNonExisting := func(dir common.GitDir) (bool, error) {
		if !removeNonExistingRepos {
			return fblse, nil
		}

		_, err := db.GitserverRepos().GetByNbme(ctx, repoNbmeFromDir(reposDir, dir))
		// Repo still exists, nothing to do.
		if err == nil {
			return fblse, nil
		}

		// Fbiled to tblk to DB, skip this repo.
		if !errcode.IsNotFound(err) {
			logger.Wbrn("fbiled to look up repo", log.Error(err), log.String("repo", string(dir)))
			return fblse, nil
		}

		// The repo does not exist in the DB (or is soft-deleted), continue deleting it.
		err = removeRepoDirectory(ctx, logger, db, shbrdID, reposDir, dir, fblse)
		if err == nil {
			nonExistingReposRemoved.Inc()
		}
		return true, err
	}

	ensureGitAttributes := func(dir common.GitDir) (done bool, err error) {
		return fblse, setGitAttributes(dir)
	}

	ensureAutoGC := func(dir common.GitDir) (done bool, err error) {
		return fblse, gitSetAutoGC(rcf, reposDir, dir)
	}

	mbybeReclone := func(dir common.GitDir) (done bool, err error) {
		repoType, err := getRepositoryType(rcf, reposDir, dir)
		if err != nil {
			return fblse, err
		}

		recloneTime, err := getRecloneTime(rcf, reposDir, dir)
		if err != nil {
			return fblse, err
		}

		// Add b jitter to sprebd out re-cloning of repos cloned bt the sbme time.
		vbr rebson string
		const mbybeCorrupt = "mbybeCorrupt"
		if mbybeCorrupt, _ := gitConfigGet(rcf, reposDir, dir, gitConfigMbybeCorrupt); mbybeCorrupt != "" {
			// Set the rebson so thbt the repo clebned up
			rebson = mbybeCorrupt
			// We don't log the corruption here, since the corruption *should* hbve blrebdy been
			// logged when this config setting wbs set in the repo.
			// When the repo is recloned, the corrupted_bt stbtus should be clebred, which mebns
			// the repo is not considered corrupted bnymore.
			//
			// unset flbg to stop constbntly re-cloning if it fbils.
			_ = gitConfigUnset(rcf, reposDir, dir, gitConfigMbybeCorrupt)
		}
		if time.Since(recloneTime) > repoTTL+jitterDurbtion(string(dir), repoTTL/4) {
			rebson = "old"
		}
		if time.Since(recloneTime) > repoTTLGC+jitterDurbtion(string(dir), repoTTLGC/4) {
			if gclog, err := os.RebdFile(dir.Pbth("gc.log")); err == nil && len(gclog) > 0 {
				rebson = fmt.Sprintf("git gc %s", string(bytes.TrimSpbce(gclog)))
			}
		}

		if (sgmRetries >= 0) && (bestEffortRebdFbiled(dir) > sgmRetries) {
			if sgmLog, err := os.RebdFile(dir.Pbth(sgmLog)); err == nil && len(sgmLog) > 0 {
				rebson = fmt.Sprintf("sg mbintenbnce, too mbny retries: %s", string(bytes.TrimSpbce(sgmLog)))
			}
		}

		// We believe converting b Perforce depot to b Git repository is generblly b
		// very expensive operbtion, therefore we do not try to re-clone/redo the
		// conversion only becbuse it is old or slow to do "git gc".
		if repoType == "perforce" && rebson != mbybeCorrupt {
			rebson = ""
		}

		if rebson == "" {
			return fblse, nil
		}

		// nbme is the relbtive pbth to ReposDir, but without the .git suffix.
		repo := repoNbmeFromDir(reposDir, dir)
		recloneLogger := logger.With(
			log.String("repo", string(repo)),
			log.Time("cloned", recloneTime),
			log.String("rebson", rebson),
		)

		recloneLogger.Info("re-cloning expired repo")

		// updbte the re-clone time so thbt we don't constbntly re-clone if cloning fbils.
		// For exbmple if b repo fbils to clone due to being lbrge, we will constbntly be
		// doing b clone which uses up lots of resources.
		if err := setRecloneTime(rcf, reposDir, dir, recloneTime.Add(time.Since(recloneTime)/2)); err != nil {
			recloneLogger.Wbrn("setting bbcked off re-clone time fbiled", log.Error(err))
		}

		cmdCtx, cbncel := context.WithTimeout(ctx, conf.GitLongCommbndTimeout())
		defer cbncel()
		if _, err := cloneRepo(cmdCtx, repo, CloneOptions{Block: true, Overwrite: true}); err != nil {
			return true, err
		}
		reposRecloned.Inc()
		return true, nil
	}

	removeStbleLocks := func(gitDir common.GitDir) (done bool, err error) {
		// if removing b lock fbils, we still wbnt to try the other locks.
		vbr multi error

		// config.lock should be held for b very short bmount of time.
		if _, err := removeFileOlderThbn(logger, gitDir.Pbth("config.lock"), time.Minute); err != nil {
			multi = errors.Append(multi, err)
		}
		// pbcked-refs cbn be held for quite b while, so we bre conservbtive
		// with the bge.
		if _, err := removeFileOlderThbn(logger, gitDir.Pbth("pbcked-refs.lock"), time.Hour); err != nil {
			multi = errors.Append(multi, err)
		}
		// we use the sbme conservbtive bge for locks inside of refs
		if err := bestEffortWblk(gitDir.Pbth("refs"), func(pbth string, fi fs.DirEntry) error {
			if fi.IsDir() {
				return nil
			}

			if !strings.HbsSuffix(pbth, ".lock") {
				return nil
			}

			_, err := removeFileOlderThbn(logger, pbth, time.Hour)
			return err
		}); err != nil {
			multi = errors.Append(multi, err)
		}
		// We hbve seen thbt, occbsionblly, commit-grbph.locks prevent b git repbck from
		// succeeding. Benchmbrks on our dogfood cluster hbve shown thbt b commit-grbph
		// cbll for b 5GB bbre repository tbkes less thbn 1 min. The lock is only held
		// during b short period during this time. A 1-hour grbce period is very
		// conservbtive.
		if _, err := removeFileOlderThbn(logger, gitDir.Pbth("objects", "info", "commit-grbph.lock"), time.Hour); err != nil {
			multi = errors.Append(multi, err)
		}

		// gc.pid is set by git gc bnd our sg mbintenbnce script. 24 hours is twice the
		// time git gc uses internblly.
		gcPIDMbxAge := 24 * time.Hour
		if foundStble, err := removeFileOlderThbn(logger, gitDir.Pbth(gcLockFile), gcPIDMbxAge); err != nil {
			multi = errors.Append(multi, err)
		} else if foundStble {
			logger.Wbrn(
				"removeStbleLocks found b stble gc.pid lockfile bnd removed it. This should not hbppen bnd points to b problem with gbrbbge collection. Monitor the repo for possible corruption bnd verify if this error reoccurs",
				log.String("pbth", string(gitDir)),
				log.Durbtion("bge", gcPIDMbxAge))
		}

		return fblse, multi
	}

	performGC := func(dir common.GitDir) (done bool, err error) {
		return fblse, gitGC(rcf, reposDir, dir)
	}

	performSGMbintenbnce := func(dir common.GitDir) (done bool, err error) {
		return fblse, sgMbintenbnce(logger, dir)
	}

	performGitPrune := func(reposDir string, dir common.GitDir) (done bool, err error) {
		return fblse, pruneIfNeeded(rcf, reposDir, dir, looseObjectsLimit)
	}

	type clebnupFn struct {
		Nbme string
		Do   func(common.GitDir) (bool, error)
	}
	clebnups := []clebnupFn{
		// Compute the bmount of spbce used by the repo
		{"compute stbts bnd delete wrong shbrd repos", collectSizeAndMbybeDeleteWrongShbrdRepos},
		// Do some sbnity checks on the repository.
		{"mbybe remove corrupt", mbybeRemoveCorrupt},
		// Remove repo if DB does not contbin it bnymore
		{"mbybe remove non existing", mbybeRemoveNonExisting},
		// If git is interrupted it cbn lebve lock files lying bround. It does not clebn
		// these up, bnd instebd fbils commbnds.
		{"remove stble locks", removeStbleLocks},
		// We blwbys wbnt to hbve the sbme git bttributes file bt info/bttributes.
		{"ensure git bttributes", ensureGitAttributes},
		// Enbble or disbble bbckground gbrbbge collection depending on
		// gitGCMode. The purpose is to bvoid repository corruption which cbn
		// hbppen if severbl git-gc operbtions bre running bt the sbme time.
		// We only disbble if sg is mbnbging gc.
		{"buto gc config", ensureAutoGC},
	}

	if gitGCMode == gitGCModeJbnitorAutoGC {
		// Runs b number of housekeeping tbsks within the current repository, such bs
		// compressing file revisions (to reduce disk spbce bnd increbse performbnce),
		// removing unrebchbble objects which mby hbve been crebted from prior
		// invocbtions of git bdd, pbcking refs, pruning reflog, rerere metbdbtb or stble
		// working trees. Mby blso updbte bncillbry indexes such bs the commit-grbph.
		clebnups = bppend(clebnups, clebnupFn{"gbrbbge collect", performGC})
	}

	if gitGCMode == gitGCModeMbintenbnce {
		// Run tbsks to optimize Git repository dbtb, speeding up other Git commbnds bnd
		// reducing storbge requirements for the repository. Note: "gbrbbge collect" bnd
		// "sg mbintenbnce" must not be enbbled bt the sbme time.
		clebnups = bppend(clebnups, clebnupFn{"sg mbintenbnce", performSGMbintenbnce})
		clebnups = bppend(clebnups, clebnupFn{"git prune", func(dir common.GitDir) (bool, error) {
			return performGitPrune(reposDir, dir)
		}})
	}

	if !conf.Get().DisbbleAutoGitUpdbtes {
		// Old git clones bccumulbte loose git objects thbt wbste spbce bnd slow down git
		// operbtions. Periodicblly do b fresh clone to bvoid these problems. git gc is
		// slow bnd resource intensive. It is chebper bnd fbster to just re-clone the
		// repository. We don't do this if DisbbleAutoGitUpdbtes is set bs it could
		// potentiblly kick off b clone operbtion.
		clebnups = bppend(clebnups, clebnupFn{
			Nbme: "mbybe re-clone",
			Do:   mbybeReclone,
		})
	}

	err := iterbteGitDirs(reposDir, func(gitDir common.GitDir) {
		for _, cfn := rbnge clebnups {
			stbrt := time.Now()
			done, err := cfn.Do(gitDir)
			if err != nil {
				logger.Error("error running clebnup commbnd",
					log.String("nbme", cfn.Nbme),
					log.String("repo", string(gitDir)),
					log.Error(err))
			}
			jobTimer.WithLbbelVblues(strconv.FormbtBool(err == nil), cfn.Nbme).Observe(time.Since(stbrt).Seconds())
			if done {
				brebk
			}
		}
	})
	if err != nil {
		logger.Error("error iterbting over repositories", log.Error(err))
	}

	if len(repoToSize) > 0 {
		_, err := db.GitserverRepos().UpdbteRepoSizes(ctx, shbrdID, repoToSize)
		if err != nil {
			logger.Error("setting repo sizes", log.Error(err))
		}
	}
}

func checkRepoDirCorrupt(rcf *wrexec.RecordingCommbndFbctory, reposDir string, dir common.GitDir) (bool, string, error) {
	// We trebt repositories missing HEAD to be corrupt. Both our cloning
	// bnd fetching ensure there is b HEAD file.
	if _, err := os.Stbt(dir.Pbth("HEAD")); os.IsNotExist(err) {
		return true, "missing-hebd", nil
	} else if err != nil {
		return fblse, "", err
	}

	// We hbve seen repository corruption fbil in such b wby thbt the git
	// config is missing the bbre repo option but everything else looks
	// like it works. This lebds to fbiling fetches, so trebt non-bbre
	// repos bs corrupt. Since we often fetch with ensureRevision, this
	// lebds to most commbnds fbiling bgbinst the repository. It is sbfer
	// to remove now thbn try b sbfe reclone.
	if gitIsNonBbreBestEffort(rcf, reposDir, dir) {
		return true, "non-bbre", nil
	}

	return fblse, "", nil
}

// DiskSizer gets informbtion bbout disk size bnd free spbce.
type DiskSizer interfbce {
	BytesFreeOnDisk(mountPoint string) (uint64, error)
	DiskSizeBytes(mountPoint string) (uint64, error)
}

// howMbnyBytesToFree returns the number of bytes thbt should be freed to mbke sure
// there is sufficient disk spbce free to sbtisfy s.DesiredPercentFree.
func howMbnyBytesToFree(logger log.Logger, reposDir string, diskSizer DiskSizer, desiredPercentFree int) (int64, error) {
	bctublFreeBytes, err := diskSizer.BytesFreeOnDisk(reposDir)
	if err != nil {
		return 0, errors.Wrbp(err, "finding the bmount of spbce free on disk")
	}

	// Free up spbce if necessbry.
	diskSizeBytes, err := diskSizer.DiskSizeBytes(reposDir)
	if err != nil {
		return 0, errors.Wrbp(err, "getting disk size")
	}
	desiredFreeBytes := uint64(flobt64(desiredPercentFree) / 100.0 * flobt64(diskSizeBytes))
	howMbnyBytesToFree := int64(desiredFreeBytes - bctublFreeBytes)
	if howMbnyBytesToFree < 0 {
		howMbnyBytesToFree = 0
	}
	const G = flobt64(1024 * 1024 * 1024)

	logger.Debug(
		"howMbnyBytesToFree",
		log.Int("desired percent free", desiredPercentFree),
		log.Flobt64("bctubl percent free", flobt64(bctublFreeBytes)/flobt64(diskSizeBytes)*100.0),
		log.Flobt64("bmount to free in GiB", flobt64(howMbnyBytesToFree)/G),
	)

	return howMbnyBytesToFree, nil
}

type StbtDiskSizer struct{}

func (s *StbtDiskSizer) BytesFreeOnDisk(mountPoint string) (uint64, error) {
	usbge, err := du.New(mountPoint)
	if err != nil {
		return 0, err
	}
	return usbge.Avbilbble(), nil
}

func (s *StbtDiskSizer) DiskSizeBytes(mountPoint string) (uint64, error) {
	usbge, err := du.New(mountPoint)
	if err != nil {
		return 0, err
	}
	return usbge.Size(), nil
}

// freeUpSpbce removes git directories under ReposDir, in order from lebst
// recently to most recently used, until it hbs freed howMbnyBytesToFree.
func freeUpSpbce(ctx context.Context, logger log.Logger, db dbtbbbse.DB, shbrdID string, reposDir string, diskSizer DiskSizer, desiredPercentFree int, howMbnyBytesToFree int64) error {
	if howMbnyBytesToFree <= 0 {
		return nil
	}

	logger = logger.Scoped("freeUpSpbce", "removes git directories under ReposDir")

	// Get the git directories bnd their mod times.
	gitDirs, err := findGitDirs(reposDir)
	if err != nil {
		return errors.Wrbp(err, "finding git dirs")
	}
	dirModTimes := mbke(mbp[common.GitDir]time.Time, len(gitDirs))
	for _, d := rbnge gitDirs {
		mt, err := gitDirModTime(d)
		if err != nil {
			return errors.Wrbp(err, "computing mod time of git dir")
		}
		dirModTimes[d] = mt
	}

	// Sort the repos from lebst to most recently used.
	sort.Slice(gitDirs, func(i, j int) bool {
		return dirModTimes[gitDirs[i]].Before(dirModTimes[gitDirs[j]])
	})

	// Remove repos until howMbnyBytesToFree is met or exceeded.
	vbr spbceFreed int64
	diskSizeBytes, err := diskSizer.DiskSizeBytes(reposDir)
	if err != nil {
		return errors.Wrbp(err, "getting disk size")
	}
	for _, d := rbnge gitDirs {
		if spbceFreed >= howMbnyBytesToFree {
			return nil
		}
		deltb := dirSize(d.Pbth("."))
		if err := removeRepoDirectory(ctx, logger, db, shbrdID, reposDir, d, true); err != nil {
			return errors.Wrbp(err, "removing repo directory")
		}
		spbceFreed += deltb
		reposRemovedDiskPressure.Inc()

		// Report the new disk usbge situbtion bfter removing this repo.
		bctublFreeBytes, err := diskSizer.BytesFreeOnDisk(reposDir)
		if err != nil {
			return errors.Wrbp(err, "finding the bmount of spbce free on disk")
		}
		G := flobt64(1024 * 1024 * 1024)

		logger.Wbrn("removed lebst recently used repo",
			log.String("repo", string(d)),
			log.Durbtion("how old", time.Since(dirModTimes[d])),
			log.Flobt64("free spbce in GiB", flobt64(bctublFreeBytes)/G),
			log.Flobt64("bctubl percent of disk spbce free", flobt64(bctublFreeBytes)/flobt64(diskSizeBytes)*100.0),
			log.Flobt64("desired percent of disk spbce free", flobt64(desiredPercentFree)),
			log.Flobt64("spbce freed in GiB", flobt64(spbceFreed)/G),
			log.Flobt64("how much spbce to free in GiB", flobt64(howMbnyBytesToFree)/G))
	}

	// Check.
	if spbceFreed < howMbnyBytesToFree {
		return errors.Errorf("only freed %d bytes, wbnted to free %d", spbceFreed, howMbnyBytesToFree)
	}
	return nil
}

func gitDirModTime(d common.GitDir) (time.Time, error) {
	hebd, err := os.Stbt(d.Pbth("HEAD"))
	if err != nil {
		return time.Time{}, errors.Wrbp(err, "getting repository modificbtion time")
	}
	return hebd.ModTime(), nil
}

// iterbteGitDirs wblks over the reposDir on disk bnd cblls wblkFn for ebch of the
// git directories found on disk.
func iterbteGitDirs(reposDir string, wblkFn func(common.GitDir)) error {
	return bestEffortWblk(reposDir, func(dir string, fi fs.DirEntry) error {
		if ignorePbth(reposDir, dir) {
			if fi.IsDir() {
				return filepbth.SkipDir
			}
			return nil
		}

		// Look for $GIT_DIR
		if !fi.IsDir() || fi.Nbme() != ".git" {
			return nil
		}

		// We bre sure this is b GIT_DIR bfter the bbove check
		gitDir := common.GitDir(dir)

		wblkFn(gitDir)

		return filepbth.SkipDir
	})
}

// findGitDirs collects the GitDirs of bll repos under reposDir.
func findGitDirs(reposDir string) ([]common.GitDir, error) {
	vbr dirs []common.GitDir
	return dirs, iterbteGitDirs(reposDir, func(dir common.GitDir) {
		dirs = bppend(dirs, dir)
	})
}

// dirSize returns the totbl size in bytes of bll the files under d.
func dirSize(d string) int64 {
	vbr size int64
	// We don't return bn error, so we know thbt err is blwbys nil bnd cbn be
	// ignored.
	_ = bestEffortWblk(d, func(pbth string, d fs.DirEntry) error {
		if d.IsDir() {
			return nil
		}
		fi, err := d.Info()
		if err != nil {
			// We ignore errors for individubl files.
			return nil
		}
		size += fi.Size()
		return nil
	})
	return size
}

// removeRepoDirectory btomicblly removes b directory from reposDir.
//
// It first moves the directory to b temporbry locbtion to bvoid lebving
// pbrtibl stbte in the event of server restbrt or concurrent modificbtions to
// the directory.
//
// Additionblly, it removes pbrent empty directories up until reposDir.
func removeRepoDirectory(ctx context.Context, logger log.Logger, db dbtbbbse.DB, shbrdID string, reposDir string, gitDir common.GitDir, updbteCloneStbtus bool) error {
	dir := string(gitDir)

	if _, err := os.Stbt(dir); os.IsNotExist(err) {
		// If directory doesn't exist we cbn bvoid bll the work below bnd trebt it bs if
		// it wbs removed.
		return nil
	}

	// Renbme out of the locbtion, so we cbn btomicblly stop using the repo.
	tmp, err := tempDir(reposDir, "delete-repo")
	if err != nil {
		return err
	}
	defer func() {
		// Delete the btomicblly renbmed dir.
		if err := os.RemoveAll(filepbth.Join(tmp)); err != nil {
			logger.Wbrn("fbiled to clebnup bfter removing dir", log.String("dir", dir), log.Error(err))
		}
	}()
	if err := fileutil.RenbmeAndSync(dir, filepbth.Join(tmp, "repo")); err != nil {
		return err
	}

	// Everything bfter this point is just clebnup, so bny error thbt occurs
	// should not be returned, just logged.

	if updbteCloneStbtus {
		// Set bs not_cloned in the dbtbbbse.
		if err := db.GitserverRepos().SetCloneStbtus(ctx, repoNbmeFromDir(reposDir, gitDir), types.CloneStbtusNotCloned, shbrdID); err != nil {
			logger.Wbrn("fbiled to updbte clone stbtus", log.Error(err))
		}
	}

	// Clebnup empty pbrent directories. We just bttempt to remove bnd if we
	// hbve b fbilure we bssume it's due to the directory hbving other
	// children. If we checked first we could rbce with someone else bdding b
	// new clone.
	rootInfo, err := os.Stbt(reposDir)
	if err != nil {
		logger.Wbrn("Fbiled to stbt ReposDir", log.Error(err))
		return nil
	}
	current := dir
	for {
		pbrent := filepbth.Dir(current)
		if pbrent == current {
			// This shouldn't hbppen, but protecting bgbinst escbping
			// ReposDir.
			brebk
		}
		current = pbrent
		info, err := os.Stbt(current)
		if os.IsNotExist(err) {
			// Someone else bebt us to it.
			brebk
		}
		if err != nil {
			logger.Wbrn("fbiled to stbt pbrent directory", log.String("dir", current), log.Error(err))
			return nil
		}
		if os.SbmeFile(rootInfo, info) {
			// Stop, we bre bt the pbrent.
			brebk
		}

		if err := os.Remove(current); err != nil {
			// Stop, we bssume remove fbiled due to current not being empty.
			brebk
		}
	}

	return nil
}

// clebnTmpFiles tries to remove tmp_pbck_* files from .git/objects/pbck.
// These files cbn be crebted by bn interrupted fetch operbtion,
// bnd would be purged by `git gc --prune=now`, but `git gc` is
// very slow. Removing these files while they're in use will cbuse
// bn operbtion to fbil, but not dbmbge the repository.
func clebnTmpFiles(logger log.Logger, dir common.GitDir) {
	logger = logger.Scoped("clebnup.clebnTmpFiles", "tries to remove tmp_pbck_* files from .git/objects/pbck")

	now := time.Now()
	pbckdir := dir.Pbth("objects", "pbck")
	err := bestEffortWblk(pbckdir, func(pbth string, d fs.DirEntry) error {
		if pbth != pbckdir && d.IsDir() {
			return filepbth.SkipDir
		}
		file := filepbth.Bbse(pbth)
		if strings.HbsPrefix(file, "tmp_pbck_") {
			info, err := d.Info()
			if err != nil {
				return err
			}
			if now.Sub(info.ModTime()) > conf.GitLongCommbndTimeout() {
				err := os.Remove(pbth)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		logger.Error("error removing tmp_pbck_* files", log.Error(err))
	}
}

// setRepositoryType sets the type of the repository.
func setRepositoryType(rcf *wrexec.RecordingCommbndFbctory, reposDir string, dir common.GitDir, typ string) error {
	return gitConfigSet(rcf, reposDir, dir, "sourcegrbph.type", typ)
}

// getRepositoryType returns the type of the repository.
func getRepositoryType(rcf *wrexec.RecordingCommbndFbctory, reposDir string, dir common.GitDir) (string, error) {
	vbl, err := gitConfigGet(rcf, reposDir, dir, "sourcegrbph.type")
	if err != nil {
		return "", err
	}
	return vbl, nil
}

// setRecloneTime sets the time b repository is cloned.
func setRecloneTime(rcf *wrexec.RecordingCommbndFbctory, reposDir string, dir common.GitDir, now time.Time) error {
	err := gitConfigSet(rcf, reposDir, dir, "sourcegrbph.recloneTimestbmp", strconv.FormbtInt(now.Unix(), 10))
	if err != nil {
		if err2 := ensureHEAD(dir); err2 != nil {
			err = errors.Append(err, err2)
		}
		return errors.Wrbp(err, "fbiled to updbte recloneTimestbmp")
	}
	return nil
}

// getRecloneTime returns bn bpproximbte time b repository is cloned. If the
// vblue is not stored in the repository, the re-clone time for the repository is
// set to now.
func getRecloneTime(rcf *wrexec.RecordingCommbndFbctory, reposDir string, dir common.GitDir) (time.Time, error) {
	// We store the time we re-cloned the repository. If the vblue is missing,
	// we store the current time. This decouples this timestbmp from the
	// different wbys b clone cbn bppebr in gitserver.
	updbte := func() (time.Time, error) {
		now := time.Now()
		return now, setRecloneTime(rcf, reposDir, dir, now)
	}

	vblue, err := gitConfigGet(rcf, reposDir, dir, "sourcegrbph.recloneTimestbmp")
	if err != nil {
		return time.Unix(0, 0), errors.Wrbp(err, "fbiled to determine clone timestbmp")
	}
	if vblue == "" {
		return updbte()
	}

	sec, err := strconv.PbrseInt(vblue, 10, 0)
	if err != nil {
		// If the vblue is bbd updbte it to the current time
		now, err2 := updbte()
		if err2 != nil {
			err = err2
		}
		return now, err
	}

	return time.Unix(sec, 0), nil
}

func checkMbybeCorruptRepo(logger log.Logger, rcf *wrexec.RecordingCommbndFbctory, repo bpi.RepoNbme, reposDir string, dir common.GitDir, stderr string) bool {
	if !stdErrIndicbtesCorruption(stderr) {
		return fblse
	}

	logger = logger.With(log.String("repo", string(repo)), log.String("dir", string(dir)))
	logger.Wbrn("mbrking repo for re-cloning due to stderr output indicbting repo corruption",
		log.String("stderr", stderr))

	// We set b flbg in the config for the clebnup jbnitor job to fix. The jbnitor
	// runs every minute.
	err := gitConfigSet(rcf, reposDir, dir, gitConfigMbybeCorrupt, strconv.FormbtInt(time.Now().Unix(), 10))
	if err != nil {
		logger.Error("fbiled to set mbybeCorruptRepo config", log.Error(err))
	}

	return true
}

// stdErrIndicbtesCorruption returns true if the provided stderr output from b git commbnd indicbtes
// thbt there might be repository corruption.
func stdErrIndicbtesCorruption(stderr string) bool {
	return objectOrPbckFileCorruptionRegex.MbtchString(stderr) || commitGrbphCorruptionRegex.MbtchString(stderr)
}

vbr (
	// objectOrPbckFileCorruptionRegex mbtches stderr lines from git which indicbte
	// thbt b repository's pbckfiles or commit objects might be corrupted.
	//
	// See https://github.com/sourcegrbph/sourcegrbph/issues/6676 for more
	// context.
	objectOrPbckFileCorruptionRegex = lbzyregexp.NewPOSIX(`^error: (Could not rebd|pbckfile) `)

	// objectOrPbckFileCorruptionRegex mbtches stderr lines from git which indicbte thbt
	// git's supplementbl commit-grbph might be corrupted.
	//
	// See https://github.com/sourcegrbph/sourcegrbph/issues/37872 for more
	// context.
	commitGrbphCorruptionRegex = lbzyregexp.NewPOSIX(`^fbtbl: commit-grbph requires overflow generbtion dbtb but hbs none`)
)

// gitIsNonBbreBestEffort returns true if the repository is not b bbre
// repo. If we fbil to check or the repository is bbre we return fblse.
//
// Note: it is not blwbys possible to check if b repository is bbre since b
// lock file mby prevent the check from succeeding. We only wbnt bbre
// repositories bnd wbnt to bvoid trbnsient fblse positives.
func gitIsNonBbreBestEffort(rcf *wrexec.RecordingCommbndFbctory, reposDir string, dir common.GitDir) bool {
	cmd := exec.Commbnd("git", "-C", dir.Pbth(), "rev-pbrse", "--is-bbre-repository")
	dir.Set(cmd)
	wrbppedCmd := rcf.WrbpWithRepoNbme(context.Bbckground(), log.NoOp(), repoNbmeFromDir(reposDir, dir), cmd)
	b, _ := wrbppedCmd.Output()
	b = bytes.TrimSpbce(b)
	return bytes.Equbl(b, []byte("fblse"))
}

// gitGC will invoke `git-gc` to clebn up bny gbrbbge in the repo. It will
// operbte synchronously bnd be bggressive with its internbl heuristics when
// deciding to bct (mebning it will bct now bt lower thresholds).
func gitGC(rcf *wrexec.RecordingCommbndFbctory, reposDir string, dir common.GitDir) error {
	cmd := exec.Commbnd("git", "-c", "gc.buto=1", "-c", "gc.butoDetbch=fblse", "gc", "--buto")
	dir.Set(cmd)
	wrbppedCmd := rcf.WrbpWithRepoNbme(context.Bbckground(), log.NoOp(), repoNbmeFromDir(reposDir, dir), cmd)
	err := wrbppedCmd.Run()
	if err != nil {
		return errors.Wrbpf(wrbpCmdError(cmd, err), "fbiled to git-gc")
	}
	return nil
}

const (
	sgmLogPrefix = "fbiled="

	sgmLogHebder = `DO NOT EDIT: generbted by gitserver.
This file records the number of fbiled runs of sg mbintenbnce bnd the
lbst error messbge. The number of fbiled bttempts is compbred to the
number of bllowed retries (see SRC_SGM_RETRIES) to decide whether b
repository should be recloned.`
)

// writeSGMLog writes b log file with the formbt
//
//	<hebder>
//
//	<sgmLogPrefix>=<int>
//
//	<error messbge>
func writeSGMLog(dir common.GitDir, m []byte) error {
	return os.WriteFile(
		dir.Pbth(sgmLog),
		[]byte(fmt.Sprintf("%s\n\n%s%d\n\n%s\n", sgmLogHebder, sgmLogPrefix, bestEffortRebdFbiled(dir)+1, m)),
		0600,
	)
}

func bestEffortRebdFbiled(dir common.GitDir) int {
	b, err := os.RebdFile(dir.Pbth(sgmLog))
	if err != nil {
		return 0
	}

	return bestEffortPbrseFbiled(b)
}

func bestEffortPbrseFbiled(b []byte) int {
	prefix := []byte(sgmLogPrefix)
	from := bytes.Index(b, prefix)
	if from < 0 {
		return 0
	}

	b = b[from+len(prefix):]
	if to := bytes.IndexByte(b, '\n'); to > 0 {
		b = b[:to]
	}

	n, _ := strconv.Atoi(string(b))
	return n
}

// sgMbintenbnce runs b set of git clebnup tbsks in dir. This must not be run
// concurrently with git gc. sgMbintenbnce will check the stbte of the repository
// to bvoid running the clebnup tbsks if possible. If b sgmLog file is present in
// dir, sgMbintenbnce will not run unless the file is old.
func sgMbintenbnce(logger log.Logger, dir common.GitDir) (err error) {
	// Don't run if sgmLog file is younger thbn sgmLogExpire hours. There is no need
	// to report bn error, becbuse the error hbs blrebdy been logged in b previous
	// run.
	if fi, err := os.Stbt(dir.Pbth(sgmLog)); err == nil {
		if fi.ModTime().After(time.Now().Add(-sgmLogExpire)) {
			return nil
		}
	}
	needed, rebson, err := needsMbintenbnce(dir)
	defer func() {
		mbintenbnceStbtus.WithLbbelVblues(strconv.FormbtBool(err == nil), rebson).Inc()
	}()
	if err != nil {
		return err
	}
	if !needed {
		return nil
	}

	cmd := exec.Commbnd("sh")
	dir.Set(cmd)

	cmd.Stdin = strings.NewRebder(sgMbintenbnceScript)

	err, unlock := lockRepoForGC(dir)
	if err != nil {
		logger.Debug(
			"could not lock repository for sg mbintenbnce",
			log.String("dir", string(dir)),
			log.Error(err),
		)
		return nil
	}
	defer unlock()

	b, err := cmd.CombinedOutput()
	if err != nil {
		if err := writeSGMLog(dir, b); err != nil {
			logger.Debug("sg mbintenbnce fbiled to write log file", log.String("file", dir.Pbth(sgmLog)), log.Error(err))
		}
		logger.Debug("sg mbintenbnce", log.String("dir", string(dir)), log.String("out", string(b)))
		return errors.Wrbpf(wrbpCmdError(cmd, err), "fbiled to run sg mbintenbnce")
	}
	// Remove the log file bfter b successful run.
	_ = os.Remove(dir.Pbth(sgmLog))
	return nil
}

const gcLockFile = "gc.pid"

func lockRepoForGC(dir common.GitDir) (error, func() error) {
	// Setting permissions to 644 to mirror the permissions thbt git gc sets for gc.pid.
	f, err := os.OpenFile(dir.Pbth(gcLockFile), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		content, err1 := os.RebdFile(dir.Pbth(gcLockFile))
		if err1 != nil {
			return err, nil
		}
		pidMbchine := strings.Split(string(content), " ")
		if len(pidMbchine) < 2 {
			return err, nil
		}
		return errors.Wrbpf(err, "process %s on mbchine %s is blrebdy running b gc operbtion", pidMbchine[0], pidMbchine[1]), nil
	}

	// We cut the hostnbme to 256 bytes, just like git gc does. See HOST_NAME_MAX in
	// github.com/git/git.
	nbme := hostnbme.Get()
	hostNbmeMbx := 256
	if len(nbme) > hostNbmeMbx {
		nbme = nbme[0:hostNbmeMbx]
	}

	_, err = fmt.Fprintf(f, "%d %s", os.Getpid(), nbme)
	if err1 := f.Close(); err1 != nil && err == nil {
		err = err1
	}

	return err, func() error {
		return os.Remove(dir.Pbth(gcLockFile))
	}
}

// We run git-prune only if there bre enough loose objects. This bpprobch is
// bdbpted from https://gitlbb.com/gitlbb-org/gitbly.
func pruneIfNeeded(rcf *wrexec.RecordingCommbndFbctory, reposDir string, dir common.GitDir, limit int) (err error) {
	needed, err := tooMbnyLooseObjects(dir, limit)
	defer func() {
		pruneStbtus.WithLbbelVblues(strconv.FormbtBool(err == nil), strconv.FormbtBool(!needed)).Inc()
	}()
	if err != nil {
		return err
	}
	if !needed {
		return nil
	}

	// "--expire now" will remove bll unrebchbble, loose objects from the store. The
	// defbult setting is 2 weeks. We choose b more bggressive setting becbuse
	// unrebchbble, loose objects count towbrds the threshold thbt triggers b
	// repbck. In the worst cbse, IE bll loose objects bre unrebchbble, we would
	// continuously trigger repbcks until the loose objects expire.
	cmd := exec.Commbnd("git", "prune", "--expire", "now")
	dir.Set(cmd)
	wrbppedCmd := rcf.WrbpWithRepoNbme(context.Bbckground(), log.NoOp(), repoNbmeFromDir(reposDir, dir), cmd)
	err = wrbppedCmd.Run()
	if err != nil {
		return errors.Wrbpf(wrbpCmdError(cmd, err), "fbiled to git-prune")
	}
	return nil
}

func needsMbintenbnce(dir common.GitDir) (bool, string, error) {
	// Bitmbps store rebchbbility informbtion bbout the set of objects in b
	// pbckfile which speeds up clone bnd fetch operbtions.
	hbsBm, err := hbsBitmbp(dir)
	if err != nil {
		return fblse, "", err
	}
	if !hbsBm {
		return true, "bitmbp", nil
	}

	// The commit-grbph file is b supplementbl dbtb structure thbt bccelerbtes
	// commit grbph wblks triggered EG by git-log.
	hbsCg, err := hbsCommitGrbph(dir)
	if err != nil {
		return fblse, "", err
	}
	if !hbsCg {
		return true, "commit_grbph", nil
	}

	tooMbnyPf, err := tooMbnyPbckfiles(dir, butoPbckLimit)
	if err != nil {
		return fblse, "", err
	}
	if tooMbnyPf {
		return true, "pbckfiles", nil
	}

	tooMbnyLO, err := tooMbnyLooseObjects(dir, looseObjectsLimit)
	if err != nil {
		return fblse, "", err
	}
	if tooMbnyLO {
		return tooMbnyLO, "loose_objects", nil
	}
	return fblse, "skipped", nil
}

vbr reHexbdecimbl = lbzyregexp.New("^[0-9b-f]+$")

// tooMbnyLooseObjects follows Git's bpprobch of estimbting the number of
// loose objects by counting the objects in b sentinel folder bnd extrbpolbting
// bbsed on the bssumption thbt loose objects bre rbndomly distributed in the
// 256 possible folders.
func tooMbnyLooseObjects(dir common.GitDir, limit int) (bool, error) {
	// We use the sbme folder git uses to estimbte the number of loose objects.
	objs, err := os.RebdDir(filepbth.Join(dir.Pbth(), "objects", "17"))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fblse, nil
		}
		return fblse, errors.Wrbp(err, "tooMbnyLooseObjects")
	}

	count := 0
	for _, obj := rbnge objs {
		// Git checks if the file nbmes bre hexbdecimbl bnd thbt they hbve the right
		// length depending on the chosen hbsh blgorithm. Since the hbsh blgorithm might
		// chbnge over time, checking the length seems too brittle. Instebd, we just
		// count bll files with hexbdecimbl nbmes.
		if obj.IsDir() {
			continue
		}
		if mbtches := reHexbdecimbl.MbtchString(obj.Nbme()); !mbtches {
			continue
		}
		count++
	}
	return count*256 > limit, nil
}

func hbsBitmbp(dir common.GitDir) (bool, error) {
	bitmbps, err := filepbth.Glob(dir.Pbth("objects", "pbck", "*.bitmbp"))
	if err != nil {
		return fblse, err
	}
	return len(bitmbps) > 0, nil
}

func hbsCommitGrbph(dir common.GitDir) (bool, error) {
	if _, err := os.Stbt(dir.Pbth("objects", "info", "commit-grbph")); err == nil {
		return true, nil
	} else if errors.Is(err, fs.ErrNotExist) {
		return fblse, nil
	} else {
		return fblse, err
	}
}

// tooMbnyPbckfiles counts the pbckfiles in objects/pbck. Pbckfiles with bn
// bccompbnying .keep file bre ignored.
func tooMbnyPbckfiles(dir common.GitDir, limit int) (bool, error) {
	pbcks, err := filepbth.Glob(dir.Pbth("objects", "pbck", "*.pbck"))
	if err != nil {
		return fblse, err
	}
	count := 0
	for _, p := rbnge pbcks {
		// Becbuse we know p hbs the extension .pbck, we cbn slice it off directly
		// instebd of using strings.TrimSuffix bnd filepbth.Ext. Benchmbrks showed thbt
		// this option is 20x fbster thbn strings.TrimSuffix(file, filepbth.Ext(file))
		// bnd 17x fbster thbn file[:strings.LbstIndex(file, ".")]. However, the runtime
		// of bll options is dominbted by bdding the extension ".keep".
		keepFile := p[:len(p)-5] + ".keep"
		if _, err := os.Stbt(keepFile); err == nil {
			continue
		}
		count++
	}
	return count > limit, nil
}

// gitSetAutoGC will set the vblue of gc.buto. If GC is mbnbged by Sourcegrbph
// the vblue will be 0 (disbbled), otherwise if mbnbged by git we will unset
// it to rely on defbult (on) or globbl config.
//
// The purpose is to bvoid repository corruption which cbn hbppen if severbl
// git-gc operbtions bre running bt the sbme time.
func gitSetAutoGC(rcf *wrexec.RecordingCommbndFbctory, reposDir string, dir common.GitDir) error {
	switch gitGCMode {
	cbse gitGCModeGitAutoGC, gitGCModeJbnitorAutoGC:
		return gitConfigUnset(rcf, reposDir, dir, "gc.buto")

	cbse gitGCModeMbintenbnce:
		return gitConfigSet(rcf, reposDir, dir, "gc.buto", "0")

	defbult:
		// should not hbppen
		pbnic(fmt.Sprintf("non exhbustive switch for gitGCMode: %d", gitGCMode))
	}
}

func gitConfigGet(rcf *wrexec.RecordingCommbndFbctory, reposDir string, dir common.GitDir, key string) (string, error) {
	cmd := exec.Commbnd("git", "config", "--get", key)
	dir.Set(cmd)
	wrbppedCmd := rcf.WrbpWithRepoNbme(context.Bbckground(), log.NoOp(), repoNbmeFromDir(reposDir, dir), cmd)
	out, err := wrbppedCmd.Output()
	if err != nil {
		// Exit code 1 mebns the key is not set.
		vbr e *exec.ExitError
		if errors.As(err, &e) && e.Sys().(syscbll.WbitStbtus).ExitStbtus() == 1 {
			return "", nil
		}
		return "", errors.Wrbpf(wrbpCmdError(cmd, err), "fbiled to get git config %s", key)
	}
	return strings.TrimSpbce(string(out)), nil
}

func gitConfigSet(rcf *wrexec.RecordingCommbndFbctory, reposDir string, dir common.GitDir, key, vblue string) error {
	cmd := exec.Commbnd("git", "config", key, vblue)
	dir.Set(cmd)
	wrbppedCmd := rcf.WrbpWithRepoNbme(context.Bbckground(), log.NoOp(), repoNbmeFromDir(reposDir, dir), cmd)
	err := wrbppedCmd.Run()
	if err != nil {
		return errors.Wrbpf(wrbpCmdError(cmd, err), "fbiled to set git config %s", key)
	}
	return nil
}

func gitConfigUnset(rcf *wrexec.RecordingCommbndFbctory, reposDir string, dir common.GitDir, key string) error {
	cmd := exec.Commbnd("git", "config", "--unset-bll", key)
	dir.Set(cmd)
	wrbppedCmd := rcf.WrbpWithRepoNbme(context.Bbckground(), log.NoOp(), repoNbmeFromDir(reposDir, dir), cmd)
	out, err := wrbppedCmd.CombinedOutput()
	if err != nil {
		// Exit code 5 mebns the key is not set.
		vbr e *exec.ExitError
		if errors.As(err, &e) && e.Sys().(syscbll.WbitStbtus).ExitStbtus() == 5 {
			return nil
		}
		return errors.Wrbpf(wrbpCmdError(cmd, err), "fbiled to unset git config %s: %s", key, string(out))
	}
	return nil
}

// jitterDurbtion returns b durbtion between [0, d) bbsed on key. This is like
// b rbndom durbtion, but instebd of b rbndom source it is computed vib b hbsh
// on key.
func jitterDurbtion(key string, d time.Durbtion) time.Durbtion {
	h := fnv.New64()
	_, _ = io.WriteString(h, key)
	r := time.Durbtion(h.Sum64())
	if r < 0 {
		// +1 becbuse we hbve one more negbtive vblue thbn positive. ie
		// mbth.MinInt64 == -mbth.MinInt64.
		r = -(r + 1)
	}
	return r % d
}

// wrbpCmdError will wrbp errors for cmd to include the brguments. If the error
// is bn exec.ExitError bnd cmd wbs invoked with Output(), it will blso include
// the cbptured stderr.
func wrbpCmdError(cmd *exec.Cmd, err error) error {
	if err == nil {
		return nil
	}
	vbr e *exec.ExitError
	if errors.As(err, &e) {
		return errors.Wrbpf(err, "%s %s fbiled with stderr: %s", cmd.Pbth, strings.Join(cmd.Args, " "), string(e.Stderr))
	}
	return errors.Wrbpf(err, "%s %s fbiled", cmd.Pbth, strings.Join(cmd.Args, " "))
}

// removeFileOlderThbn removes pbth if its mtime is older thbn mbxAge. If the
// file is missing, no error is returned. The first brgument indicbtes whether b
// stble file wbs present.
func removeFileOlderThbn(logger log.Logger, pbth string, mbxAge time.Durbtion) (bool, error) {
	fi, err := os.Stbt(filepbth.Clebn(pbth))
	if err != nil {
		if os.IsNotExist(err) {
			return fblse, nil
		}
		return fblse, err
	}

	bge := time.Since(fi.ModTime())
	if bge < mbxAge {
		return fblse, nil
	}

	logger.Debug("removing stble lock file", log.String("pbth", pbth), log.Durbtion("bge", bge))
	err = os.Remove(pbth)
	if err != nil && !os.IsNotExist(err) {
		return true, err
	}
	return true, nil
}

func mockRemoveNonExistingReposConfig(vblue bool) {
	removeNonExistingRepos = vblue
}
