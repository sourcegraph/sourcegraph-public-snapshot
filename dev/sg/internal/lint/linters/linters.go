// Package linters defines all available linters.
package linters

import (
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"
)

// Targets lists all available linter targets. Each target consists of multiple linters.
var Targets = []lint.Target{
	{
		Name: "urls",
		Help: "Check for broken urls in the codebase",
		Linters: []lint.Runner{
			lint.RunScript("Broken urls", "dev/check/broken-urls.bash"),
		},
	},
	{
		Name: "go",
		Help: "Check go code for linting errors, forbidden imports, generated files, etc",
		Linters: []lint.Runner{
			lint.RunScript("Go format", "dev/check/gofmt.sh"),
			lintGoGenerate,
			lint.RunScript("Go lint", "dev/check/go-lint.sh"),
			lint.RunScript("Go pkg/database/dbconn", "dev/check/go-dbconn-import.sh"),
			lint.RunScript("Go enterprise imports in OSS", "dev/check/go-enterprise-import.sh"),
			lintGoDirectives,
		},
	},
	{
		Name: "go-custom",
		Help: "Custom checks for Go, will be migrated to the default go check set in the future",
		Linters: []lint.Runner{
			lintLoggingLibraries(),
			goModGuards(),
		},
	},
	{
		Name: "docsite",
		Help: "Check the code powering docs.sourcegraph.com for broken links and linting errors",
		Linters: []lint.Runner{
			lint.RunScript("Docsite lint", "dev/check/docsite.sh"),
		},
	},
	{
		Name: "docker",
		Help: "Check Dockerfiles for Sourcegraph best practices",
		Linters: []lint.Runner{
			hadolint(),
			customDockerfileLinters(),
		},
	},
	{
		Name: "client",
		Help: "Check client code for linting errors, forbidden imports, etc",
		Linters: []lint.Runner{
			lint.RunScript("Typescript imports in OSS", "dev/check/ts-enterprise-import.sh"),
			lint.RunScript("Inline templates", "dev/check/template-inlines.sh"),
			lint.RunScript("Yarn duplicate", "dev/check/yarn-deduplicate.sh"),
			lint.RunScript("SVG Compression", "dev/check/svgo.sh"),
		},
	},
	{
		Name: "shell",
		Help: "Check shell code for linting errors, formatting, etc",
		Linters: []lint.Runner{
			lint.RunScript("Shell formatting", "dev/check/shfmt.sh"),
			lint.RunScript("Shell lint", "dev/check/shellcheck.sh"),
		},
	},
}
