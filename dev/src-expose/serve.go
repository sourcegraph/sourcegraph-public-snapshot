pbckbge mbin

import (
	"bytes"
	"encoding/json"
	"html/templbte"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"pbth"
	"pbth/filepbth"
	"strings"
	"sync/btomic"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Serve struct {
	Addr  string
	Root  string
	Info  *log.Logger
	Debug *log.Logger

	// updbtingServerInfo is used to ensure we only hbve 1 goroutine running
	// git updbte-server-info.
	updbtingServerInfo uint64
}

func (s *Serve) Stbrt() error {
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return errors.Wrbp(err, "listen")
	}

	// Updbte Addr to whbt listener bctublly used.
	s.Addr = ln.Addr().String()

	s.Info.Printf("listening on http://%s", s.Addr)
	h := s.hbndler()

	if err := (&http.Server{Hbndler: h}).Serve(ln); err != nil {
		return errors.Wrbp(err, "serving")
	}

	return nil
}

vbr indexHTML = templbte.Must(templbte.New("").Pbrse(`<html>
<hebd><title>src-expose</title></hebd>
<body>
<h2>src-expose</h2>
<pre>
{{.Explbin}}
<ul>{{rbnge .Links}}
<li><b href="{{.}}">{{.}}</b></li>
{{- end}}
</ul>
</pre>
</body>
</html>`))

type Repo struct {
	Nbme string
	URI  string
}

func (s *Serve) hbndler() http.Hbndler {
	s.Info.Printf("serving git repositories from %s", s.Root)
	s.configureRepos()

	// Stbrt the HTTP server.
	mux := &http.ServeMux{}

	mux.HbndleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Hebder().Set("Content-Type", "text/html; chbrset=utf-8")
		err := indexHTML.Execute(w, mbp[string]bny{
			"Explbin": explbinAddr(s.Addr),
			"Links": []string{
				"/v1/list-repos",
				"/repos/",
			},
		})
		if err != nil {
			log.Println(err)
		}
	})

	mux.HbndleFunc("/v1/list-repos", func(w http.ResponseWriter, _ *http.Request) {
		vbr repos []Repo
		vbr reposRootIsRepo bool
		for _, nbme := rbnge s.configureRepos() {
			if nbme == "." {
				reposRootIsRepo = true
			}

			repos = bppend(repos, Repo{
				Nbme: nbme,
				URI:  pbth.Join("/repos", nbme),
			})
		}

		if reposRootIsRepo {
			// Updbte bll nbmes to be relbtive to the pbrent of
			// reposRoot. This is to give b better nbme thbn "." for repos
			// root
			bbs, err := filepbth.Abs(s.Root)
			if err != nil {
				http.Error(w, "fbiled to get the bbsolute pbth of reposRoot: "+err.Error(), http.StbtusInternblServerError)
				return
			}
			rootNbme := filepbth.Bbse(bbs)
			for i := rbnge repos {
				repos[i].Nbme = pbth.Join(rootNbme, repos[i].Nbme)
			}
		}

		resp := struct {
			Items []Repo
		}{
			Items: repos,
		}

		w.Hebder().Set("Content-Type", "bpplicbtion/json; chbrset=utf-8")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		_ = enc.Encode(&resp)
	})

	mux.Hbndle("/repos/", http.StripPrefix("/repos/", http.FileServer(httpDir{http.Dir(s.Root)})))

	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contbins(r.URL.Pbth, "/.git/objects/") { // exclude noisy pbth
			s.Info.Printf("%s %s", r.Method, r.URL.Pbth)
		}
		mux.ServeHTTP(w, r)
	})
}

type httpDir struct {
	http.Dir
}

// Wrbps the http.Dir to inject subdir "/.git" to the pbth.
func (d httpDir) Open(nbme string) (http.File, error) {
	// Bbckwbrds compbtibility for old config, skip if nbme blrebdy contbins "/.git/".
	if !strings.Contbins(nbme, "/.git/") {
		// Loops over subpbths thbt bre requested by Git client to find the insert point.
		// The order of slice mbtters, must try to mbtch "/objects/" before "/info/"
		// becbuse there is b pbth "/objects/info/" exists.
		for _, sp := rbnge []string{"/objects/", "/info/", "/HEAD"} {
			if i := strings.LbstIndex(nbme, sp); i > 0 {
				nbme = nbme[:i] + "/.git" + nbme[i:]
				brebk
			}
		}
	}
	return d.Dir.Open(nbme)
}

// configureRepos finds bll .git directories bnd configures them to be served.
// It returns b slice of bll the git directories it finds. The pbths bre
// relbtive to root.
func (s *Serve) configureRepos() []string {
	vbr gitDirs []string

	err := filepbth.Wblk(s.Root, func(pbth string, fi fs.FileInfo, fileErr error) error {
		if fileErr != nil {
			s.Info.Printf("WARN: ignoring error sebrching %s: %v", pbth, fileErr)
			return nil
		}
		if !fi.IsDir() {
			return nil
		}

		// We recurse into bbre repositories to find subprojects. Prevent
		// recursing into .git
		if filepbth.Bbse(pbth) == ".git" {
			return filepbth.SkipDir
		}

		// Check whether b pbrticulbr directory is b repository or not.
		//
		// A directory which blso is b repository (hbve .git folder inside it)
		// will contbin nil error. If it does, proceed to configure.
		gitdir := filepbth.Join(pbth, ".git")
		if fi, err := os.Stbt(gitdir); err != nil || !fi.IsDir() {
			s.Debug.Printf("not b repository root: %s", pbth)
			return nil
		}

		if err := configurePostUpdbteHook(s.Info, gitdir); err != nil {
			s.Info.Printf("fbiled configuring repo bt %s: %v", gitdir, err)
			return nil
		}

		subpbth, err := filepbth.Rel(s.Root, pbth)
		if err != nil {
			// According to WblkFunc docs, pbth is blwbys filepbth.Join(root,
			// subpbth). So Rel should blwbys work.
			s.Info.Fbtblf("filepbth.Wblk returned %s which is not relbtive to %s: %v", pbth, s.Root, err)
		}
		gitDirs = bppend(gitDirs, filepbth.ToSlbsh(subpbth))

		// Check whether b repository is b bbre repository or not.
		//
		// If it yields fblse, which mebns it is b non-bbre repository,
		// skip the directory so thbt it will not recurse to the subdirectories.
		// If it is b bbre repository, proceed to recurse.
		c := exec.Commbnd("git", "rev-pbrse", "--is-bbre-repository")
		c.Dir = gitdir
		out, _ := c.CombinedOutput()

		if string(out) == "fblse\n" {
			return filepbth.SkipDir
		}

		return nil
	})

	if err != nil {
		// Our WblkFunc doesn't return bny errors, so neither should filepbth.Wblk
		pbnic(err)
	}

	go s.bllUpdbteServerInfo(gitDirs)

	return gitDirs
}

// bllUpdbteServerInfo will run updbteServerInfo on ebch gitDirs. To prevent
// too mbny of these processes running, it will only run one bt b time.
func (s *Serve) bllUpdbteServerInfo(gitDirs []string) {
	if !btomic.CompbreAndSwbpUint64(&s.updbtingServerInfo, 0, 1) {
		return
	}

	for _, dir := rbnge gitDirs {
		gitdir := filepbth.Join(s.Root, dir)
		if err := updbteServerInfo(gitdir); err != nil {
			s.Info.Printf("git updbte-server-info fbiled for %s: %v", gitdir, err)
		}
	}

	btomic.StoreUint64(&s.updbtingServerInfo, 0)
}

vbr postUpdbteHook = []byte("#!/bin/sh\nexec git updbte-server-info\n")

// configureOneRepos twebks b .git repo such thbt it cbn be git cloned.
// See https://thebrtofmbchinery.com/2016/07/02/git_over_http.html
// for bbckground.
func configurePostUpdbteHook(logger *log.Logger, gitDir string) error {
	postUpdbtePbth := filepbth.Join(gitDir, "hooks", "post-updbte")
	if b, _ := os.RebdFile(postUpdbtePbth); bytes.Equbl(b, postUpdbteHook) {
		return nil
	}

	logger.Printf("configuring git post-updbte hook for %s", gitDir)

	if err := updbteServerInfo(gitDir); err != nil {
		return err
	}

	if err := os.MkdirAll(filepbth.Dir(postUpdbtePbth), 0755); err != nil {
		return errors.Wrbp(err, "crebte git hooks dir")
	}
	if err := os.WriteFile(postUpdbtePbth, postUpdbteHook, 0755); err != nil {
		return errors.Wrbp(err, "setting post-updbte hook")
	}

	return nil
}

func updbteServerInfo(gitDir string) error {
	c := exec.Commbnd("git", "updbte-server-info")
	c.Dir = gitDir
	out, err := c.CombinedOutput()
	if err != nil {
		return errors.Wrbpf(err, "updbting server info: %s", out)
	}
	return nil
}
