package catalog

import (
	"sort"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

type Component struct {
	Kind            string
	Name            string
	Description     string
	Lifecycle       string
	SourceLocations []SourceLocation

	UsagePatterns []UsagePattern
	APIDefPath    string

	Tags []string

	Links []Link // TODO(sqs): use this instead of hardcoding

	// Edges
	DependsOn []string
	OwnedBy   string
}

type SourceLocation struct {
	Repo  api.RepoName
	Paths []string

	// TODO(sqs): how to handle if the paths move in a commit?
}

type Link struct {
	Title string
	URL   string
}

type UsagePattern struct {
	Query string
}

type Package struct {
	Name string
}

type Group struct {
	Name        string
	Title       string
	Description string
	ParentGroup string
	Members     []string
}

type Edge struct {
	Type    EntityRelationType
	Out, In string // entity names
}

type EntityRelationType string

const (
	DependsOnRelation    = "DEPENDS_ON"
	DependencyOfRelation = "DEPENDENCY_OF"

	PartOfRelation  = "PART_OF"
	HasPartRelation = "HAS_PART"

	MemberOfRelation  = "MEMBER_OF"
	HasMemberRelation = "HAS_MEMBER"

	//

	LifecycleProduction   = "PRODUCTION"
	LifecycleExperimental = "EXPERIMENTAL"
)

func newQueryUsagePattern(query string) UsagePattern {
	return UsagePattern{
		Query: `count:1000 repo:^github\.com/sourcegraph/sourcegraph$ ` + query,
	}
}

type ComputedLabel struct {
	Key          string
	ValueQueries map[string]string
}

var (
	components     []Component
	groups         []Group
	edges          []Edge
	computedLabels []ComputedLabel
)

func init() {
	components, groups, edges, computedLabels = data()
}

func data() ([]Component, []Group, []Edge, []ComputedLabel) {
	components := []Component{
		{
			Kind:            "SERVICE",
			Name:            "frontend",
			Description:     "Serves the web app, public APIs, and internal APIs.",
			Lifecycle:       LifecycleProduction,
			SourceLocations: []SourceLocation{{Repo: "github.com/sourcegraph/sourcegraph", Paths: []string{"cmd/frontend", "enterprise/cmd/frontend"}}},
			UsagePatterns: []UsagePattern{
				newQueryUsagePattern(`\.api/graphql patterntype:literal`),
				newQueryUsagePattern(`lang:typescript requestGraphQL\(|useConnection\(|useQuery\(|gql` + "`" + ` patterntype:regexp`),
			},
			APIDefPath: "cmd/frontend/graphqlbackend/schema.graphql",
			DependsOn:  []string{"gitserver", "client-web", "repo-updater", "executor", "github-proxy", "precise-code-intel-worker", "query-runner", "searcher", "sitemap", "symbols"},
			OwnedBy:    "search-core",
			Tags:       []string{"public-api", "ui"},
		},
		{
			Kind:            "SERVICE",
			Name:            "gitserver",
			Description:     "Mirrors repositories from their code host.",
			Lifecycle:       LifecycleProduction,
			SourceLocations: []SourceLocation{{Repo: "github.com/sourcegraph/sourcegraph", Paths: []string{"cmd/gitserver"}}},
			UsagePatterns: []UsagePattern{
				newQueryUsagePattern(`lang:go \bgitserver\.Client\b patterntype:regexp`),
				newQueryUsagePattern(`lang:go \bgit\.[A-Z]\w+\(ctx, patterntype:regexp`),
				newQueryUsagePattern(`lang:go "github.com/sourcegraph/sourcegraph/internal/vcs/git" AND git. patterntype:literal`),
			},
			APIDefPath: "internal/gitserver/protocol/gitserver.go",
			DependsOn:  []string{"repo-updater", "searcher"},
			OwnedBy:    "repo-mgmt",
			Tags:       []string{"infra"},
		},
		{
			Kind:            "SERVICE",
			Name:            "repo-updater",
			Lifecycle:       LifecycleProduction,
			SourceLocations: []SourceLocation{{Repo: "github.com/sourcegraph/sourcegraph", Paths: []string{"cmd/repo-updater", "enterprise/cmd/repo-updater"}}},
			UsagePatterns: []UsagePattern{
				newQueryUsagePattern(`lang:go REPO_UPDATER`),
				newQueryUsagePattern(`lang:go repoupdater.`),
				newQueryUsagePattern(`lang:go "github.com/sourcegraph/sourcegraph/internal/repoupdater" AND repoupdater patterntype:literal`),
			},
			DependsOn: []string{"gitserver", "github-proxy"},
			OwnedBy:   "repo-mgmt",
			Tags:      []string{"infra"},
		},
		{
			Kind:            "SERVICE",
			Name:            "searcher",
			Description:     "Provides on-demand unindexed search for repositories",
			Lifecycle:       LifecycleProduction,
			SourceLocations: []SourceLocation{{Repo: "github.com/sourcegraph/sourcegraph", Paths: []string{"cmd/searcher"}}},
			UsagePatterns: []UsagePattern{
				newQueryUsagePattern(`lang:go \bsearcher\.Search\( patterntype:regexp`),
			},
			DependsOn: []string{"gitserver"},
			OwnedBy:   "search-core",
			Tags:      []string{"search"},
		},
		{
			Kind:            "SERVICE",
			Name:            "executor",
			Lifecycle:       LifecycleExperimental,
			SourceLocations: []SourceLocation{{Repo: "github.com/sourcegraph/sourcegraph", Paths: []string{"enterprise/cmd/executor"}}},
			UsagePatterns: []UsagePattern{
				newQueryUsagePattern(`lang:go \bexecutor\b patterntype:regexp`),
			},
			DependsOn: []string{"frontend"},
			OwnedBy:   "code-intel",
		},
		{
			Kind:            "SERVICE",
			Name:            "precise-code-intel-worker",
			Lifecycle:       LifecycleProduction,
			SourceLocations: []SourceLocation{{Repo: "github.com/sourcegraph/sourcegraph", Paths: []string{"enterprise/cmd/precise-code-intel-worker"}}},
			UsagePatterns: []UsagePattern{
				newQueryUsagePattern(`lang:go \bprecise-code-intel-worker\b patterntype:regexp`),
			},
			DependsOn: []string{"frontend", "worker"},
			OwnedBy:   "code-intel",
			Tags:      []string{"worker"},
		},
		{
			Kind:            "SERVICE",
			Name:            "github-proxy",
			Lifecycle:       LifecycleProduction,
			SourceLocations: []SourceLocation{{Repo: "github.com/sourcegraph/sourcegraph", Paths: []string{"cmd/github-proxy"}}},
			UsagePatterns: []UsagePattern{
				newQueryUsagePattern(`GITHUB_PROXY`),
			},
			OwnedBy: "repo-mgmt",
		},
		{
			Kind:            "SERVICE",
			Name:            "query-runner",
			Lifecycle:       LifecycleProduction,
			SourceLocations: []SourceLocation{{Repo: "github.com/sourcegraph/sourcegraph", Paths: []string{"cmd/query-runner"}}},
			DependsOn:       []string{"frontend"},
			OwnedBy:         "search-product",
			Tags:            []string{"search"},
		},
		{
			Kind:            "SERVICE",
			Name:            "worker",
			Lifecycle:       LifecycleProduction,
			SourceLocations: []SourceLocation{{Repo: "github.com/sourcegraph/sourcegraph", Paths: []string{"cmd/worker", "enterprise/cmd/worker"}}},
			OwnedBy:         "code-intel",
			Tags:            []string{"worker"},
		},
		{
			Kind:            "SERVICE",
			Name:            "server",
			Lifecycle:       LifecycleProduction,
			SourceLocations: []SourceLocation{{Repo: "github.com/sourcegraph/sourcegraph", Paths: []string{"cmd/server", "enterprise/cmd/server"}}},
			DependsOn:       []string{"frontend", "repo-updater", "symbols", "query-runner", "gitserver"},
			OwnedBy:         "delivery",
		},
		{
			Kind:            "SERVICE",
			Name:            "symbols",
			Lifecycle:       LifecycleProduction,
			SourceLocations: []SourceLocation{{Repo: "github.com/sourcegraph/sourcegraph", Paths: []string{"cmd/symbols"}}},
			DependsOn:       []string{"gitserver", "frontend"},
			OwnedBy:         "code-intel",
		},
		{
			Kind:            "TOOL",
			Name:            "sitemap",
			Description:     "This tool is run offline to generate the sitemap files served at https://sourcegraph.com/sitemap.xml.",
			Lifecycle:       LifecycleExperimental,
			SourceLocations: []SourceLocation{{Repo: "github.com/sourcegraph/sourcegraph", Paths: []string{"cmd/sitemap"}}},
			DependsOn:       []string{"frontend"},
			OwnedBy:         "cloud-growth",
		},
		{
			Kind:            "TOOL",
			Name:            "sg",
			Description:     "The Sourcegraph developer tool",
			Lifecycle:       LifecycleProduction,
			SourceLocations: []SourceLocation{{Repo: "github.com/sourcegraph/sourcegraph", Paths: []string{"dev/sg"}}},
			UsagePatterns: []UsagePattern{
				newQueryUsagePattern(`lang:markdown ` + "`" + `sg[` + "`" + `\s] patterntype:regexp`),
				newQueryUsagePattern(`lang:markdown (^|\s*\$ )sg\s patterntype:regexp`),
			},
			OwnedBy: "dev-experience",
		},
		{
			Kind:        "TOOL",
			Name:        "src-cli",
			Description: "Sourcegraph CLI",
			Lifecycle:   LifecycleProduction,
			// Only the gitlab mirror of this repo is loaded by the default dev-private config.
			SourceLocations: []SourceLocation{{Repo: "github.com/sourcegraph/src-cli", Paths: []string{"."}}},
			UsagePatterns: []UsagePattern{
				newQueryUsagePattern(`lang:markdown ` + "`" + `src[` + "`" + `\s] patterntype:regexp`),
				newQueryUsagePattern(`lang:markdown (^|\s*\$ )src\s patterntype:regexp`),
			},
			DependsOn: []string{"frontend"},
			OwnedBy:   "batch-changes",
		},
		{
			Kind:            "LIBRARY",
			Name:            "client-web",
			Description:     "Main web app UI",
			Lifecycle:       LifecycleProduction,
			SourceLocations: []SourceLocation{{Repo: "github.com/sourcegraph/sourcegraph", Paths: []string{"client/web"}}},
			DependsOn:       []string{"extension-api", "client-shared", "wildcard", "frontend"},
			OwnedBy:         "frontend-platform",
			Tags:            []string{"ui"},
		},
		{
			Kind:            "LIBRARY",
			Name:            "client-browser",
			Description:     "Browser extension and native code host extension",
			Lifecycle:       LifecycleProduction,
			SourceLocations: []SourceLocation{{Repo: "github.com/sourcegraph/sourcegraph", Paths: []string{"client/browser"}}},
			DependsOn:       []string{"extension-api", "client-shared", "wildcard", "frontend"},
			OwnedBy:         "extensibility",
			Tags:            []string{"ui"},
		},
		{
			Kind:            "LIBRARY",
			Name:            "client-shared",
			Description:     "Frontend code shared by the web app and browser extension",
			Lifecycle:       LifecycleProduction,
			SourceLocations: []SourceLocation{{Repo: "github.com/sourcegraph/sourcegraph", Paths: []string{"client/shared"}}},
			UsagePatterns: []UsagePattern{
				newQueryUsagePattern(`lang:typescript import @sourcegraph/shared patterntype:regexp`),
			},
			DependsOn: []string{"extension-api", "wildcard", "frontend"},
			OwnedBy:   "frontend-platform",
			Tags:      []string{"ui"},
		},
		{
			Kind:            "LIBRARY",
			Name:            "wildcard",
			Description:     "The Wildcard component library is a collection of design-approved reusable components that are suitable for use within the Sourcegraph codebase.",
			Lifecycle:       LifecycleProduction,
			SourceLocations: []SourceLocation{{Repo: "github.com/sourcegraph/sourcegraph", Paths: []string{"client/wildcard"}}},
			UsagePatterns: []UsagePattern{
				newQueryUsagePattern(`lang:typescript import @sourcegraph/wildcard patterntype:regexp`),
			},
			OwnedBy: "frontend-platform",
			Tags:    []string{"ui"},
		},
		{
			Kind:            "LIBRARY",
			Name:            "extension-api",
			Description:     "Public TypeScript API for Sourcegraph extensions",
			Lifecycle:       LifecycleProduction,
			SourceLocations: []SourceLocation{{Repo: "github.com/sourcegraph/sourcegraph", Paths: []string{"client/extension-api"}}},
			UsagePatterns: []UsagePattern{
				newQueryUsagePattern(`lang:typescript import from ['"]sourcegraph['"] patterntype:regexp`),
			},
			OwnedBy: "extensibility",
			Tags:    []string{"public-api"},
		},
	}

	sampleRepos := []api.RepoName{
		"github.com/sourcegraph/PSSourcegraph",
		//  SLOW "github.com/sourcegraph/about",
		"github.com/sourcegraph/asdf-ds-to-dhall",
		"github.com/sourcegraph/automation-testing",
		"github.com/sourcegraph/campaigns-client",
		"github.com/sourcegraph/checkup",
		"github.com/sourcegraph/codenav-bash",
		"github.com/sourcegraph/codenotify",
		"github.com/sourcegraph/coif-to-lsif",
		"github.com/sourcegraph/create",
		"github.com/sourcegraph/create-extension",
		"github.com/sourcegraph/datadog-test-services",
		"github.com/sourcegraph/deploy-k8s-helper",
		"github.com/sourcegraph/deploy-sourcegraph",
		"github.com/sourcegraph/deploy-sourcegraph-docker",
		"github.com/sourcegraph/distribution-tools",
		"github.com/sourcegraph/docsite",
		"github.com/sourcegraph/ds-to-dhall",
		"github.com/sourcegraph/eslint-formatter-lsif",
		"github.com/sourcegraph/extension-api-classes",
		"github.com/sourcegraph/ghdump",
		"github.com/sourcegraph/go-ctags",
		"github.com/sourcegraph/go-diff",
		"github.com/sourcegraph/go-imports-search",
		"github.com/sourcegraph/go-jsonschema",
		"github.com/sourcegraph/go-langserver",
		"github.com/sourcegraph/gosyntect",
		"github.com/sourcegraph/handbook",
		"github.com/sourcegraph/jsonrpc2",
		"github.com/sourcegraph/jsonschemadoc",
		"github.com/sourcegraph/jsonx",
		"github.com/sourcegraph/learn",
		"github.com/sourcegraph/lsif-clang",
		"github.com/sourcegraph/lsif-demos",
		"github.com/sourcegraph/lsif-go",
		"github.com/sourcegraph/lsif-java",
		"github.com/sourcegraph/lsif-jsonnet",
		"github.com/sourcegraph/lsif-static-doc",
		"github.com/sourcegraph/lsp-adapter",
		"github.com/sourcegraph/nodepacker",
		"github.com/sourcegraph/org-search",
		"github.com/sourcegraph/prettierrc",
		"github.com/sourcegraph/promgrep",
		"github.com/sourcegraph/prototools",
		"github.com/sourcegraph/puppeteer-interceptor-poc",
		"github.com/sourcegraph/query-performance-testing",
		"github.com/sourcegraph/react-loading-spinner",
		"github.com/sourcegraph/renovate-config",
		"github.com/sourcegraph/renovate-dependency-test",
		"github.com/sourcegraph/resource-estimator",
		"github.com/sourcegraph/search-badger",
		"github.com/sourcegraph/shortcut",
		"github.com/sourcegraph/sourcegraph-atom",
		"github.com/sourcegraph/sourcegraph-code-ownership",
		"github.com/sourcegraph/sourcegraph-datadog-metrics",
		"github.com/sourcegraph/sourcegraph-hubspot",
		"github.com/sourcegraph/sourcegraph-jira",
		"github.com/sourcegraph/sourcegraph-lightstep",
		"github.com/sourcegraph/sourcegraph-npm-audit",
		"github.com/sourcegraph/sourcegraph-open-in-atom",
		"github.com/sourcegraph/sourcegraph-open-in-intellij",
		"github.com/sourcegraph/sourcegraph-open-in-sublime",
		"github.com/sourcegraph/sourcegraph-open-in-vscode",
		"github.com/sourcegraph/sourcegraph-project-view",
		"github.com/sourcegraph/sourcegraph-search-insights",
		"github.com/sourcegraph/sourcegraph-server-gen",
		"github.com/sourcegraph/sourcegraph-test-leaderboard",
		"github.com/sourcegraph/sourcegraph-typescript",
		"github.com/sourcegraph/sourcegraph-vscode",
		"github.com/sourcegraph/stylelint-config",
		"github.com/sourcegraph/themes",
		"github.com/sourcegraph/tsconfig",
		"github.com/sourcegraph/zoekt",
	}
	for _, repo := range sampleRepos {
		name := strings.TrimPrefix(string(repo), "github.com/sourcegraph/")

		var kind string
		if strings.HasPrefix(name, "sourcegraph-") || strings.HasPrefix(name, "create") || name == "go-langserver" {
			kind = "APPLICATION"
		} else if strings.HasPrefix(name, "deploy-") {
			kind = "OTHER"
		} else if name == "handbook" || name == "about" {
			kind = "WEBSITE"
		} else {
			kind = "LIBRARY"
		}

		components = append(components, Component{
			Kind:            kind,
			Name:            name,
			Lifecycle:       LifecycleProduction,
			SourceLocations: []SourceLocation{{Repo: repo, Paths: []string{""}}},
			OwnedBy:         "",
			Tags:            []string{"repo"},
		})
	}
	sort.Slice(components, func(i, j int) bool { return components[i].Name < components[j].Name })

	var edges []Edge
	for _, c := range components {
		for _, dependsOn := range c.DependsOn {
			edges = append(edges, Edge{
				Type: DependsOnRelation,
				Out:  c.Name,
				In:   dependsOn,
			})
			edges = append(edges, Edge{
				Type: DependencyOfRelation,
				Out:  dependsOn,
				In:   c.Name,
			})
		}
	}

	groups := []Group{
		{
			Name:        "product-eng",
			Title:       "Product and engineering",
			Description: "The entire product & engineering team.",
			Members:     []string{"nick-snyder", "christina-forney", "beyang-liu"},
		},

		{
			Name:        "code-graph",
			Title:       "Code graph",
			Description: "The Code Graph org’s mission is to build the code graph to make working with code easier, regardless of how much you have, how complex it is, where you store it, or even how technical you are.",
			ParentGroup: "product-eng",
			Members:     []string{"yink-teo"},
		},
		{
			Name:        "search-core",
			Title:       "Search core",
			Description: "The search core team owns all parts of Sourcegraph that map an interpreted search query to a set of results: indexed and unindexed search (Zoekt & Searcher), diff/commit search, and result ranking.",
			ParentGroup: "code-graph",
			Members:     []string{"jeffwarner", "ryanhitchman", "stefanhengl", "tomas", "keegancs"},
		},
		{
			Name:        "search-product",
			Title:       "Search product",
			Description: "The search product team owns all parts of Sourcegraph that help users Compose search queries and navigate search results: search field, search results UI, search contexts, query language (including structural search), the search homepage, homepage panels, and repogroup pages. It also owns a subset of features built on top of Sourcegraph search: code monitoring and saved searches.",
			ParentGroup: "code-graph",
			Members:     []string{"lguychard", "fkling", "ccheek", "rok", "juliana", "rijnard"},
		},
		{
			Name:        "code-intel",
			Title:       "Code intelligence",
			Description: "The Code Intelligence team builds tools and services that provide contextual information around code, taking into account its lexical, syntactic, and semantic structure.",
			ParentGroup: "code-graph",
			Members:     []string{"oconvey", "vgandhi", "cesarj", "chrismwendt", "teej", "olaf", "noahsc", "efritz"},
		},
		{
			Name:        "batch-changes",
			Title:       "Batch Changes",
			Description: "Batch Changes is a tool to find code that needs to be changed and change it at scale by running code. ",
			ParentGroup: "code-graph",
			Members:     []string{"chris-pine", "kelli-rockwell", "adeola-akinsiku", "adam-harvey", "erik-seliger", "thorsten-ball"},
		},
		{
			Name:        "code-insights",
			Title:       "Code insights",
			Description: "The code insights team is responsible for building and delivering code insights to engineering leaders, empowering data-driven decisions in engineering organizations.",
			ParentGroup: "code-graph",
			Members:     []string{"felix-becker", "cristina-birkel", "justin-boyson", "coury-clark", "vova-kulikov"},
		},

		{
			Name:        "enablement",
			Title:       "Enablement",
			Description: "Technical foundations critical to the business, our customers, and our products. We do this by ensuring we have the best tools, processes, and services in place, for use by both customers and our own engineers when using or developing Sourcegraph.",
			ParentGroup: "product-eng",
			Members:     []string{"jean-du-plessis"},
		},
		{
			Name:        "repo-mgmt",
			Title:       "Repo management",
			Description: "Maintain and evolve the methods by which code is pulled into Sourcegraph from code hosts, in a way that supports all required functionality while maximizing performance, reliability, and ease of use.",
			ParentGroup: "enablement",
			Members:     []string{"jplahn", "indrag", "rslade", "mweitzel", "alex-ostrikov"},
		},
		{
			Name:        "delivery",
			Title:       "Delivery",
			Description: "Enable any Sourcegraph customer or user to trial or run (in production) Sourcegraph in a way fits within their environment, supports their level of technical expertise and allows them to easily access the value that our product provides.",
			ParentGroup: "enablement",
			Members:     []string{"jean-du-plessis", "kevin-wojkovich", "crystal-augustus"},
		},
		{
			Name:        "dev-experience",
			Title:       "Dev experience",
			Description: "The Dev Experience team, or DevX for short, is a team focused on improving the developer experience of Sourcegraph.",
			ParentGroup: "enablement",
			Members:     []string{"kristen-stretch", "jh-chabran", "robert-lin", "dave-try"},
		},
		{
			Name:        "frontend-platform",
			Title:       "Frontend platform",
			Description: "The Frontend Platform team defines and maintains the standards and tools for web development at Sourcegraph.",
			ParentGroup: "enablement",
			Members:     []string{"patrick-dubroy", "valeryb", "tomross"},
		},

		{
			Name:        "cloud",
			Title:       "Cloud",
			Description: "Offer the fastest, most seamless way for development teams to bring Sourcegraph into their workflows, wherever they are.",
			ParentGroup: "product-eng",
			Members:     []string{"billcreager"},
		},
		{
			Name:        "cloud-growth",
			Title:       "Cloud growth",
			Description: "The cloud growth team focuses on the growth of Sourcegraph.com, both as a SaaS product and lead generator for Sourcegraph as an on-prem product.",
			ParentGroup: "cloud",
			Members:     []string{"stephengutekanst"},
		},
		{
			Name:        "extensibility",
			Title:       "Extensibility",
			Description: "The extensibility team owns our code host and third-party integrations (including our browser extension) and our Sourcegraph extensions. Our mission is to bring the value of Sourcegraph to everywhere you work with code and to bring the value of other developer tools into Sourcegraph.",
			ParentGroup: "cloud",
			Members:     []string{"murat-sutunc", "erzhan-torokulov", "beatrix-woo", "tharuntej-kandala"},
		},
		{
			Name:        "security",
			Title:       "Security",
			Description: "We think that security is an enabler for the business. Sourcegraph is committed to proactive security, and addressing vulnerabilities in a timely manner. We approach security with a can-do philosophy, and look to achieve product goals while maintaining a positive posture, and increasing our security stance over time.",
			ParentGroup: "cloud",
			Members:     []string{"diego-comas", "lauren-chapman", "david-sandy", "mohammad-alam", "andre-eleuterio"},
		},
		{
			Name:        "devops",
			Title:       "DevOps",
			Description: "The two primary pillars of the Cloud Team are Availability and Observability as defined in RFC 498. This team ensures that Sourcegraph.com has the same reliability and availability as other world-class SaaS offerings. This team is also responsible for Observability monitoring and tooling to ensure that we are meeting these goals.",
			ParentGroup: "cloud",
			Members:     []string{"jennifer-mitchell", "dax-mcdonald", "daniel-dides"},
		},
		{
			Name:        "cloud-saas",
			Title:       "Cloud SaaS",
			Description: "The Cloud Software-as-a-Service (SaaS) Team is responsible for the service management part of the Sourcegraph Cloud product. The team provides both customer-facing and internal capabilities that enable our customers to use the Sourcegraph Cloud product as a service.",
			ParentGroup: "cloud",
			Members:     []string{"rafal-leszczynski", "rafal-gajdulewicz", "milan-freml", "artem-ruts", "joe-cheng"},
		},
	}
	for _, g := range groups {
		if g.ParentGroup != "" {
			edges = append(edges, Edge{
				Type: PartOfRelation,
				Out:  g.Name,
				In:   g.ParentGroup,
			})
			edges = append(edges, Edge{
				Type: HasPartRelation,
				Out:  g.ParentGroup,
				In:   g.Name,
			})
		}
	}

	computedLabels := []ComputedLabel{
		{
			Key: "lang",
			ValueQueries: map[string]string{
				// TODO(sqs): want to be able to define it like this
				//
				// "go":         `component:contains.file(\.go$)`,
				// "typescript": `component:contains.file(\.tsx?$)`,
				"go":         `file:\.go$`,
				"typescript": `file:\.tsx?$`,
				"javascript": `file:\.jsx?$`,
			},
		},
		{
			Key: "stale",
			ValueQueries: map[string]string{
				"!true": `type:diff after:2weeks`,
			},
		},
		{
			Key: "no-license",
			ValueQueries: map[string]string{
				"!true": `file:^LICENSE$`,
			},
		},
		{
			Key: "codecov",
			ValueQueries: map[string]string{
				"true": `file:(^|/)\.?codecov\.ya?ml$ OR https://codecov\.io/bash`,
			},
		},
	}

	return components, groups, edges, computedLabels
}

func Components() []Component { return append([]Component(nil), components...) }
func ComponentByName(name string) *Component {
	for _, c := range Components() {
		if c.Name == name {
			return &c
		}
	}
	return nil
}

func Groups() []Group { return append([]Group(nil), groups...) }
func GroupByName(name string) *Group {
	for _, g := range Groups() {
		if g.Name == name {
			return &g
		}
	}
	return nil
}

func Edges() []Edge { return append([]Edge(nil), edges...) }

func ComputedLabels() []ComputedLabel { return append([]ComputedLabel(nil), computedLabels...) }
