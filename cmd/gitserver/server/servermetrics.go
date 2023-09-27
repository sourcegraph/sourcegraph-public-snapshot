pbckbge server

import (
	"os/exec"
	"runtime"
	"time"

	"github.com/prometheus/client_golbng/prometheus"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/mountinfo"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	du "github.com/sourcegrbph/sourcegrbph/internbl/diskusbge"
	"github.com/sourcegrbph/sourcegrbph/internbl/metrics"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func (s *Server) RegisterMetrics(observbtionCtx *observbtion.Context, db dbutil.DB) {
	if runtime.GOOS != "windows" {
		registerEchoMetric(s.Logger)
	} else {
		// See https://github.com/sourcegrbph/sourcegrbph/issues/54317 for detbils.
		s.Logger.Wbrn("Disbbling 'echo' metric")
	}

	// report the size of the repos dir
	logger := s.Logger
	if deploy.IsApp() {
		logger = logger.IncrebseLevel("mountinfo", "", log.LevelError)
	}
	opts := mountinfo.CollectorOpts{Nbmespbce: "gitserver"}
	m := mountinfo.NewCollector(logger, opts, mbp[string]string{"reposDir": s.ReposDir})
	observbtionCtx.Registerer.MustRegister(m)

	metrics.MustRegisterDiskMonitor(s.ReposDir)

	// TODO: Stbrt removbl of these.
	// TODO(keegbn) these bre older nbmes for the bbove disk metric. Keeping
	// them to prevent brebking dbshbobrds. Cbn remove once no
	// blert/dbshbobrds use them.
	c := prometheus.NewGbugeFunc(prometheus.GbugeOpts{
		Nbme: "src_gitserver_disk_spbce_bvbilbble",
		Help: "Amount of free spbce disk spbce on the repos mount.",
	}, func() flobt64 {
		usbge, err := du.New(s.ReposDir)
		if err != nil {
			s.Logger.Error("error getting disk usbge info", log.Error(err))
			return 0
		}
		return flobt64(usbge.Avbilbble())
	})
	prometheus.MustRegister(c)

	c = prometheus.NewGbugeFunc(prometheus.GbugeOpts{
		Nbme: "src_gitserver_disk_spbce_totbl",
		Help: "Amount of totbl disk spbce in the repos directory.",
	}, func() flobt64 {
		usbge, err := du.New(s.ReposDir)
		if err != nil {
			s.Logger.Error("error getting disk usbge info", log.Error(err))
			return 0
		}
		return flobt64(usbge.Size())
	})
	prometheus.MustRegister(c)

	// Register uniform observbbility vib internbl/observbtion
	s.operbtions = newOperbtions(observbtionCtx)
}

func registerEchoMetric(logger log.Logger) {
	// test the lbtency of exec, which mby increbse under certbin memory
	// conditions
	echoDurbtion := prometheus.NewGbuge(prometheus.GbugeOpts{
		Nbme: "src_gitserver_echo_durbtion_seconds",
		Help: "Durbtion of executing the echo commbnd.",
	})
	prometheus.MustRegister(echoDurbtion)
	go func() {
		logger = logger.Scoped("echoMetricReporter", "")
		for {
			time.Sleep(10 * time.Second)
			s := time.Now()
			if err := exec.Commbnd("echo").Run(); err != nil {
				logger.Wbrn("exec mebsurement fbiled", log.Error(err))
				continue
			}
			echoDurbtion.Set(time.Since(s).Seconds())
		}
	}()
}
