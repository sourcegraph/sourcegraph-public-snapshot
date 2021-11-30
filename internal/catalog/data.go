package catalog

import (
	"sort"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

type Component struct {
	Kind          string
	Name          string
	Description   string
	System        *string
	SourceRepo    api.RepoName
	SourceCommit  api.CommitID
	SourcePaths   []string
	UsagePatterns []UsagePattern
	APIDefPath    string
}

type UsagePattern struct {
	Query string
}

func newQueryUsagePattern(query string) UsagePattern {
	return UsagePattern{
		Query: `count:13 repo:^github\.com/sourcegraph/sourcegraph$ ` + query,
	}
}

func Components() []Component {
	const (
		sourceRepo   = "github.com/sourcegraph/sourcegraph"
		sourceCommit = "2ada4911722e2c812cc4f1bbfb6d5d1756891392"
	)
	components := []Component{
		{
			Kind:         "SERVICE",
			Name:         "frontend",
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"cmd/frontend", "enterprise/cmd/frontend"},
			UsagePatterns: []UsagePattern{
				newQueryUsagePattern(`\.api/graphql patterntype:literal`),
				newQueryUsagePattern(`lang:typescript requestGraphQL\(|useConnection\(|useQuery\(|gql` + "`" + ` patterntype:regexp`),
			},
			APIDefPath: "cmd/frontend/graphqlbackend/schema.graphql",
		},
		{
			Kind:         "SERVICE",
			Name:         "gitserver",
			Description:  "Mirrors repositories from their code host.",
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"cmd/gitserver"},
			UsagePatterns: []UsagePattern{
				newQueryUsagePattern(`lang:go \bgitserver\.Client\b patterntype:regexp`),
				newQueryUsagePattern(`lang:go \bgit\.[A-Z]\w+\(ctx, patterntype:regexp`),
				newQueryUsagePattern(`lang:go "github.com/sourcegraph/sourcegraph/internal/vcs/git" patterntype:literal`),
			},
			APIDefPath: "internal/gitserver/protocol/gitserver.go",
		},
		{
			Kind:         "SERVICE",
			Name:         "repo-updater",
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"cmd/repo-updater", "enterprise/cmd/repo-updater"},
		},
		{
			Kind:         "SERVICE",
			Name:         "searcher",
			Description:  "Provides on-demand unindexed search for repositories",
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"cmd/searcher"},
			UsagePatterns: []UsagePattern{
				newQueryUsagePattern(`lang:go \bsearcher\.Search\( patterntype:regexp`),
			},
		},
		{
			Kind:         "SERVICE",
			Name:         "executor",
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"enterprise/cmd/executor"},
		},
		{
			Kind:         "SERVICE",
			Name:         "precise-code-intel-worker",
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"enterprise/cmd/precise-code-intel-worker"},
		},
		{
			Kind:         "SERVICE",
			Name:         "github-proxy",
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"cmd/github-proxy"},
		},
		{
			Kind:         "SERVICE",
			Name:         "query-runner",
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"cmd/query-runner"},
		},
		{
			Kind:         "SERVICE",
			Name:         "worker",
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"cmd/worker", "enterprise/cmd/worker"},
		},
		{
			Kind:         "SERVICE",
			Name:         "server",
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"cmd/server", "enterprise/cmd/server"},
		},
		{
			Kind:         "SERVICE",
			Name:         "symbols",
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"cmd/symbols"},
		},
		{
			Kind:         "SERVICE",
			Name:         "sitemap",
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"cmd/sitemap"},
		},
		{
			Kind:         "TOOL",
			Name:         "sg",
			Description:  "The Sourcegraph developer tool",
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"dev/sg"},
			UsagePatterns: []UsagePattern{
				newQueryUsagePattern(`lang:markdown ` + "`" + `sg[` + "`" + `\s] patterntype:regexp`),
				newQueryUsagePattern(`lang:markdown (^|\s*\$ )sg\s patterntype:regexp`),
			},
		},
		{
			Kind:        "TOOL",
			Name:        "src-cli",
			Description: "Sourcegraph CLI",
			// Only the gitlab mirror of this repo is loaded by the default dev-private config.
			SourceRepo:    "gitlab.sgdev.org/sourcegraph/src-cli",
			SourceCommit:  "4a4341bc1c53fc5306f09bdcb31e8892ee40e6c7",
			SourcePaths:   []string{"."},
			UsagePatterns: []UsagePattern{
				// newQueryUsagePattern(`lang:markdown ` + "`" + `src[` + "`" + `\s] patterntype:regexp`),
				// newQueryUsagePattern(`lang:markdown (^|\s*\$ )src\s patterntype:regexp`),
			},
		},
		{
			Kind:         "LIBRARY",
			Name:         "client-web",
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"client/web"},
		},
		{
			Kind:         "LIBRARY",
			Name:         "client-browser",
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"client/browser"},
		},
		{
			Kind:         "LIBRARY",
			Name:         "client-shared",
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"client/shared"},
		},
		{
			Kind:         "LIBRARY",
			Name:         "wildcard",
			Description:  "The Wildcard component library is a collection of design-approved reusable components that are suitable for use within the Sourcegraph codebase.",
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"client/wildcard"},
			UsagePatterns: []UsagePattern{
				newQueryUsagePattern(`lang:typescript import @sourcegraph/wildcard patterntype:regexp`),
			},
		},
		{
			Kind:         "LIBRARY",
			Name:         "extension-api",
			SourceRepo:   sourceRepo,
			SourceCommit: sourceCommit,
			SourcePaths:  []string{"client/extension-api"},
			UsagePatterns: []UsagePattern{
				newQueryUsagePattern(`lang:typescript import from ['"]sourcegraph['"] patterntype:regexp`),
			},
		},
	}
	sort.Slice(components, func(i, j int) bool { return components[i].Name < components[j].Name })
	return components
}

func ComponentByName(name string) *Component {
	for _, c := range Components() {
		if c.Name == name {
			return &c
		}
	}
	return nil
}
