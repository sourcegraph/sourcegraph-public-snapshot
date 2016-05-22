package cli

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/rogpeppe/rog-go/parallel"

	"golang.org/x/net/context"

	"strings"
	"sync"

	"sourcegraph.com/sqs/pbtypes"

	"sourcegraph.com/sourcegraph/go-flags"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/util/expvarutil"
)

func initStressTest(g *flags.Command) {
	_, err := g.AddCommand("stress",
		"run stress-test of server",
		"The stress subcommand performs a configurable stress-test on the server. If no specific stress tests are specified, all stress tests are performed.",
		&stressCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type stressCmd struct {
	Git bool `long:"git" description:"perform git ops (push, pull, clone) stress test"`

	RepoPage bool `long:"repo-page" description:"perform stress test of fetching app repo page"`
	FilePage bool `long:"file-page" description:"perform stress test of fetching app code file page"`

	Repo  string              `long:"repo" description:"(repo-page, file-page) URI of repo to fetch (default: fetch for all repos)"`
	repos []*sourcegraph.Repo // fetched based on Repo field

	File  string            `long:"file" description:"(file-page) filename of file to fetch (default: arbitrarily chosen file)"`
	files map[string]string // fetched based on File field; map of repo URI to the file for that repo to fetch

	N int `short:"n" description:"number of times to repeat stress tests" default:"1"`
	P int `short:"p" description:"number of parallel stress tests to perform" default:"1"`

	Log bool `short:"v" long:"log" description:"print log output during stress tests"`

	site   *sourcegraph.ServerConfig
	appURL *url.URL // parsed site.AppURL

	expvar *expvarutil.Client
}

func (c *stressCmd) quiet() bool { return !c.Log }

func (c *stressCmd) Execute(args []string) error {
	ctx := cliContext
	cl := cliClient

	allOps := !(c.Git || c.RepoPage || c.FilePage)

	var err error
	c.site, err = cl.Meta.Config(ctx, &pbtypes.Void{})
	if err != nil {
		return err
	}

	c.appURL, err = url.Parse(c.site.AppURL)
	if err != nil {
		return err
	}

	log.Printf("## Stress-testing server at URL %s", c.site.AppURL)
	log.Println()

	host, _, _ := net.SplitHostPort(c.appURL.Host)
	if host == "" {
		host = c.appURL.Host
	}
	c.expvar = expvarutil.NewClient((&url.URL{Scheme: c.appURL.Scheme, Host: host + ":6060", Path: "/debug/vars"}).String())

	if c.RepoPage || c.FilePage || allOps {
		if err := c.fetchRepos(ctx); err != nil {
			return err
		}
	}
	if c.FilePage || allOps {
		if err := c.fetchFiles(ctx); err != nil {
			return err
		}
	}

	// Pre memstats
	if err := c.expvar.GC(); err != nil {
		return err
	}
	var pre runtime.MemStats
	if err := c.expvar.Get("memstats", &pre); err != nil {
		return err
	}

	start := time.Now()
	var ndoneMu sync.Mutex
	ndone := 0
	done := func() {
		ndoneMu.Lock()
		defer ndoneMu.Unlock()
		ndone++
		if c.quiet() {
			const width = 60
			fmt.Printf("\r%d/%d\t|%- *s| %.1f%%",
				ndone, c.N,
				width, strings.Repeat("=", int(float64(ndone)/float64(c.N)*width)),
				100*float64(ndone)/float64(c.N),
			)
		}
	}
	if c.quiet() {
		log.SetOutput(ioutil.Discard)
	}
	par := parallel.NewRun(c.P)
	const sep = "----------------------------------------------------------------"
	for ii := 0; ii < c.N; ii++ {
		i := ii
		par.Do(func() error {
			if i == 0 {
				log.Println(sep)
			}
			if c.N > 1 {
				log.Printf("## iteration %d/%d", i+1, c.N)
			}

			if c.Git || allOps {
				log.Println("# Git stress test")
				results, err := c.runGitOps(ctx)
				if err != nil {
					return err
				}
				if results != "" {
					fmt.Println(results)
				}
			}

			if c.RepoPage || allOps {
				log.Println("# Repo page stress test")
				results, err := c.runRepoPage(ctx)
				if err != nil {
					return err
				}
				if results != "" {
					fmt.Println(results)
				}
			}

			if c.FilePage || allOps {
				log.Println("# File page stress test")
				results, err := c.runFilePage(ctx)
				if err != nil {
					return err
				}
				if results != "" {
					fmt.Println(results)
				}
			}

			log.Println(sep)

			done()

			return nil
		})
	}
	if err := par.Wait(); err != nil {
		return err
	}
	if c.quiet() {
		fmt.Println()
		log.SetOutput(os.Stderr)
	}

	log.Println()

	elapsed := time.Since(start)
	log.Printf("# Time: %s (%s per iteration)", elapsed, elapsed/time.Duration(c.N))
	log.Println("#")

	// Post memstats
	if err := c.expvar.GC(); err != nil {
		return err
	}
	var post runtime.MemStats
	if err := c.expvar.Get("memstats", &post); err != nil {
		return err
	}
	log.Println("# Memory statistics")
	log.Printf("Allocated: %s", humanizeByteDelta(pre.Alloc, post.Alloc, c.N))

	return nil
}

func humanizeByteDelta(pre, post uint64, iters int) string {
	var min, max uint64
	var sign string
	if post > pre {
		min = pre
		max = post
	} else {
		min = post
		max = pre
		sign = "-"
	}

	Δ := max - min
	s := fmt.Sprintf("Δ %s%s (%s → %s)", sign, humanize.Bytes(Δ), humanize.Bytes(pre), humanize.Bytes(post))
	if iters > 1 {
		s += fmt.Sprintf(" %s per iteration", humanize.Bytes(Δ/uint64(iters)))
	}
	return s
}

func (c *stressCmd) fetchRepos(ctx context.Context) error {
	cl := cliClient
	if c.Repo == "" {
		allRepos, err := cl.Repos.List(ctx, &sourcegraph.RepoListOptions{
			ListOptions: sourcegraph.ListOptions{PerPage: 20},
		})
		if err != nil {
			return err
		}
		c.repos = allRepos.Repos
	} else {
		repo, err := cl.Repos.Get(ctx, &sourcegraph.RepoSpec{URI: c.Repo})
		if err != nil {
			return err
		}
		c.repos = []*sourcegraph.Repo{repo}
	}
	return nil
}

func (c *stressCmd) fetchFiles(ctx context.Context) error {
	cl := cliClient
	c.files = make(map[string]string, len(c.repos))
	if c.File == "" {
		for _, repo := range c.repos {
			res, err := cl.Repos.ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{
				Repo: repo.RepoSpec(),
				Rev:  repo.DefaultBranch,
			})
			if err != nil {
				return err
			}

			tree, err := cl.RepoTree.Get(ctx, &sourcegraph.RepoTreeGetOp{
				Entry: sourcegraph.TreeEntrySpec{
					RepoRev: sourcegraph.RepoRevSpec{RepoSpec: repo.RepoSpec(), CommitID: res.CommitID},
					Path:    ".",
				},
			})
			if err != nil {
				return err
			}

			// Find the first file (not dir).
			for _, e := range tree.Entries {
				if e.Type == sourcegraph.FileEntry {
					c.files[repo.URI] = e.Name
					break
				}
			}
			if _, found := c.files[repo.URI]; !found {
				return fmt.Errorf("repo %s has no files for file-page stress test", repo.URI)
			}
		}
	} else {
		for _, repo := range c.repos {
			c.files[repo.URI] = c.File
		}
	}
	return nil
}

func (c *stressCmd) runGitOps(ctx context.Context) (string, error) {
	return "", nil
}

func (c *stressCmd) runRepoPage(ctx context.Context) (string, error) {
	for _, repo := range c.repos {
		t0 := time.Now()
		err := func() error {
			repoPageURL := c.appURL.ResolveReference(router.Rel.URLToRepo(repo.URI)).String()
			resp, err := http.Get(repoPageURL)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if want := http.StatusOK; resp.StatusCode != want {
				return fmt.Errorf("non-200 HTTP response (%d): %s", resp.StatusCode, repoPageURL)
			}
			return nil
		}()
		if err != nil {
			return "", err
		}
		log.Printf("fetched %s repo page in %s", repo.URI, time.Since(t0))
	}
	return "", nil
}

func (c *stressCmd) runFilePage(ctx context.Context) (string, error) {
	for _, repo := range c.repos {
		t0 := time.Now()
		err := func() error {
			filePageURL := c.appURL.ResolveReference(router.Rel.URLToRepoTreeEntry(repo.URI, "", c.files[repo.URI])).String()
			resp, err := http.Get(filePageURL)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if want := http.StatusOK; resp.StatusCode != want {
				return fmt.Errorf("non-200 HTTP response (%d): %s", resp.StatusCode, filePageURL)
			}
			return nil
		}()
		if err != nil {
			return "", err
		}
		log.Printf("fetched %s file %s page in %s", repo.URI, c.files[repo.URI], time.Since(t0))
	}
	return "", nil
}
