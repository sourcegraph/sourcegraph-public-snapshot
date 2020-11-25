package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/grafana-tools/sdk"
	"gopkg.in/yaml.v2"
)

var isDev, _ = strconv.ParseBool(os.Getenv("DEV"))
var noPrune, _ = strconv.ParseBool(os.Getenv("NO_PRUNE"))

const alertSuffix = "_alert_rules.yml"

// Generate is the main Sourcegraph monitoring generator entrypoint.
func Generate(containers ...*Container) {
	println("Regenerating monitoring...")

	grafanaDir, ok := os.LookupEnv("GRAFANA_DIR")
	if !ok {
		grafanaDir = "../docker-images/grafana/config/provisioning/dashboards/sourcegraph/"
	}
	prometheusDir, ok := os.LookupEnv("PROMETHEUS_DIR")
	if !ok {
		prometheusDir = "../docker-images/prometheus/config/"
	}
	docSolutionsFile, ok := os.LookupEnv("DOC_SOLUTIONS_FILE")
	if !ok {
		docSolutionsFile = "../doc/admin/observability/alert_solutions.md"
	}

	reloadValue, ok := os.LookupEnv("RELOAD")
	if !ok && isDev {
		reloadValue = "true"
	}
	reload, _ := strconv.ParseBool(reloadValue)

	var filelist []string
	for _, container := range containers {
		if err := container.validate(); err != nil {
			log.Fatal(fmt.Sprintf("container %q: %+v", container.Name, err))
		}
		if grafanaDir != "" {
			board := container.dashboard()
			data, err := json.MarshalIndent(board, "", "  ")
			if err != nil {
				log.Fatal(err)
			}
			// #nosec G306  prometheus runs as nobody
			err = ioutil.WriteFile(filepath.Join(grafanaDir, container.Name+".json"), data, 0666)
			if err != nil {
				log.Fatal(err)
			}
			filelist = append(filelist, container.Name+".json")

			if reload {
				ctx := context.Background()
				client := sdk.NewClient("http://127.0.0.1:3370", "admin:admin", sdk.DefaultHTTPClient)
				_, err := client.SetDashboard(ctx, *board, sdk.SetDashboardParams{Overwrite: true})
				if err != nil {
					log.Fatal("updating dashboard:", err)
				}
			}
		}

		if prometheusDir != "" {
			promAlertsFile := container.promAlertsFile()
			data, err := yaml.Marshal(promAlertsFile)
			if err != nil {
				log.Fatal(err)
			}
			fileName := strings.Replace(container.Name, "-", "_", -1) + alertSuffix
			filelist = append(filelist, fileName)
			// #nosec G306  grafana runs as UID 472
			err = ioutil.WriteFile(filepath.Join(prometheusDir, fileName), data, 0666)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	if !noPrune {
		deleteRemnants(filelist, grafanaDir, prometheusDir)
	}

	if prometheusDir != "" && reload {
		resp, err := http.Post("http://127.0.0.1:9090/-/reload", "", nil)
		if err != nil {
			log.Fatal("reloading Prometheus rules, got error:", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Fatal("reloading Prometheus rules, got status code:", resp.StatusCode)
		}
	}
	if reload && grafanaDir != "" && prometheusDir != "" {
		fmt.Println("Reloaded Prometheus rules & Grafana dashboards")
	}

	if docSolutionsFile != "" {
		solutions := generateDocs(containers)
		// #nosec G306
		err := ioutil.WriteFile(docSolutionsFile, solutions, 0666)
		if err != nil {
			log.Fatal(err)
		}
	}
}
