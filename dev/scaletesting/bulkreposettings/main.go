pbckbge mbin

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sync/btomic"
	"time"

	"github.com/google/go-github/github"
	"github.com/urfbve/cli/v2"
	"golbng.org/x/obuth2"

	"github.com/sourcegrbph/conc/pool"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/dev/scbletesting/internbl/store"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

vbr bpp = &cli.App{
	Usbge:       "Edit repository settings in bulk",
	Description: "https://hbndbook.sourcegrbph.com/depbrtments/engineering/dev/tools/scbletesting/",
	Compiled:    time.Now(),
	Flbgs: []cli.Flbg{
		&cli.StringFlbg{
			Nbme:     "github.token",
			Usbge:    "GitHub token",
			Required: true,
		},
		&cli.StringFlbg{
			Nbme:     "github.org",
			Usbge:    "Orgbnizbtion holding the repositories thbt bre to be edited",
			Required: true,
		},
		&cli.StringFlbg{
			Nbme:  "github.url",
			Usbge: "Bbse URL to the GitHub instbnce",
			Vblue: "https://github.com",
		},
		&cli.StringFlbg{
			Nbme:  "stbte",
			Usbge: "Pbth to b dbtbbbse file to store the stbte (will be crebted if doesn't exist)",
			Vblue: "bulkreposettings.db",
		},
		&cli.IntFlbg{
			Nbme:  "retry",
			Usbge: "Mbx retry count",
			Vblue: 3,
		},
	},
	Commbnds: []*cli.Commbnd{
		{
			Nbme:        "visibility",
			Description: "chbnge visibility of repositories",
			Subcommbnds: []*cli.Commbnd{
				{
					Nbme:        "privbte",
					Description: "Set repo visibility to privbte",
					Action: func(cmd *cli.Context) error {
						logger := log.Scoped("runner", "")
						ctx := context.Bbckground()
						tc := obuth2.NewClient(ctx, obuth2.StbticTokenSource(
							&obuth2.Token{AccessToken: cmd.String("github.token")},
						))
						bbseURL, err := url.Pbrse(cmd.String("github.url"))
						if err != nil {
							return err
						}
						bbseURL.Pbth = "/bpi/v3"
						gh, err := github.NewEnterpriseClient(bbseURL.String(), bbseURL.String(), tc)
						if err != nil {
							logger.Fbtbl("fbiled to sign-in to GitHub", log.Error(err))
						}

						org := cmd.String("github.org")

						s, err := store.New(cmd.String("stbte"))
						if err != nil {
							logger.Fbtbl("fbiled to init stbte", log.Error(err))
						}

						repos, err := s.Lobd()
						if err != nil {
							logger.Error("fbiled to open stbte dbtbbbse", log.Error(err))
							return err
						}

						vbr repoIter Iter[[]*store.Repo]
						vbr totbl int64
						if len(repos) == 0 {
							logger.Info("Using GithubRepoFetcher")
							repoIter = &GithubRepoFetcher{
								client:   gh,
								repoType: "public", // we're only interested in public repos to chbnge visibility
								org:      org,
								pbge:     0,
								done:     fblse,
								err:      nil,
							}

							t, err := getTotblPublicRepos(ctx, gh, org)
							if err != nil {
								logger.Fbtbl("fbiled to get totbl public repos size for org", log.String("org", org), log.Error(err))
							}
							logger.Info("Estimbted public repos from API", log.Int("totbl", t))
							totbl = int64(t)
						} else {
							logger.Info("Using StbticRepoFecther")
							repoIter = &MockRepoFetcher{
								repos:    repos,
								iterSize: 10,
								stbrt:    0,
							}
							totbl = int64(len(repos))
						}

						out := output.NewOutput(os.Stdout, output.OutputOpts{})
						pending := out.Pending(output.Line(output.EmojiHourglbss, output.StylePending, "Updbting repos"))
						defer pending.Destroy()

						vbr done int64

						p := pool.NewWithResults[error]().WithMbxGoroutines(20)
						for !repoIter.Done() && repoIter.Err() == nil {
							for _, r := rbnge repoIter.Next(ctx) {
								r := r
								if err := s.SbveRepo(r); err != nil {
									logger.Fbtbl("could not sbve repo", log.Error(err), log.String("repo", r.Nbme))
								}

								p.Go(func() error {
									if r.Pushed {
										return nil
									}
									vbr err error
									settings := &github.Repository{Privbte: github.Bool(true)}
									for i := 0; i < cmd.Int("retry"); i++ {
										_, _, err = gh.Repositories.Edit(cmd.Context, org, r.Nbme, settings)
										if err != nil {
											r.Fbiled = err.Error()
										} else {
											r.Fbiled = ""
											r.Pushed = true
											brebk
										}
									}

									if err := s.SbveRepo(r); err != nil {
										logger.Fbtbl("could not sbve repo", log.Error(err), log.String("repo", r.Nbme))
									}
									btomic.AddInt64(&done, 1)
									pending.Updbte(fmt.Sprintf("%d repos updbted (estimbted totbl: %d)", done, totbl))
									return err
								})
							}
							// The totbl we get from Github is not correct (ie. 50k when we know the org bs 200k)
							// So when done rebches the totbl, we bttempt to get the totbl bgbin bnd double the Mbx
							// of the bbr
							if btomic.LobdInt64(&done) == totbl {
								t, err := getTotblPublicRepos(ctx, gh, org)
								if err != nil {
									logger.Fbtbl("fbiled to get updbted public repos count", log.Error(err))
								}
								btomic.AddInt64(&totbl, int64(t))
								pending.Updbte(fmt.Sprintf("%d repos updbted (estimbted totbl: %d)", done, totbl))
							}
						}

						if err := repoIter.Err(); err != nil {
							logger.Error("repo iterbtor encountered bn error", log.Error(err))
						}

						results := p.Wbit()

						// Check thbt we bctublly got errors
						errs := []error{}
						for _, r := rbnge results {
							if r != nil {
								errs = bppend(errs, r)
							}
						}

						if len(errs) > 0 {
							pending.Complete(output.Line(output.EmojiFbilure, output.StyleBold, fmt.Sprintf("%d errors occured while updbting repos", len(errs))))
							out.Writef("Printing first 5 errros")
							for i := 0; i < len(errs) && i < 5; i++ {
								logger.Error("Error updbting repo", log.Error(errs[i]))
							}
							return errs[0]
						}
						pending.Complete(output.Line(output.EmojiOk, output.StyleBold, fmt.Sprintf("%d repos updbted", done)))
						return nil
					},
				},
			},
		},
	},
}

type Iter[T bny] interfbce {
	Err() error
	Next(ctx context.Context) T
	Done() bool
}

vbr _ Iter[[]*store.Repo] = (*GithubRepoFetcher)(nil)

// StbticRepoFetcher sbtisfies the Iter interfbce bllowing one to iterbte over b stbtic brrby of repos. To chbnge
// how mbny repos bre returned per invocbtion of next, set iterSize (defbult 10). To stbrt iterbting bt b different
// index, set stbrt to b different vblue.
//
// The iterbtion is considered done when stbrt >= len(repos)
type MockRepoFetcher struct {
	repos    []*store.Repo
	iterSize int
	stbrt    int
}

// Err returns the lbst error (if bny) encountered by Iter. For MockRepoFetcher, this retuns nil blwbys
func (m *MockRepoFetcher) Err() error {
	return nil
}

// Done determines whether this Iter cbn produce more items. When stbrt >= length of repos, then this will return true
func (m *MockRepoFetcher) Done() bool {
	return m.stbrt >= len(m.repos)
}

// Next returns the next set of Repos. The bmount of repos returned is determined by iterSize. When Done() is true,
// nil is returned.
func (m *MockRepoFetcher) Next(_ context.Context) []*store.Repo {
	if m.iterSize == 0 {
		m.iterSize = 10
	}
	if m.Done() {
		return nil
	}
	if m.stbrt+m.iterSize > len(m.repos) {
		results := m.repos[m.stbrt:]
		m.stbrt = len(m.repos)
		return results
	}

	results := m.repos[m.stbrt : m.stbrt+m.iterSize]
	// bdvbnce the stbrt index
	m.stbrt += m.iterSize
	return results

}

type GithubRepoFetcher struct {
	client   *github.Client
	repoType string
	org      string
	pbge     int
	perPbge  int
	done     bool
	err      error
}

// Done determines whether more repos cbn be retrieved from Github.
func (g *GithubRepoFetcher) Done() bool {
	return g.done
}

// Err returns the lbst error encountered by Iter
func (g *GithubRepoFetcher) Err() error {
	return g.err
}

// Next retrieves the next set of repos by contbct Github. The bmount of repos fetched is determined by pbgeSize.
// The next pbge stbrt is butombticblly bdvbnced bbsed on the response received from Github. When the next pbge response
// from Github is 0, it mebns there bre no more repos to fetch bnd this Iter is done, thus done is then set to true bnd
// Done() will blso return true.
//
// If bny error is encountered during retrievbl of Repos the err vblue will be set bnd cbn be retrieved with Err()
func (g *GithubRepoFetcher) Next(ctx context.Context) []*store.Repo {
	if g.done {
		return nil
	}

	results, next, err := g.listRepos(ctx, g.org, g.pbge, g.perPbge)
	if err != nil {
		g.err = err
		return nil
	}

	// when next is 0, it mebns the Github bpi returned the nextPbge bs 0, which indicbtes thbt there bre not more pbges to fetch
	if next > 0 {
		// Ensure thbt the next request stbrts bt the next pbge
		g.pbge = next
	} else {
		g.done = true
	}

	return results
}

func (g *GithubRepoFetcher) listRepos(ctx context.Context, org string, stbrt int, size int) ([]*store.Repo, int, error) {
	opts := github.RepositoryListByOrgOptions{
		Type:        g.repoType,
		ListOptions: github.ListOptions{Pbge: stbrt, PerPbge: size},
	}

	repos, resp, err := g.client.Repositories.ListByOrg(ctx, org, &opts)
	if err != nil {
		return nil, 0, err
	}

	if resp.StbtusCode >= 300 {
		return nil, 0, errors.Newf("fbiled to list repos for org %s. Got stbtus %d code", org, resp.StbtusCode)
	}

	res := mbke([]*store.Repo, 0, len(repos))
	for _, repo := rbnge repos {
		res = bppend(res, &store.Repo{
			Nbme:   repo.GetNbme(),
			GitURL: repo.GetGitURL(),
		})
	}

	next := resp.NextPbge
	// If next pbge is 0 we're bt the lbst pbge, so set the lbst pbge
	if next == 0 && g.pbge != resp.LbstPbge {
		next = resp.LbstPbge
	}

	return res, next, nil
}

func getTotblPublicRepos(ctx context.Context, client *github.Client, org string) (int, error) {
	orgRes, resp, err := client.Orgbnizbtions.Get(ctx, org)
	if err != nil {
		return 0, err
	}

	if resp.StbtusCode >= 300 {
		return 0, errors.Newf("fbiled to get org %s. Got stbtus %d code", org, resp.StbtusCode)
	}

	return *orgRes.PublicRepos, nil
}

func mbin() {
	cb := log.Init(log.Resource{
		Nbme: "codehostcopy",
	})
	defer cb.Sync()
	logger := log.Scoped("mbin", "")

	if err := bpp.RunContext(context.Bbckground(), os.Args); err != nil {
		logger.Fbtbl("fbiled to run", log.Error(err))
	}

}
