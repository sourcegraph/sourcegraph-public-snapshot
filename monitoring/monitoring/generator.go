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
	"github.com/prometheus/prometheus/model/labels"
	"gopkg.in/yaml.v2"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring/internal/grafana"
)

// GenerateOptions declares options for the monitoring generator.
type GenerateOptions struct {
	// Toggles pruning of dangling generated assets through simple heuristic, should be disabled during builds
	DisablePrune bool
	// Trigger reload of active Prometheus or Grafana instance (requires respective output directories)
	Reload bool

	// Output directory for generated Grafana assets
	GrafanaDir string
	// GrafanaURL is the address for the Grafana instance to reload
	GrafanaURL string
	// GrafanaCredentials is the basic auth credentials for the Grafana instance at
	// GrafanaURL, e.g. "admin:admin"
	GrafanaCredentials string

	// Output directory for generated Prometheus assets
	PrometheusDir string
	// PrometheusURL is the address for the Prometheus instance to reload
	PrometheusURL string

	// Output directory for generated documentation
	DocsDir string

	// InjectLabelMatchers specifies labels to inject into all selectors in Prometheus
	// expressions - this includes dashboard template variables, observable queries,
	// alert queries, and so on - using internal/promql.Inject(...).
	InjectLabelMatchers []*labels.Matcher
}

// Generate is the main Sourcegraph monitoring generator entrypoint.
func Generate(logger log.Logger, opts GenerateOptions, dashboards ...*Dashboard) error {
	logger.Info("Regenerating monitoring", log.String("options", fmt.Sprintf("%+v", opts)))

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

	// Generate Grafana content for all dashboards
	for _, dashboard := range dashboards {
		// Logger for dashboard
		dlog := logger.With(log.String("dashboard", dashboard.Name))

		// Prepare Grafana assets
		if opts.GrafanaDir != "" {
			glog := dlog.Scoped("grafana", "grafana dashboard generation").
				With(log.String("instance", opts.GrafanaURL))
			os.MkdirAll(opts.GrafanaDir, os.ModePerm)

			glog.Debug("Rendering Grafana assets")
			board, err := dashboard.renderDashboard(opts.InjectLabelMatchers)
			if err != nil {
				return errors.Wrapf(err, "Failed to render dashboard %q", dashboard.Name)
			}
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
				glog.Debug("Reloading Grafana instance")
				client, err := sdk.NewClient(opts.GrafanaURL, opts.GrafanaCredentials, sdk.DefaultHTTPClient)
				if err != nil {
					return errors.Wrapf(err, "Failed to initialize Grafana client for dashboard %q", dashboard.Title)
				}
				_, err = client.SetDashboard(context.Background(), *board, sdk.SetDashboardParams{Overwrite: true})
				if err != nil {
					return errors.Wrapf(err, "Could not reload Grafana instance for dashboard %q", dashboard.Title)
				} else {
					glog.Info("Reloaded Grafana instance")
				}
			}
		}

		// Generate home page
		if opts.GrafanaDir != "" {
			data, err := grafana.Home(opts.InjectLabelMatchers)
			if err != nil {
				return errors.Wrap(err, "failed to generate home dashboard")
			}
			generatedDashboard := "home.json"
			generatedAssets = append(generatedAssets, generatedDashboard)
			if err = os.WriteFile(filepath.Join(opts.GrafanaDir, generatedDashboard), data, os.ModePerm); err != nil {
				return errors.Wrap(err, "failed to generate home dashboard")
			}
		}

		// Prepare Prometheus assets
		if opts.PrometheusDir != "" {
			plog := dlog.Scoped("prometheus", "prometheus rules generation")

			os.MkdirAll(opts.PrometheusDir, os.ModePerm)

			plog.Debug("Rendering Prometheus assets")
			promAlertsFile, err := dashboard.renderRules(opts.InjectLabelMatchers)
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

	// Generate additional Prometheus assets
	if opts.PrometheusDir != "" {
		customRules, err := customPrometheusRules(opts.InjectLabelMatchers)
		if err != nil {
			return errors.Wrap(err, "failed to generate custom rules")
		}
		data, err := yaml.Marshal(customRules)
		if err != nil {
			return errors.Wrapf(err, "Invalid custom rules")
		}
		fileName := "src_custom_rules.yml"
		generatedAssets = append(generatedAssets, fileName)
		err = os.WriteFile(filepath.Join(opts.PrometheusDir, fileName), data, os.ModePerm)
		if err != nil {
			return errors.Wrap(err, "Could not write custom rules")
		}
	}

	// Reload all Prometheus rules
	if opts.PrometheusDir != "" && opts.Reload {
		rlog := logger.Scoped("prometheus", "prometheus alerts generation").
			With(log.String("instance", opts.PrometheusURL))
		// Reload all Prometheus rules
		rlog.Debug("Reloading Prometheus instance")
		resp, err := http.Post(opts.PrometheusURL+"/-/reload", "", nil)
		if err != nil {
			return errors.Wrapf(err, "Could not reload Prometheus at %q", opts.PrometheusURL)
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
		os.MkdirAll(opts.DocsDir, os.ModePerm)

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
