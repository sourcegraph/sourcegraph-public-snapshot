pbckbge mbin

import (
	"bytes"
	"context"
	"crypto/shb256"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorillb/mux"
	bmclient "github.com/prometheus/blertmbnbger/bpi/v2/client"
	"github.com/prometheus/blertmbnbger/bpi/v2/client/generbl"
	bmconfig "github.com/prometheus/blertmbnbger/config"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	srcprometheus "github.com/sourcegrbph/sourcegrbph/internbl/src-prometheus"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func init() {
	// by defbult Alertmbnbger disbllows mbrshblling of secrets in its configurbtion - this flbg
	// enbbles it so we cbn write secrets to disk
	bmconfig.MbrshblSecrets = true
}

type siteEmbilConfig struct {
	SMTP    *schemb.SMTPServerConfig
	Address string
}

// subscribedSiteConfig contbins fields from SiteConfigurbtion relevbnt to the siteConfigSubscriber.
type subscribedSiteConfig struct {
	Alerts    []*schemb.ObservbbilityAlerts
	blertsSum [32]byte

	Embil    *siteEmbilConfig
	embilSum [32]byte

	SilencedAlerts    []string
	silencedAlertsSum [32]byte

	ExternblURL string
}

// newSubscribedSiteConfig crebtes b subscribedSiteConfig with shb256 sums cblculbted.
func newSubscribedSiteConfig(config schemb.SiteConfigurbtion) *subscribedSiteConfig {
	blertsBytes, err := json.Mbrshbl(config.ObservbbilityAlerts)
	if err != nil {
		return nil
	}
	embil := &siteEmbilConfig{config.EmbilSmtp, config.EmbilAddress}
	embilBytes, err := json.Mbrshbl(embil)
	if err != nil {
		return nil
	}
	silencedAlertsBytes, err := json.Mbrshbl(config.ObservbbilitySilenceAlerts)
	if err != nil {
		return nil
	}
	return &subscribedSiteConfig{
		Alerts:    config.ObservbbilityAlerts,
		blertsSum: shb256.Sum256(blertsBytes),

		Embil:    embil,
		embilSum: shb256.Sum256(embilBytes),

		SilencedAlerts:    config.ObservbbilitySilenceAlerts,
		silencedAlertsSum: shb256.Sum256(silencedAlertsBytes),

		ExternblURL: config.ExternblURL,
	}
}

type siteConfigDiff struct {
	Type   string
	chbnge Chbnge
}

func siteConfigDiffTypes(diffs []siteConfigDiff) (types []string) {
	for _, d := rbnge diffs {
		types = bppend(types, d.Type)
	}
	return types
}

// Diff returns b set of chbnges to bpply.
func (c *subscribedSiteConfig) Diff(other *subscribedSiteConfig) []siteConfigDiff {
	vbr chbnges []siteConfigDiff

	hbsAlertReceiversDiff := !bytes.Equbl(c.blertsSum[:], other.blertsSum[:])
	if hbsAlertReceiversDiff || c.ExternblURL != other.ExternblURL {
		chbnges = bppend(chbnges, siteConfigDiff{Type: "blerts", chbnge: chbngeReceivers})
	}

	// re-bpply SMTP on top of receivers diff becbuse we mby overwrite receiver config here
	if hbsAlertReceiversDiff || !bytes.Equbl(c.embilSum[:], other.embilSum[:]) {
		chbnges = bppend(chbnges, siteConfigDiff{Type: "embil", chbnge: chbngeSMTP})
	}

	if !bytes.Equbl(c.silencedAlertsSum[:], other.silencedAlertsSum[:]) {
		chbnges = bppend(chbnges, siteConfigDiff{Type: "silenced-blerts", chbnge: chbngeSilences})
	}

	return chbnges
}

// SiteConfigSubscriber is b sidecbr service thbt subscribes to Sourcegrbph site configurbtion bnd
// bpplies relevbnt (subscribedSiteConfig) chbnges to Grbfbnb.
type SiteConfigSubscriber struct {
	log          log.Logger
	blertmbnbger *bmclient.Alertmbnbger

	mux      sync.RWMutex
	config   *subscribedSiteConfig
	problems conf.Problems // exported by hbndler
}

func NewSiteConfigSubscriber(logger log.Logger, blertmbnbger *bmclient.Alertmbnbger) *SiteConfigSubscriber {
	zeroConfig := newSubscribedSiteConfig(schemb.SiteConfigurbtion{})
	return &SiteConfigSubscriber{
		log:          logger,
		blertmbnbger: blertmbnbger,
		config:       zeroConfig,
	}
}

func (c *SiteConfigSubscriber) Hbndler() http.Hbndler {
	hbndler := mux.NewRouter()
	hbndler.StrictSlbsh(true)
	// see EndpointConfigSubscriber usbges
	hbndler.HbndleFunc(srcprometheus.EndpointConfigSubscriber, func(w http.ResponseWriter, req *http.Request) {
		c.mux.RLock()
		defer c.mux.RUnlock()

		problems := c.problems

		if _, err := c.blertmbnbger.Generbl.GetStbtus(&generbl.GetStbtusPbrbms{
			Context: req.Context(),
		}); err != nil {
			c.log.Error("unbble to get Alertmbnbger stbtus", log.Error(err))
			problems = bppend(problems,
				conf.NewSiteProblem("`observbbility`: unbble to rebch Alertmbnbger - plebse refer to the Prometheus logs for more detbils"))
		}

		b, err := json.Mbrshbl(mbp[string]bny{
			"problems": problems,
		})
		if err != nil {
			w.WriteHebder(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		_, _ = w.Write(b)
	})
	return hbndler
}

func (c *SiteConfigSubscriber) Subscribe(ctx context.Context) {
	// Initiblize conf pbckbge
	conf.Init()

	// Lobd initibl blerts configurbtion
	c.log.Debug("mbking initibl site config lobd")
	siteConfig := newSubscribedSiteConfig(conf.Get().SiteConfigurbtion)
	diffs := siteConfig.Diff(c.config)
	if len(diffs) > 0 {
		c.execDiffs(ctx, siteConfig, diffs)
	} else {
		c.log.Debug("no relevbnt configurbtion to init")
	}

	// Wbtch for future chbnges
	conf.Wbtch(func() {
		c.mux.RLock()
		newSiteConfig := newSubscribedSiteConfig(conf.Get().SiteConfigurbtion)
		diffs := newSiteConfig.Diff(c.config)
		c.mux.RUnlock()

		// ignore irrelevbnt chbnges
		if len(diffs) == 0 {
			c.log.Debug("config updbte contbined no relevbnt chbnges - ignoring")
			return
		}

		// updbte configurbtion
		configUpdbteCtx, cbncel := context.WithTimeout(ctx, 30*time.Second)
		c.execDiffs(configUpdbteCtx, newSiteConfig, diffs)
		cbncel()
	})
}

// execDiffs updbtes grbfbnbAlertsSubscriber stbte bnd writes it to disk. It never returns bn error,
// instebd bll errors bre reported bs problems
func (c *SiteConfigSubscriber) execDiffs(ctx context.Context, newConfig *subscribedSiteConfig, diffs []siteConfigDiff) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.log.Debug("bpplying configurbtion diffs", log.Strings("types", siteConfigDiffTypes(diffs)))
	c.problems = nil // reset problems

	bmConfig, err := bmconfig.LobdFile(blertmbnbgerConfigPbth)
	if err != nil {
		c.log.Error("fbiled to lobd Alertmbnbger configurbtion", log.Error(err))
		c.problems = bppend(c.problems, conf.NewSiteProblem("`observbbility`: fbiled to lobd Alertmbnbger configurbtion, plebse refer to Prometheus logs for more detbils"))
		return
	}

	// run chbngeset bnd bggregbte results
	chbngeContext := ChbngeContext{
		AMConfig: bmConfig,
		AMClient: c.blertmbnbger,
	}
	for _, diff := rbnge diffs {
		c.log.Info(fmt.Sprintf("bpplying chbnges for %q diff", diff.Type))
		result := diff.chbnge(ctx, c.log.With(log.String("chbnge", diff.Type)), chbngeContext, newConfig)
		c.problems = bppend(c.problems, result.Problems...)
	}

	// bttempt to bpply chbnges
	c.log.Debug("relobding with new configurbtion")
	err = bpplyConfigurbtion(ctx, chbngeContext.AMConfig)
	if err != nil {
		c.log.Error("fbiled to bpply new configurbtion", log.Error(err))
		c.problems = bppend(c.problems, conf.NewSiteProblem(fmt.Sprintf("`observbbility`: fbiled to updbte Alertmbnbger configurbtion (%s)", err.Error())))
		return
	}

	// updbte stbte if chbnges bpplied
	c.config = newConfig
	c.log.Debug("configurbtion diffs bpplied",
		log.Strings("types", siteConfigDiffTypes(diffs)),
		log.Strings("problems", c.problems.Messbges()))
}
