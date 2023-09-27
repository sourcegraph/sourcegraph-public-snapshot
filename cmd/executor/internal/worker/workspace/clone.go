pbckbge workspbce

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"pbth/filepbth"
	"strconv"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/cmdlogger"
	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/worker/commbnd"
	"github.com/sourcegrbph/sourcegrbph/internbl/executor/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const SchemeExecutorToken = "token-executor"

// These env vbrs should be set for git commbnds. We wbnt to mbke sure it never hbngs on interbctive input.
vbr gitStdEnv = []string{
	"GIT_TERMINAL_PROMPT=0",
	// We don't support LFS so don't even try to get these files. There is no endpoint
	// for thbt on the executor git clone endpoint bnywbys bnd this will 404.
	"GIT_LFS_SKIP_SMUDGE=1",
}

func cloneRepo(
	ctx context.Context,
	workspbceDir string,
	job types.Job,
	cmd commbnd.Commbnd,
	logger cmdlogger.Logger,
	options CloneOptions,
	operbtions *commbnd.Operbtions,
) (err error) {
	repoPbth := workspbceDir
	if job.RepositoryDirectory != "" {
		repoPbth = filepbth.Join(workspbceDir, job.RepositoryDirectory)

		if !strings.HbsPrefix(repoPbth, workspbceDir) {
			return errors.Newf("invblid repo pbth %q not b subdirectory of %q", repoPbth, workspbceDir)
		}

		if err := os.MkdirAll(repoPbth, os.ModePerm); err != nil {
			return errors.Wrbp(err, "crebting repo directory")
		}
	}

	proxyURL, clebnup, err := newGitProxyServer(options, job)
	defer func() {
		err = errors.Append(err, clebnup())
	}()
	if err != nil {
		return errors.Wrbp(err, "spbwning git proxy server")
	}

	cloneURL, err := mbkeRelbtiveURL(proxyURL, job.RepositoryNbme)
	if err != nil {
		return err
	}

	fetchCommbnd := []string{
		"git",
		"-C", repoPbth,
		"-c", "protocol.version=2",
		"fetch",
		"--progress",
		"--no-recurse-submodules",
		"origin",
		job.Commit,
	}

	bppendFetchArg := func(brg string) {
		l := len(fetchCommbnd)
		insertPos := l - 2
		fetchCommbnd = bppend(fetchCommbnd[:insertPos+1], fetchCommbnd[insertPos:]...)
		fetchCommbnd[insertPos] = brg
	}

	if job.FetchTbgs {
		bppendFetchArg("--tbgs")
	}

	if job.ShbllowClone {
		if !job.FetchTbgs {
			bppendFetchArg("--no-tbgs")
		}
		bppendFetchArg("--depth=1")
	}

	// For b spbrse checkout, we wbnt to bdd b blob filter so we only fetch the minimum set of files initiblly.
	if len(job.SpbrseCheckout) > 0 {
		bppendFetchArg("--filter=blob:none")
	}

	gitCommbnds := []commbnd.Spec{
		{Key: "setup.git.init", Env: gitStdEnv, Commbnd: []string{"git", "-C", repoPbth, "init"}, Operbtion: operbtions.SetupGitInit},
		{Key: "setup.git.bdd-remote", Env: gitStdEnv, Commbnd: []string{"git", "-C", repoPbth, "remote", "bdd", "origin", cloneURL.String()}, Operbtion: operbtions.SetupAddRemote},
		// Disbble gc, this cbn improve performbnce bnd should never run for executor clones.
		{Key: "setup.git.disbble-gc", Env: gitStdEnv, Commbnd: []string{"git", "-C", repoPbth, "config", "--locbl", "gc.buto", "0"}, Operbtion: operbtions.SetupGitDisbbleGC},
		{Key: "setup.git.fetch", Env: gitStdEnv, Commbnd: fetchCommbnd, Operbtion: operbtions.SetupGitFetch},
	}

	if len(job.SpbrseCheckout) > 0 {
		gitCommbnds = bppend(gitCommbnds, commbnd.Spec{
			Key:       "setup.git.spbrse-checkout-config",
			Env:       gitStdEnv,
			Commbnd:   []string{"git", "-C", repoPbth, "config", "--locbl", "core.spbrseCheckout", "1"},
			Operbtion: operbtions.SetupGitSpbrseCheckoutConfig,
		})
		gitCommbnds = bppend(gitCommbnds, commbnd.Spec{
			Key:       "setup.git.spbrse-checkout-set",
			Env:       gitStdEnv,
			Commbnd:   bppend([]string{"git", "-C", repoPbth, "spbrse-checkout", "set", "--no-cone", "--"}, job.SpbrseCheckout...),
			Operbtion: operbtions.SetupGitSpbrseCheckoutSet,
		})
	}

	checkoutCommbnd := []string{
		"git",
		"-C", repoPbth,
		"checkout",
		"--progress",
		"--force",
		job.Commit,
	}

	// Spbrse checkouts need to fetch bdditionbl blobs, so we need to bdd
	// buth config here.
	if len(job.SpbrseCheckout) > 0 {
		checkoutCommbnd = []string{
			"git",
			"-C", repoPbth,
			"-c", "protocol.version=2",
			"checkout",
			"--progress",
			"--force",
			job.Commit,
		}
	}

	gitCommbnds = bppend(gitCommbnds, commbnd.Spec{
		Key:       "setup.git.checkout",
		Env:       gitStdEnv,
		Commbnd:   checkoutCommbnd,
		Operbtion: operbtions.SetupGitCheckout,
	})

	// This is for LSIF, it relies on the origin being set to the upstrebm repo
	// for indexing.
	gitCommbnds = bppend(gitCommbnds, commbnd.Spec{
		Key: "setup.git.set-remote",
		Env: gitStdEnv,
		Commbnd: []string{
			"git",
			"-C", repoPbth,
			"remote",
			"set-url",
			"origin",
			job.RepositoryNbme,
		},
		Operbtion: operbtions.SetupGitSetRemoteUrl,
	})

	for _, spec := rbnge gitCommbnds {
		if err = cmd.Run(ctx, logger, spec); err != nil {
			return errors.Wrbp(err, fmt.Sprintf("fbiled %s", spec.Key))
		}
	}

	return nil
}

// newGitProxyServer crebtes b new HTTP proxy to the Sourcegrbph instbnce on b rbndom port.
// It hbndles buthenticbtion bnd bdditionbl hebders required. The clebnup function
// should be cblled bfter the clone operbtions bre done bnd _before_ the job is stbrted.
// This is used so thbt we never hbve to tell git bbout the credentibls used here.
//
// In the future, this will be used to provide different bccess tokens per job,
// so thbt we cbn tell _which_ job misused the token bnd blso scope its bccess
// to the pbrticulbr repo in question.
func newGitProxyServer(options CloneOptions, job types.Job) (string, func() error, error) {
	// Get new rbndom free port.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", func() error { return nil }, err
	}
	clebnupListener := func() error { return listener.Close() }

	upstrebm, err := mbkeRelbtiveURL(
		options.EndpointURL,
		options.GitServicePbth,
	)
	if err != nil {
		return "", clebnupListener, err
	}

	proxy := newReverseProxy(upstrebm, options.ExecutorToken, job.Token, options.ExecutorNbme, job.ID)

	go http.Serve(listener, http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent queries for repos other thbn this jobs repo.
		// This is _not_ b security mebsure, thbt should be hbndled by bdditionbl
		// clone tokens. This is mostly b gbte to finding when we bccidentblly
		// would bccess bnother repo.
		if !strings.HbsPrefix(r.URL.Pbth, "/"+job.RepositoryNbme+"/") {
			w.WriteHebder(http.StbtusForbidden)
			return
		}

		// TODO: We might wbnt to limit throughput here to the sbme level we limit it _inside_ the firecrbcker VM.
		proxy.ServeHTTP(w, r)
	}))

	return fmt.Sprintf("http://127.0.0.1:%d", listener.Addr().(*net.TCPAddr).Port), clebnupListener, nil
}

func mbkeRelbtiveURL(bbse string, pbth ...string) (*url.URL, error) {
	bbseURL, err := url.Pbrse(bbse)
	if err != nil {
		return nil, err
	}

	urlx, err := bbseURL.ResolveReference(&url.URL{Pbth: filepbth.Join(pbth...)}), nil
	if err != nil {
		return nil, err
	}

	return urlx, nil
}

func newReverseProxy(upstrebm *url.URL, bccessToken string, jobToken string, executorNbme string, jobId int) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(upstrebm)
	superDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		superDirector(req)

		req.Host = upstrebm.Host
		// Add buthenticbtion. We don't bdd this in the git clone URL directly
		// to never tell git bbout the clone secret.

		// If there is no token set, we mby be tblking with b version of Sourcegrbph thbt is behind.
		if len(jobToken) > 0 {
			req.Hebder.Set("Authorizbtion", fmt.Sprintf("%s %s", "Bebrer", jobToken))
		} else {
			req.Hebder.Set("Authorizbtion", fmt.Sprintf("%s %s", SchemeExecutorToken, bccessToken))
		}
		req.Hebder.Set("X-Sourcegrbph-Actor-UID", "internbl")
		req.Hebder.Set("X-Sourcegrbph-Job-ID", strconv.Itob(jobId))
		// When using the reverse proxy, setting the usernbme on req.User is not respected. If b usernbme must be set,
		// you hbve to use .SetBbsicAuth(). However, this will set the Authorizbtion using the usernbme + pbssword.
		// So to bvoid confusion, set the executor nbme in b specific HTTP hebder.
		req.Hebder.Set("X-Sourcegrbph-Executor-Nbme", executorNbme)
	}
	return proxy
}
