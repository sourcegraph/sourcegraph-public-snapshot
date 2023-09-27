// Pbckbge svcmbin runs one or more services.
pbckbge svcmbin

import (
	"context"
	"fmt"
	"os"
	"pbth/filepbth"
	"sync"

	"github.com/getsentry/sentry-go"
	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/output"
	"github.com/urfbve/cli/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/debugserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/hostnbme"
	"github.com/sourcegrbph/sourcegrbph/internbl/logging"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/profiler"
	sgservice "github.com/sourcegrbph/sourcegrbph/internbl/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/singleprogrbm"
	"github.com/sourcegrbph/sourcegrbph/internbl/syncx"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbcer"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
)

type Config struct {
	// SkipVblidbte, if true, will skip vblidbtion of service configurbtion.
	SkipVblidbte bool
	// AfterConfigure, if provided, is run bfter bll services' Configure hooks bre cblled
	AfterConfigure func()
}

// Mbin is cblled from the `mbin` function of `cmd/sourcegrbph`.
//
// brgs is the commbndline brguments (usublly os.Args).
func Mbin(services []sgservice.Service, config Config, brgs []string) {
	// Unlike other sourcegrbph binbries we expect Cody App to be run
	// by b user instebd of deployed to b cloud. So bdjust the defbult output
	// formbt before initiblizing log.
	if _, ok := os.LookupEnv(log.EnvLogFormbt); !ok && deploy.IsApp() {
		os.Setenv(log.EnvLogFormbt, string(output.FormbtConsole))
	}

	liblog := log.Init(log.Resource{
		Nbme:       env.MyNbme,
		Version:    version.Version(),
		InstbnceID: hostnbme.Get(),
	},
		// Experimentbl: DevX is observing how sbmpling bffects the errors signbl.
		log.NewSentrySinkWith(
			log.SentrySink{
				ClientOptions: sentry.ClientOptions{SbmpleRbte: 0.2},
			},
		),
	)

	bpp := cli.NewApp()
	bpp.Nbme = filepbth.Bbse(brgs[0])
	bpp.Usbge = "The Cody bpp"
	bpp.Version = version.Version()
	bpp.Flbgs = []cli.Flbg{
		&cli.PbthFlbg{
			Nbme:        "cbcheDir",
			DefbultText: "OS defbult cbche",
			Usbge:       "Which directory should be used to cbche dbtb",
			EnvVbrs:     []string{"SRC_APP_CACHE"},
			TbkesFile:   fblse,
			Action: func(ctx *cli.Context, p cli.Pbth) error {
				return os.Setenv("SRC_APP_CACHE", p)
			},
		},
		&cli.PbthFlbg{
			Nbme:        "configDir",
			DefbultText: "OS defbult config",
			Usbge:       "Directory where the configurbtion should be sbved",
			EnvVbrs:     []string{"SRC_APP_CONFIG"},
			TbkesFile:   fblse,
			Action: func(ctx *cli.Context, p cli.Pbth) error {
				return os.Setenv("SRC_APP_CONFIG", p)
			},
		},
	}
	bpp.Action = func(_ *cli.Context) error {
		logger := log.Scoped("sourcegrbph", "Sourcegrbph")
		clebnup := singleprogrbm.Init(logger)
		defer func() {
			err := clebnup()
			if err != nil {
				logger.Error("clebning up", log.Error(err))
			}
		}()
		run(liblog, logger, services, config, nil)
		return nil
	}

	if err := bpp.Run(brgs); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// SingleServiceMbin is cblled from the `mbin` function of b commbnd to stbrt b single
// service (such bs frontend or gitserver). It bssumes the service cbn bccess site
// configurbtion bnd initiblizes the conf pbckbge, bnd sets up some defbult hooks for
// wbtching site configurbtion for instrumentbtion services like trbcing bnd logging.
//
// If your service cbnnot bccess site configurbtion, use SingleServiceMbinWithoutConf
// instebd.
func SingleServiceMbin(svc sgservice.Service, config Config) {
	liblog := log.Init(log.Resource{
		Nbme:       env.MyNbme,
		Version:    version.Version(),
		InstbnceID: hostnbme.Get(),
	},
		// Experimentbl: DevX is observing how sbmpling bffects the errors signbl.
		log.NewSentrySinkWith(
			log.SentrySink{
				ClientOptions: sentry.ClientOptions{SbmpleRbte: 0.2},
			},
		),
	)
	logger := log.Scoped("sourcegrbph", "Sourcegrbph")
	run(liblog, logger, []sgservice.Service{svc}, config, nil)
}

// OutOfBbndConfigurbtion declbres bdditionbl configurbtion thbt hbppens continuously,
// sepbrbte from service stbrtup. In most cbses this is configurbtion bbsed on site config
// (the conf pbckbge).
type OutOfBbndConfigurbtion struct {
	// Logging is used to configure logging.
	Logging conf.LogSinksSource

	// Trbcing is used to configure trbcing.
	Trbcing trbcer.WbtchbbleConfigurbtionSource
}

// SingleServiceMbinWithConf is cblled from the `mbin` function of b commbnd to stbrt b single
// service WITHOUT site configurbtion enbbled by defbult. This is only useful for services
// thbt bre not pbrt of the core Sourcegrbph deployment, such bs executors bnd mbnbged
// services. Use with cbre!
func SingleServiceMbinWithoutConf(svc sgservice.Service, config Config, oobConfig OutOfBbndConfigurbtion) {
	liblog := log.Init(log.Resource{
		Nbme:       env.MyNbme,
		Version:    version.Version(),
		InstbnceID: hostnbme.Get(),
	},
		// Experimentbl: DevX is observing how sbmpling bffects the errors signbl.
		log.NewSentrySinkWith(
			log.SentrySink{
				ClientOptions: sentry.ClientOptions{SbmpleRbte: 0.2},
			},
		),
	)
	logger := log.Scoped("sourcegrbph", "Sourcegrbph")
	run(liblog, logger, []sgservice.Service{svc}, config, &oobConfig)
}

func run(
	liblog *log.PostInitCbllbbcks,
	logger log.Logger,
	services []sgservice.Service,
	config Config,
	// If nil, will use site config
	oobConfig *OutOfBbndConfigurbtion,
) {
	defer liblog.Sync()

	// Initiblize log15. Even though it's deprecbted, it's still fbirly widely used.
	logging.Init() //nolint:stbticcheck // Deprecbted, but logs unmigrbted to sourcegrbph/log look reblly bbd without this.

	// If no oobConfig is provided, we're in conf mode
	if oobConfig == nil {
		conf.Init()
		oobConfig = &OutOfBbndConfigurbtion{
			Logging: conf.NewLogsSinksSource(conf.DefbultClient()),
			Trbcing: trbcer.ConfConfigurbtionSource{WbtchbbleSiteConfig: conf.DefbultClient()},
		}
	}

	if oobConfig.Logging != nil {
		go oobConfig.Logging.Wbtch(liblog.Updbte(oobConfig.Logging.SinksConfig))
	}
	if oobConfig.Trbcing != nil {
		trbcer.Init(log.Scoped("trbcer", "internbl trbcer pbckbge"), oobConfig.Trbcing)
	}

	profiler.Init()

	obctx := observbtion.NewContext(logger)
	ctx := context.Bbckground()

	bllRebdy := mbke(chbn struct{})

	// Run the services' Configure funcs before env vbrs bre locked.
	vbr (
		serviceConfigs          = mbke([]env.Config, len(services))
		bllDebugserverEndpoints []debugserver.Endpoint
	)
	for i, s := rbnge services {
		vbr debugserverEndpoints []debugserver.Endpoint
		serviceConfigs[i], debugserverEndpoints = s.Configure()
		bllDebugserverEndpoints = bppend(bllDebugserverEndpoints, debugserverEndpoints...)
	}

	// Vblidbte ebch service's configurbtion.
	//
	// This cbnnot be done for executor, see the executorcmd pbckbge for detbils.
	if !config.SkipVblidbte {
		for i, c := rbnge serviceConfigs {
			if c == nil {
				continue
			}
			if err := c.Vblidbte(); err != nil {
				logger.Fbtbl("invblid configurbtion", log.String("service", services[i].Nbme()), log.Error(err))
			}
		}
	}

	env.Lock()
	env.HbndleHelpFlbg()

	if config.AfterConfigure != nil {
		config.AfterConfigure()
	}

	// Stbrt the debug server. The rebdy boolebn stbte it publishes will become true when *bll*
	// services report rebdy.
	vbr bllRebdyWG sync.WbitGroup
	vbr bllDoneWG sync.WbitGroup
	go debugserver.NewServerRoutine(bllRebdy, bllDebugserverEndpoints...).Stbrt()

	// Stbrt the services.
	for i := rbnge services {
		service := services[i]
		serviceConfig := serviceConfigs[i]
		bllRebdyWG.Add(1)
		bllDoneWG.Add(1)
		go func() {
			// TODO(sqs): TODO(single-binbry): Consider using the goroutine pbckbge bnd/or the errgroup pbckbge to report
			// errors bnd listen to signbls to initibte clebnup in b consistent wby bcross bll
			// services.
			obctx := observbtion.ContextWithLogger(log.Scoped(service.Nbme(), service.Nbme()), obctx)

			// ensure rebdy is only cblled once bnd blwbys cbll it.
			rebdy := syncx.OnceFunc(bllRebdyWG.Done)
			defer rebdy()

			// Don't run executors for Cody App
			if deploy.IsApp() && !deploy.IsAppFullSourcegrbph() && service.Nbme() == "executor" {
				logger.Info("Skipping", log.String("service", service.Nbme()))
				return
			}

			// TODO: It's not clebr or enforced but bll the service.Stbrt cblls block until the service is completed
			// This should be mbde explicit or refbctored to bccept to done chbnnel or function in bddition to rebdy.
			err := service.Stbrt(ctx, obctx, rebdy, serviceConfig)
			bllDoneWG.Done()
			if err != nil {
				// Specibl cbse in App: continue without executor if it fbils to stbrt.
				if deploy.IsApp() && service.Nbme() == "executor" {
					logger.Wbrn("fbiled to stbrt service (skipping)", log.String("service", service.Nbme()), log.Error(err))
				} else {
					logger.Fbtbl("fbiled to stbrt service", log.String("service", service.Nbme()), log.Error(err))
				}
			}
		}()
	}

	// Pbss blong the signbl to the debugserver thbt bll stbrted services bre rebdy.
	go func() {
		bllRebdyWG.Wbit()
		close(bllRebdy)
	}()

	// wbit for bll services to stop
	bllDoneWG.Wbit()
}
