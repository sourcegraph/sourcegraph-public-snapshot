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

	grafanasdk "github.com/grafana-tools/sdk"
	"github.com/prometheus/prometheus/model/labels"
	"gopkg.in/yaml.v2"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring/internal/grafana"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring/internal/headertransport"
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
	// GrafanaHeaders are additional HTTP headers to add to all requests to the target Grafana instance
	GrafanaHeaders map[string]string
	// GrafanaFolder is the folder on the destination Grafana instance to upload the dashboards to
	GrafanaFolder string

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
	ctx := context.TODO()

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

	// Set up disk directories
	if opts.GrafanaDir != "" {
		os.MkdirAll(opts.GrafanaDir, os.ModePerm)
	}
	if opts.PrometheusDir != "" {
		os.MkdirAll(opts.PrometheusDir, os.ModePerm)
	}
	if opts.DocsDir != "" {
		os.MkdirAll(opts.DocsDir, os.ModePerm)
	}

	// Generate Grafana content for all dashboards. If grafanaClient is not nil, Grafana
	// should be reloaded.
	var grafanaClient *grafanasdk.Client
	var grafanaFolderID int
	if opts.GrafanaURL != "" && opts.Reload {
		gclog := logger.Scoped("grafana.client", "grafana client setup")

		// DefaultHTTPClient is used unless additional headers are requested
		httpClient := grafanasdk.DefaultHTTPClient
		if len(opts.GrafanaHeaders) > 0 {
			gclog.Debug("Adding additional headers to Grafana requests",
				log.String("headers", fmt.Sprintf("%v", opts.GrafanaHeaders)))
			httpClient.Transport = headertransport.New(httpClient.Transport, opts.GrafanaHeaders)
		}

		// Init Grafana client
		var err error
		grafanaClient, err = grafanasdk.NewClient(opts.GrafanaURL, opts.GrafanaCredentials, httpClient)
		if err != nil {
			return errors.Wrap(err, "Failed to initialize Grafana client")
		}
		if opts.GrafanaFolder != "" {
			gclog.Debug("Preparing dashboard folder", log.String("folder", opts.GrafanaFolder))

			// Get all the folders and look up the customer by the customer name (title)
			folders, err := grafanaClient.GetAllFolders(ctx)
			if err != nil {
				return errors.Wrap(err, "Unable to get all folders from Grafana API")
			}
			for _, folder := range folders {
				if folder.Title == opts.GrafanaFolder {
					gclog.Debug("Found existing folder", log.Int("folder.id", folder.ID))
					grafanaFolderID = folder.ID
				}
			}

			// folderId is not found, create it
			if grafanaFolderID == 0 {
				gclog.Debug("No existing folder found, creating a new one")
				folder, err := grafanaClient.CreateFolder(ctx, grafanasdk.Folder{
					Title: opts.GrafanaFolder,
				})
				if err != nil {
					return errors.Wrapf(err, "Error creating new folder %s", opts.GrafanaFolder)
				}

				gclog.Debug("Created folder",
					log.String("folder.title", folder.Title),
					log.Int("folder.id", folder.ID))
				grafanaFolderID = folder.ID
			}
		}
	}

	// Generate Garafana home dasboard "Overview"
	{
		data, err := grafana.Home(opts.InjectLabelMatchers)
		if err != nil {
			return errors.Wrap(err, "failed to generate home dashboard")
		}
		if opts.GrafanaDir != "" {
			generatedDashboard := "home.json"
			generatedAssets = append(generatedAssets, generatedDashboard)
			if err = os.WriteFile(filepath.Join(opts.GrafanaDir, generatedDashboard), data, os.ModePerm); err != nil {
				return errors.Wrap(err, "failed to generate home dashboard")
			}
		}
		if grafanaClient != nil {
			logger.Debug("Reloading Grafana dashboard",
				log.Int("folder.id", grafanaFolderID))
			if _, err := grafanaClient.SetRawDashboardWithParam(ctx, grafanasdk.RawBoardRequest{
				Dashboard: data,
				Parameters: grafanasdk.SetDashboardParams{
					Overwrite: true,
					FolderID:  grafanaFolderID,
				},
			}); err != nil {
				return errors.Wrapf(err, "Could not reload Grafana dashboard 'Overview'")
			} else {
				logger.Info("Reloaded Grafana dashboard")
			}
		}
	}

	// Generate per-dashboard assets
	for _, dashboard := range dashboards {
		// Logger for dashboard
		dlog := logger.With(log.String("dashboard", dashboard.Name))

		glog := dlog.Scoped("grafana", "grafana dashboard generation").
			With(log.String("instance", opts.GrafanaURL))

		glog.Debug("Rendering Grafana assets")
		board, err := dashboard.renderDashboard(opts.InjectLabelMatchers, opts.GrafanaFolder)
		if err != nil {
			return errors.Wrapf(err, "Failed to render dashboard %q", dashboard.Name)
		}

		// Prepare Grafana assets
		if opts.GrafanaDir != "" {
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
		}
		// Reload specific dashboard
		if grafanaClient != nil {
			glog.Debug("Reloading Grafana dashboard",
				log.Int("folder.id", grafanaFolderID))
			if _, err := grafanaClient.SetDashboard(ctx, *board, grafanasdk.SetDashboardParams{
				Overwrite: true,
				FolderID:  grafanaFolderID,
			}); err != nil {
				return errors.Wrapf(err, "Could not reload Grafana dashboard %q", dashboard.Title)
			} else {
				glog.Info("Reloaded Grafana dashboard")
			}
		}

		// Prepare Prometheus assets
		if opts.PrometheusDir != "" {
			plog := dlog.Scoped("prometheus", "prometheus rules generation")

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
	logger.Info("generated assets", log.Strings("files", generatedAssets))
	if !opts.DisablePrune {
		logger.Debug("Pruning dangling assets")
		if err := pruneAssets(logger, generatedAssets, opts.GrafanaDir, opts.PrometheusDir); err != nil {
			return errors.Wrap(err, "Failed to prune assets, resolve manually or disable pruning")
		}
	}

	return nil
}
