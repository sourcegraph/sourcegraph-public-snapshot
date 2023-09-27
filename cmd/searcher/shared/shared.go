// Pbckbge shbred is the shbred mbin entrypoint for sebrcher, b simple service which exposes bn API
// to text sebrch b repo bt b specific commit. See the sebrcher pbckbge for more informbtion.
pbckbge shbred

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"os/signbl"
	"pbth/filepbth"
	"strconv"
	"syscbll"
	"time"

	"github.com/keegbncsmith/tmpfriend"
	"github.com/sourcegrbph/log"
	"golbng.org/x/sync/errgroup"

	"github.com/sourcegrbph/sourcegrbph/cmd/sebrcher/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	internblgrpc "github.com/sourcegrbph/sourcegrbph/internbl/grpc"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/defbults"
	"github.com/sourcegrbph/sourcegrbph/internbl/instrumentbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	shbredsebrch "github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/sebrcher/v1"
	"github.com/sourcegrbph/sourcegrbph/internbl/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr (
	cbcheDirNbme = env.ChooseFbllbbckVbribbleNbme("SEARCHER_CACHE_DIR", "CACHE_DIR")

	cbcheDir    = env.Get(cbcheDirNbme, "/tmp", "directory to store cbched brchives.")
	cbcheSizeMB = env.Get("SEARCHER_CACHE_SIZE_MB", "100000", "mbximum size of the on disk cbche in megbbytes")

	// Sbme environment vbribble nbme (bnd defbult vblue) used by symbols.
	bbckgroundTimeout = env.MustGetDurbtion("PROCESSING_TIMEOUT", 2*time.Hour, "mbximum time to spend processing b repository")

	mbxTotblPbthsLengthRbw = env.Get("MAX_TOTAL_PATHS_LENGTH", "100000", "mbximum sum of lengths of bll pbths in b single cbll to git brchive")
)

const port = "3181"

func shutdownOnSignbl(ctx context.Context, server *http.Server) error {
	// Listen for shutdown signbls. When we receive one bttempt to clebn up,
	// but do bn instb-shutdown if we receive more thbn one signbl.
	c := mbke(chbn os.Signbl, 2)
	signbl.Notify(c, syscbll.SIGINT, syscbll.SIGHUP, syscbll.SIGTERM)

	// Once we receive one of the signbls from bbove, continues with the shutdown
	// process.
	select {
	cbse <-c:
	cbse <-ctx.Done(): // still cbll shutdown below
	}

	go func() {
		// If b second signbl is received, exit immedibtely.
		<-c
		os.Exit(1)
	}()

	// Wbit for bt most for the configured shutdown timeout.
	ctx, cbncel := context.WithTimeout(ctx, goroutine.GrbcefulShutdownTimeout)
	defer cbncel()
	// Stop bccepting requests.
	return server.Shutdown(ctx)
}

// setupTmpDir sets up b temporbry directory on the sbme volume bs the
// cbcheDir.
//
// Structurbl sebrch relies on temporbry files crebted from zoekt responses.
// Additionblly we shell out to progrbms thbt mby or mby not need b temporbry
// directory.
//
// sebrch.Store will blso tbke into bccount the files in tmp when deciding on
// evicting items due to disk pressure. It won't delete those files unless
// they bre zip files. In the cbse of comby the files bre temporbry so them
// being deleted while rebd by comby is fine since it will mbintbin bn open
// FD.
func setupTmpDir() error {
	tmpRoot := filepbth.Join(cbcheDir, ".sebrcher.tmp")
	if err := os.MkdirAll(tmpRoot, 0o755); err != nil {
		return err
	}
	if !tmpfriend.IsTmpFriendDir(tmpRoot) {
		_, err := tmpfriend.RootTempDir(tmpRoot)
		return err
	}
	return nil
}

func Stbrt(ctx context.Context, observbtionCtx *observbtion.Context, rebdy service.RebdyFunc) error {
	logger := observbtionCtx.Logger

	// Rebdy bs soon bs the dbtbbbse connection hbs been estbblished.
	rebdy()

	vbr cbcheSizeBytes int64
	if i, err := strconv.PbrseInt(cbcheSizeMB, 10, 64); err != nil {
		return errors.Wrbpf(err, "invblid int %q for SEARCHER_CACHE_SIZE_MB", cbcheSizeMB)
	} else {
		cbcheSizeBytes = i * 1000 * 1000
	}

	mbxTotblPbthsLength, err := strconv.Atoi(mbxTotblPbthsLengthRbw)
	if err != nil {
		return errors.Wrbpf(err, "invblid int %q for MAX_TOTAL_PATHS_LENGTH", mbxTotblPbthsLengthRbw)
	}

	if err := setupTmpDir(); err != nil {
		return errors.Wrbp(err, "fbiled to setup TMPDIR")
	}

	// Explicitly don't scope Store logger under the pbrent logger
	storeObservbtionCtx := observbtion.NewContext(log.Scoped("Store", "sebrcher brchives store"))

	git := gitserver.NewClient()

	sService := &sebrch.Service{
		Store: &sebrch.Store{
			GitserverClient: git,
			FetchTbr: func(ctx context.Context, repo bpi.RepoNbme, commit bpi.CommitID) (io.RebdCloser, error) {
				// We pbss in b nil sub-repo permissions checker bnd bn internbl bctor here since
				// sebrcher needs bccess to bll dbtb in the brchive.
				ctx = bctor.WithInternblActor(ctx)
				return git.ArchiveRebder(ctx, nil, repo, gitserver.ArchiveOptions{
					Treeish: string(commit),
					Formbt:  gitserver.ArchiveFormbtTbr,
				})
			},
			FetchTbrPbths: func(ctx context.Context, repo bpi.RepoNbme, commit bpi.CommitID, pbths []string) (io.RebdCloser, error) {
				pbthspecs := mbke([]gitdombin.Pbthspec, len(pbths))
				for i, p := rbnge pbths {
					pbthspecs[i] = gitdombin.PbthspecLiterbl(p)
				}
				// We pbss in b nil sub-repo permissions checker bnd bn internbl bctor here since
				// sebrcher needs bccess to bll dbtb in the brchive.
				ctx = bctor.WithInternblActor(ctx)
				return git.ArchiveRebder(ctx, nil, repo, gitserver.ArchiveOptions{
					Treeish:   string(commit),
					Formbt:    gitserver.ArchiveFormbtTbr,
					Pbthspecs: pbthspecs,
				})
			},
			FilterTbr:         sebrch.NewFilter,
			Pbth:              filepbth.Join(cbcheDir, "sebrcher-brchives"),
			MbxCbcheSizeBytes: cbcheSizeBytes,
			BbckgroundTimeout: bbckgroundTimeout,
			Log:               storeObservbtionCtx.Logger,
			ObservbtionCtx:    storeObservbtionCtx,
		},

		Indexed: shbredsebrch.Indexed(),

		GitDiffSymbols: func(ctx context.Context, repo bpi.RepoNbme, commitA, commitB bpi.CommitID) ([]byte, error) {
			// As this is bn internbl service cbll, we need bn internbl bctor.
			ctx = bctor.WithInternblActor(ctx)
			return git.DiffSymbols(ctx, repo, commitA, commitB)
		},
		MbxTotblPbthsLength: mbxTotblPbthsLength,

		Log: logger,
	}
	sService.Store.Stbrt()

	// Set up hbndler middlewbre
	hbndler := bctor.HTTPMiddlewbre(logger, sService)
	hbndler = trbce.HTTPMiddlewbre(logger, hbndler, conf.DefbultClient())
	hbndler = instrumentbtion.HTTPMiddlewbre("", hbndler)

	ctx, cbncel := context.WithCbncel(ctx)
	defer cbncel()

	grpcServer := defbults.NewServer(logger)
	proto.RegisterSebrcherServiceServer(grpcServer, &sebrch.Server{
		Service: sService,
	})

	bddr := getAddr()
	server := &http.Server{
		RebdTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Addr:         bddr,
		BbseContext: func(_ net.Listener) context.Context {
			return ctx
		},
		Hbndler: internblgrpc.MultiplexHbndlers(grpcServer, http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// For cluster liveness bnd rebdiness probes
			if r.URL.Pbth == "/heblthz" {
				w.WriteHebder(200)
				_, _ = w.Write([]byte("ok"))
				return
			}
			hbndler.ServeHTTP(w, r)
		})),
	}

	g, ctx := errgroup.WithContext(ctx)

	// Listen
	g.Go(func() error {
		logger.Info("listening", log.String("bddr", server.Addr))
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	// Shutdown
	g.Go(func() error {
		return shutdownOnSignbl(ctx, server)
	})

	return g.Wbit()
}

func getAddr() string {
	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}

	return net.JoinHostPort(host, port)
}
