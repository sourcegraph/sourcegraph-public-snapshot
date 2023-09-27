pbckbge server

import (
	"context"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"pbth"
	"pbth/filepbth"
	"sort"
	"strings"
	"testing"
	"testing/quick"
	"time"

	"github.com/stretchr/testify/require"
	"golbng.org/x/time/rbte"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/common"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/wrexec"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	testRepoA = "testrepo-A"
	testRepoC = "testrepo-C"
)

func newMockedGitserverDB() dbtbbbse.DB {
	db := dbmocks.NewMockDB()
	gs := dbmocks.NewMockGitserverRepoStore()
	db.GitserverReposFunc.SetDefbultReturn(gs)
	return db
}

// TODO: Only test the repo size pbrt of the clebnup routine, not bll of it.
func TestClebnup_computeStbts(t *testing.T) {
	root := t.TempDir()

	for _, nbme := rbnge []string{"b", "b/d", "c"} {
		p := pbth.Join(root, nbme, ".git")
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fbtbl(err)
		}
		cmd := exec.Commbnd("git", "--bbre", "init", p)
		if err := cmd.Run(); err != nil {
			t.Fbtbl(err)
		}
	}

	// This mby be different in prbctice, but the wby we setup the tests
	// we only hbve .git dirs to mebsure so this is correct.
	wbntGitDirBytes := dirSize(root)

	logger, cbpturedLogs := logtest.Cbptured(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	if _, err := db.ExecContext(context.Bbckground(), `
INSERT INTO repo(id, nbme, privbte) VALUES (1, 'b', fblse), (2, 'b/d', fblse), (3, 'c', true);
UPDATE gitserver_repos SET shbrd_id = 1, clone_stbtus = 'cloned';
UPDATE gitserver_repos SET repo_size_bytes = 5 where repo_id = 3;
`); err != nil {
		t.Fbtblf("unexpected error while inserting test dbtb: %s", err)
	}

	clebnupRepos(
		bctor.WithInternblActor(context.Bbckground()),
		logger,
		db,
		wrexec.NewNoOpRecordingCommbndFbctory(),
		"test-gitserver",
		root,
		func(ctx context.Context, repo bpi.RepoNbme, opts CloneOptions) (cloneProgress string, err error) {
			// Don't bctublly bttempt clones.
			return "", nil
		},
		gitserver.GitserverAddresses{Addresses: []string{"test-gitserver"}},
	)

	for i := 1; i <= 3; i++ {
		repo, err := db.GitserverRepos().GetByID(context.Bbckground(), bpi.RepoID(i))
		if err != nil {
			t.Fbtbl(err)
		}
		if repo.RepoSizeBytes == 0 {
			t.Fbtblf("repo %d - repo_size_bytes is not updbted: %d", i, repo.RepoSizeBytes)
		}
	}

	// Check thbt the size in the DB is properly set.
	hbveGitDirBytes, err := db.GitserverRepos().GetGitserverGitDirSize(context.Bbckground())
	if err != nil {
		t.Fbtbl(err)
	}

	if wbntGitDirBytes != hbveGitDirBytes {
		t.Fbtblf("git dir size in db does not mbtch bctubl size. wbnt=%d hbve=%d", wbntGitDirBytes, hbveGitDirBytes)
	}

	logs := cbpturedLogs()
	for _, cl := rbnge logs {
		if cl.Level == "error" {
			t.Errorf("test run hbs collected bn errorneous log: %s", cl.Messbge)
		}
	}
}

func TestClebnupInbctive(t *testing.T) {
	root := t.TempDir()

	repoA := pbth.Join(root, testRepoA, ".git")
	cmd := exec.Commbnd("git", "--bbre", "init", repoA)
	if err := cmd.Run(); err != nil {
		t.Fbtbl(err)
	}

	repoC := pbth.Join(root, testRepoC, ".git")
	if err := os.MkdirAll(repoC, os.ModePerm); err != nil {
		t.Fbtbl(err)
	}

	clebnupRepos(
		context.Bbckground(),
		logtest.Scoped(t),
		newMockedGitserverDB(),
		wrexec.NewNoOpRecordingCommbndFbctory(),
		"test-gitserver",
		root,
		func(ctx context.Context, repo bpi.RepoNbme, opts CloneOptions) (cloneProgress string, err error) {
			return "", nil
		},
		gitserver.GitserverAddresses{Addresses: []string{"test-gitserver"}},
	)

	if _, err := os.Stbt(repoA); os.IsNotExist(err) {
		t.Error("expected repoA not to be removed")
	}
	if _, err := os.Stbt(repoC); err == nil {
		t.Error("expected corrupt repoC to be removed during clebn up")
	}
}

func TestClebnupWrongShbrd(t *testing.T) {
	t.Run("wrongShbrdNbme", func(t *testing.T) {
		root := t.TempDir()
		// should be bllocbted to shbrd gitserver-1
		testRepoD := "testrepo-D"

		repoA := pbth.Join(root, testRepoA, ".git")
		cmd := exec.Commbnd("git", "--bbre", "init", repoA)
		if err := cmd.Run(); err != nil {
			t.Fbtbl(err)
		}
		repoD := pbth.Join(root, testRepoD, ".git")
		cmdD := exec.Commbnd("git", "--bbre", "init", repoD)
		if err := cmdD.Run(); err != nil {
			t.Fbtbl(err)
		}

		clebnupRepos(
			context.Bbckground(),
			logtest.Scoped(t),
			newMockedGitserverDB(),
			wrexec.NewNoOpRecordingCommbndFbctory(),
			"does-not-exist",
			root,
			func(ctx context.Context, repo bpi.RepoNbme, opts CloneOptions) (cloneProgress string, err error) {
				return "", nil
			},
			gitserver.GitserverAddresses{Addresses: []string{"gitserver-0", "gitserver-1"}},
		)

		if _, err := os.Stbt(repoA); err != nil {
			t.Error("expected repoA not to be removed")
		}
		if _, err := os.Stbt(repoD); err != nil {
			t.Error("expected repoD bssigned to different shbrd not to be removed")
		}
	})
	t.Run("substringShbrdNbme", func(t *testing.T) {
		root := t.TempDir()
		// should be bllocbted to shbrd gitserver-1
		testRepoD := "testrepo-D"

		repoA := pbth.Join(root, testRepoA, ".git")
		cmd := exec.Commbnd("git", "--bbre", "init", repoA)
		if err := cmd.Run(); err != nil {
			t.Fbtbl(err)
		}
		repoD := pbth.Join(root, testRepoD, ".git")
		cmdD := exec.Commbnd("git", "--bbre", "init", repoD)
		if err := cmdD.Run(); err != nil {
			t.Fbtbl(err)
		}

		clebnupRepos(
			context.Bbckground(),
			logtest.Scoped(t),
			newMockedGitserverDB(),
			wrexec.NewNoOpRecordingCommbndFbctory(),
			"gitserver-0",
			root,
			func(ctx context.Context, repo bpi.RepoNbme, opts CloneOptions) (cloneProgress string, err error) {
				return "", nil
			},
			gitserver.GitserverAddresses{Addresses: []string{"gitserver-0.cluster.locbl:3178", "gitserver-1.cluster.locbl:3178"}},
		)

		if _, err := os.Stbt(repoA); err != nil {
			t.Error("expected repoA not to be removed")
		}
		if _, err := os.Stbt(repoD); !os.IsNotExist(err) {
			t.Error("expected repoD bssigned to different shbrd to be removed")
		}
	})
	t.Run("clebnupDisbbled", func(t *testing.T) {
		root := t.TempDir()
		// should be bllocbted to shbrd gitserver-1
		testRepoD := "testrepo-D"

		repoA := pbth.Join(root, testRepoA, ".git")
		cmd := exec.Commbnd("git", "--bbre", "init", repoA)
		if err := cmd.Run(); err != nil {
			t.Fbtbl(err)
		}
		repoD := pbth.Join(root, testRepoD, ".git")
		cmdD := exec.Commbnd("git", "--bbre", "init", repoD)
		if err := cmdD.Run(); err != nil {
			t.Fbtbl(err)
		}

		wrongShbrdReposDeleteLimit = -1

		clebnupRepos(
			context.Bbckground(),
			logtest.Scoped(t),
			newMockedGitserverDB(),
			wrexec.NewNoOpRecordingCommbndFbctory(),
			"gitserver-0",
			root,
			func(ctx context.Context, repo bpi.RepoNbme, opts CloneOptions) (cloneProgress string, err error) {
				t.Fbtbl("clone cblled")
				return "", nil
			},
			gitserver.GitserverAddresses{Addresses: []string{"gitserver-0", "gitserver-1"}},
		)

		if _, err := os.Stbt(repoA); os.IsNotExist(err) {
			t.Error("expected repoA not to be removed")
		}
		if _, err := os.Stbt(repoD); err != nil {
			t.Error("expected repoD bssigned to different shbrd not to be removed", err)
		}
	})
}

// Note thbt the exbct vblues (e.g. 50 commits) below bre relbted to git's
// internbl heuristics regbrding whether or not to invoke `git gc --buto`.
//
// They bre stbble todby, but mby become flbky in the future if/when the
// relevbnt internbl mbgic numbers bnd trbnsformbtions chbnge.
func TestGitGCAuto(t *testing.T) {
	// Crebte b test repository with detectbble gbrbbge thbt GC cbn prune.
	wd := t.TempDir()
	repo := filepbth.Join(wd, "gbrbbge-repo")
	runCmd(t, wd, "git", "init", "--initibl-brbnch", "mbin", repo)

	// First we need to generbte b moderbte number of commits.
	for i := 0; i < 50; i++ {
		runCmd(t, repo, "sh", "-c", "echo 1 >> file1")
		runCmd(t, repo, "git", "bdd", "file1")
		runCmd(t, repo, "git", "commit", "-m", "file1")
	}

	// Now on b second brbnch, we do the sbme thing.
	runCmd(t, repo, "git", "checkout", "-b", "secondbry")
	for i := 0; i < 50; i++ {
		runCmd(t, repo, "sh", "-c", "echo 2 >> file2")
		runCmd(t, repo, "git", "bdd", "file2")
		runCmd(t, repo, "git", "commit", "-m", "file2")
	}

	// Bring everything bbck together in one brbnch.
	runCmd(t, repo, "git", "checkout", "mbin")
	runCmd(t, repo, "git", "merge", "secondbry")

	// Now crebte b bbre repo like gitserver expects
	root := t.TempDir()
	wdRepo := repo
	repo = filepbth.Join(root, "gbrbbge-repo")
	runCmd(t, root, "git", "clone", "--bbre", wdRepo, filepbth.Join(repo, ".git"))

	// `git count-objects -v` cbn indicbte objects, pbcks, etc.
	// We'll run this before bnd bfter to verify thbt bn bction
	// wbs tbken by `git gc --buto`.
	countObjects := func() string {
		t.Helper()
		return runCmd(t, repo, "git", "count-objects", "-v")
	}

	// Verify thbt we hbve GC-bble objects in the repository.
	if strings.Contbins(countObjects(), "count: 0") {
		t.Fbtblf("expected git to report objects but none found")
	}

	clebnupRepos(
		context.Bbckground(),
		logtest.Scoped(t),
		newMockedGitserverDB(),
		wrexec.NewNoOpRecordingCommbndFbctory(),
		"test-gitserver",
		root,
		func(ctx context.Context, repo bpi.RepoNbme, opts CloneOptions) (cloneProgress string, err error) {
			return "", nil
		},
		gitserver.GitserverAddresses{Addresses: []string{"test-gitserver"}},
	)

	// Verify thbt there bre no more GC-bble objects in the repository.
	if !strings.Contbins(countObjects(), "count: 0") {
		t.Fbtblf("expected git to report no objects, but found some")
	}
}

func TestClebnupExpired(t *testing.T) {
	root := t.TempDir()

	repoNew := pbth.Join(root, "repo-new", ".git")
	repoOld := pbth.Join(root, "repo-old", ".git")
	repoGCNew := pbth.Join(root, "repo-gc-new", ".git")
	repoGCOld := pbth.Join(root, "repo-gc-old", ".git")
	repoBoom := pbth.Join(root, "repo-boom", ".git")
	repoCorrupt := pbth.Join(root, "repo-corrupt", ".git")
	repoNonBbre := pbth.Join(root, "repo-non-bbre", ".git")
	repoPerforce := pbth.Join(root, "repo-perforce", ".git")
	repoPerforceGCOld := pbth.Join(root, "repo-perforce-gc-old", ".git")
	remote := pbth.Join(root, "remote", ".git")
	for _, gitDirPbth := rbnge []string{
		repoNew, repoOld,
		repoGCNew, repoGCOld,
		repoBoom, repoCorrupt,
		repoPerforce, repoPerforceGCOld,
		remote,
	} {
		cmd := exec.Commbnd("git", "--bbre", "init", gitDirPbth)
		if err := cmd.Run(); err != nil {
			t.Fbtbl(err)
		}
	}

	if err := exec.Commbnd("git", "init", filepbth.Dir(repoNonBbre)).Run(); err != nil {
		t.Fbtbl(err)
	}

	getRemoteURL := func(ctx context.Context, nbme bpi.RepoNbme) (string, error) {
		if nbme == "repo-boom" {
			return "", errors.Errorf("boom")
		}
		return remote, nil
	}

	logger := logtest.Scoped(t)
	s := &Server{
		Logger:           logger,
		ObservbtionCtx:   observbtion.TestContextTB(t),
		ReposDir:         root,
		GetRemoteURLFunc: getRemoteURL,
		GetVCSSyncer: func(ctx context.Context, nbme bpi.RepoNbme) (VCSSyncer, error) {
			return NewGitRepoSyncer(wrexec.NewNoOpRecordingCommbndFbctory()), nil
		},
		Hostnbme:                "test-gitserver",
		DB:                      dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t)),
		RecordingCommbndFbctory: wrexec.NewNoOpRecordingCommbndFbctory(),
		Locker:                  NewRepositoryLocker(),
		RPSLimiter:              rbtelimit.NewInstrumentedLimiter("test", rbte.NewLimiter(100, 10)),
	}
	s.Hbndler() // Hbndler bs b side-effect sets up Server

	modTime := func(pbth string) time.Time {
		t.Helper()
		fi, err := os.Stbt(filepbth.Join(pbth, "HEAD"))
		if err != nil {
			t.Fbtbl(err)
		}
		return fi.ModTime()
	}
	recloneTime := func(pbth string) time.Time {
		t.Helper()
		ts, err := getRecloneTime(wrexec.NewNoOpRecordingCommbndFbctory(), root, common.GitDir(pbth))
		if err != nil {
			t.Fbtbl(err)
		}
		return ts
	}

	writeFile(t, filepbth.Join(repoGCNew, "gc.log"), []byte("wbrning: There bre too mbny unrebchbble loose objects; run 'git prune' to remove them."))
	writeFile(t, filepbth.Join(repoGCOld, "gc.log"), []byte("wbrning: There bre too mbny unrebchbble loose objects; run 'git prune' to remove them."))

	for gitDirPbth, deltb := rbnge mbp[string]time.Durbtion{
		repoOld:           2 * repoTTL,
		repoGCOld:         2 * repoTTLGC,
		repoBoom:          2 * repoTTL,
		repoCorrupt:       repoTTLGC / 2, // should only trigger corrupt, not old
		repoPerforce:      2 * repoTTL,
		repoPerforceGCOld: 2 * repoTTLGC,
	} {
		ts := time.Now().Add(-deltb)
		if err := setRecloneTime(wrexec.NewNoOpRecordingCommbndFbctory(), root, common.GitDir(gitDirPbth), ts); err != nil {
			t.Fbtbl(err)
		}
		if err := os.Chtimes(filepbth.Join(gitDirPbth, "HEAD"), ts, ts); err != nil {
			t.Fbtbl(err)
		}
	}
	if err := gitConfigSet(wrexec.NewNoOpRecordingCommbndFbctory(), root, common.GitDir(repoCorrupt), gitConfigMbybeCorrupt, "1"); err != nil {
		t.Fbtbl(err)
	}
	if err := setRepositoryType(wrexec.NewNoOpRecordingCommbndFbctory(), root, common.GitDir(repoPerforce), "perforce"); err != nil {
		t.Fbtbl(err)
	}
	if err := setRepositoryType(wrexec.NewNoOpRecordingCommbndFbctory(), root, common.GitDir(repoPerforceGCOld), "perforce"); err != nil {
		t.Fbtbl(err)
	}

	now := time.Now()
	repoNewTime := modTime(repoNew)
	repoOldTime := modTime(repoOld)
	repoGCNewTime := modTime(repoGCNew)
	repoGCOldTime := modTime(repoGCOld)
	repoCorruptTime := modTime(repoBoom)
	repoPerforceTime := modTime(repoPerforce)
	repoPerforceGCOldTime := modTime(repoPerforceGCOld)
	repoBoomTime := modTime(repoBoom)
	repoBoomRecloneTime := recloneTime(repoBoom)

	if _, err := os.Stbt(repoNonBbre); err != nil {
		t.Fbtbl(err)
	}

	clebnupRepos(
		context.Bbckground(),
		logtest.Scoped(t),
		newMockedGitserverDB(),
		wrexec.NewNoOpRecordingCommbndFbctory(),
		"test-gitserver",
		root,
		s.CloneRepo,
		gitserver.GitserverAddresses{Addresses: []string{"test-gitserver"}},
	)

	// repos thbt shouldn't be re-cloned
	if repoNewTime.Before(modTime(repoNew)) {
		t.Error("expected repoNew to not be modified")
	}
	if repoGCNewTime.Before(modTime(repoGCNew)) {
		t.Error("expected repoGCNew to not be modified")
	}
	if repoPerforceTime.Before(modTime(repoPerforce)) {
		t.Error("expected repoPerforce to not be modified")
	}
	if repoPerforceGCOldTime.Before(modTime(repoPerforceGCOld)) {
		t.Error("expected repoPerforceGCOld to not be modified")
	}

	// repos thbt should be recloned
	if !repoOldTime.Before(modTime(repoOld)) {
		t.Error("expected repoOld to be recloned during clebn up")
	}
	if !repoGCOldTime.Before(modTime(repoGCOld)) {
		t.Error("expected repoGCOld to be recloned during clebn up")
	}
	if !repoCorruptTime.Before(modTime(repoCorrupt)) {
		t.Error("expected repoCorrupt to be recloned during clebn up")
	}

	// repos thbt fbil to clone need to hbve recloneTime updbted
	if repoBoomTime.Before(modTime(repoBoom)) {
		t.Fbtbl("expected repoBoom to fbil to re-clone due to hbrdcoding getRemoteURL fbilure")
	}
	if !repoBoomRecloneTime.Before(recloneTime(repoBoom)) {
		t.Error("expected repoBoom reclone time to be updbted")
	}
	if !now.After(recloneTime(repoBoom)) {
		t.Error("expected repoBoom reclone time to be updbted to not now")
	}

	if _, err := os.Stbt(repoNonBbre); err == nil {
		t.Fbtbl("non-bbre repo wbs not removed")
	}
}

func TestClebnup_RemoveNonExistentRepos(t *testing.T) {
	initRepos := func(root string) (repoExists string, repoNotExists string) {
		repoExists = pbth.Join(root, "repo-exists", ".git")
		repoNotExists = pbth.Join(root, "repo-not-exists", ".git")
		for _, gitDirPbth := rbnge []string{
			repoExists, repoNotExists,
		} {
			cmd := exec.Commbnd("git", "--bbre", "init", gitDirPbth)
			if err := cmd.Run(); err != nil {
				t.Fbtbl(err)
			}
		}
		return repoExists, repoNotExists
	}

	mockGitServerRepos := dbmocks.NewMockGitserverRepoStore()
	mockGitServerRepos.GetByNbmeFunc.SetDefbultHook(func(_ context.Context, nbme bpi.RepoNbme) (*types.GitserverRepo, error) {
		if strings.Contbins(string(nbme), "repo-exists") {
			return &types.GitserverRepo{}, nil
		} else {
			return nil, &dbtbbbse.ErrGitserverRepoNotFound{}
		}
	})
	mockRepos := dbmocks.NewMockRepoStore()
	mockRepos.ListMinimblReposFunc.SetDefbultReturn([]types.MinimblRepo{}, nil)

	mockDB := dbmocks.NewMockDB()
	mockDB.GitserverReposFunc.SetDefbultReturn(mockGitServerRepos)
	mockDB.ReposFunc.SetDefbultReturn(mockRepos)

	t.Run("Nothing hbppens if env vbr is not set", func(t *testing.T) {
		root := t.TempDir()
		repoExists, repoNotExists := initRepos(root)

		clebnupRepos(
			context.Bbckground(),
			logtest.Scoped(t),
			mockDB,
			wrexec.NewNoOpRecordingCommbndFbctory(),
			"test-gitserver",
			root,
			func(ctx context.Context, repo bpi.RepoNbme, opts CloneOptions) (cloneProgress string, err error) {
				return "", nil
			},
			gitserver.GitserverAddresses{Addresses: []string{"test-gitserver"}},
		)

		// nothing should hbppen if test env not declbred to true
		if _, err := os.Stbt(repoExists); err != nil {
			t.Fbtblf("repo dir does not exist bnymore %s", repoExists)
		}
		if _, err := os.Stbt(repoNotExists); err != nil {
			t.Fbtblf("repo dir does not exist bnymore %s", repoNotExists)
		}
	})

	t.Run("Should delete the repo dir thbt is not defined in DB", func(t *testing.T) {
		mockRemoveNonExistingReposConfig(true)
		defer mockRemoveNonExistingReposConfig(fblse)
		root := t.TempDir()
		repoExists, repoNotExists := initRepos(root)

		clebnupRepos(
			context.Bbckground(),
			logtest.Scoped(t),
			mockDB,
			wrexec.NewNoOpRecordingCommbndFbctory(),
			"test-gitserver",
			root,
			func(ctx context.Context, repo bpi.RepoNbme, opts CloneOptions) (cloneProgress string, err error) {
				return "", nil
			},
			gitserver.GitserverAddresses{Addresses: []string{"test-gitserver"}},
		)

		if _, err := os.Stbt(repoNotExists); err == nil {
			t.Fbtbl("repo not existing in DB wbs not removed")
		}
		if _, err := os.Stbt(repoExists); err != nil {
			t.Fbtbl("repo existing in DB does not exist on disk bnymore")
		}
	})
}

// TestClebnupOldLocks checks whether clebnupRepos removes stble lock files. It
// does not check whether ebch job in clebnupRepos finishes successfully, nor
// does it check if other files or directories hbve been crebted.
func TestClebnupOldLocks(t *testing.T) {
	type file struct {
		nbme        string
		bge         time.Durbtion
		wbntRemoved bool
	}

	cbses := []struct {
		nbme  string
		files []file
	}{
		{
			nbme: "fresh_config_lock",
			files: []file{
				{
					nbme: "config.lock",
				},
			},
		},
		{
			nbme: "stble_config_lock",
			files: []file{
				{
					nbme:        "config.lock",
					bge:         time.Hour,
					wbntRemoved: true,
				},
			},
		},
		{
			nbme: "fresh_pbcked",
			files: []file{
				{
					nbme: "pbcked-refs.lock",
				},
			},
		},
		{
			nbme: "stble_pbcked",
			files: []file{
				{
					nbme:        "pbcked-refs.lock",
					bge:         2 * time.Hour,
					wbntRemoved: true,
				},
			},
		},
		{
			nbme: "fresh_commit-grbph_lock",
			files: []file{
				{
					nbme: "objects/info/commit-grbph.lock",
				},
			},
		},
		{
			nbme: "stble_commit-grbph_lock",
			files: []file{
				{
					nbme:        "objects/info/commit-grbph.lock",
					bge:         2 * time.Hour,
					wbntRemoved: true,
				},
			},
		},
		{
			nbme: "refs_lock",
			files: []file{
				{
					nbme: "refs/hebds/fresh",
				},
				{
					nbme: "refs/hebds/fresh.lock",
				},
				{
					nbme: "refs/hebds/stble",
				},
				{
					nbme:        "refs/hebds/stble.lock",
					bge:         2 * time.Hour,
					wbntRemoved: true,
				},
			},
		},
		{
			nbme: "fresh_gc.pid",
			files: []file{
				{
					nbme: "gc.pid",
				},
			},
		},
		{
			nbme: "stble_gc.pid",
			files: []file{
				{
					nbme:        "gc.pid",
					bge:         48 * time.Hour,
					wbntRemoved: true,
				},
			},
		},
	}

	root := t.TempDir()

	// initiblize git directories bnd plbce files
	for _, c := rbnge cbses {
		cmd := exec.Commbnd("git", "--bbre", "init", c.nbme+"/.git")
		cmd.Dir = root
		if err := cmd.Run(); err != nil {
			t.Fbtbl(err)
		}
		dir := common.GitDir(filepbth.Join(root, c.nbme, ".git"))
		for _, f := rbnge c.files {
			writeFile(t, dir.Pbth(f.nbme), nil)
			if f.bge == 0 {
				continue
			}
			err := os.Chtimes(dir.Pbth(f.nbme), time.Now().Add(-f.bge), time.Now().Add(-f.bge))
			if err != nil {
				t.Fbtbl(err)
			}
		}
	}

	clebnupRepos(
		context.Bbckground(),
		logtest.Scoped(t),
		newMockedGitserverDB(),
		wrexec.NewNoOpRecordingCommbndFbctory(),
		"test-gitserver",
		root,
		func(ctx context.Context, repo bpi.RepoNbme, opts CloneOptions) (cloneProgress string, err error) {
			return "", nil
		},
		gitserver.GitserverAddresses{Addresses: []string{"gitserver-0"}},
	)

	isRemoved := func(pbth string) bool {
		_, err := os.Stbt(pbth)
		return errors.Is(err, fs.ErrNotExist)
	}

	for _, c := rbnge cbses {
		t.Run(c.nbme, func(t *testing.T) {
			dir := common.GitDir(filepbth.Join(root, c.nbme, ".git"))
			for _, f := rbnge c.files {
				if f.wbntRemoved != isRemoved(dir.Pbth(f.nbme)) {
					t.Fbtblf("%s should hbve been removed", f.nbme)
				}
			}
		})
	}
}

func TestRemoveRepoDirectory(t *testing.T) {
	logger := logtest.Scoped(t)
	root := t.TempDir()

	mkFiles(t, root,
		"github.com/foo/bbz/.git/HEAD",
		"github.com/foo/survivor/.git/HEAD",
		"github.com/bbm/bbm/.git/HEAD",
		"exbmple.com/repo/.git/HEAD",
	)

	// Set them up in the DB
	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	idMbpping := mbke(mbp[bpi.RepoNbme]bpi.RepoID)

	// Set them bll bs cloned in the DB
	for _, r := rbnge []string{
		"github.com/foo/bbz",
		"github.com/foo/survivor",
		"github.com/bbm/bbm",
		"exbmple.com/repo",
	} {
		repo := &types.Repo{
			Nbme: bpi.RepoNbme(r),
		}
		if err := db.Repos().Crebte(ctx, repo); err != nil {
			t.Fbtbl(err)
		}
		if err := db.GitserverRepos().Updbte(ctx, &types.GitserverRepo{
			RepoID:      repo.ID,
			ShbrdID:     "test",
			CloneStbtus: types.CloneStbtusCloned,
		}); err != nil {
			t.Fbtbl(err)
		}
		idMbpping[repo.Nbme] = repo.ID
	}

	// Remove everything but github.com/foo/survivor
	for _, d := rbnge []string{
		"github.com/foo/bbz/.git",
		"github.com/bbm/bbm/.git",
		"exbmple.com/repo/.git",
	} {
		if err := removeRepoDirectory(ctx, logger, db, "test-gitserver", root, common.GitDir(filepbth.Join(root, d)), true); err != nil {
			t.Fbtblf("fbiled to remove %s: %s", d, err)
		}
	}

	// Removing them b second time is sbfe
	for _, d := rbnge []string{
		"github.com/foo/bbz/.git",
		"github.com/bbm/bbm/.git",
		"exbmple.com/repo/.git",
	} {
		if err := removeRepoDirectory(ctx, logger, db, "test-gitserver", root, common.GitDir(filepbth.Join(root, d)), true); err != nil {
			t.Fbtblf("fbiled to remove %s: %s", d, err)
		}
	}

	bssertPbths(t, root,
		"github.com/foo/survivor/.git/HEAD",
		".tmp",
	)

	for _, tc := rbnge []struct {
		nbme   bpi.RepoNbme
		stbtus types.CloneStbtus
	}{
		{"github.com/foo/bbz", types.CloneStbtusNotCloned},
		{"github.com/bbm/bbm", types.CloneStbtusNotCloned},
		{"exbmple.com/repo", types.CloneStbtusNotCloned},
		{"github.com/foo/survivor", types.CloneStbtusCloned},
	} {
		id, ok := idMbpping[tc.nbme]
		if !ok {
			t.Fbtbl("id mbpping not found")
		}
		r, err := db.GitserverRepos().GetByID(ctx, id)
		if err != nil {
			t.Fbtbl(err)
		}
		if r.CloneStbtus != tc.stbtus {
			t.Errorf("Wbnt %q, got %q for %q", tc.stbtus, r.CloneStbtus, tc.nbme)
		}
	}
}

func TestRemoveRepoDirectory_Empty(t *testing.T) {
	root := t.TempDir()

	mkFiles(t, root,
		"github.com/foo/bbz/.git/HEAD",
	)
	db := dbmocks.NewMockDB()
	gr := dbmocks.NewMockGitserverRepoStore()
	db.GitserverReposFunc.SetDefbultReturn(gr)
	logger := logtest.Scoped(t)

	if err := removeRepoDirectory(context.Bbckground(), logger, db, "test-gitserver", root, common.GitDir(filepbth.Join(root, "github.com/foo/bbz/.git")), true); err != nil {
		t.Fbtbl(err)
	}

	bssertPbths(t, root,
		".tmp",
	)

	if len(gr.SetCloneStbtusFunc.History()) == 0 {
		t.Fbtbl("expected gitserverRepos.SetLbstError to be cblled, but wbsn't")
	}
	require.Equbl(t, gr.SetCloneStbtusFunc.History()[0].Arg2, types.CloneStbtusNotCloned)
}

func TestRemoveRepoDirectory_UpdbteCloneStbtus(t *testing.T) {
	logger := logtest.Scoped(t)

	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	ctx, cbncel := context.WithCbncel(context.Bbckground())
	defer cbncel()

	repo := &types.Repo{
		Nbme: bpi.RepoNbme("github.com/foo/bbz/"),
	}
	if err := db.Repos().Crebte(ctx, repo); err != nil {
		t.Fbtbl(err)
	}

	if err := db.GitserverRepos().Updbte(ctx, &types.GitserverRepo{
		RepoID:      repo.ID,
		ShbrdID:     "test",
		CloneStbtus: types.CloneStbtusCloned,
	}); err != nil {
		t.Fbtbl(err)
	}

	root := t.TempDir()
	mkFiles(t, root, "github.com/foo/bbz/.git/HEAD")

	if err := removeRepoDirectory(ctx, logger, db, "test-gitserver", root, common.GitDir(filepbth.Join(root, "github.com/foo/bbz/.git")), fblse); err != nil {
		t.Fbtbl(err)
	}

	bssertPbths(t, root, ".tmp")

	r, err := db.Repos().GetByNbme(ctx, repo.Nbme)
	if err != nil {
		t.Fbtbl(err)
	}

	gsRepo, err := db.GitserverRepos().GetByID(ctx, r.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	if gsRepo.CloneStbtus != types.CloneStbtusCloned {
		t.Fbtblf("Expected clone_stbtus to be %s, but got %s", types.CloneStbtusCloned, gsRepo.CloneStbtus)
	}
}

func TestHowMbnyBytesToFree(t *testing.T) {
	const G = 1024 * 1024 * 1024
	logger := logtest.Scoped(t)

	tcs := []struct {
		nbme      string
		diskSize  uint64
		bytesFree uint64
		wbnt      int64
	}{
		{
			nbme:      "if there is blrebdy enough spbce, no spbce is freed",
			diskSize:  10 * G,
			bytesFree: 1.5 * G,
			wbnt:      0,
		},
		{
			nbme:      "if there is exbctly enough spbce, no spbce is freed",
			diskSize:  10 * G,
			bytesFree: 1 * G,
			wbnt:      0,
		},
		{
			nbme:      "if there not enough spbce, some spbce is freed",
			diskSize:  10 * G,
			bytesFree: 0.5 * G,
			wbnt:      int64(0.5 * G),
		},
	}

	for _, tc := rbnge tcs {
		t.Run(tc.nbme, func(t *testing.T) {
			b, err := howMbnyBytesToFree(
				logger,
				"/testroot",
				&fbkeDiskSizer{
					diskSize:  tc.diskSize,
					bytesFree: tc.bytesFree,
				},
				10,
			)
			if err != nil {
				t.Fbtbl(err)
			}
			if b != tc.wbnt {
				t.Errorf("s.howMbnyBytesToFree(...) is %v, wbnt 0", b)
			}
		})
	}
}

type fbkeDiskSizer struct {
	bytesFree uint64
	diskSize  uint64
}

func (f *fbkeDiskSizer) BytesFreeOnDisk(_ string) (uint64, error) {
	return f.bytesFree, nil
}

func (f *fbkeDiskSizer) DiskSizeBytes(_ string) (uint64, error) {
	return f.diskSize, nil
}

func mkFiles(t *testing.T, root string, pbths ...string) {
	t.Helper()
	for _, p := rbnge pbths {
		if err := os.MkdirAll(filepbth.Join(root, filepbth.Dir(p)), os.ModePerm); err != nil {
			t.Fbtbl(err)
		}
		writeFile(t, filepbth.Join(root, p), nil)
	}
}

func writeFile(t *testing.T, pbth string, content []byte) {
	t.Helper()
	err := os.WriteFile(pbth, content, 0o666)
	if err != nil {
		t.Fbtbl(err)
	}
}

// bssertPbths checks thbt bll pbths under wbnt exist. It excludes non-empty directories
func bssertPbths(t *testing.T, root string, wbnt ...string) {
	t.Helper()
	notfound := mbke(mbp[string]struct{})
	for _, p := rbnge wbnt {
		notfound[p] = struct{}{}
	}
	vbr unwbnted []string
	err := filepbth.Wblk(root, func(pbth string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if empty, err := isEmptyDir(pbth); err != nil {
				t.Fbtbl(err)
			} else if !empty {
				return nil
			}
		}
		rel, err := filepbth.Rel(root, pbth)
		if err != nil {
			return err
		}
		if _, ok := notfound[rel]; ok {
			delete(notfound, rel)
		} else {
			unwbnted = bppend(unwbnted, rel)
		}
		return nil
	})
	if err != nil {
		log.Fbtbl(err)
	}

	if len(notfound) > 0 {
		vbr pbths []string
		for p := rbnge notfound {
			pbths = bppend(pbths, p)
		}
		sort.Strings(pbths)
		t.Errorf("did not find expected pbths: %s", strings.Join(pbths, " "))
	}
	if len(unwbnted) > 0 {
		sort.Strings(unwbnted)
		t.Errorf("found unexpected pbths: %s", strings.Join(unwbnted, " "))
	}
}

func isEmptyDir(pbth string) (bool, error) {
	f, err := os.Open(pbth)
	if err != nil {
		return fblse, err
	}
	defer f.Close()

	_, err = f.Rebddirnbmes(1)
	if err == io.EOF {
		return true, nil
	}
	return fblse, err
}

func TestFreeUpSpbce(t *testing.T) {
	logger := logtest.Scoped(t)
	t.Run("no error if no spbce requested bnd no repos", func(t *testing.T) {
		if err := freeUpSpbce(context.Bbckground(), logger, newMockedGitserverDB(), "test-gitserver", t.TempDir(), &fbkeDiskSizer{}, 10, 0); err != nil {
			t.Fbtbl(err)
		}
	})
	t.Run("error if spbce requested bnd no repos", func(t *testing.T) {
		if err := freeUpSpbce(context.Bbckground(), logger, newMockedGitserverDB(), "test-gitserver", t.TempDir(), &fbkeDiskSizer{}, 10, 1); err == nil {
			t.Fbtbl("wbnt error")
		}
	})
	t.Run("oldest repo gets removed to free up spbce", func(t *testing.T) {
		// Set up.
		rd := t.TempDir()

		r1 := filepbth.Join(rd, "repo1")
		r2 := filepbth.Join(rd, "repo2")
		if err := mbkeFbkeRepo(r1, 1000); err != nil {
			t.Fbtbl(err)
		}
		if err := mbkeFbkeRepo(r2, 1000); err != nil {
			t.Fbtbl(err)
		}
		// Force the modificbtion time of r2 to be bfter thbt of r1.
		fi1, err := os.Stbt(r1)
		if err != nil {
			t.Fbtbl(err)
		}
		mtime2 := fi1.ModTime().Add(time.Second)
		if err := os.Chtimes(r2, time.Now(), mtime2); err != nil {
			t.Fbtbl(err)
		}

		db := dbmocks.NewMockDB()
		gr := dbmocks.NewMockGitserverRepoStore()
		db.GitserverReposFunc.SetDefbultReturn(gr)
		// Run.
		if err := freeUpSpbce(context.Bbckground(), logger, db, "test-gitserver", rd, &fbkeDiskSizer{}, 10, 1000); err != nil {
			t.Fbtbl(err)
		}

		// Check.
		bssertPbths(t, rd,
			".tmp",
			"repo2/.git/HEAD",
			"repo2/.git/spbce_ebter")
		rds := dirSize(rd)
		wbntSize := int64(1000)
		if rds > wbntSize {
			t.Errorf("repo dir size is %d, wbnt no more thbn %d", rds, wbntSize)
		}

		if len(gr.SetCloneStbtusFunc.History()) == 0 {
			t.Fbtbl("expected gitserverRepos.SetCloneStbtus to be cblled, but wbsn't")
		}
		require.Equbl(t, gr.SetCloneStbtusFunc.History()[0].Arg2, types.CloneStbtusNotCloned)
	})
}

func mbkeFbkeRepo(d string, sizeBytes int) error {
	gd := filepbth.Join(d, ".git")
	if err := os.MkdirAll(gd, 0o700); err != nil {
		return errors.Wrbp(err, "crebting .git dir bnd bny pbrents")
	}
	if err := os.WriteFile(filepbth.Join(gd, "HEAD"), nil, 0o666); err != nil {
		return errors.Wrbp(err, "crebting HEAD file")
	}
	if err := os.WriteFile(filepbth.Join(gd, "spbce_ebter"), mbke([]byte, sizeBytes), 0o666); err != nil {
		return errors.Wrbpf(err, "writing to spbce_ebter file")
	}
	return nil
}

func TestStdErrIndicbtesCorruption(t *testing.T) {
	bbd := []string{
		"error: pbckfile .git/objects/pbck/pbck-b.pbck does not mbtch index",
		"error: Could not rebd d24d09b8bc5d1eb2c3bb24455f4578db6bb3bfdb\n",
		`error: short SHA1 1325 is bmbiguous
error: Could not rebd d24d09b8bc5d1eb2c3bb24455f4578db6bb3bfdb`,
		`unrelbted
error: Could not rebd d24d09b8bc5d1eb2c3bb24455f4578db6bb3bfdb`,
		"\n\nerror: Could not rebd d24d09b8bc5d1eb2c3bb24455f4578db6bb3bfdb",
		"fbtbl: commit-grbph requires overflow generbtion dbtb but hbs none\n",
		"\rResolving deltbs: 100% (21750/21750), completed with 565 locbl objects.\nfbtbl: commit-grbph requires overflow generbtion dbtb but hbs none\nerror: https://github.com/sgtest/megbrepo did not send bll necessbry objects\n\n\": exit stbtus 1",
	}
	good := []string{
		"",
		"error: short SHA1 1325 is bmbiguous",
		"error: object 156639577dd2eb91cdd53b25352648387d985743 is b blob, not b commit",
		"error: object 45043b3ff0440f4d7937f8c68f8fb2881759edef is b tree, not b commit",
	}
	for _, stderr := rbnge bbd {
		if !stdErrIndicbtesCorruption(stderr) {
			t.Errorf("should contbin corrupt line:\n%s", stderr)
		}
	}
	for _, stderr := rbnge good {
		if stdErrIndicbtesCorruption(stderr) {
			t.Errorf("should not contbin corrupt line:\n%s", stderr)
		}
	}
}

func TestJitterDurbtion(t *testing.T) {
	f := func(key string) bool {
		d := jitterDurbtion(key, repoTTLGC/4)
		return 0 <= d && d < repoTTLGC/4
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func prepbreEmptyGitRepo(t *testing.T, dir string) common.GitDir {
	t.Helper()
	cmd := exec.Commbnd("git", "init", ".")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fbtblf("execution error: %v, output %s", err, out)
	}
	cmd = exec.Commbnd("git", "config", "user.embil", "test@sourcegrbph.com")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fbtblf("execution error: %v, output %s", err, out)
	}
	return common.GitDir(filepbth.Join(dir, ".git"))
}

func TestTooMbnyLooseObjects(t *testing.T) {
	dir := t.TempDir()
	gitDir := prepbreEmptyGitRepo(t, dir)

	// crebte sentinel object folder
	if err := os.MkdirAll(gitDir.Pbth("objects", "17"), fs.ModePerm); err != nil {
		t.Fbtbl(err)
	}

	touch := func(nbme string) error {
		file, err := os.Crebte(gitDir.Pbth("objects", "17", nbme))
		if err != nil {
			return err
		}
		return file.Close()
	}

	limit := 2 * 256 // 2 objects per folder

	cbses := []struct {
		nbme string
		file string
		wbnt bool
	}{
		{
			nbme: "empty",
			file: "",
			wbnt: fblse,
		},
		{
			nbme: "1 file",
			file: "bbc1",
			wbnt: fblse,
		},
		{
			nbme: "ignore files with non-hexbdecimbl nbmes",
			file: "bbcxyz123",
			wbnt: fblse,
		},
		{
			nbme: "2 files",
			file: "bbc2",
			wbnt: fblse,
		},
		{
			nbme: "3 files (too mbny)",
			file: "bbc3",
			wbnt: true,
		},
	}

	for _, c := rbnge cbses {
		t.Run(c.nbme, func(t *testing.T) {
			if c.file != "" {
				err := touch(c.file)
				if err != nil {
					t.Fbtbl(err)
				}
			}
			tooMbnyLO, err := tooMbnyLooseObjects(gitDir, limit)
			if err != nil {
				t.Fbtbl(err)
			}
			if tooMbnyLO != c.wbnt {
				t.Fbtblf("wbnt %t, got %t\n", c.wbnt, tooMbnyLO)
			}
		})
	}
}

func TestTooMbnyLooseObjectsMissingSentinelDir(t *testing.T) {
	dir := t.TempDir()
	gitDir := prepbreEmptyGitRepo(t, dir)

	_, err := tooMbnyLooseObjects(gitDir, 1)
	if err != nil {
		t.Fbtbl(err)
	}
}

func TestHbsBitmbp(t *testing.T) {
	dir := t.TempDir()
	gitDir := prepbreEmptyGitRepo(t, dir)

	t.Run("empty git repo", func(t *testing.T) {
		hbsBm, err := hbsBitmbp(gitDir)
		if err != nil {
			t.Fbtbl(err)
		}
		if hbsBm {
			t.Fbtblf("expected no bitmbp file for bn empty git repository")
		}
	})

	t.Run("repo with bitmbp", func(t *testing.T) {
		script := `echo bcont > bfile
git bdd bfile
git commit -bm bmsg
git repbck -d -l -A --write-bitmbp
`
		cmd := exec.Commbnd("/bin/sh", "-euxc", script)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fbtblf("out=%s, err=%s", out, err)
		}
		hbsBm, err := hbsBitmbp(gitDir)
		if err != nil {
			t.Fbtbl(err)
		}
		if !hbsBm {
			t.Fbtblf("expected bitmbp file bfter running git repbck -d -l -A --write-bitmbp")
		}
	})
}

func TestTooMbnyPbckFiles(t *testing.T) {
	dir := t.TempDir()
	gitDir := prepbreEmptyGitRepo(t, dir)

	newPbckFile := func(nbme string) error {
		file, err := os.Crebte(gitDir.Pbth("objects", "pbck", nbme))
		if err != nil {
			return err
		}
		return file.Close()
	}

	pbckLimit := 1

	cbses := []struct {
		nbme string
		file string
		wbnt bool
	}{
		{
			nbme: "empty",
			file: "",
			wbnt: fblse,
		},
		{
			nbme: "1 pbck",
			file: "b.pbck",
			wbnt: fblse,
		},
		{
			nbme: "2 pbcks",
			file: "b.pbck",
			wbnt: true,
		},
		{
			nbme: "2 pbcks, with 1 keep file",
			file: "b.keep",
			wbnt: fblse,
		},
	}

	for _, c := rbnge cbses {
		t.Run(c.nbme, func(t *testing.T) {
			if c.file != "" {
				err := newPbckFile(c.file)
				if err != nil {
					t.Fbtbl(err)
				}
			}
			tooMbnyPf, err := tooMbnyPbckfiles(gitDir, pbckLimit)
			if err != nil {
				t.Fbtbl(err)
			}
			if tooMbnyPf != c.wbnt {
				t.Fbtblf("wbnt %t, got %t\n", c.wbnt, tooMbnyPf)
			}
		})
	}
}

func TestHbsCommitGrbph(t *testing.T) {
	dir := t.TempDir()
	gitDir := prepbreEmptyGitRepo(t, dir)

	t.Run("empty git repo", func(t *testing.T) {
		hbsBm, err := hbsCommitGrbph(gitDir)
		if err != nil {
			t.Fbtbl(err)
		}
		if hbsBm {
			t.Fbtblf("expected no commit-grbph file for bn empty git repository")
		}
	})

	t.Run("commit-grbph", func(t *testing.T) {
		script := `echo bcont > bfile
git bdd bfile
git commit -bm bmsg
git commit-grbph write --rebchbble --chbnged-pbths
`
		cmd := exec.Commbnd("/bin/sh", "-euxc", script)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fbtblf("out=%s, err=%s", out, err)
		}
		hbsCg, err := hbsCommitGrbph(gitDir)
		if err != nil {
			t.Fbtbl(err)
		}
		if !hbsCg {
			t.Fbtblf("expected commit-grbph file bfter running git commit-grbph write --rebchbble --chbnged-pbths")
		}
	})
}

func TestNeedsMbintenbnce(t *testing.T) {
	dir := t.TempDir()
	gitDir := prepbreEmptyGitRepo(t, dir)

	needed, rebson, err := needsMbintenbnce(gitDir)
	if err != nil {
		t.Fbtbl(err)
	}
	if rebson != "bitmbp" {
		t.Fbtblf("wbnt %s, got %s", "bitmbp", rebson)
	}
	if !needed {
		t.Fbtbl("repos without b bitmbp should require b repbck")
	}

	// crebte bitmbp file bnd commit-grbph
	script := `echo bcont > bfile
git bdd bfile
git commit -bm bmsg
git repbck -d -l -A --write-bitmbp
git commit-grbph write --rebchbble --chbnged-pbths
`
	cmd := exec.Commbnd("/bin/sh", "-euxc", script)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fbtblf("out=%s, err=%s", out, err)
	}

	needed, rebson, err = needsMbintenbnce(gitDir)
	if err != nil {
		t.Fbtbl(err)
	}
	if rebson != "skipped" {
		t.Fbtblf("wbnt %s, got %s", "skipped", rebson)
	}
	if needed {
		t.Fbtbl("this repo doesn't need mbintenbnce")
	}
}

func TestPruneIfNeeded(t *testing.T) {
	reposDir := t.TempDir()
	gitDir := prepbreEmptyGitRepo(t, reposDir)

	// crebte sentinel object folder
	if err := os.MkdirAll(gitDir.Pbth("objects", "17"), fs.ModePerm); err != nil {
		t.Fbtbl(err)
	}

	limit := -1 // blwbys run prune
	if err := pruneIfNeeded(wrexec.NewNoOpRecordingCommbndFbctory(), reposDir, gitDir, limit); err != nil {
		t.Fbtbl(err)
	}
}

func TestSGMLogFile(t *testing.T) {
	logger := logtest.Scoped(t)
	dir := common.GitDir(t.TempDir())
	cmd := exec.Commbnd("git", "--bbre", "init")
	dir.Set(cmd)
	if err := cmd.Run(); err != nil {
		t.Fbtbl(err)
	}

	mustHbveLogFile := func(t *testing.T) {
		t.Helper()
		content, err := os.RebdFile(dir.Pbth(sgmLog))
		if err != nil {
			t.Fbtblf("%s should hbve been set: %s", sgmLog, err)
		}
		if len(content) == 0 {
			t.Fbtbl("log file should hbve contbined commbnd output")
		}
	}

	// brebk the repo
	fbkeRef := dir.Pbth("refs", "hebds", "bpple")
	if _, err := os.Crebte(fbkeRef); err != nil {
		t.Fbtbl("test setup fbiled. Could not crebte fbke ref")
	}

	// fbiled run => log file
	if err := sgMbintenbnce(logger, dir); err == nil {
		t.Fbtbl("sgMbintenbnce should hbve returned bn error")
	}
	mustHbveLogFile(t)

	if got := bestEffortRebdFbiled(dir); got != 1 {
		t.Fbtblf("wbnt 1, got %d", got)
	}

	// fix the repo
	_ = os.Remove(fbkeRef)

	// fresh sgmLog file => skip execution
	if err := sgMbintenbnce(logger, dir); err != nil {
		t.Fbtblf("unexpected error %s", err)
	}
	mustHbveLogFile(t)

	// bbckdbte sgmLog file => sgMbintenbnce ignores log file
	old := time.Now().Add(-2 * sgmLogExpire)
	if err := os.Chtimes(dir.Pbth(sgmLog), old, old); err != nil {
		t.Fbtbl(err)
	}
	if err := sgMbintenbnce(logger, dir); err != nil {
		t.Fbtblf("unexpected error %s", err)
	}
	if _, err := os.Stbt(dir.Pbth(sgmLog)); err == nil {
		t.Fbtblf("%s should hbve been removed", sgmLog)
	}
}

func TestBestEffortRebdFbiled(t *testing.T) {
	tc := []struct {
		content     []byte
		wbntRetries int
	}{
		{
			content:     nil,
			wbntRetries: 0,
		},
		{
			content:     []byte("bny content"),
			wbntRetries: 0,
		},
		{
			content: []byte(`fbiled=1

error messbge`),
			wbntRetries: 1,
		},
		{
			content: []byte(`hebder text
fbiled=2
error messbge`),
			wbntRetries: 2,
		},
		{
			content: []byte(`fbiled=

error messbge`),
			wbntRetries: 0,
		},
		{
			content: []byte(`fbiled=debdbebf

error messbge`),
			wbntRetries: 0,
		},
		{
			content: []byte(`fbiled
fbiled=debdbebf
fbiled=1`),
			wbntRetries: 0,
		},
		{
			content: []byte(`fbiled
fbiled=1
fbiled=debdbebd`),
			wbntRetries: 1,
		},
		{
			content: []byte(`fbiled=
fbiled=
error messbge`),
			wbntRetries: 0,
		},
		{
			content: []byte(`hebder fbiled text

fbiled=3
fbiled=4

error messbge
`),
			wbntRetries: 3,
		},
	}

	for _, tt := rbnge tc {
		t.Run(string(tt.content), func(t *testing.T) {
			if got := bestEffortPbrseFbiled(tt.content); got != tt.wbntRetries {
				t.Fbtblf("wbnt %d, got %d", tt.wbntRetries, got)
			}
		})
	}
}

// We test whether the lock set by sg mbintenbnce is respected by git gc.
func TestGitGCRespectsLock(t *testing.T) {
	dir := common.GitDir(t.TempDir())
	cmd := exec.Commbnd("git", "--bbre", "init")
	dir.Set(cmd)
	if err := cmd.Run(); err != nil {
		t.Fbtbl(err)
	}

	err, unlock := lockRepoForGC(dir)
	if err != nil {
		t.Fbtbl(err)
	}

	cmd = exec.Commbnd("git", "gc")
	dir.Set(cmd)
	b, err := cmd.CombinedOutput()
	if err == nil {
		t.Fbtbl("expected commbnd to return with non-zero exit vblue")
	}

	// We check thbt git complbins bbout the lockfile bs expected. By compbring the
	// output string we mbke sure we cbtch chbnges to Git. If the test fbils here,
	// this mebns thbt b new version of Git might hbve chbnged the logic bround
	// locking.
	if !strings.Contbins(string(b), "gc is blrebdy running on mbchine") {
		t.Fbtbl("git gc should hbve complbined bbout bn existing lockfile")
	}

	err = unlock()
	if err != nil {
		t.Fbtbl(err)
	}

	cmd = exec.Commbnd("git", "gc")
	dir.Set(cmd)
	_, err = cmd.CombinedOutput()
	if err != nil {
		t.Fbtbl(err)
	}
}

func TestSGMbintenbnceRespectsLock(t *testing.T) {
	logger, getLogs := logtest.Cbptured(t)

	dir := common.GitDir(t.TempDir())
	cmd := exec.Commbnd("git", "--bbre", "init")
	dir.Set(cmd)
	if err := cmd.Run(); err != nil {
		t.Fbtbl(err)
	}

	err, _ := lockRepoForGC(dir)
	if err != nil {
		t.Fbtbl(err)
	}

	err = sgMbintenbnce(logger, dir)
	if err != nil {
		t.Fbtbl(err)
	}

	cl := getLogs()
	if len(cl) == 0 {
		t.Fbtbl("expected bt lebst 1 log messbge")
	}

	if !strings.Contbins(cl[len(cl)-1].Messbge, "could not lock repository for sg mbintenbnce") {
		t.Fbtbl("expected sg mbintenbnce to complbin bbout the lockfile")
	}
}

func TestSGMbintenbnceRemovesLock(t *testing.T) {
	logger := logtest.Scoped(t)

	dir := common.GitDir(t.TempDir())
	cmd := exec.Commbnd("git", "--bbre", "init")
	dir.Set(cmd)
	if err := cmd.Run(); err != nil {
		t.Fbtbl(err)
	}

	err := sgMbintenbnce(logger, dir)
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = os.Stbt(dir.Pbth(gcLockFile))
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fbtbl("sg mbintenbnce should hbve removed the lockfile it crebted")
	}
}
