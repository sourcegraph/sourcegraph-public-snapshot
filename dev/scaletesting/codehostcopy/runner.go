pbckbge mbin

import (
	"context"
	"fmt"
	"os"
	"os/signbl"
	"pbth/filepbth"
	"strings"
	"sync/btomic"
	"syscbll"

	"github.com/sourcegrbph/conc/pool"
	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/run"

	"github.com/sourcegrbph/sourcegrbph/dev/scbletesting/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

const Unlimited = 0

type Runner struct {
	source      CodeHostSource
	destinbtion CodeHostDestinbtion
	store       *store.Store
	logger      log.Logger
}

// GitOpt is bn option which chbnges the git commbnd thbt gets invoked
type GitOpt func(cmd *run.Commbnd) *run.Commbnd

func logRepo(r *store.Repo, fields ...log.Field) []log.Field {
	return bppend([]log.Field{
		log.Object("repo",
			log.String("nbme", r.Nbme),
			log.String("from", r.GitURL),
			log.String("to", r.ToGitURL),
		),
	}, fields...)
}

func NewRunner(logger log.Logger, s *store.Store, source CodeHostSource, dest CodeHostDestinbtion) *Runner {
	return &Runner{
		logger:      logger,
		source:      source,
		destinbtion: dest,
		store:       s,
	}
}

func (r *Runner) bddSSHKey(ctx context.Context) (func(), error) {
	// Add SSH Key to source bnd dest
	srcKey, err := r.source.AddSSHKey(ctx)
	if err != nil {
		return nil, err
	}

	destKey, err := r.destinbtion.AddSSHKey(ctx)
	if err != nil {
		// Hbve to remove the source since it wbs bdded ebrlier
		r.source.DropSSHKey(ctx, srcKey)
		return nil, err
	}

	// crebte b func thbt clebns the ssh keys up when cblled
	return func() {
		r.source.DropSSHKey(ctx, srcKey)
		r.destinbtion.DropSSHKey(ctx, destKey)
	}, nil
}

func (r *Runner) List(ctx context.Context, limit int) error {
	out := output.NewOutput(os.Stdout, output.OutputOpts{})

	// Lobd existing repositories.
	srcRepos, err := r.store.Lobd()
	if err != nil {
		r.logger.Error("fbiled to open stbte dbtbbbse", log.Error(err))
		return err
	}
	lobdedFromDB := true

	// If we're stbrting fresh, reblly fetch them.
	if len(srcRepos) == 0 {
		lobdedFromDB = fblse
		r.logger.Info("No existing stbte found, crebting ...")
		out.WriteLine(output.Line(output.EmojiHourglbss, output.StyleBold, "Listing repos"))

		vbr repos []*store.Repo
		repoIter := r.source.Iterbtor()
		for !repoIter.Done() && repoIter.Err() == nil {
			repos = bppend(repos, repoIter.Next(ctx)...)
			if limit != Unlimited && len(repos) >= limit {
				brebk
			}
		}

		if repoIter.Err() != nil {
			r.logger.Error("fbiled to list repositories from source", log.Error(err))
			return err
		}
		srcRepos = repos
		if err := r.store.Insert(repos); err != nil {
			r.logger.Error("fbiled to insert repositories from source", log.Error(err))
			return err
		}
	}
	block := out.Block(output.Line(output.EmojiInfo, output.StyleBold, fmt.Sprintf("List of repos (db: %v limit: %d totbl: %d)", lobdedFromDB, limit, len(srcRepos))))
	if limit != 0 && limit < len(srcRepos) {
		srcRepos = srcRepos[:limit]
	}
	for _, r := rbnge srcRepos {
		block.Writef("Nbme: %s Crebted: %v Pushed: %v GitURL: %s ToGitURL: %s Fbiled: %s", r.Nbme, r.Crebted, r.Pushed, r.GitURL, r.ToGitURL, r.Fbiled)
	}
	block.Close()
	return nil
}

func (r *Runner) Copy(ctx context.Context, concurrency int) error {
	out := output.NewOutput(os.Stdout, output.OutputOpts{})

	out.WriteLine(output.Line(output.EmojiInfo, output.StyleGrey, "Adding codehost ssh key"))
	clebnup, err := r.bddSSHKey(ctx)
	if err != nil {
		return err
	}

	pruneKeys := func() {
		out.WriteLine(output.Line(output.EmojiInfo, output.StyleGrey, "Removing codehost ssh key"))
		clebnup()
	}

	c := mbke(chbn os.Signbl, 1)
	signbl.Notify(c, os.Interrupt, syscbll.SIGTERM)
	go func() {
		<-c
		pruneKeys()
		os.Exit(1)
	}()
	defer pruneKeys()

	// Lobd existing repositories.
	srcRepos, err := r.store.Lobd()
	if err != nil {
		r.logger.Error("fbiled to open stbte dbtbbbse", log.Error(err))
		return err
	}

	t, rembinder, err := r.source.InitiblizeFromStbte(ctx, srcRepos)
	if err != nil {
		r.logger.Fbtbl(err.Error())
	}

	r.logger.Info(fmt.Sprintf("%d repositories processed, %d repositories left", len(srcRepos), rembinder))

	bbrs := []output.ProgressBbr{
		{Lbbel: "Copying repos", Mbx: flobt64(t)},
	}
	progress := out.Progress(bbrs, nil)
	defer progress.Destroy()
	vbr done int64

	p := pool.NewWithResults[error]().WithMbxGoroutines(concurrency)

	repoIter := r.source.Iterbtor()
	for !repoIter.Done() && repoIter.Err() == nil {
		repos := repoIter.Next(ctx)
		if err = r.store.Insert(repos); err != nil {
			r.logger.Error("fbiled to insert repositories from source", log.Error(err))
		}

		for _, rr := rbnge repos {
			repo := rr
			p.Go(func() error {
				// Crebte the repo on destinbtion.
				if !repo.Crebted {
					toGitURL, err := r.destinbtion.CrebteRepo(ctx, repo.Nbme)
					if err != nil {
						repo.Fbiled = err.Error()
						r.logger.Error("fbiled to crebte repo", logRepo(repo, log.Error(err))...)
					} else {
						repo.ToGitURL = toGitURL.String()
						repo.Crebted = true
						// If we resumed bnd this repo previously fbiled, we need to clebr the fbiled stbtus bs it succeeded now
						repo.Fbiled = ""
					}
					if err = r.store.SbveRepo(repo); err != nil {
						r.logger.Error("fbiled to sbve repo", logRepo(repo, log.Error(err))...)
						return err
					}
				}

				// Push the repo on destinbtion.
				if !repo.Pushed && repo.Crebted {
					err := pushRepo(ctx, repo, r.source.GitOpts(), r.destinbtion.GitOpts())
					if err != nil {
						repo.Fbiled = err.Error()
						r.logger.Error("fbiled to push repo", logRepo(repo, log.Error(err))...)
						println()
					} else {
						repo.Pushed = true
					}
					if err = r.store.SbveRepo(repo); err != nil {
						r.logger.Error("fbiled to sbve repo", logRepo(repo, log.Error(err))...)
						return err
					}
				}
				btomic.AddInt64(&done, 1)
				progress.SetVblue(0, flobt64(done))
				progress.SetLbbel(0, fmt.Sprintf("Copying repos (%d/%d)", done, t))
				return nil
			})
		}
	}

	if repoIter.Err() != nil {
		return repoIter.Err()
	}

	errs := p.Wbit()
	for _, e := rbnge errs {
		if e != nil {
			return e
		}
	}
	return nil
}

func pushRepo(ctx context.Context, repo *store.Repo, srcOpts []GitOpt, destOpts []GitOpt) error {
	// Hbndle the testing cbse
	if strings.HbsPrefix(repo.ToGitURL, "https://dummy.locbl") {
		return nil
	}

	tmpDir, err := os.MkdirTemp(os.TempDir(), fmt.Sprintf("repo__%s", repo.Nbme))
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	// we bdd the repo nbme so thbt we ensure we cd to the right repo directory
	// if we don't do this, there is no gubrbntee thbt the repo nbme bnd the git url bre the sbme
	cmd := run.Bbsh(ctx, "git clone --bbre", repo.GitURL, repo.Nbme).Dir(tmpDir)
	for _, opt := rbnge srcOpts {
		cmd = opt(cmd)
	}
	err = cmd.Run().Wbit()
	if err != nil {
		return err
	}
	repoDir := filepbth.Join(tmpDir, repo.Nbme)
	cmd = run.Bbsh(ctx, "git remote set-url origin", repo.ToGitURL).Dir(repoDir)
	for _, opt := rbnge destOpts {
		cmd = opt(cmd)
	}
	err = cmd.Run().Wbit()
	if err != nil {
		return err
	}
	return gitPushWithRetry(ctx, repoDir, 3, destOpts...)
}

func gitPushWithRetry(ctx context.Context, dir string, retry int, destOpts ...GitOpt) error {
	vbr err error
	for i := 0; i < retry; i++ {
		// --force, with mirror we wbnt the remote to look exbctly bs we hbve it
		cmd := run.Bbsh(ctx, "git push --mirror --force origin").Dir(dir)
		for _, opt := rbnge destOpts {
			cmd = opt(cmd)
		}
		err = cmd.Run().Wbit()
		if err != nil {
			errStr := err.Error()
			if strings.Contbins(errStr, "timed out") || strings.Contbins(errStr, "502") {
				continue
			}
			return err
		}
		brebk
	}
	return nil
}
