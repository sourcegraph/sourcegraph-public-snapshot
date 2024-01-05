package operationdocs

import (
	"fmt"
	"path"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/operationdocs/internal/markdown"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Options struct {
	GenerateCommand string
}

// Render creates a Markdown string with operational guidance for a MSP specification
// and runtime properties using OutputsClient.
func Render(s spec.Spec, opts Options) (string, error) {
	md := markdown.NewBuilder()

	md.Headingf(1, "%s infrastructure operations", s.Service.GetName())

	if opts.GenerateCommand != "" {
		md.Commentf("Generated documentation; DO NOT EDIT. Regenerate using this command: '%s'",
			opts.GenerateCommand)
	}

	md.Paragraphf(`This document describes operational guidance for %s infrastructure.
This service is operated on the [Managed Services Platform (MSP)](https://handbook.sourcegraph.com/departments/engineering/teams/core-services/managed-services/platform/).`,
		s.Service.GetName())

	type environmentHeader struct {
		environmentID string
		header        string
		link          string
	}
	var environmentHeaders []environmentHeader

	md.Headingf(2, "Service overview")
	serviceKind := pointers.Deref(s.Service.Kind, spec.ServiceKindService)
	md.Table(
		[]string{"Property", "Details"},
		[][]string{
			{"Service ID", markdown.Linkf(markdown.Code(s.Service.ID),
				"https://github.com/sourcegraph/managed-services/blob/main/services/%s/service.yaml",
				s.Service.ID)},
			{"Owners", strings.Join(mapTo(s.Service.Owners, markdown.Bold), ", ")},
			{"Service kind", fmt.Sprintf("Cloud Run %s", string(serviceKind))},
			{"Environments", strings.Join(mapTo(s.Environments, func(e spec.EnvironmentSpec) string {
				l, h := markdown.HeadingLinkf("%s environment", e.ID)
				environmentHeaders = append(environmentHeaders, environmentHeader{
					environmentID: e.ID,
					header:        h,
					link:          l,
				})
				return l
			}), ", ")},
			{"Docker image", markdown.Code(s.Build.Image)},
			{"Source code", markdown.Linkf(
				fmt.Sprintf("%s - %s", markdown.Code(s.Build.Source.Repo), markdown.Code(s.Build.Source.Dir)),
				"https://%s/tree/HEAD/%s", s.Build.Source.Repo, path.Clean(s.Build.Source.Dir))},
		})

	md.Headingf(2, "Environments")
	for _, section := range environmentHeaders {
		md.Headingf(3, section.header)
		env := s.GetEnvironment(section.environmentID)

		var cloudRunURL string
		switch serviceKind {
		case spec.ServiceKindService:
			cloudRunURL = fmt.Sprintf("https://console.cloud.google.com/run?project=%s",
				env.ProjectID)
		case spec.ServiceKindJob:
			cloudRunURL = fmt.Sprintf("https://console.cloud.google.com/run/jobs?project=%s",
				env.ProjectID)
		default:
			return "", errors.Newf("unknown service kind %q", serviceKind)
		}

		// ResourceKind:env-specific header
		resourceHeadings := map[string]string{}

		overview := [][]string{
			{"Project ID", markdown.Linkf(markdown.Code(env.ProjectID), cloudRunURL)},
			{"Category", markdown.Bold(string(env.Category))},
			{"Resources", strings.Join(mapTo(env.Resources.List(), func(k string) string {
				l, h := markdown.HeadingLinkf("%s %s", env.ID, k)
				resourceHeadings[k] = h
				return l
			}), ", ")},
			{"Alerts", markdown.Linkf("GCP monitoring", "https://console.cloud.google.com/monitoring/alerting?project=%s", env.ProjectID)},
		}
		if env.EnvironmentServiceSpec != nil {
			if domain := env.Domain.GetDNSName(); domain != "" {
				overview = append(overview, []string{"Domain", markdown.Link(domain, "https://"+domain)})
				if env.Domain.Cloudflare != nil && env.Domain.Cloudflare.Proxied {
					overview = append(overview, []string{"Cloudflare WAF", "✅"})
				}
			}
			if env.Authentication != nil {
				if pointers.DerefZero(env.Authentication.Sourcegraph) {
					overview = append(overview, []string{"Authentication", "sourcegraph.com GSuite users"})
				} else {
					overview = append(overview, []string{"Authentication", "Restricted"})
				}
			}
		}
		md.Table(
			[]string{"Property", "Details"},
			overview,
		)

		md.Paragraphf(`MSP infrastructure access needs to be requested using Entitle for time-bound privileges.
Test environments have less stringent requirements.`)

		md.Table([]string{"Access", "Entitle request template"}, [][]string{
			{"GCP project read access", entitleReaderLinksByCategory[env.Category]},
			{"GCP project write access", entitleEditorLinksByCategory[env.Category]},
		})
		// TODO: Add a comment about per-project access as well?

		md.Headingf(4, "%s Cloud Run", env.ID)
		md.Table(
			[]string{"Property", "Details"},
			[][]string{
				{"Console", markdown.Linkf(
					fmt.Sprintf("Cloud Run %s", string(serviceKind)), cloudRunURL)},
				{"Logs", markdown.Link("GCP logging", ServiceLogsURL(serviceKind, env.ProjectID))},
			},
		)

		// Individual resources - add them in the same order as (EnvironmentResourcesSpec).List()
		if env.Resources != nil {
			if redis := env.Resources.Redis; redis != nil {
				md.Headingf(4, resourceHeadings[redis.ResourceKind()])
				md.Table(
					[]string{"Property", "Details"},
					[][]string{
						{"Console", markdown.Linkf("Memorystore Redis instances",
							"https://console.cloud.google.com/memorystore/redis/instances?project=%s", env.ProjectID)},
					},
				)

				// TODO: More details
			}

			if pg := env.Resources.PostgreSQL; pg != nil {
				md.Headingf(4, resourceHeadings[pg.ResourceKind()])
				md.Table(
					[]string{"Property", "Details"},
					[][]string{
						{"Console", markdown.Linkf("Cloud SQL instances",
							"https://console.cloud.google.com/sql/instances?project=%s", env.ProjectID)},
						{"Databases", strings.Join(mapTo(pg.Databases, markdown.Code), ", ")},
					},
				)

				managedServicesRepoLink := markdown.Link(markdown.Code("sourcegraph/managed-services"),
					"https://github.com/sourcegraph/managed-services")

				md.Paragraphf("To connect to the PostgreSQL instance in this environment, use %s in the %s repository:",
					markdown.Code("sg msp"), managedServicesRepoLink)

				md.CodeBlockf("bash", `# For read-only access
sg msp pg connect %[1]s %[2]s

# For write access - use with caution!
sg msp pg connect -write-access %[1]s %[2]s`, s.Service.ID, env.ID)
			}

			if bq := env.Resources.BigQueryDataset; bq != nil {
				md.Headingf(4, resourceHeadings[bq.ResourceKind()])
				md.Table(
					[]string{"Property", "Details"},
					[][]string{
						{"Dataset Project", markdown.Code(pointers.Deref(bq.ProjectID, env.ProjectID))},
						{"Dataset ID", markdown.Code(bq.GetDatasetID(s.Service.ID))},
						{"Tables", strings.Join(mapTo(bq.Tables, func(t string) string {
							return markdown.Linkf(markdown.Code(t),
								"https://github.com/sourcegraph/managed-services/blob/main/services/%s/%s.bigquerytable.json",
								s.Service.ID, t)
						}), ", ")},
					},
				)

				// TODO: more details
			}
		}
	}

	return md.String(), nil
}

func mapTo[InputT any, OutputT any](input []InputT, mapFn func(InputT) OutputT) []OutputT {
	var output []OutputT
	for _, i := range input {
		output = append(output, mapFn(i))
	}
	return output
}
