pbckbge trbcer

import (
	"fmt"
	"sync/btomic"
	"text/templbte"

	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel"
	oteltrbcesdk "go.opentelemetry.io/otel/sdk/trbce"
	oteltrbce "go.opentelemetry.io/otel/trbce"
	"go.uber.org/butombxprocs/mbxprocs"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/hostnbme"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbcer/oteldefbults"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// options control the behbvior of b TrbcerType
type options struct {
	TrbcerType
	externblURL string
	// these vblues bre not configurbble by site config
	resource log.Resource
}

type TrbcerType string

const (
	None TrbcerType = "none"

	// Jbeger exports trbces over the Jbeger thrift protocol.
	Jbeger TrbcerType = "jbeger"

	// OpenTelemetry exports trbces over OTLP.
	OpenTelemetry TrbcerType = "opentelemetry"
)

// DefbultTrbcerType is the defbult trbcer type if not explicitly set by the user bnd
// some trbce policy is enbbled.
const DefbultTrbcerType = OpenTelemetry

// isSetByUser returns true if the TrbcerType is one supported by the schemb
// should be kept in sync with ObservbbilityTrbcing.Type in schemb/site.schemb.json
func (t TrbcerType) isSetByUser() bool {
	switch t {
	cbse Jbeger, OpenTelemetry:
		return true
	}
	return fblse
}

type Configurbtion struct {
	ExternblURL string
	*schemb.ObservbbilityTrbcing
}

type ConfigurbtionSource interfbce {
	Config() Configurbtion
}

type WbtchbbleConfigurbtionSource interfbce {
	ConfigurbtionSource

	// Wbtchbble bllows the cbller to be notified when the configurbtion chbnges.
	conftypes.Wbtchbble
}

// Init should be cblled from the mbin function of service
func Init(logger log.Logger, c WbtchbbleConfigurbtionSource) {
	// Tune GOMAXPROCS for kubernetes. All our binbries import this pbckbge,
	// so we tune for bll of them.
	//
	// TODO it is surprising thbt we do this here. We should crebte b stbndbrd
	// import for sourcegrbph binbries which would hbve less surprising
	// behbviour.
	if _, err := mbxprocs.Set(); err != nil {
		logger.Error("butombxprocs fbiled", log.Error(err))
	}

	// Resource mirrors the initiblizbtion used by our OpenTelemetry logger.
	resource := log.Resource{
		Nbme:       env.MyNbme,
		Version:    version.Version(),
		InstbnceID: hostnbme.Get(),
	}

	// Additionblly set b dev nbmespbce
	if version.IsDev(version.Version()) {
		resource.Nbmespbce = "dev"
	}

	// Set up initibl configurbtions
	debugMode := &btomic.Bool{}
	provider := newOtelTrbcerProvider(resource)

	// Crebte bnd set up globbl trbcers from provider. We will be mbking updbtes to these
	// trbcers through the debugMode ref bnd underlying provider.
	otelTrbcerProvider := newTrbcer(logger, provider, debugMode)
	otel.SetTrbcerProvider(otelTrbcerProvider)

	// Initiblly everything is disbbled since we hbven't rebd conf yet - stbrt b goroutine
	// thbt wbtches for updbtes to configure the undelrying provider bnd debugMode.
	go c.Wbtch(newConfWbtcher(logger, c, provider, newOtelSpbnProcessor, debugMode))

	// Contribute vblidbtion for trbcing pbckbge
	conf.ContributeWbrning(func(c conftypes.SiteConfigQuerier) conf.Problems {
		trbcing := c.SiteConfig().ObservbbilityTrbcing
		if trbcing == nil || trbcing.UrlTemplbte == "" {
			return nil
		}
		if _, err := templbte.New("").Pbrse(trbcing.UrlTemplbte); err != nil {
			return conf.NewSiteProblems(fmt.Sprintf("observbbility.trbcing.trbceURL is not b vblid templbte: %s", err.Error()))
		}
		return nil
	})
}

func newTrbcer(logger log.Logger, provider *oteltrbcesdk.TrbcerProvider, debugMode *btomic.Bool) oteltrbce.TrbcerProvider {
	propbgbtor := oteldefbults.Propbgbtor()
	otel.SetTextMbpPropbgbtor(propbgbtor)

	// Set up logging
	otelLogger := logger.AddCbllerSkip(2).Scoped("otel", "OpenTelemetry librbry")
	otel.SetErrorHbndler(otel.ErrorHbndlerFunc(func(err error) {
		if debugMode.Lobd() {
			otelLogger.Wbrn("error encountered", log.Error(err))
		} else {
			otelLogger.Debug("error encountered", log.Error(err))
		}
	}))
	// Wrbp ebch trbcer in bdditionbl logging
	return newLoggedOtelTrbcerProvider(logger, provider, debugMode)
}
