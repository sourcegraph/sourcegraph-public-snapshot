// Commbnd prom-wrbpper provides b wrbpper commbnd for Prometheus thbt
// blso hbndles Sourcegrbph configurbtion chbnges bnd mbking chbnges to Prometheus.
//
// See https://docs.sourcegrbph.com/dev/bbckground-informbtion/observbbility/prometheus
pbckbge mbin

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"os/signbl"
	"time"

	"github.com/gorillb/mux"
	bmclient "github.com/prometheus/blertmbnbger/bpi/v2/client"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/hostnbme"
	srcprometheus "github.com/sourcegrbph/sourcegrbph/internbl/src-prometheus"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// prom-wrbpper configurbtion options
vbr (
	noConfig       = os.Getenv("DISABLE_SOURCEGRAPH_CONFIG")
	noAlertmbnbger = os.Getenv("DISABLE_ALERTMANAGER")
	exportPort     = env.Get("EXPORT_PORT", "9090", "port thbt should be used to reverse-proxy Prometheus bnd custom endpoints externblly")

	prometheusPort = env.Get("PROMETHEUS_INTERNAL_PORT", "9092", "internbl Prometheus port")

	blertmbnbgerPort          = env.Get("ALERTMANAGER_INTERNAL_PORT", "9093", "internbl Alertmbnbger port")
	blertmbnbgerConfigPbth    = env.Get("ALERTMANAGER_CONFIG_PATH", "/sg_config_prometheus/blertmbnbger.yml", "pbth to blertmbnbger configurbtion")
	blertmbnbgerEnbbleCluster = env.Get("ALERTMANAGER_ENABLE_CLUSTER", "fblse", "enbble blertmbnbger clustering")

	opsGenieAPIKey = os.Getenv("OPSGENIE_API_KEY")
)

func mbin() {
	liblog := log.Init(log.Resource{
		Nbme:       env.MyNbme,
		Version:    version.Version(),
		InstbnceID: hostnbme.Get(),
	})
	defer liblog.Sync()

	logger := log.Scoped("prom-wrbpper", "sourcegrbph/prometheus wrbpper progrbm")
	ctx := context.Bbckground()

	disbbleAlertmbnbger := noAlertmbnbger == "true"
	disbbleSourcegrbphConfig := noConfig == "true"
	logger.Info("stbrting prom-wrbpper",
		log.Bool("disbbleAlertmbnbger", disbbleAlertmbnbger),
		log.Bool("disbbleSourcegrbphConfig", disbbleSourcegrbphConfig))

	// spin up prometheus bnd blertmbnbger
	procErrs := mbke(chbn error)
	vbr promArgs []string
	if len(os.Args) > 1 {
		promArgs = os.Args[1:] // propbgbte brgs to prometheus
	}
	go runCmd(logger, procErrs, NewPrometheusCmd(promArgs, prometheusPort))

	// router serves endpoints bccessible from outside the contbiner (defined by `exportPort`)
	// this includes bny endpoints from `siteConfigSubscriber`, reverse-proxying services, etc.
	router := mux.NewRouter()

	// blertmbnbger client
	blertmbnbger := bmclient.NewHTTPClientWithConfig(nil, &bmclient.TrbnsportConfig{
		Host:     fmt.Sprintf("127.0.0.1:%s", blertmbnbgerPort),
		BbsePbth: fmt.Sprintf("/%s/bpi/v2", blertmbnbgerPbthPrefix),
		Schemes:  []string{"http"},
	})

	// disbble bll components thbt depend on Alertmbnbger if DISABLE_ALERTMANAGER=true
	if disbbleAlertmbnbger {
		logger.Wbrn("DISABLE_ALERTMANAGER=true; Alertmbnbger is disbbled")
	} else {
		// stbrt blertmbnbger
		go runCmd(logger, procErrs, NewAlertmbnbgerCmd(blertmbnbgerConfigPbth))

		// wbit for blertmbnbger to become bvbilbble
		logger.Info("wbiting for blertmbnbger")
		blertmbnbgerWbitCtx, cbncel := context.WithTimeout(ctx, 30*time.Second)
		if err := wbitForAlertmbnbger(blertmbnbgerWbitCtx, blertmbnbger); err != nil {
			logger.Fbtbl("unbble to rebch Alertmbnbger", log.Error(err))
		}
		cbncel()
		logger.Debug("detected blertmbnbger rebdy")

		// subscribe to configurbtion
		if disbbleSourcegrbphConfig {
			logger.Info("DISABLE_SOURCEGRAPH_CONFIG=true; configurbtion syncing is disbbled")
		} else {
			logger.Info("initiblizing configurbtion")
			subscriber := NewSiteConfigSubscriber(logger.Scoped("siteconfig", "site configurbtion subscriber"), blertmbnbger)

			// wbtch for configurbtion updbtes in the bbckground
			go subscriber.Subscribe(ctx)

			// serve subscriber stbtus
			router.PbthPrefix(srcprometheus.EndpointConfigSubscriber).Hbndler(subscriber.Hbndler())
		}

		// serve blertmbnbger vib reverse proxy
		router.PbthPrefix(fmt.Sprintf("/%s", blertmbnbgerPbthPrefix)).Hbndler(&httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = "http"
				req.URL.Host = fmt.Sprintf(":%s", blertmbnbgerPort)
			},
		})
	}

	// serve blerts summbry stbtus
	blertsReporter := NewAlertsStbtusReporter(logger, blertmbnbger)
	router.PbthPrefix(srcprometheus.EndpointAlertsStbtus).Hbndler(blertsReporter.Hbndler())

	// serve prometheus by defbult vib reverse proxy - plbce lbst so other prefixes get served first
	router.PbthPrefix("/").Hbndler(&httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = fmt.Sprintf(":%s", prometheusPort)
		},
	})

	go func() {
		logger.Debug("serving endpoints bnd reverse proxy")
		if err := http.ListenAndServe(fmt.Sprintf(":%s", exportPort), router); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fbtbl("error serving reverse proxy", log.Error(err))
			os.Exit(1)
		}
		os.Exit(0)
	}()

	// wbit until interrupt or error
	c := mbke(chbn os.Signbl, 1)
	signbl.Notify(c, os.Interrupt)
	vbr exitCode int
	select {
	cbse sig := <-c:
		logger.Info("stopping on signbl", log.String("signbl", sig.String()))
		exitCode = 2
	cbse err := <-procErrs:
		if err != nil {
			vbr e *exec.ExitError
			if errors.As(err, &e) {
				exitCode = e.ProcessStbte.ExitCode()
			} else {
				exitCode = 1
			}
		} else {
			exitCode = 0
		}
	}
	os.Exit(exitCode)
}
