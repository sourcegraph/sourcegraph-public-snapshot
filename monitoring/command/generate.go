package command

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/hcl/strconv"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/urfave/cli/v3"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/cliutil/completions"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/monitoring/definitions"
	"github.com/sourcegraph/sourcegraph/monitoring/monitoring"
)

// Generate creates a 'generate' command that generates the default monitoring dashboards.
func Generate(cmdRoot string, sgRoot string) *cli.Command {
	return &cli.Command{
		Name:      "generate",
		ArgsUsage: "<dashboard>",
		UsageText: fmt.Sprintf(`
# Generate all monitoring with default configuration into a temporary directory
%[1]s generate -all.dir /tmp/monitoring

# Generate and reload local instances of Grafana, Prometheus, etc.
%[1]s generate -reload

# Render dashboards in a custom directory, and disable rendering of docs
%[1]s generate -grafana.dir /tmp/my-dashboards -docs.dir ''
`, cmdRoot),
		Usage: "Generate monitoring assets - dashboards, alerts, and more",
		// Flags should correspond to monitoring.GenerateOpts
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "no-prune",
				Sources: cli.EnvVars("NO_PRUNE"),
				Usage:   "Toggles pruning of dangling generated assets through simple heuristic - should be disabled during builds.",
			},
			&cli.BoolFlag{
				Name:    "reload",
				Sources: cli.EnvVars("RELOAD"),
				Usage:   "Trigger reload of active Prometheus or Grafana instance (requires respective output directories)",
			},

			&cli.StringFlag{
				Name:  "all.dir",
				Usage: "Override all other '-*.dir' directories",
			},

			&cli.StringFlag{
				Name:    "grafana.dir",
				Sources: cli.EnvVars("GRAFANA_DIR"),
				Value:   "$SG_ROOT/docker-images/grafana/config/provisioning/dashboards/sourcegraph/",
				Usage:   "Output directory for generated Grafana assets",
			},
			&cli.StringFlag{
				Name:  "grafana.url",
				Value: "http://127.0.0.1:3370",
				Usage: "Address for the Grafana instance to reload",
			},
			&cli.StringFlag{
				Name:  "grafana.creds",
				Value: "admin:admin",
				Usage: "Credentials for the Grafana instance to reload",
			},
			&cli.StringSliceFlag{
				Name:    "grafana.headers",
				Sources: cli.EnvVars("GRAFANA_HEADERS"),
				Usage:   "Additional headers for HTTP requests to the Grafana instance",
			},
			&cli.StringFlag{
				Name:  "grafana.folder",
				Usage: "Folder on Grafana instance to put generated dashboards in",
			},

			&cli.StringFlag{
				Name:    "prometheus.dir",
				Sources: cli.EnvVars("PROMETHEUS_DIR"),
				Value:   "$SG_ROOT/docker-images/prometheus/config/",
				Usage:   "Output directory for generated Prometheus assets",
			},
			&cli.StringFlag{
				Name:  "prometheus.url",
				Value: "http://127.0.0.1:9090",
				Usage: "Address for the Prometheus instance to reload",
			},

			&cli.StringFlag{
				Name:    "docs.dir",
				Sources: cli.EnvVars("DOCS_DIR"),
				Value:   "$SG_ROOT/doc/admin/observability/",
				Usage:   "Output directory for generated documentation",
			},
			&cli.StringSliceFlag{
				Name:    "inject-label-matcher",
				Sources: cli.EnvVars("INJECT_LABEL_MATCHERS"),
				Usage:   "Labels to inject into all selectors in Prometheus expressions: observable queries, dashboard template variables, etc.",
			},
			&cli.StringSliceFlag{
				Name:    "multi-instance-groupings",
				Sources: cli.EnvVars("MULTI_INSTANCE_GROUPINGS"),
				Usage:   "If non-empty, indicates whether or not to generate multi-instance assets with the provided labels to group on. The standard per-instance monitoring assets will NOT be generated.",
			},
		},
		ShellComplete: completions.CompleteArgs(func() (options []string) {
			return definitions.Default().Names()
		}),
		Action: func(ctx context.Context, cmd *cli.Command) error {
			logger := log.Scoped(cmd.Name)

			// expandErr is set from within expandWithSgRoot
			var expandErr error
			expandWithSgRoot := func(key string) string {
				// Lookup first, to allow overrides of SG_ROOT
				if v, set := os.LookupEnv(key); set {
					return v
				}
				if key == "SG_ROOT" {
					if sgRoot == "" {
						expandErr = errors.New("$SG_ROOT is required to use the default paths")
					}
					return sgRoot
				}
				return ""
			}

			options := monitoring.GenerateOptions{
				DisablePrune: cmd.Bool("no-prune"),
				Reload:       cmd.Bool("reload"),

				GrafanaDir:         os.Expand(cmd.String("grafana.dir"), expandWithSgRoot),
				GrafanaURL:         cmd.String("grafana.url"),
				GrafanaCredentials: cmd.String("grafana.creds"),
				GrafanaFolder:      cmd.String("grafana.folder"),
				GrafanaHeaders: func() map[string]string {
					h := make(map[string]string)
					for _, entry := range cmd.StringSlice("grafana.headers") {
						if len(entry) == 0 {
							continue
						}

						parts := strings.Split(entry, "=")
						if len(parts) != 2 {
							logger.Error("discarding invalid grafana.headers entry",
								log.String("entry", entry))
							continue
						}
						header := parts[0]
						value, err := strconv.Unquote(parts[1])
						if err != nil {
							value = parts[1]
						}
						h[header] = value
					}
					return h
				}(),

				PrometheusDir: os.Expand(cmd.String("prometheus.dir"), expandWithSgRoot),
				PrometheusURL: cmd.String("prometheus.url"),

				DocsDir: os.Expand(cmd.String("docs.dir"), expandWithSgRoot),

				InjectLabelMatchers: func() []*labels.Matcher {
					var matchers []*labels.Matcher
					for _, entry := range cmd.StringSlice("inject-label-matcher") {
						if len(entry) == 0 {
							continue
						}

						parts := strings.Split(entry, "=")
						if len(parts) != 2 {
							logger.Error("discarding invalid INJECT_LABEL_MATCHERS entry",
								log.String("entry", entry))
							continue
						}

						label := parts[0]
						value, err := strconv.Unquote(parts[1])
						if err != nil {
							value = parts[1]
						}
						matcher, err := labels.NewMatcher(labels.MatchEqual, label, value)
						if err != nil {
							logger.Error("discarding invalid INJECT_LABEL_MATCHERS entry",
								log.String("entry", entry),
								log.Error(err))
							continue
						}
						matchers = append(matchers, matcher)
					}
					return matchers
				}(),

				MultiInstanceDashboardGroupings: cmd.StringSlice("multi-instance-groupings"),
			}

			// If 'all.dir' is set, override all other '*.dir' flags and ignore expansion
			// errors.
			if allDir := cmd.String("all.dir"); allDir != "" {
				logger.Info("overriding all directory flags with 'all.dir'", log.String("all.dir", allDir))
				options.GrafanaDir = filepath.Join(allDir, "grafana")
				options.PrometheusDir = filepath.Join(allDir, "prometheus")
				options.DocsDir = filepath.Join(allDir, "docs")
			} else if expandErr != nil {
				return expandErr
			}

			// Decide which dashboards to generate
			var dashboards definitions.Dashboards
			if cmd.Args().Len() == 0 {
				dashboards = definitions.Default()
			} else {
				for _, arg := range cmd.Args().Slice() {
					d := definitions.Default().GetByName(cmd.Args().First())
					if d == nil {
						return errors.Newf("Dashboard %q not found", arg)
					}
					dashboards = append(dashboards, d)
				}
			}

			logger.Info("generating dashboards",
				log.Strings("dashboards", dashboards.Names()))

			return monitoring.Generate(logger, options, dashboards...)
		},
	}

}
