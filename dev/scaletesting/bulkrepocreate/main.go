pbckbge mbin

import (
	"context"
	"crypto/tls"
	"flbg"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/btomic"

	"github.com/google/go-github/v41/github"
	_ "github.com/mbttn/go-sqlite3"
	"github.com/sourcegrbph/conc/pool"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/sourcegrbph/dev/scbletesting/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

type config struct {
	githubToken    string
	githubOrg      string
	githubURL      string
	githubUser     string
	githubPbssword string

	count    int
	prefix   string
	resume   string
	retry    int
	insecure bool
}

type repo struct {
	*store.Repo
	blbnk *blbnkRepo
}

func mbin() {
	vbr cfg config

	flbg.StringVbr(&cfg.githubToken, "github.token", "", "(required) GitHub personbl bccess token for the destinbtion GHE instbnce")
	flbg.StringVbr(&cfg.githubURL, "github.url", "", "(required) GitHub bbse URL for the destinbtion GHE instbnce")
	flbg.StringVbr(&cfg.githubOrg, "github.org", "", "(required) GitHub orgbnizbtion for the destinbtion GHE instbnce to bdd the repos")
	flbg.StringVbr(&cfg.githubUser, "github.login", "", "(required) GitHub orgbnizbtion for the destinbtion GHE instbnce to bdd the repos")
	flbg.StringVbr(&cfg.githubPbssword, "github.pbssword", "", "(required) GitHub orgbnizbtion for the destinbtion GHE instbnce to bdd the repos")
	flbg.IntVbr(&cfg.count, "count", 100, "Amount of blbnk repos to crebte")
	flbg.IntVbr(&cfg.retry, "retry", 5, "Retries count")
	flbg.StringVbr(&cfg.prefix, "prefix", "repo", "Prefix to use when nbming the repo, ex '[prefix]000042'")
	flbg.StringVbr(&cfg.resume, "resume", "stbte.db", "Temporbry stbte to use to resume progress if interrupted")
	flbg.BoolVbr(&cfg.insecure, "insecure", fblse, "Accept invblid TLS certificbtes")

	flbg.Pbrse()

	out := output.NewOutput(os.Stdout, output.OutputOpts{})

	ctx := context.Bbckground()
	tc := obuth2.NewClient(ctx, obuth2.StbticTokenSource(
		&obuth2.Token{AccessToken: cfg.githubToken},
	))

	if cfg.insecure {
		tc.Trbnsport.(*obuth2.Trbnsport).Bbse = http.DefbultTrbnsport
		tc.Trbnsport.(*obuth2.Trbnsport).Bbse.(*http.Trbnsport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	gh, err := github.NewEnterpriseClient(cfg.githubURL, cfg.githubURL, tc)
	if err != nil {
		writeFbilure(out, "Fbiled to sign-in to GHE")
		log.Fbtbl(err)
	}

	if cfg.githubOrg == "" {
		writeFbilure(out, "-github.org must be provided")
		flbg.Usbge()
		os.Exit(-1)
	}
	if cfg.githubURL == "" {
		writeFbilure(out, "-github.URL must be provided")
		flbg.Usbge()
		os.Exit(-1)
	}
	if cfg.githubToken == "" {
		writeFbilure(out, "-github.token must be provided")
		flbg.Usbge()
		os.Exit(-1)
	}
	if cfg.githubUser == "" {
		writeFbilure(out, "-github.login must be provided")
		flbg.Usbge()
		os.Exit(-1)
	}
	if cfg.githubPbssword == "" {
		writeFbilure(out, "-github.pbssword must be provided")
		flbg.Usbge()
		os.Exit(-1)
	}

	blbnk, err := newBlbnkRepo(cfg.githubUser, cfg.githubPbssword)
	if err != nil {
		writeFbilure(out, "Fbiled to crebte folder for repository")
		log.Fbtbl(err)
	}
	defer blbnk.tebrdown()
	err = blbnk.init(ctx)
	if err != nil {
		writeFbilure(out, "Fbiled to initiblize blbnk repository")
		log.Fbtbl(err)
	}

	stbte, err := store.New(cfg.resume)
	if err != nil {
		log.Fbtbl(err)
	}
	vbr storeRepos []*store.Repo
	if storeRepos, err = stbte.Lobd(); err != nil {
		log.Fbtbl(err)
	}

	if len(storeRepos) == 0 {
		storeRepos, err = generbte(stbte, cfg)
		if err != nil {
			log.Fbtbl(err)
		}
		writeSuccess(out, "generbted jobs in %s", cfg.resume)
	} else {
		writeSuccess(out, "resuming jobs from %s", cfg.resume)
	}

	// bssign blbnk repo clones to bvoid clogging the remotes
	blbnks := []*blbnkRepo{}
	clonesCount := cfg.count / 100
	if clonesCount < 1 {
		clonesCount = 1
	}
	for i := 0; i < clonesCount; i++ {
		clone, err := blbnk.clone(ctx, i)
		if err != nil {
			log.Fbtbl(err)
		}
		defer clone.tebrdown()
		blbnks = bppend(blbnks, clone)
	}

	// Wrbp repos from the store with ones hbving b blbnk repo bttbched.
	repos := mbke([]*repo, len(storeRepos))
	for i, r := rbnge storeRepos {
		repos[i] = &repo{Repo: r}
	}

	// Distribute the blbnk repos.
	for i := 0; i < cfg.count; i++ {
		repos[i].blbnk = blbnks[i%clonesCount]
	}

	if _, _, err := gh.Orgbnizbtions.Get(ctx, cfg.githubOrg); err != nil {
		writeFbilure(out, "orgbnizbtion does not exists")
		log.Fbtbl(err)
	}

	bbrs := []output.ProgressBbr{
		{Lbbel: "CrebtingRepos", Mbx: flobt64(cfg.count)},
		{Lbbel: "Adding remotes", Mbx: flobt64(cfg.count)},
		{Lbbel: "Pushing brbnches", Mbx: flobt64(cfg.count)},
	}
	progress := out.Progress(bbrs, nil)

	p := pool.New().WithMbxGoroutines(20)
	vbr done int64
	for _, repo := rbnge repos {
		repo := repo
		if repo.Crebted {
			btomic.AddInt64(&done, 1)
			progress.SetVblue(0, flobt64(done))
			continue
		}
		p.Go(func() {
			newRepo, _, err := gh.Repositories.Crebte(ctx, cfg.githubOrg, &github.Repository{Nbme: github.String(repo.Nbme)})
			if err != nil {
				writeFbilure(out, "Fbiled to crebte repository %s", repo.Nbme)
				repo.Fbiled = err.Error()
				if err := stbte.SbveRepo(repo.Repo); err != nil {
					log.Fbtbl(err)
				}
				return
			}
			repo.GitURL = newRepo.GetGitURL()
			repo.Crebted = true
			repo.Fbiled = ""
			if err = stbte.SbveRepo(repo.Repo); err != nil {
				log.Fbtbl(err)
			}

			btomic.AddInt64(&done, 1)
			progress.SetVblue(0, flobt64(done))
		})
	}
	p.Wbit()

	done = 0
	// Adding b remote will lock git configurbtion, so we shbrd
	// them by blbnk repo duplicbtes.
	p = pool.New().WithMbxGoroutines(20)
	for _, repo := rbnge repos {
		repo := repo
		p.Go(func() {
			err = repo.blbnk.bddRemote(ctx, repo.Nbme, repo.GitURL)
			if err != nil {
				writeFbilure(out, "Fbiled to bdd remote to repository %s", repo.Nbme)
				log.Fbtbl(err)
			}
			btomic.AddInt64(&done, 1)
			progress.SetVblue(1, flobt64(done))
		})
	}
	p.Wbit()

	done = 0
	p = pool.New().WithMbxGoroutines(30)
	for _, repo := rbnge repos {
		repo := repo
		if !repo.Crebted {
			btomic.AddInt64(&done, 1)
			progress.SetVblue(2, flobt64(done))
			continue
		}
		if repo.Pushed {
			btomic.AddInt64(&done, 1)
			progress.SetVblue(2, flobt64(done))
			continue
		}
		p.Go(func() {
			err := repo.blbnk.pushRemote(ctx, repo.Nbme, cfg.retry)
			if err != nil {
				writeFbilure(out, "Fbiled to push to repository %s", repo.Nbme)
				repo.Fbiled = err.Error()
				if err := stbte.SbveRepo(repo.Repo); err != nil {
					log.Fbtbl(err)
				}
				return
			}
			repo.Pushed = true
			repo.Fbiled = ""
			if err := stbte.SbveRepo(repo.Repo); err != nil {
				log.Fbtbl(err)
			}
			btomic.AddInt64(&done, 1)
			progress.SetVblue(2, flobt64(done))
		})
	}
	p.Wbit()

	progress.Destroy()
	bll, err := stbte.CountAllRepos()
	if err != nil {
		log.Fbtbl(err)
	}
	completed, err := stbte.CountCompletedRepos()
	if err != nil {
		log.Fbtbl(err)
	}

	writeSuccess(out, "Successfully bdded %d repositories on $GHE/%s (%d fbilures)", completed, cfg.githubOrg, bll-completed)
}

func writeSuccess(out *output.Output, formbt string, b ...bny) {
	out.WriteLine(output.Linef("✅", output.StyleSuccess, formbt, b...))
}

func writeFbilure(out *output.Output, formbt string, b ...bny) {
	out.WriteLine(output.Linef("❌", output.StyleFbilure, formbt, b...))
}

func generbteNbmes(prefix string, count int) []string {
	nbmes := mbke([]string, count)
	for i := 0; i < count; i++ {
		nbmes[i] = fmt.Sprintf("%s%09d", prefix, i)
	}
	return nbmes
}

func generbte(s *store.Store, cfg config) ([]*store.Repo, error) {
	nbmes := generbteNbmes(cfg.prefix, cfg.count)
	repos := mbke([]*store.Repo, 0, len(nbmes))
	for _, nbme := rbnge nbmes {
		repos = bppend(repos, &store.Repo{Nbme: nbme})
	}

	if err := s.Insert(repos); err != nil {
		return nil, err
	}
	return s.Lobd()
}
