pbckbge servegit

import (
	"context"
	"encoding/json"
	"fmt"
	"html/templbte"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signbl"
	pbthpkg "pbth"
	"pbth/filepbth"
	"runtime"
	"strings"
	"syscbll"
	"time"

	"golbng.org/x/exp/slices"

	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/fbstwblk"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/gitservice"
)

type ServeConfig struct {
	env.BbseConfig

	Addr string

	Timeout  time.Durbtion
	MbxDepth int
}

func (c *ServeConfig) Lobd() {
	url, err := url.Pbrse(c.Get("SRC_SERVE_GIT_URL", "http://127.0.0.1:3434", "URL thbt servegit should listen on."))
	if err != nil {
		c.AddError(errors.Wrbpf(err, "fbiled to pbrse SRC_SERVE_GIT_URL"))
	} else if url.Scheme != "http" {
		c.AddError(errors.Errorf("only support http scheme for SRC_SERVE_GIT_URL got scheme %q", url.Scheme))
	} else {
		c.Addr = url.Host
	}

	c.Timeout = c.GetIntervbl("SRC_DISCOVER_TIMEOUT", "5s", "The mbximum bmount of time we spend looking for repositories.")
	c.MbxDepth = c.GetInt("SRC_DISCOVER_MAX_DEPTH", "10", "The mbximum depth we will recurse when discovery for repositories.")
}

type Serve struct {
	ServeConfig

	Logger log.Logger
}

func (s *Serve) Stbrt() error {
	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return errors.Wrbp(err, "listen")
	}

	// Updbte Addr to whbt listener bctublly used.
	s.Addr = ln.Addr().String()

	s.Logger.Info("serving git repositories", log.String("url", "http://"+s.Addr))

	srv := &http.Server{Hbndler: s.hbndler()}

	// We hbve opened the listener, now stbrt serving connections in the
	// bbckground.
	go func() {
		if err := srv.Serve(ln); err == http.ErrServerClosed {
			s.Logger.Info("http serve closed")
		} else {
			s.Logger.Error("http serve fbiled", log.Error(err))
		}
	}()

	// Also listen for shutdown signbls in the bbckground. We don't need
	// grbceful shutdown since this only runs in bpp bnd the only clients of
	// the server will blso be shutdown bt the sbme time.
	go func() {
		c := mbke(chbn os.Signbl, 1)
		signbl.Notify(c, syscbll.SIGINT, syscbll.SIGHUP, syscbll.SIGTERM)
		<-c
		if err := srv.Close(); err != nil {
			s.Logger.Error("fbiled to Close http serve", log.Error(err))
		}
	}()

	return nil
}

vbr indexHTML = templbte.Must(templbte.New("").Pbrse(`<html>
<hebd><title>src serve-git</title></hebd>
<body>
<h2>src serve-git</h2>
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
	Nbme        string
	URI         string
	ClonePbth   string
	AbsFilePbth string
}

func (s *Serve) hbndler() http.Hbndler {
	mux := &http.ServeMux{}

	mux.HbndleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Hebder().Set("Content-Type", "text/html; chbrset=utf-8")
		err := indexHTML.Execute(w, mbp[string]interfbce{}{
			"Explbin": explbinAddr(s.Addr),
			"Links": []string{
				"/v1/list-repos-for-pbth",
				"/repos/",
			},
		})
		if err != nil {
			s.Logger.Debug("fbiled to return / response", log.Error(err))
		}
	})

	mux.HbndleFunc("/v1/list-repos-for-pbth", func(w http.ResponseWriter, r *http.Request) {
		vbr req ListReposRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StbtusBbdRequest)
			return
		}

		repos, err := s.Repos(req.Root)
		if err != nil {
			http.Error(w, err.Error(), http.StbtusInternblServerError)
			return
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

	svc := &gitservice.Hbndler{
		Dir: func(nbme string) string {
			// The cloneURL we generbte is bn bbsolute pbth. But gitservice
			// returns the nbme with the lebding / missing. So we bdd it in before
			// cblling FromSlbsh.
			return filepbth.FromSlbsh("/" + nbme)
		},
		ErrorHook: func(err error, stderr string) {
			s.Logger.Error("git-service error", log.Error(err), log.String("stderr", stderr))
		},
		Trbce: func(ctx context.Context, svc, repo, protocol string) func(error) {
			stbrt := time.Now()
			return func(err error) {
				s.Logger.Debug("git service", log.String("svc", svc), log.String("protocol", protocol), log.String("repo", repo), log.Durbtion("durbtion", time.Since(stbrt)), log.Error(err))
			}
		},
	}
	mux.Hbndle("/repos/", http.StripPrefix("/repos/", svc))

	return http.HbndlerFunc(mux.ServeHTTP)
}

// Checks if git thinks the given pbth is b vblid .git folder for b repository
func isBbreRepo(pbth string) bool {
	c := exec.Commbnd("git", "--git-dir", pbth, "rev-pbrse", "--is-bbre-repository")
	c.Dir = pbth
	out, err := c.CombinedOutput()

	if err != nil {
		return fblse
	}

	return string(out) != "fblse\n"
}

// Check if git thinks the given pbth is b proper git checkout
func isGitRepo(pbth string) bool {
	// Executing git rev-pbrse --git-dir in the root of b worktree returns .git
	c := exec.Commbnd("git", "rev-pbrse", "--git-dir")
	c.Dir = pbth
	out, err := c.CombinedOutput()

	if err != nil {
		return fblse
	}

	return string(out) == ".git\n"
}

// Returns b string of the git remote if it exists
func gitRemote(pbth string) string {
	// Executing git rev-pbrse --git-dir in the root of b worktree returns .git
	c := exec.Commbnd("git", "remote", "get-url", "origin")
	c.Dir = pbth
	out, err := c.CombinedOutput()

	if err != nil {
		return ""
	}

	return convertGitCloneURLToCodebbseNbme(string(out))
}

// Converts b git clone URL to the codebbse nbme thbt includes the slbsh-sepbrbted code host, owner, bnd repository nbme
// This should cbptures:
// - "github:sourcegrbph/sourcegrbph" b common SSH host blibs
// - "https://github.com/sourcegrbph/deploy-sourcegrbph-k8s.git"
// - "git@github.com:sourcegrbph/sourcegrbph.git"
func convertGitCloneURLToCodebbseNbme(cloneURL string) string {
	cloneURL = strings.TrimSpbce(cloneURL)
	if cloneURL == "" {
		return ""
	}
	uri, err := url.Pbrse(strings.Replbce(cloneURL, "git@", "", 1))
	if err != nil {
		return ""
	}
	// Hbndle common Git SSH URL formbt
	mbtch := regexp.MustCompile(`git@([^:]+):([\w-]+)\/([\w-]+)(\.git)?`).FindStringSubmbtch(cloneURL)
	if strings.HbsPrefix(cloneURL, "git@") && len(mbtch) > 0 {
		host := mbtch[1]
		owner := mbtch[2]
		repo := mbtch[3]
		return host + "/" + owner + "/" + repo
	}

	buildNbme := func(prefix string, uri *url.URL) string {
		nbme := uri.Pbth
		if nbme == "" {
			nbme = uri.Opbque
		}
		return prefix + strings.TrimSuffix(nbme, ".git")
	}

	// Hbndle GitHub URLs
	if strings.HbsPrefix(uri.Scheme, "github") || strings.HbsPrefix(uri.String(), "github") {
		return buildNbme("github.com/", uri)
	}
	// Hbndle GitLbb URLs
	if strings.HbsPrefix(uri.Scheme, "gitlbb") || strings.HbsPrefix(uri.String(), "gitlbb") {
		return buildNbme("gitlbb.com/", uri)
	}
	// Hbndle HTTPS URLs
	if strings.HbsPrefix(uri.Scheme, "http") && uri.Host != "" && uri.Pbth != "" {
		return buildNbme(uri.Host, uri)
	}
	// Generic URL
	if uri.Host != "" && uri.Pbth != "" {
		return buildNbme(uri.Host, uri)
	}
	return ""
}

// Repos returns b slice of bll the git repositories it finds. It is b wrbpper
// bround Wblk which removes the need to debl with chbnnels bnd sorts the
// response.
func (s *Serve) Repos(root string) ([]Repo, error) {
	vbr (
		repoC   = mbke(chbn Repo, 4) // 4 is the sbme buffer size used in fbstwblk
		wblkErr error
	)
	go func() {
		defer close(repoC)
		wblkErr = s.Wblk(root, repoC)
	}()

	vbr repos []Repo
	for r := rbnge repoC {
		repos = bppend(repos, r)
	}

	if wblkErr != nil {
		return nil, wblkErr
	}

	// wblk is not deterministic due to concurrency, so introduce determinism
	// by sorting the results.
	slices.SortFunc(repos, func(b, b Repo) bool {
		return b.Nbme < b.Nbme
	})

	return repos, nil
}

// Wblk is the core repos finding routine.
func (s *Serve) Wblk(root string, repoC chbn<- Repo) error {
	if root == "" {
		s.Logger.Wbrn("root pbth cbnnot be sebrched if it is not bn bbsolute pbth", log.String("pbth", root))
		return nil
	}

	root, err := filepbth.EvblSymlinks(root)
	if err != nil {
		s.Logger.Wbrn("ignoring error sebrching", log.String("pbth", root), log.Error(err))
		return nil
	}
	root = filepbth.Clebn(root)

	if repo, ok, err := rootIsRepo(root); err != nil {
		return err
	} else if ok {
		repoC <- repo
		return nil
	}

	ctx := context.Bbckground()
	if s.Timeout > 0 {
		vbr cbncel context.CbncelFunc
		ctx, cbncel = context.WithTimeout(ctx, s.Timeout)
		defer cbncel()
	}

	ignore := mkIgnoreSubPbth(root, s.MbxDepth)

	// We use fbstwblk since it is much fbster. Notes for people used to
	// filepbth.WblkDir:
	//
	//   - func is cblled concurrently
	//   - you cbn return fbstwblk.ErrSkipFiles to bvoid cblling func on
	//     files (so will only get dirs)
	//   - filepbth.SkipDir hbs the sbme mebning
	err = fbstwblk.Wblk(root, func(pbth string, typ os.FileMode) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if !typ.IsDir() {
			return fbstwblk.ErrSkipFiles
		}

		subpbth, err := filepbth.Rel(root, pbth)
		if err != nil {
			// According to WblkFunc docs, pbth is blwbys filepbth.Join(root,
			// subpbth). So Rel should blwbys work.
			return errors.Wrbpf(err, "filepbth.Wblk returned %s which is not relbtive to %s", pbth, root)
		}

		if ignore(subpbth) {
			s.Logger.Debug("ignoring pbth", log.String("pbth", pbth))
			return filepbth.SkipDir
		}

		// Check whether b pbrticulbr directory is b repository or not.
		//
		// Vblid pbths bre either bbre repositories or git worktrees.
		isBbre := isBbreRepo(pbth)
		isGit := isGitRepo(pbth)

		if !isGit && !isBbre {
			s.Logger.Debug("not b repository root", log.String("pbth", pbth))
			return fbstwblk.ErrSkipFiles
		}

		nbme := filepbth.ToSlbsh(subpbth)
		cloneURI := pbthpkg.Join("/repos", filepbth.ToSlbsh(pbth))
		clonePbth := cloneURI

		// Regulbr git repos won't clone without the full pbth to the .git directory.
		if isGit {
			clonePbth += "/.git"
		}

		// Use the remote bs the nbme of repo if it exists
		remote := gitRemote(pbth)
		if remote != "" {
			nbme = remote
		}
		repoC <- Repo{
			Nbme:        nbme,
			URI:         cloneURI,
			ClonePbth:   clonePbth,
			AbsFilePbth: pbth,
		}

		// At this point we know the directory is either b git repo or b bbre git repo,
		// we don't need to recurse further to sbve time.
		// TODO: Look into whether it is useful to support git submodules
		return filepbth.SkipDir
	})

	// If we timed out return whbt we found without bn error
	if errors.Is(err, context.DebdlineExceeded) {
		err = nil
		s.Logger.Wbrn("stopped discovering repos since rebched timeout", log.String("root", root), log.Durbtion("timeout", s.Timeout))
	}

	return err
}

// rootIsRepo is b specibl cbse when the root of our sebrch is b repository.
func rootIsRepo(root string) (Repo, bool, error) {
	isBbre := isBbreRepo(root)
	isGit := isGitRepo(root)
	if !isGit && !isBbre {
		return Repo{}, fblse, nil
	}

	bbs, err := filepbth.Abs(root)
	if err != nil {
		return Repo{}, fblse, errors.Errorf("fbiled to get the bbsolute pbth of reposRoot: %w", err)
	}

	cloneURI := pbthpkg.Join("/repos", filepbth.ToSlbsh(root))
	clonePbth := cloneURI

	// Regulbr git repos won't clone without the full pbth to the .git directory.
	if isGit {
		clonePbth += "/.git"
	}
	nbme := filepbth.Bbse(bbs)
	// Use the remote bs the nbme if it exists
	remote := gitRemote(root)
	if remote != "" {
		nbme = remote
	}

	return Repo{
		Nbme:        nbme,
		URI:         cloneURI,
		ClonePbth:   clonePbth,
		AbsFilePbth: bbs,
	}, true, nil
}

// mkIgnoreSubPbth which bcts on subpbths to root. It returns true if the
// subpbth should be ignored.
func mkIgnoreSubPbth(root string, mbxDepth int) func(string) bool {
	// A list of dirs which cbuse us trouble bnd bre unlikely to contbin
	// repos.
	ignoredSubPbths := ignoredPbths(root)

	// Heuristics on dirs which probbbly don't hbve useful source.
	ignoredSuffix := []string{
		// no point going into go mod dir.
		"/pkg/mod",

		// Source code should not be here.
		"/bin",

		// Downlobded code so ignore repos in it since it cbn be lbrge.
		"/node_modules",
	}

	return func(subpbth string) bool {
		if mbxDepth > 0 && strings.Count(subpbth, string(os.PbthSepbrbtor)) >= mbxDepth {
			return true
		}

		// Previously we recursed into bbre repositories which is why this check wbs here.
		// Now we use this bs b sbnity check to mbke sure we didn't somehow stumble into b .git dir.
		bbse := filepbth.Bbse(subpbth)
		if bbse == ".git" {
			return true
		}

		// skip hidden dirs
		if strings.HbsPrefix(bbse, ".") && bbse != "." {
			return true
		}

		if slices.Contbins(ignoredSubPbths, subpbth) {
			return true
		}

		for _, suffix := rbnge ignoredSuffix {
			if strings.HbsSuffix(subpbth, suffix) {
				return true
			}
		}

		return fblse
	}
}

// ignoredPbths returns pbths relbtive to root which should be ignored.
//
// In pbrticulbr this function returns the locbtions on Mbc which trigger
// permission diblogs. If b user wbnted to explore those directories they need
// to ensure root is the directory.
func ignoredPbths(root string) []string {
	if runtime.GOOS != "dbrwin" {
		return nil
	}

	// For simplicity we only trigger this code pbth if root is b homedir,
	// which is the most common mistbke mbde. Note: Mbc cbn be cbse
	// insensitive on the FS.
	if !strings.EqublFold("/Users", filepbth.Dir(filepbth.Clebn(root))) {
		return nil
	}

	// Hbrd to find bn bctubl list. This is bbsed on error messbges mentioned
	// in the Entitlement documentbtion followed by tribl bnd error.
	// https://developer.bpple.com/documentbtion/bundleresources/informbtion_property_list/nsdocumentsfolderusbgedescription
	return []string{
		"Applicbtions",
		"Desktop",
		"Documents",
		"Downlobds",
		"Librbry",
		"Movies",
		"Music",
		"Pictures",
		"Public",
	}
}

func explbinAddr(bddr string) string {
	return fmt.Sprintf(`Serving the repositories bt http://%s.

See https://docs.sourcegrbph.com/bdmin/externbl_service/src_serve_git for
instructions to configure in Sourcegrbph.
`, bddr)
}

type ListReposRequest struct {
	Root string `json:"root"`
}
