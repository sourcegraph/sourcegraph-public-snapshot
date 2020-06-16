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

type subscribedSiteConfig struct {
	Alerts    []*schema.ObservabilityAlerts
	alertsSum [32]byte

	SMTP    *schema.SMTPServerConfig
	smtpSum [32]byte
}

func newSubscribedSiteConfig(config schema.SiteConfiguration) *subscribedSiteConfig {
	alertsBytes, err := json.Marshal(config.ObservabilityAlerts)
	if err != nil {
		return nil
	}
	smtpBytes, err := json.Marshal(config.EmailSmtp)
	if err != nil {
		return nil
	}
	return &subscribedSiteConfig{
		Alerts:    config.ObservabilityAlerts,
		alertsSum: sha256.Sum256(alertsBytes),

		SMTP:    config.EmailSmtp,
		smtpSum: sha256.Sum256(smtpBytes),
	}
}

func (c *subscribedSiteConfig) Sum() []byte {
	return append(c.alertsSum[:], c.smtpSum[:]...)
}

func (c *subscribedSiteConfig) Diff(other *subscribedSiteConfig) []GrafanaChange {
	var changes []GrafanaChange
	if !bytes.Equal(c.alertsSum[:], other.alertsSum[:]) {
		changes = append(changes, grafanaChangeNotifiers)
	}
	if !bytes.Equal(c.smtpSum[:], other.smtpSum[:]) {
		changes = append(changes, grafanaChangeSMTP)
	}
	return changes
}

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

	// Need grafana to be ready to intialize alerts
	log.Info("waiting for grafana")
	if err := grafana.WaitForServer(ctx); err != nil {
		return nil, err
	}
	log.Debug("detected grafana ready")

	// Load initial alerts configuration
	siteConfig := newSubscribedSiteConfig(conf.Get().SiteConfiguration)
	sum := siteConfig.Sum()

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

	subscriber := &siteConfigSubscriber{log: log, grafana: grafana}
	subscriber.updateGrafanaConfig(ctx, siteConfig, sum)
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
		newSum := newSiteConfig.Sum()
		isUnchanged := bytes.Equal(c.configSum, newSum)
		c.mux.RUnlock()

		// ignore irrelevant changes
		if isUnchanged {
			c.log.Debug("config updated contained no relevant changes - ignoring")
			return
		}

		// update configuration
		c.mux.Lock()
		configUpdateCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		c.updateGrafanaConfig(configUpdateCtx, newSiteConfig, newSum)
		cancel()
		c.mux.Unlock()
	})
}

// updateGrafanaConfig updates grafanaAlertsSubscriber state and writes it to disk. It never returns an error,
// instead all errors are reported as problems
func (c *siteConfigSubscriber) updateGrafanaConfig(ctx context.Context, newConfig *subscribedSiteConfig, newSum []byte) {
	c.log.Debug("updating grafana configuration")
	c.problems = nil

	// run changeset and aggregate results
	aggregated := GrafanaChangeResult{}
	grafanaChanges := c.config.Diff(newConfig)
	for _, change := range grafanaChanges {
		result := change(ctx, c.log, c.grafana.Client, c.config, newConfig)
		aggregated.Problems = append(aggregated.Problems, result.Problems...)
		if result.ShouldRestartGrafana {
			aggregated.ShouldRestartGrafana = true
		}
	}

	// restart if needed
	if aggregated.ShouldRestartGrafana {
		// TODO what do if fail
		newFailedToRestartProblem := func(e error) *conf.Problem {
			return conf.NewSiteProblem(fmt.Sprintf("observability: Grafana failed to restart for configuration changes: %v", e))
		}
		if err := c.grafana.RunServer(); err != nil {
			aggregated.Problems = append(aggregated.Problems, newFailedToRestartProblem(err))
			return
		}
		if err := c.grafana.WaitForServer(ctx); err != nil {
			aggregated.Problems = append(aggregated.Problems, newFailedToRestartProblem(err))
			return
		}
	}

	// update state
	c.config = newConfig
	c.configSum = newSum
	c.problems = aggregated.Problems
	c.log.Debug("updated grafana configuration")
}
