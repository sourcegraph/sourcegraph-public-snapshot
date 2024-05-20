package operationdocs

import (
	"fmt"
	"slices"
	"sort"

	"golang.org/x/exp/maps"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/operationdocs/internal/markdown"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
)

// IndexNotionPageID designates where in Notion the contents of
// operationdocs.RenderIndexPage should go.
// https://www.notion.so/sourcegraph/Managed-Services-0d0b709881674eee9dca4202de9f93b1
func IndexNotionPageID() string { return "0d0b709881674eee9dca4202de9f93b1" }

var (
	// https://www.notion.so/sourcegraph/Sourcegraph-Managed-Services-Platform-MSP-712a0389f54c4d3a90d069aa2d979a59
	mspNotionPageURL = NotionHandbookURL("712a0389f54c4d3a90d069aa2d979a59")
	// https://www.notion.so/sourcegraph/Core-Services-team-ed8af5ecf15545b292816ebba261a93c
	coreServicesNotionPageURL = NotionHandbookURL("ed8af5ecf15545b292816ebba261a93c")
)

func NotionHandbookURL(pageID string) string {
	return fmt.Sprintf("https://sourcegraph.notion.site/%s", pageID)
}

// RenderIndexPage renders an index page for use at HandbookPath, assuming that
// operationdocs.Render contents are stored
func RenderIndexPage(services []*spec.Spec, opts Options) []byte {
	md := markdown.NewBuilder()

	opts.AddDocumentNote(md)

	generalGuidanceLink, generalGuidance := markdown.HeadingLinkf("General guidance")
	md.Paragraphf(`These pages contain generated operational guidance for the infrastructure of the %d %s services (across %d environments) currently in operation at Sourcegraph. `+
		`This includes information about each service, configured environments, Entitle requests, common tasks, monitoring, custom documentation provided by service operators, and so on. `+
		`In addition to service-specific guidance, %s is also available.`,
		len(services),
		markdown.Link("Managed Services Platform (MSP)", mspNotionPageURL),
		specSet(services).countEnvironments(),
		generalGuidanceLink)

	md.Paragraphf(`MSP is owned by %s, but individual teams are responsible for the services they operate on the platform.`,
		markdown.Link("Core Services", coreServicesNotionPageURL))

	md.Paragraphf("Services are defined in %s, though service source code may live elsewhere.",
		markdown.Link(markdown.Code("sourcegraph/managed-services"), "https://github.com/sourcegraph/managed-services"))

	md.Admonitionf(markdown.AdmonitionImportant,
		"This page may be out of date if a service or environment was recently added or updated - reach out to #discuss-core-services for help updating these pages, or use %s to view the generated documentation in your terminal.",
		markdown.Code("sg msp operations $SERVICE_ID"))

	if opts.Notion {
		addNotionWarning(md)
	}

	owners, byOwner := collectByOwner(services)
	for _, o := range owners {
		md.Headingf(1, o)
		md.Paragraphf("Managed Services Platform services owned by %s:", markdown.Code(o))
		md.List(mapTo(byOwner[o], func(s *spec.Spec) string {
			if s.Service.NotionPageID != nil {
				return markdown.Linkf(s.Service.GetName(), NotionHandbookURL(*s.Service.NotionPageID))
			}
			return fmt.Sprintf("%s (no Notion page provided in service specification for generated docs)", s.Service.GetName())
		}))
	}

	md.Headingf(1, generalGuidance)

	md.Headingf(2, "Infrastructure access")
	md.Paragraphf(`For MSP service environments other than %s, access needs to be requested through Entitle. `+
		`Test environments are placed in the "Engineering Projects" GCP folder, which should have access granted to engineers by default.`,
		markdown.Code("category: test"))

	md.Paragraphf(`Entitle access to a production MSP project is generally provisioned through the %s and %s custom GCP roles, which provide read-only and editing access respectively. `+
		`Convenience links for requesting these roles are available in the per-service operation pages above, based on each environment.`,
		markdown.Code("mspServiceReader"), markdown.Code("mspServiceEditor"))

	md.Paragraphf(`You can also choose to request access to an individual project in Entitle by following these steps:`)

	md.List([]any{
		`Go to [app.entitle.io/request](https://app.entitle.io/request) and select **Specific Permission**`,
		`Fill out the following:`, []string{
			`Integration: **GCP Production Projects**`,
			`Resource types: **Project**`,
			`Resource: name of MSP project you are interested in`,
			fmt.Sprintf(`Role: %s (or %s if you need additional privileges - use with care!)`,
				markdown.Code("mspServiceReader"), markdown.Code("mspServiceEditor")),
			`Duration: choose your own adventure!`,
		},
	})

	md.Paragraphf(`The custom roles used for MSP infrastructure access are [configured in %s](https://github.com/sourcegraph/infrastructure/blob/main/gcp/custom-roles/msp.tf).`,
		markdown.Code("sourcegraph/infrastructure"))

	md.Headingf(2, "Terraform Cloud access")
	md.Paragraphf(`Terraform Cloud (TFC) workspaces for MSP [can be found using the %s workspace tag](https://app.terraform.io/app/sourcegraph/workspaces?tag=msp).`,
		markdown.Code("msp"))

	md.Paragraphf(`To gain access to MSP project TFC workspaces, [log in to Terraform Cloud](https://app.terraform.io/app/sourcegraph) and _then_ [request membership to the %s TFC team via Entitle](%s). `+
		`This TFC team has access to all MSP workspaces, and is [configured here](https://sourcegraph.sourcegraph.com/github.com/sourcegraph/infrastructure/-/blob/terraform-cloud/terraform.tfvars?L44:1-48:4).`,
		markdown.Code("Managed Services Platform Operators"),
		"https://app.entitle.io/request?data=eyJkdXJhdGlvbiI6IjM2MDAiLCJqdXN0aWZpY2F0aW9uIjoiRU5URVIgSlVTVElGSUNBVElPTiBIRVJFIiwicm9sZUlkcyI6W3siaWQiOiJiMzg3MzJjYy04OTUyLTQ2Y2QtYmIxZS1lZjI2ODUwNzIyNmIiLCJ0aHJvdWdoIjoiYjM4NzMyY2MtODk1Mi00NmNkLWJiMWUtZWYyNjg1MDcyMjZiIiwidHlwZSI6InJvbGUifV19")

	md.Paragraphf(`Note that you **must [log in to Terraform Cloud](https://app.terraform.io/app/sourcegraph) before making your Entitle request**. ` +
		`If you make your Entitle request, then log in, you will be removed from any team memberships granted through Entitle by Terraform Cloud's SSO implementation.`)

	md.Paragraphf(`For more details, also see [creating and configuring services](https://github.com/sourcegraph/managed-services#operations).`)

	return []byte(md.String())
}

func collectByOwner(services []*spec.Spec) ([]string, map[string]specSet) {
	m := make(map[string]specSet)
	for _, s := range services {
		for _, o := range s.Service.Owners {
			m[o] = append(m[o], s)
		}
	}

	owners := maps.Keys(m)
	slices.Sort(owners)
	for _, o := range owners {
		sort.Sort(m[o])
	}
	return owners, m
}

type specSet []*spec.Spec

func (s specSet) Len() int { return len(s) }
func (s specSet) Less(i, j int) bool {
	return s[i].Service.ID < s[j].Service.ID
}
func (s specSet) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s specSet) countEnvironments() int {
	var environments int
	for _, sp := range s {
		environments += len(sp.Environments)
	}
	return environments
}
