package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	amclient "github.com/prometheus/alertmanager/api/v2/client"
	"github.com/prometheus/alertmanager/api/v2/client/general"
	amconfig "github.com/prometheus/alertmanager/config"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	srcprometheus "github.com/sourcegraph/sourcegraph/internal/src-prometheus"
	"github.com/sourcegraph/sourcegraph/schema"
)

func init() {
	// by default Alertmanager disallows marshalling of secrets in its configuration - this flag
	// enables it so we can write secrets to disk
	amconfig.MarshalSecrets = true
}

type siteEmailConfig struct {
	SMTP    *schema.SMTPServerConfig
	Address string
}

// subscribedSiteConfig contains fields from SiteConfiguration relevant to the siteConfigSubscriber.
type subscribedSiteConfig struct {
	Alerts    []*schema.ObservabilityAlerts
	alertsSum [32]byte

	Email    *siteEmailConfig
	emailSum [32]byte

	SilencedAlerts    []string
	silencedAlertsSum [32]byte

	ExternalURL string
}

// newSubscribedSiteConfig creates a subscribedSiteConfig with sha256 sums calculated.
func newSubscribedSiteConfig(config schema.SiteConfiguration) *subscribedSiteConfig {
	alertsBytes, err := json.Marshal(config.ObservabilityAlerts)
	if err != nil {
		return nil
	}
	email := &siteEmailConfig{config.EmailSmtp, config.EmailAddress}
	emailBytes, err := json.Marshal(email)
	if err != nil {
		return nil
	}
	silencedAlertsBytes, err := json.Marshal(config.ObservabilitySilenceAlerts)
	if err != nil {
		return nil
	}
	return &subscribedSiteConfig{
		Alerts:    config.ObservabilityAlerts,
		alertsSum: sha256.Sum256(alertsBytes),

		Email:    email,
		emailSum: sha256.Sum256(emailBytes),

		SilencedAlerts:    config.ObservabilitySilenceAlerts,
		silencedAlertsSum: sha256.Sum256(silencedAlertsBytes),

		ExternalURL: config.ExternalURL,
	}
}

type siteConfigDiff struct {
	Type   string
	change Change
}

func siteConfigDiffTypes(diffs []siteConfigDiff) (types []string) {
	for _, d := range diffs {
		types = append(types, d.Type)
	}
	return types
}

// Diff returns a set of changes to apply.
func (c *subscribedSiteConfig) Diff(other *subscribedSiteConfig) []siteConfigDiff {
	var changes []siteConfigDiff

	hasAlertReceiversDiff := !bytes.Equal(c.alertsSum[:], other.alertsSum[:])
	if hasAlertReceiversDiff || c.ExternalURL != other.ExternalURL {
		changes = append(changes, siteConfigDiff{Type: "alerts", change: changeReceivers})
	}

	// re-apply SMTP on top of receivers diff because we may overwrite receiver config here
	if hasAlertReceiversDiff || !bytes.Equal(c.emailSum[:], other.emailSum[:]) {
		changes = append(changes, siteConfigDiff{Type: "email", change: changeSMTP})
	}

	if !bytes.Equal(c.silencedAlertsSum[:], other.silencedAlertsSum[:]) {
		changes = append(changes, siteConfigDiff{Type: "silenced-alerts", change: changeSilences})
	}

	return changes
}

// SiteConfigSubscriber is a sidecar service that subscribes to Sourcegraph site configuration and
// applies relevant (subscribedSiteConfig) changes to Grafana.
type SiteConfigSubscriber struct {
	log          log.Logger
	alertmanager *amclient.Alertmanager

	mux      sync.RWMutex
	config   *subscribedSiteConfig
	problems conf.Problems // exported by handler
}

func NewSiteConfigSubscriber(logger log.Logger, alertmanager *amclient.Alertmanager) *SiteConfigSubscriber {
	zeroConfig := newSubscribedSiteConfig(schema.SiteConfiguration{})
	return &SiteConfigSubscriber{
		log:          logger,
		alertmanager: alertmanager,
		config:       zeroConfig,
	}
}

func (c *SiteConfigSubscriber) Handler() http.Handler {
	handler := mux.NewRouter()
	handler.StrictSlash(true)
	// see EndpointConfigSubscriber usages
	handler.HandleFunc(srcprometheus.EndpointConfigSubscriber, func(w http.ResponseWriter, req *http.Request) {
		c.mux.RLock()
		defer c.mux.RUnlock()

		problems := c.problems

		if _, err := c.alertmanager.General.GetStatus(&general.GetStatusParams{
			Context: req.Context(),
		}); err != nil {
			c.log.Error("unable to get Alertmanager status", log.Error(err))
			problems = append(problems,
				conf.NewSiteProblem("`observability`: unable to reach Alertmanager - please refer to the Prometheus logs for more details"))
		}

		b, err := json.Marshal(map[string]any{
			"problems": problems,
		})
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte(err.Error()))
			return
		}
		_, _ = w.Write(b)
	})
	return handler
}

func (c *SiteConfigSubscriber) Subscribe(ctx context.Context) {
	// Initialize conf package
	conf.Init()

	// Load initial alerts configuration
	c.log.Debug("making initial site config load")
	siteConfig := newSubscribedSiteConfig(conf.Get().SiteConfiguration)
	diffs := siteConfig.Diff(c.config)
	if len(diffs) > 0 {
		c.execDiffs(ctx, siteConfig, diffs)
	} else {
		c.log.Debug("no relevant configuration to init")
	}

	// Watch for future changes
	conf.Watch(func() {
		c.mux.RLock()
		newSiteConfig := newSubscribedSiteConfig(conf.Get().SiteConfiguration)
		diffs := newSiteConfig.Diff(c.config)
		c.mux.RUnlock()

		// ignore irrelevant changes
		if len(diffs) == 0 {
			c.log.Debug("config update contained no relevant changes - ignoring")
			return
		}

		// update configuration
		configUpdateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		c.execDiffs(configUpdateCtx, newSiteConfig, diffs)
		cancel()
	})
}

// execDiffs updates grafanaAlertsSubscriber state and writes it to disk. It never returns an error,
// instead all errors are reported as problems
func (c *SiteConfigSubscriber) execDiffs(ctx context.Context, newConfig *subscribedSiteConfig, diffs []siteConfigDiff) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.log.Debug("applying configuration diffs", log.Strings("types", siteConfigDiffTypes(diffs)))
	c.problems = nil // reset problems

	amConfig, err := amconfig.LoadFile(alertmanagerConfigPath)
	if err != nil {
		c.log.Error("failed to load Alertmanager configuration", log.Error(err))
		c.problems = append(c.problems, conf.NewSiteProblem("`observability`: failed to load Alertmanager configuration, please refer to Prometheus logs for more details"))
		return
	}

	// run changeset and aggregate results
	changeContext := ChangeContext{
		AMConfig: amConfig,
		AMClient: c.alertmanager,
	}
	for _, diff := range diffs {
		c.log.Info(fmt.Sprintf("applying changes for %q diff", diff.Type))
		result := diff.change(ctx, c.log.With(log.String("change", diff.Type)), changeContext, newConfig)
		c.problems = append(c.problems, result.Problems...)
	}

	// attempt to apply changes
	c.log.Debug("reloading with new configuration")
	err = applyConfiguration(ctx, changeContext.AMConfig)
	if err != nil {
		c.log.Error("failed to apply new configuration", log.Error(err))
		c.problems = append(c.problems, conf.NewSiteProblem(fmt.Sprintf("`observability`: failed to update Alertmanager configuration (%s)", err.Error())))
		return
	}

	// update state if changes applied
	c.config = newConfig
	c.log.Debug("configuration diffs applied",
		log.Strings("types", siteConfigDiffTypes(diffs)),
		log.Strings("problems", c.problems.Messages()))
}
