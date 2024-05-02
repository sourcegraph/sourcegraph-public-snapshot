package operationdocs

import (
	"bytes"
	"fmt"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/operationdocs/internal/markdown"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/operationdocs/terraform"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"golang.org/x/exp/maps"
)

type Options struct {
	// ManagedServicesRevision is the revision of sourcegraph/managed-services
	// used to generate page.
	ManagedServicesRevision string
	// GenerateCommand is the command used to generate this documentation -
	// it will be included in the generated output in a "DO NOT EDIT" comment.
	GenerateCommand string
	// Handbook indicates we are generating output for sourcegraph/handbook.
	Handbook bool
	// AlertPolicies is a deduplicated map of alert policies defined for all
	// environments of a service
	AlertPolicies map[string]terraform.AlertPolicy
}

// AddDocumentComment adds a comment to the markdown document with details about
// how this documentation was generated.
func (o Options) AddDocumentComment(md *markdown.Builder) {
	generatedFromComment := fmt.Sprintf("Generated from: https://github.com/sourcegraph/managed-services/tree/%s",
		o.ManagedServicesRevision)
	if o.ManagedServicesRevision == "" {
		generatedFromComment = "Generated from: unknown revision of https://github.com/sourcegraph/managed-services"
	}

	if o.GenerateCommand != "" {
		md.Commentf(`Generated documentation; DO NOT EDIT. Regenerate using this command: '%s'

Last updated: %s
%s`,
			o.GenerateCommand,
			time.Now().UTC().String(),
			generatedFromComment)
	} else {
		md.Commentf(generatedFromComment)
	}
}

// Render creates a Markdown string with operational guidance for a MSP specification
// and runtime properties using OutputsClient.
func Render(s spec.Spec, opts Options) (string, error) {
	md := markdown.NewBuilder()

	md.Headingf(1, "%s infrastructure operations", s.Service.GetName())

	opts.AddDocumentComment(md)

	mspURL := "https://handbook.sourcegraph.com/departments/engineering/teams/core-services/managed-services/platform/"
	coreServicesURL := "https://handbook.sourcegraph.com/departments/engineering/teams/core-services/"
	if opts.Handbook {
		mspURL = relativePathToMSPPage
		coreServicesURL = relativePathToCoreServicesPage
	}

	md.Paragraphf(`This document describes operational guidance for %s infrastructure.
This service is operated on the %s.`,
		s.Service.GetName(),
		markdown.Link("Managed Services Platform (MSP)", mspURL))

	md.Admonitionf(markdown.AdmonitionImportant, "If this is your first time here, you must follow the %s as well to clone the service definitions repository and set up the prerequisite tooling.",
		markdown.Link("sourcegraph/managed-services 'Tooling setup' guide", "https://github.com/sourcegraph/managed-services/blob/main/README.md"))

	md.Paragraphf("If you need assistance with MSP infrastructure, reach out to the %s team in #discuss-core-services.",
		markdown.Link("Core Services", coreServicesURL))

	type environmentHeader struct {
		environmentID string
		header        string
		link          string
	}
	var environmentHeaders []environmentHeader

	md.Headingf(2, "Service overview")
	serviceKind := pointers.Deref(s.Service.Kind, spec.ServiceKindService)
	serviceDirURL := fmt.Sprintf("https://github.com/sourcegraph/managed-services/blob/main/services/%s", s.Service.ID)
	serviceConfigURL := fmt.Sprintf("%s/service.yaml", serviceDirURL)

	md.Table(
		[]string{"Property", "Details"},
		[][]string{
			{"Service ID", fmt.Sprintf("%s (%s)",
				markdown.Code(s.Service.ID), markdown.Link("specification", serviceConfigURL))},
			// TODO: See service.Description docstring
			// {"Description", s.Service.Description},
			{"Owners", strings.Join(mapTo(s.Service.Owners, markdown.Bold), ", ")},
			{"Service kind", fmt.Sprintf("Cloud Run %s", string(serviceKind))},
			{"Environments", strings.Join(mapTo(s.Environments, func(e spec.EnvironmentSpec) string {
				l, h := markdown.HeadingLinkf(e.ID)
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

	if len(s.README) > 0 {
		md.Commentf("Automatically generated from the service README: %s", fmt.Sprintf("%s/README.md", serviceDirURL))

		readme := string(s.README)
		lines := strings.Split(readme, "\n")
		for i, line := range lines {
			// Increase all headers by 1 so that they fit nicely into the
			// generated page.
			if strings.HasPrefix(line, "##") {
				lines[i] = "#" + line
			}
		}
		md.Paragraphf(strings.Join(lines, "\n"))
	}

	if s.Rollout != nil {
		md.Headingf(2, "Rollouts")
		region := "us-central1"
		var rolloutDetails [][]string
		// Get final stage to generate pipeline url
		finalStageEnv := s.Rollout.Stages[len(s.Rollout.Stages)-1].EnvironmentID
		finalStageProject := s.GetEnvironment(finalStageEnv).ProjectID
		rolloutDetails = append(rolloutDetails, []string{"Delivery pipeline", markdown.Linkf(fmt.Sprintf("`%s-%s-rollout`", s.Service.ID, region),
			"https://console.cloud.google.com/deploy/delivery-pipelines/%[1]s/%[2]s-%[1]s-rollout?project=%[3]s", region, s.Service.ID, finalStageProject)})

		var stages []string
		for _, stage := range s.Rollout.Stages {
			envIndex := slices.IndexFunc(environmentHeaders, func(env environmentHeader) bool {
				return stage.EnvironmentID == env.environmentID
			})
			stages = append(stages, environmentHeaders[envIndex].link)
		}
		rolloutDetails = append(rolloutDetails, []string{"Stages", strings.Join(stages, " -> ")})

		md.Table([]string{"Property", "Details"}, rolloutDetails)
		md.Paragraphf("Changes to %[1]s are continuously delivered to the first stage (%[2]s) of the delivery pipeline.", *s.Service.Name, stages[0])
		if len(stages) > 1 {
			md.Paragraphf("Promotion of a release to the next stage in the pipeline must be done manually using the GCP Delivery pipeline UI.")
		}
	}

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

		sentryLink := markdown.Linkf("Sentry "+markdown.Codef("%s-%s", s.Service.ID, env.ID), "https://sourcegraph.sentry.io/projects/%s-%s/", s.Service.ID, env.ID)
		slackChannelName := fmt.Sprintf("alerts-%s-%s", s.Service.ID, env.ID)
		overview := [][]string{
			{"Project ID", markdown.Linkf(markdown.Code(env.ProjectID), cloudRunURL)},
			{"Category", markdown.Bold(string(env.Category))},
			{"Deployment type", fmt.Sprintf("`%s`", env.Deploy.Type)},
			{"Resources", strings.Join(mapTo(env.Resources.List(), func(k string) string {
				l, h := markdown.HeadingLinkf("%s %s", env.ID, k)
				resourceHeadings[k] = h
				return l
			}), ", ")},
			{"Slack notifications", markdown.Linkf("#"+slackChannelName, "https://sourcegraph.slack.com/archives/"+slackChannelName)},
			{"Alert policies",
				fmt.Sprintf("%s, %s",
					markdown.Linkf("GCP Monitoring alert policies list", "https://console.cloud.google.com/monitoring/alerting/policies?project=%s", env.ProjectID),
					AlertPolicyDashboardURL(env.ProjectID))},
			{"Errors", sentryLink},
		}
		if env.EnvironmentServiceSpec != nil {
			if domain := env.Domain.GetDNSName(); domain != "" {
				overview = append(overview, []string{"Domain", markdown.Link(domain, "https://"+domain)})
				if env.Domain.Cloudflare != nil && env.Domain.Cloudflare.ShouldProxy() {
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

		entitleIntro := `MSP infrastructure access needs to be requested using Entitle for time-bound privileges.`
		if env.Category == spec.EnvironmentCategoryTest {
			entitleIntro += " Test environments may have less stringent requirements."
		}
		md.Paragraphf(entitleIntro)

		md.Table([]string{"Access", "Entitle request template"}, [][]string{
			{"GCP project read access", entitleReaderLinksByCategory[env.Category]},
			{"GCP project write access", entitleEditorLinksByCategory[env.Category]},
		})

		terraformCloudSectionLink, terraformCloudSectionHeading := markdown.HeadingLinkf("%s Terraform Cloud", env.ID)
		md.Paragraphf("For Terraform Cloud access, see %s.", terraformCloudSectionLink)

		_, cloudRunSectionLink := md.Headingf(4, "%s Cloud Run", env.ID)

		// It's not immediately obvious to new users that Cloud Run is where
		// their service "runs".
		md.Paragraphf("The %s %s service implementation is deployed on %s.",
			s.Service.GetName(), env.ID,
			markdown.Link("Google Cloud Run", "https://cloud.google.com/run"))

		md.Table(
			[]string{"Property", "Details"},
			[][]string{
				{"Console", markdown.Linkf(
					fmt.Sprintf("Cloud Run %s", string(serviceKind)), cloudRunURL)},
				{"Service logs", markdown.Link("GCP logging", ServiceLogsURL(serviceKind, env.ProjectID))},
				{"Service traces", markdown.Linkf("Cloud Trace", "https://console.cloud.google.com/traces/list?project=%s", env.ProjectID)},
				{"Service errors", sentryLink},
			},
		)

		md.Paragraphf("You can also use %s to quickly open a link to your service logs:", markdown.Code("sg msp"))
		md.CodeBlockf("bash", `sg msp logs %s %s`, s.Service.ID, env.ID)

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

				md.Admonitionf(markdown.AdmonitionNote, "The %s is required for BOTH read-only and write access to the database.",
					entitleEditorLinksByCategory[env.Category])

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

		md.Headingf(4, "%s Architecture Diagram", env.ID)
		md.Paragraphf("!" + markdown.Linkf("Architecture Diagram", "./%s-%s.svg", s.Service.ID, env.ID))

		md.Headingf(4, terraformCloudSectionHeading)

		md.Paragraphf(`This service's configuration is defined in %s, and %s generates the required infrastructure configuration for this environment in Terraform.
Terraform Cloud (TFC) workspaces specific to each service then provisions the required infrastructure from this configuration.
You may want to check your service environment's TFC workspaces if a Terraform apply fails (reported via GitHub commit status checks in the %s repository, or in #alerts-msp-tfc).`,
			markdown.Link(markdown.Codef("sourcegraph/managed-services/services/%s/service.yaml", s.Service.ID), serviceConfigURL),
			markdown.Codef("sg msp generate %s %s", s.Service.ID, env.ID),
			markdown.Link(markdown.Code("sourcegraph/managed-services"), "https://github.com/sourcegraph/managed-services"))

		md.Admonitionf(markdown.AdmonitionNote, `If you are looking for service logs, see the %s section instead. In general:

%s`,
			cloudRunSectionLink, markdown.List([]string{
				fmt.Sprintf("check service logs (%s) if your service has gone down or is misbehaving", cloudRunSectionLink),
				"check TFC workspaces for infrastructure provisioning or configuration issues",
			}))

		md.Paragraphf(`To access this environment's Terraform Cloud workspaces, you will need to [log in to Terraform Cloud](https://app.terraform.io/app/sourcegraph) and then [request Entitle access to membership in the "Managed Services Platform Operator" TFC team](https://app.entitle.io/request?data=eyJkdXJhdGlvbiI6IjM2MDAiLCJqdXN0aWZpY2F0aW9uIjoiSlVTVElGSUNBVElPTiBIRVJFIiwicm9sZUlkcyI6W3siaWQiOiJiMzg3MzJjYy04OTUyLTQ2Y2QtYmIxZS1lZjI2ODUwNzIyNmIiLCJ0aHJvdWdoIjoiYjM4NzMyY2MtODk1Mi00NmNkLWJiMWUtZWYyNjg1MDcyMjZiIiwidHlwZSI6InJvbGUifV19).
The "Managed Services Platform Operator" team has access to all MSP TFC workspaces.`)

		md.Admonitionf(markdown.AdmonitionWarning, `You **must [log in to Terraform Cloud](https://app.terraform.io/app/sourcegraph) before making your Entitle request**.
If you make your Entitle request, then log in, you will be removed from any team memberships granted through Entitle by Terraform Cloud's SSO implementation.`)

		md.Paragraphf("The Terraform Cloud workspaces for this service environment are %s, or you can use:",
			markdown.Linkf(fmt.Sprintf("grouped under the %s tag", markdown.Codef("msp-%s-%s", s.Service.ID, env.ID)),
				"https://app.terraform.io/app/sourcegraph/workspaces?tag=msp-%s-%s", s.Service.ID, env.ID))
		md.CodeBlockf("bash", `sg msp tfc view %s %s`, s.Service.ID, env.ID)
	}

	md.Headingf(2, "Alert Policies")

	md.Paragraphf("The following alert policies are defined for each of this service's environments.")

	// Render alerts
	// Sort the map keys to make order deterministic
	alertKeys := maps.Keys(opts.AlertPolicies)
	slices.Sort(alertKeys)
	for _, key := range alertKeys {
		policy := opts.AlertPolicies[key]
		md.Headingf(4, policy.DisplayName)
		// We need to remove the footer text we add to each alert policy description
		b := []byte(policy.Documentation.Content)
		lastParagraphIndex := bytes.LastIndex(b, []byte("\n\n"))
		if lastParagraphIndex != -1 {
			b = b[:lastParagraphIndex]
		}

		md.CodeBlock("md", string(b))
		md.Paragraphf("Severity: %s", policy.Severity)
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
