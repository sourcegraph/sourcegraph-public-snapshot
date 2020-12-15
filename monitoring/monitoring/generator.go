package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/grafana-tools/sdk"
	"gopkg.in/yaml.v2"
)

const alertSuffix = "_alert_rules.yml"

const alertSolutionsFile = "alert_solutions.md"

// GenerateOptions declares options for the monitoring generator.
type GenerateOptions struct {
	// Toggles pruning of dangling generated assets through simple heuristic, should be disabled during builds
	DisablePrune bool
	// Trigger reload of active Prometheus or Grafana instance (requires respective output directories)
	LiveReload bool

	// Output directory for generated Grafana assets
	GrafanaDir string
	// Output directory for generated Prometheus assets
	PrometheusDir string
	// Output directory for generated documentation
	DocsDir string
}

// Generate is the main Sourcegraph monitoring generator entrypoint.
func Generate(opts GenerateOptions, containers ...*Container) {
	println("Regenerating monitoring...")

	var filelist []string
	for _, container := range containers {
		if err := container.validate(); err != nil {
			log.Fatal(fmt.Sprintf("container %q: %+v", container.Name, err))
		}
		if opts.GrafanaDir != "" {
			board := container.dashboard()
			data, err := json.MarshalIndent(board, "", "  ")
			if err != nil {
				log.Fatal(err)
			}
			// #nosec G306  prometheus runs as nobody
			err = ioutil.WriteFile(filepath.Join(opts.GrafanaDir, container.Name+".json"), data, 0666)
			if err != nil {
				log.Fatal(err)
			}
			filelist = append(filelist, container.Name+".json")

			// Reload specific dashboard
			if opts.LiveReload {
				ctx := context.Background()
				client := sdk.NewClient("http://127.0.0.1:3370", "admin:admin", sdk.DefaultHTTPClient)
				_, err := client.SetDashboard(ctx, *board, sdk.SetDashboardParams{Overwrite: true})
				if err != nil {
					log.Fatal("updating dashboard:", err)
				}
			}
		}

		if opts.PrometheusDir != "" {
			promAlertsFile := container.promAlertsFile()
			data, err := yaml.Marshal(promAlertsFile)
			if err != nil {
				log.Fatal(err)
			}
			fileName := strings.Replace(container.Name, "-", "_", -1) + alertSuffix
			filelist = append(filelist, fileName)
			// #nosec G306  grafana runs as UID 472
			err = ioutil.WriteFile(filepath.Join(opts.PrometheusDir, fileName), data, 0666)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	if !opts.DisablePrune {
		deleteRemnants(filelist, opts.GrafanaDir, opts.PrometheusDir)
	}

	if opts.PrometheusDir != "" && opts.LiveReload {
		// Reload all Prometheus rules
		resp, err := http.Post("http://127.0.0.1:9090/-/reload", "", nil)
		if err != nil {
			log.Fatal("reloading Prometheus rules, got error:", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Fatal("reloading Prometheus rules, got status code:", resp.StatusCode)
		}
	}
	if opts.LiveReload && opts.GrafanaDir != "" && opts.PrometheusDir != "" {
		fmt.Println("Reloaded Prometheus rules & Grafana dashboards")
	}

	if opts.DocsDir != "" {
		solutions := generateDocs(containers)
		// #nosec G306
		err := ioutil.WriteFile(filepath.Join(opts.DocsDir, alertSolutionsFile), solutions, 0666)
		if err != nil {
			log.Fatal(err)
		}
	}
}
