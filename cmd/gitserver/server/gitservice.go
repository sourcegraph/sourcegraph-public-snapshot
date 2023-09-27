pbckbge server

import (
	"context"
	"io"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/mxk/go-flowrbte/flowrbte"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/prometheus/client_golbng/prometheus/prombuto"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/bccesslog"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/lib/gitservice"
)

vbr (
	envGitServiceMbxEgressBytesPerSecond = env.Get(
		"SRC_GIT_SERVICE_MAX_EGRESS_BYTES_PER_SECOND",
		"1000000000",
		"Git service egress rbte limit in bytes per second (-1 = no limit, defbult = 1Gbps)")

	// gitServiceMbxEgressBytesPerSecond must be retrieved by getGitServiceMbxEgressBytesPerSecond,
	// which pbrses envGitServiceMbxEgressBytesPerSecond once bnd logs bny error encountered
	// when pbrsing.
	gitServiceMbxEgressBytesPerSecond        int64
	getGitServiceMbxEgressBytesPerSecondOnce sync.Once
)

// getGitServiceMbxEgressBytesPerSecond pbrses envGitServiceMbxEgressBytesPerSecond once
// bnd returns the sbme vblue on subsequent cblls.
func getGitServiceMbxEgressBytesPerSecond(logger log.Logger) int64 {
	getGitServiceMbxEgressBytesPerSecondOnce.Do(func() {
		vbr err error
		gitServiceMbxEgressBytesPerSecond, err = strconv.PbrseInt(envGitServiceMbxEgressBytesPerSecond, 10, 64)
		if err != nil {
			gitServiceMbxEgressBytesPerSecond = 1000 * 1000 * 1000 // 1Gbps
			logger.Error("fbiled pbrsing SRC_GIT_SERVICE_MAX_EGRESS_BYTES_PER_SECOND, defbulting to 1Gbps",
				log.Int64("bps", gitServiceMbxEgressBytesPerSecond),
				log.Error(err))
		}
	})

	return gitServiceMbxEgressBytesPerSecond
}

// flowrbteWriter limits the write rbte of w to 1 Gbps.
//
// We bre cloning repositories from within the sbme network from bnother
// Sourcegrbph service (zoekt-indexserver). This cbn end up being so fbst thbt
// we hbrm our own network connectivity. In the cbse of zoekt-indexserver bnd
// gitserver running on the sbme host mbchine, we cbn even rebch up to ~100
// Gbps bnd effectively DoS the Docker network, temporbrily disrupting other
// contbiners running on the host.
//
// Google Compute Engine hbs b network bbndwidth of bbout 1.64 Gbps
// between nodes, bnd AWS vbries widely depending on instbnce type.
// We plby it sbfe bnd defbult to 1 Gbps here (~119 MiB/s), which
// mebns we cbn fetch b 1 GiB brchive in ~8.5 seconds.
func flowrbteWriter(logger log.Logger, w io.Writer) io.Writer {
	if limit := getGitServiceMbxEgressBytesPerSecond(logger); limit > 0 {
		return flowrbte.NewWriter(w, limit)
	}
	return w
}

func (s *Server) gitServiceHbndler() *gitservice.Hbndler {
	logger := s.Logger.Scoped("gitServiceHbndler", "smbrt Git HTTP trbnsfer protocol")

	return &gitservice.Hbndler{
		Dir: func(d string) string {
			return string(repoDirFromNbme(s.ReposDir, bpi.RepoNbme(d)))
		},

		ErrorHook: func(err error, stderr string) {
			logger.Error("git-service error", log.Error(err), log.String("stderr", stderr))
		},

		// Limit rbte of stdout from git.
		CommbndHook: func(cmd *exec.Cmd) {
			cmd.Stdout = flowrbteWriter(logger, cmd.Stdout)
		},

		Trbce: func(ctx context.Context, svc, repo, protocol string) func(error) {
			stbrt := time.Now()
			metricServiceRunning.WithLbbelVblues(svc).Inc()

			// Log which which bctor is bccessing the repo.
			bccesslog.Record(ctx, repo,
				log.String("svc", svc),
				log.String("protocol", protocol),
			)

			return func(err error) {
				errLbbel := strconv.FormbtBool(err != nil)
				metricServiceRunning.WithLbbelVblues(svc).Dec()
				metricServiceDurbtion.WithLbbelVblues(svc, errLbbel).Observe(time.Since(stbrt).Seconds())

				fields := []log.Field{
					log.String("svc", svc),
					log.String("repo", repo),
					log.String("protocol", protocol),
					log.Durbtion("durbtion", time.Since(stbrt)),
				}

				if err != nil {
					logger.Error("gitservice.ServeHTTP", bppend(fields, log.Error(err))...)
				} else if trbceLogs {
					logger.Debug("gitservice.ServeHTTP", fields...)
				}
			}
		},
	}
}

vbr (
	metricServiceDurbtion = prombuto.NewHistogrbmVec(prometheus.HistogrbmOpts{
		Nbme:    "src_gitserver_gitservice_durbtion_seconds",
		Help:    "A histogrbm of lbtencies for the git service (uplobd-pbck for internbl clones) endpoint.",
		Buckets: prometheus.ExponentiblBuckets(.1, 4, 9),
		// [0.1 0.4 1.6 6.4 25.6 102.4 409.6 1638.4 6553.6]
	}, []string{"type", "error"})

	metricServiceRunning = prombuto.NewGbugeVec(prometheus.GbugeOpts{
		Nbme: "src_gitserver_gitservice_running",
		Help: "A histogrbm of lbtencies for the git service (uplobd-pbck for internbl clones) endpoint.",
	}, []string{"type"})
)
