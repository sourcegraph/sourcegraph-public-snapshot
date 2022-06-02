// Package linters defines all available linters.
package linters

import (
	"github.com/Masterminds/semver"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"
)

// Targets lists all available linter targets. Each target consists of multiple linters.
//
// These should align with the names in 'enterprise/dev/ci/internal/ci/changed'
var Targets = []lint.Target{
	{
		Name: "urls",
		Help: "Check for broken urls in the codebase",
		Linters: []lint.Linter{
			lint.ScriptCheck("Broken urls", "dev/check/broken-urls.bash"),
		},
	},
	{
		Name: "go",
		Help: "Check go code for linting errors, forbidden imports, generated files, etc",
		Linters: []lint.Linter{
			goFmt,
			&goGenerateLinter{},
			lint.FuncCheck(goLint()),
			goDBConnImport,
			goEnterpriseImport,
			noLocalHost,
			lint.FuncCheck(lintGoDirectives),
			newLoggingLibraryLinter(),
			&goModVersionsLinter{
				maxVersions: map[string]*semver.Version{
					// Any version past this version is not yet released in any version of Alertmanager,
					// and causes incompatibility in prom-wrapper.
					//
					// https://github.com/sourcegraph/zoekt/pull/330#issuecomment-1116857568
					"github.com/prometheus/common": semver.MustParse("v0.32.1"),
				},
			},
			lint.FuncCheck(lintSGExit()),
		},
	},
	{
		Name: "docs",
		Help: "Documentation checks",
		Linters: []lint.Linter{
			lint.ScriptCheck("Docsite lint", "dev/docsite.sh check"),
		},
	},
	{
		Name: "dockerfiles",
		Help: "Check Dockerfiles for Sourcegraph best practices",
		Linters: []lint.Linter{
			lint.FuncCheck(hadolint()),
			lint.FuncCheck(customDockerfileLinters()),
		},
	},
	{
		Name: "client",
		Help: "Check client code for linting errors, forbidden imports, etc",
		Linters: []lint.Linter{
			tsEnterpriseImport,
			inlineTemplates,
			lint.ScriptCheck("Yarn duplicate", "dev/check/yarn-deduplicate.sh"),
			lint.FuncCheck(checkUnversionedDocsLinks()),
		},
	},
	{
		Name: "svg",
		Help: "Check svg assets",
		Linters: []lint.Linter{
			lint.FuncCheck(checkSVGCompression()),
		},
	},
	{
		Name: "shell",
		Help: "Check shell code for linting errors, formatting, etc",
		Linters: []lint.Linter{
			shFmt,
			shellCheck,
			bashSyntax,
		},
	},
}
