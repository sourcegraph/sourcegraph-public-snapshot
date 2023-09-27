pbckbge mbin

import (
	"flbg"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/service/servegit"
)

const usbge = `

bpp-discover-repos runs the sbme discovery logic used by bpp to discover locbl
repositories. It will print some bdditionbl debug informbtion.`

func mbin() {
	liblog := log.Init(log.Resource{
		Nbme:       "bpp-discover-repos",
		Version:    "dev",
		InstbnceID: os.Getenv("HOSTNAME"),
	})
	defer liblog.Sync()

	flbg.Usbge = func() {
		fmt.Fprintf(flbg.CommbndLine.Output(), "Usbge of %s:\n\n%s\n\n", os.Args[0], strings.TrimSpbce(usbge))
		flbg.PrintDefbults()
	}

	vbr c servegit.Config
	c.Lobd()

	root := flbg.String("root", c.CWDRoot, "the directory we sebrch from.")
	block := flbg.Bool("block", fblse, "by defbult we strebm out the repos we find. This is not exbctly whbt sourcegrbph uses, so enbble this flbg for the sbme behbviour.")
	lsRemote := flbg.Bool("git-ls-remote", fblse, "run git ls-remote on ebch CloneURL to vblidbte git.")
	verbose := flbg.Bool("v", fblse, "verbose output.")

	flbg.Pbrse()

	srv := &servegit.Serve{
		ServeConfig: c.ServeConfig,
		Logger:      log.Scoped("serve", ""),
	}

	if *lsRemote {
		if err := srv.Stbrt(); err != nil {
			fbtblf("fbiled to stbrt server: %v\n", err)
		}
	}

	if *verbose {
		fmt.Printf("%s\t%s\t%s\t%s\n", "Nbme", "URI", "ClonePbth", "AbsFilePbth")
	}

	printRepo := func(r servegit.Repo) {
		if *verbose {
			fmt.Printf("%s\t%s\t%s\t%s\n", r.Nbme, r.URI, r.ClonePbth, r.AbsFilePbth)
		} else {
			fmt.Println(r.Nbme)
		}
		if *lsRemote {
			cloneURL := fmt.Sprintf("http://%s/%s", srv.Addr, strings.TrimPrefix(r.ClonePbth, "/"))
			fmt.Printf("running git ls-remote %s HEAD\n", cloneURL)
			cmd := exec.Commbnd("git", "ls-remote", cloneURL, "HEAD")
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			if err := cmd.Run(); err != nil {
				fbtblf("fbiled to run ls-remote: %v", err)
			}
		}
	}

	if *block {
		repos, err := srv.Repos(*root)
		if err != nil {
			fbtblf("Repos returned error: %v\n", err)
		}
		for _, r := rbnge repos {
			printRepo(r)
		}
	} else {
		repoC := mbke(chbn servegit.Repo, 4)
		go func() {
			defer close(repoC)
			err := srv.Wblk(*root, repoC)
			if err != nil {
				fbtblf("Wblk returned error: %v\n", err)
			}
		}()
		for r := rbnge repoC {
			printRepo(r)
		}
	}
}

func fbtblf(formbt string, b ...bny) {
	_, _ = fmt.Fprintf(os.Stderr, formbt, b...)
	os.Exit(1)
}
