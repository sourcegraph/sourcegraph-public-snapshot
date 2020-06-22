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

	"github.com/grafana-tools/sdk"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

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

	// flag to indicate this config is for container startup
	isStartup bool
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
	return &subscribedSiteConfig{
		Alerts:    config.ObservabilityAlerts,
		alertsSum: sha256.Sum256(alertsBytes),

		Email:    email,
		emailSum: sha256.Sum256(emailBytes),
	}
}

type siteConfigDiff struct {
	Type   string
	change GrafanaChange
}

// Diff returns a set of changes to apply to Grafana. If the provided config has isStartup=true,
// it is assumed that this diff is for initial Grafana startup.
func (c *subscribedSiteConfig) Diff(other *subscribedSiteConfig) []siteConfigDiff {
	var changes []siteConfigDiff

	// apply notifer changes on startup, since they can persist from Grafana's database
	if other.isStartup || !bytes.Equal(c.alertsSum[:], other.alertsSum[:]) {
		changes = append(changes, siteConfigDiff{Type: "alerts", change: grafanaChangeNotifiers})
	}

	if !bytes.Equal(c.emailSum[:], other.emailSum[:]) {
		changes = append(changes, siteConfigDiff{Type: "email", change: grafanaChangeSMTP})
	}

	return changes
}

// siteConfigSubscriber is a sidecar service that subscribes to Sourcegraph site configuration and
// applies relevant (subscribedSiteConfig) changes to Grafana.
type siteConfigSubscriber struct {
	log     log15.Logger
	grafana *grafanaController

	mux       sync.RWMutex
	config    *subscribedSiteConfig
	configSum []byte
	problems  conf.Problems // exported by handler
}

func newSiteConfigSubscriber(ctx context.Context, logger log15.Logger, grafana *grafanaController) (*siteConfigSubscriber, error) {
	log := logger.New("logger", "config-subscriber")

	// Syncing relies on access to frontend, so wait until it is ready
	log.Info("waiting for frontend", "url", api.InternalClient.URL)
	if err := api.InternalClient.WaitForFrontend(ctx); err != nil {
		return nil, err
	}
	log.Debug("detected frontend ready")

	// Need grafana to be ready to initialize alerts
	log.Info("waiting for grafana")
	if err := grafana.WaitForServer(ctx); err != nil {
		return nil, err
	}
	log.Debug("detected grafana ready")

	// Load initial alerts configuration
	siteConfig := newSubscribedSiteConfig(conf.Get().SiteConfiguration)

	// Set up overview dashboard if it does not exist. We attach alerts to a copy of the
	// default home dashboard, because dashboards provisioned from disk cannot be edited.
	if _, _, err := grafana.GetDashboardByUID(ctx, overviewDashboardUID); err != nil {
		homeBoard, err := getOverviewDashboard()
		if err != nil {
			return nil, fmt.Errorf("failed to generate alerts overview dashboard: %w", err)
		}
		if _, err := grafana.SetDashboard(ctx, *homeBoard, sdk.SetDashboardParams{}); err != nil {
			return nil, fmt.Errorf("failed to set up alerts overview dashboard: %w", err)
		}
	}

	// set initial grafana state using a zero-value site config
	subscriber := &siteConfigSubscriber{log: log, grafana: grafana}
	zeroConfig := newSubscribedSiteConfig(schema.SiteConfiguration{})
	zeroConfig.isStartup = true
	subscriber.updateGrafanaConfig(ctx, siteConfig, siteConfig.Diff(zeroConfig))

	return subscriber, nil
}

func (c *siteConfigSubscriber) Handler() http.Handler {
	handler := http.NewServeMux()
	handler.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		c.mux.RLock()
		defer c.mux.RUnlock()

		b, err := json.Marshal(map[string]interface{}{
			"problems": c.problems,
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

func (c *siteConfigSubscriber) Subscribe(ctx context.Context) {
	conf.Watch(func() {
		c.mux.RLock()
		newSiteConfig := newSubscribedSiteConfig(conf.Get().SiteConfiguration)
		diffs := newSiteConfig.Diff(c.config)
		c.mux.RUnlock()

		// ignore irrelevant changes
		if len(diffs) == 0 {
			c.log.Debug("config updated contained no relevant changes - ignoring")
			return
		}

		// update configuration
		configUpdateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		c.updateGrafanaConfig(configUpdateCtx, newSiteConfig, diffs)
		cancel()
	})
}

// updateGrafanaConfig updates grafanaAlertsSubscriber state and writes it to disk. It never returns an error,
// instead all errors are reported as problems
func (c *siteConfigSubscriber) updateGrafanaConfig(ctx context.Context, newConfig *subscribedSiteConfig, diffs []siteConfigDiff) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.log.Debug("updating grafana configuration", "diffs", diffs)
	c.problems = nil

	// load grafana config
	grafanaConfig, err := getGrafanaConfig(grafanaConfigPath)
	if err != nil {
		c.problems = append(c.problems,
			conf.NewSiteProblem(fmt.Sprintf("observability: failed to load Grafana configuration: %v", err)))
		return
	}

	// run changeset and aggregate results
	configChange := false
	for _, diff := range diffs {
		c.log.Info(fmt.Sprintf("applying changes for %q diff", diff.Type))
		result := diff.change(ctx, c.log, GrafanaContext{
			Client: c.grafana.Client,
			Config: grafanaConfig,
		}, newConfig)
		c.problems = append(c.problems, result.Problems...)
		if result.ConfigChange {
			configChange = true
		}
	}

	// restart if needed
	if configChange {
		// TODO what do if fail
		if err := grafanaConfig.SaveTo(grafanaConfigPath); err != nil {
			c.problems = append(c.problems, conf.NewSiteProblem(fmt.Sprintf("failed to save Grafana config: %v", err)))
			return
		}

		newFailedToRestartProblem := func(e error) *conf.Problem {
			return conf.NewSiteProblem(fmt.Sprintf("observability: Grafana failed to restart for configuration changes: %v", e))
		}
		if err := c.grafana.Stop(); err != nil {
			c.problems = append(c.problems, newFailedToRestartProblem(err))
		}
		if err := c.grafana.RunServer(); err != nil {
			c.problems = append(c.problems, newFailedToRestartProblem(err))
		} else if err := c.grafana.WaitForServer(ctx); err != nil {
			c.problems = append(c.problems, newFailedToRestartProblem(err))
		}
	}

	// update state
	c.config = newConfig
	c.log.Debug("updated grafana configuration", "diffs", diffs)
}
