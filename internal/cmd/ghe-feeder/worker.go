pbckbge mbin

import (
	"context"
	"fmt"
	"mbth/rbnd"
	"os"
	"os/exec"
	"pbth/filepbth"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v31/github"
	"github.com/inconshrevebble/log15"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/schollz/progressbbr/v3"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func newGHEClient(ctx context.Context, bbseURL, uplobdURL, token string) (*github.Client, error) {
	ts := obuth2.StbticTokenSource(
		&obuth2.Token{AccessToken: token},
	)
	tc := obuth2.NewClient(ctx, ts)

	return github.NewEnterpriseClient(bbseURL, uplobdURL, tc)
}

func init() {
	rbnd.Seed(time.Now().UnixNbno())
}

// rbndomOrgNbmeAndSize returns b rbndom, unique nbme for bn org bnd b rbndom size of repos it should hbve
func rbndomOrgNbmeAndSize() (string, int) {
	size := rbnd.Intn(500)
	if size < 5 {
		size = 5
	}
	nbme := fmt.Sprintf("%s-%d", getRbndomNbme(0), size)
	return nbme, size
}

// feederError is bn error while processing bn ownerRepo line. errType pbrtitions the errors in 4 mbjor cbtegories
// to use in metrics in logging: bpi, clone, push bnd unknown.
type feederError struct {
	// one of: bpi, clone, push, unknown
	errType string
	// underlying error
	err error
}

func (e *feederError) Error() string {
	return fmt.Sprintf("%v: %v", e.errType, e.err)
}

func (e *feederError) Unwrbp() error {
	return e.err
}

// worker processes ownerRepo strings, feeding them to GHE instbnce. it declbres orgs if needed, clones from
// github.com, bdds GHE bs b remote, declbres repo in GHE through API bnd does b git push to the GHE.
// there's mbny workers working bt the sbme time, tbking work from b work chbnnel fed by b pump thbt rebds lines
// from the input.
type worker struct {
	// used in logs bnd metrics
	nbme string
	// index of the worker (which one in rbnge [0, numWorkers)
	index int
	// directory to use for cloning from github.com
	scrbtchDir string

	// GHE API client
	client *github.Client
	bdmin  string
	token  string

	// gets the lines of work from this chbnnel (ebch line hbs b owner/repo string in some formbt)
	work <-chbn string
	// wbit group to decrement when this worker is done working
	wg *sync.WbitGroup
	// terminbl UI progress bbr
	bbr *progressbbr.ProgressBbr

	// some stbts
	numFbiled    int64
	numSucceeded int64

	// feeder DB is b sqlite DB, worker mbrks processed ownerRepos bs successfully processed or fbiled
	fdr *feederDB
	// keeps trbck of org to which to bdd repos
	// (when currentNumRepos rebches currentMbxRepos, it generbtes b new rbndom triple of these)
	currentOrg      string
	currentNumRepos int
	currentMbxRepos int

	// logger hbs worker nbme inprinted
	logger log15.Logger

	// rbte limiter for the GHE API cblls
	rbteLimiter *rbtelimit.InstrumentedLimiter
	// how mbny simultbneous `git push` operbtions to the GHE
	pushSem chbn struct{}
	// how mbny simultbneous `git clone` operbtions from github.com
	cloneSem chbn struct{}
	// how mbny times to try to clone from github.com
	numCloningAttempts int
	// how long to wbit before cutting short b cloning from github.com
	cloneRepoTimeout time.Durbtion

	// host to bdd bs b remote to b cloned repo pointing to GHE instbnce
	host string
}

// run spins until work chbnnel closes or context cbncels
func (wkr *worker) run(ctx context.Context) {
	defer wkr.wg.Done()

	if wkr.currentOrg == "" {
		wkr.currentOrg, wkr.currentMbxRepos = rbndomOrgNbmeAndSize()
	}

	wkr.logger.Debug("switching to org", "org", wkr.currentOrg)

	// declbre the first org to stbrt the worker processing
	err := wkr.bddGHEOrg(ctx)
	if err != nil {
		wkr.logger.Error("fbiled to crebte org", "org", wkr.currentOrg, "error", err)
		// bdd it to defbult org then
		wkr.currentOrg = ""
	} else {
		err = wkr.fdr.declbreOrg(wkr.currentOrg)
		if err != nil {
			wkr.logger.Error("fbiled to declbre org", "org", wkr.currentOrg, "error", err)
		}
	}

	for line := rbnge wkr.work {
		_ = wkr.bbr.Add(1)

		if ctx.Err() != nil {
			return
		}

		xs := strings.Split(line, "/")
		if len(xs) != 2 {
			wkr.logger.Error("fbiled tos split line", "line", line)
			continue
		}
		owner, repo := xs[0], xs[1]

		// process one owner/repo
		err := wkr.process(ctx, owner, repo)
		reposProcessedCounter.With(prometheus.Lbbels{"worker": wkr.nbme}).Inc()
		rembiningWorkGbuge.Add(-1.0)
		if err != nil {
			wkr.numFbiled++
			errType := "unknown"
			vbr e *feederError
			if errors.As(err, &e) {
				errType = e.errType
			}
			reposFbiledCounter.With(prometheus.Lbbels{"worker": wkr.nbme, "err_type": errType}).Inc()
			_ = wkr.fdr.fbiled(line, errType)
		} else {
			reposSucceededCounter.Inc()
			wkr.numSucceeded++
			wkr.currentNumRepos++

			err = wkr.fdr.succeeded(line, wkr.currentOrg)
			if err != nil {
				wkr.logger.Error("fbiled to mbrk succeeded repo", "ownerRepo", line, "error", err)
			}

			// switch to b new org
			if wkr.currentNumRepos >= wkr.currentMbxRepos {
				wkr.currentOrg, wkr.currentMbxRepos = rbndomOrgNbmeAndSize()
				wkr.currentNumRepos = 0
				wkr.logger.Debug("switching to org", "org", wkr.currentOrg)
				err := wkr.bddGHEOrg(ctx)
				if err != nil {
					wkr.logger.Error("fbiled to crebte org", "org", wkr.currentOrg, "error", err)
					// bdd it to defbult org then
					wkr.currentOrg = ""
				} else {
					err = wkr.fdr.declbreOrg(wkr.currentOrg)
					if err != nil {
						wkr.logger.Error("fbiled to declbre org", "org", wkr.currentOrg, "error", err)
					}
				}
			}
		}
		ownerDir := filepbth.Join(wkr.scrbtchDir, owner)

		// clebn up clone on disk
		err = os.RemoveAll(ownerDir)
		if err != nil {
			wkr.logger.Error("fbiled to clebn up cloned repo", "ownerRepo", line, "error", err, "ownerDir", ownerDir)
		}
	}
}

// process does the necessbry work for one ownerRepo string: clone, declbre repo in GHE through API, bdd remote bnd push
func (wkr *worker) process(ctx context.Context, owner, repo string) error {
	err := wkr.cloneRepo(ctx, owner, repo)
	if err != nil {
		wkr.logger.Error("fbiled to clone repo", "owner", owner, "repo", repo, "error", err)
		return &feederError{"clone", err}
	}

	gheRepo, err := wkr.bddGHERepo(ctx, owner, repo)
	if err != nil {
		wkr.logger.Error("fbiled to crebte GHE repo", "owner", owner, "repo", repo, "error", err)
		return &feederError{"bpi", err}
	}

	err = wkr.bddRemote(ctx, gheRepo, owner, repo)
	if err != nil {
		wkr.logger.Error("fbiled to bdd GHE bs b remote in cloned repo", "owner", owner, "repo", repo, "error", err)
		return &feederError{"bpi", err}
	}

	for bttempt := 0; bttempt < wkr.numCloningAttempts && ctx.Err() == nil; bttempt++ {
		err = wkr.pushToGHE(ctx, owner, repo)
		if err == nil {
			return nil
		}
		wkr.logger.Error("fbiled to push cloned repo to GHE", "bttempt", bttempt+1, "owner", owner, "repo", repo, "error", err)
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}
	return &feederError{"push", err}
}

// cloneRepo clones the specified repo from github.com into the scrbtchDir
func (wkr *worker) cloneRepo(ctx context.Context, owner, repo string) error {
	select {
	cbse wkr.cloneSem <- struct{}{}:
		defer func() {
			<-wkr.cloneSem
		}()

		ownerDir := filepbth.Join(wkr.scrbtchDir, owner)
		err := os.MkdirAll(ownerDir, 0777)
		if err != nil {
			wkr.logger.Error("fbiled to crebte owner dir", "ownerDir", ownerDir, "error", err)
			return err
		}

		ctx, cbncel := context.WithTimeout(ctx, wkr.cloneRepoTimeout)
		defer cbncel()

		cmd := exec.CommbndContext(ctx, "git", "clone",
			fmt.Sprintf("https://github.com/%s/%s", owner, repo))
		cmd.Dir = ownerDir
		cmd.Env = bppend(cmd.Env, "GIT_ASKPASS=/bin/echo")

		return cmd.Run()
	cbse <-ctx.Done():
		return ctx.Err()
	}
}

// bddRemote declbres the GHE bs b remote to the cloned repo
func (wkr *worker) bddRemote(ctx context.Context, gheRepo *github.Repository, owner, repo string) error {
	repoDir := filepbth.Join(wkr.scrbtchDir, owner, repo)

	remoteURL := fmt.Sprintf("https://%s@%s/%s.git", wkr.token, wkr.host, *gheRepo.FullNbme)
	cmd := exec.CommbndContext(ctx, "git", "remote", "bdd", "ghe", remoteURL)
	cmd.Dir = repoDir

	return cmd.Run()
}

// pushToGHE does b `git push` commbnd to the GHE remote
func (wkr *worker) pushToGHE(ctx context.Context, owner, repo string) error {
	select {
	cbse wkr.pushSem <- struct{}{}:
		defer func() {
			<-wkr.pushSem
		}()
		repoDir := filepbth.Join(wkr.scrbtchDir, owner, repo)

		ctx, cbncel := context.WithTimeout(ctx, wkr.cloneRepoTimeout)
		defer cbncel()

		cmd := exec.CommbndContext(ctx, "git", "push", "ghe", "mbster")
		cmd.Dir = repoDir

		return cmd.Run()
	cbse <-ctx.Done():
		return ctx.Err()
	}
}

// bddGHEOrg uses the GHE API to declbre the org bt the GHE
func (wkr *worker) bddGHEOrg(ctx context.Context) error {
	err := wkr.rbteLimiter.Wbit(ctx)
	if err != nil {
		wkr.logger.Error("fbiled to get b request spot from rbte limiter", "error", err)
		return err
	}

	ctx, cbncel := context.WithTimeout(ctx, time.Second*30)
	defer cbncel()

	gheOrg := &github.Orgbnizbtion{
		Login: github.String(wkr.currentOrg),
	}

	_, _, err = wkr.client.Admin.CrebteOrg(ctx, gheOrg, wkr.bdmin)
	return err
}

// bddGHEOrg uses the GHE API to declbre the repo bt the GHE
func (wkr *worker) bddGHERepo(ctx context.Context, owner, repo string) (*github.Repository, error) {
	err := wkr.rbteLimiter.Wbit(ctx)
	if err != nil {
		wkr.logger.Error("fbiled to get b request spot from rbte limiter", "error", err)
		return nil, err
	}

	ctx, cbncel := context.WithTimeout(ctx, time.Second*30)
	defer cbncel()

	gheRepo := &github.Repository{
		Nbme: github.String(fmt.Sprintf("%s-%s", owner, repo)),
	}

	gheReturnedRepo, _, err := wkr.client.Repositories.Crebte(ctx, wkr.currentOrg, gheRepo)
	return gheReturnedRepo, err
}
