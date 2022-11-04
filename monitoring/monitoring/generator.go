package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/grafana-tools/sdk"
	"gopkg.in/yaml.v2"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const (
	localGrafanaURL         = "http://127.0.0.1:3370"
	localGrafanaCredentials = "admin:admin"

	localPrometheusURL = "http://127.0.0.1:9090"
)

// GenerateOptions declares options for the monitoring generator.
type GenerateOptions struct {
	// Toggles pruning of dangling generated assets through simple heuristic, should be disabled during builds
	DisablePrune bool
	// Trigger reload of active Prometheus or Grafana instance (requires respective output directories)
	Reload bool

	// Output directory for generated Grafana assets
	GrafanaDir string
	// Output directory for generated Prometheus assets
	PrometheusDir string
	// Output directory for generated documentation
	DocsDir string
}

// Generate is the main Sourcegraph monitoring generator entrypoint.
func Generate(logger log.Logger, opts GenerateOptions, dashboards ...*Dashboard) error {
	logger.Info("Regenerating monitoring",
		log.String("options", fmt.Sprintf("%+v", opts)),
		log.Int("dashboards", len(dashboards)))

	var generatedAssets []string

	// Verify dashboard configuration
	var validationErrors error
	for _, dashboard := range dashboards {
		if err := dashboard.validate(); err != nil {
			validationErrors = errors.Append(validationErrors,
				errors.Wrapf(err, "Invalid dashboard %q", dashboard.Name))
		}
	}
	if validationErrors != nil {
		return errors.Wrap(validationErrors, "Validation failed")
	}
	dlog := logger.Scoped("grafana", "grafana dashboard generation")
	// Generate output for all dashboards
	for _, dashboard := range dashboards {
		// Logger for dashboard
		clog := dlog.With(log.String("dashboard", dashboard.Name), log.String("instance", localGrafanaURL))
		// Prepare Grafana assets
		if opts.GrafanaDir != "" {
			clog.Debug("Rendering Grafana assets")
			board := dashboard.renderDashboard()
			data, err := json.MarshalIndent(board, "", "  ")
			if err != nil {
				return errors.Wrapf(err, "Invalid dashboard %q", dashboard.Name)
			}
			// #nosec G306  prometheus runs as nobody
			generatedDashboard := dashboard.Name + ".json"
			err = os.WriteFile(filepath.Join(opts.GrafanaDir, generatedDashboard), data, os.ModePerm)
			if err != nil {
				return errors.Wrapf(err, "Could not write dashboard %q to output", dashboard.Name)
			}
			generatedAssets = append(generatedAssets, generatedDashboard)

			// Reload specific dashboard
			if opts.Reload {
				clog.Debug("Reloading Grafana instance")
				client, err := sdk.NewClient(localGrafanaURL, localGrafanaCredentials, sdk.DefaultHTTPClient)
				if err != nil {
					return errors.Wrapf(err, "Failed to initialize Grafana client for dashboard %q", dashboard.Title)
				}
				_, err = client.SetDashboard(context.Background(), *board, sdk.SetDashboardParams{Overwrite: true})
				if err != nil {
					return errors.Wrapf(err, "Could not reload Grafana instance for dashboard %q", dashboard.Title)
				} else {
					clog.Info("Reloaded Grafana instance")
				}
			}
		}

		// Prepare Prometheus assets
		if opts.PrometheusDir != "" {
			clog.Debug("Rendering Prometheus assets")
			promAlertsFile, err := dashboard.renderRules()
			if err != nil {
				return errors.Wrapf(err, "Unable to generate alerts for dashboard %q", dashboard.Title)
			}
			data, err := yaml.Marshal(promAlertsFile)
			if err != nil {
				return errors.Wrapf(err, "Invalid rules for dashboard %q", dashboard.Title)
			}
			fileName := strings.ReplaceAll(dashboard.Name, "-", "_") + alertRulesFileSuffix
			generatedAssets = append(generatedAssets, fileName)
			err = os.WriteFile(filepath.Join(opts.PrometheusDir, fileName), data, os.ModePerm)
			if err != nil {
				return errors.Wrapf(err, "Could not write rules to output for dashboard %q", dashboard.Title)
			}
		}
	}

	// Reload all Prometheus rules
	if opts.PrometheusDir != "" && opts.Reload {
		rlog := logger.Scoped("prometheus", "prometheus alerts generation").With(log.String("instance", localPrometheusURL))
		// Reload all Prometheus rules
		rlog.Debug("Reloading Prometheus instance")
		resp, err := http.Post(localPrometheusURL+"/-/reload", "", nil)
		if err != nil {
			return errors.Wrapf(err, "Could not reload Prometheus at %q", localPrometheusURL)
		} else {
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return errors.Newf("Unexpected status code %d while reloading Prometheus rules", resp.StatusCode)
			}
			rlog.Info("Reloaded Prometheus instance")
		}
	}

	// Generate documentation
	if opts.DocsDir != "" {
		logger.Debug("Rendering docs")
		docs, err := renderDocumentation(dashboards)
		if err != nil {
			return errors.Wrap(err, "Failed to generate docs")
		}
		for _, docOut := range []struct {
			path string
			data []byte
		}{
			{path: filepath.Join(opts.DocsDir, alertsDocsFile), data: docs.alertDocs.Bytes()},
			{path: filepath.Join(opts.DocsDir, dashboardsDocsFile), data: docs.dashboards.Bytes()},
		} {
			err = os.WriteFile(docOut.path, docOut.data, os.ModePerm)
			if err != nil {
				return errors.Wrapf(err, "Could not write docs to path %q", docOut.path)
			}
			generatedAssets = append(generatedAssets, docOut.path)
		}
	}

	// Clean up dangling assets
	if !opts.DisablePrune {
		logger.Debug("Pruning dangling assets")
		if err := pruneAssets(logger, generatedAssets, opts.GrafanaDir, opts.PrometheusDir); err != nil {
			return errors.Wrap(err, "Failed to prune assets, resolve manually or disable pruning")
		}
	}

	return nil
}
