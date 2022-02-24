package main

import (
	"context"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/docker"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	lintShellFlagSet   = flag.NewFlagSet("sg lint shell", flag.ExitOnError)
	lintURLsFlagSet    = flag.NewFlagSet("sg lint urls", flag.ExitOnError)
	lintGoFlagSet      = flag.NewFlagSet("sg lint go", flag.ExitOnError)
	lintDocsiteFlagSet = flag.NewFlagSet("sg lint docsite", flag.ExitOnError)
	lintDockerFlagSet  = flag.NewFlagSet("sg lint docker", flag.ExitOnError)
	lintClientFlagSet  = flag.NewFlagSet("sg lint client", flag.ExitOnError)
)

var allLintTargets = lintTargets{
	{
		Name:    "urls",
		Help:    "Check for broken urls in the codebase.",
		FlagSet: lintURLsFlagSet,
		Linters: []lint.Runner{
			lint.RunScript("Broken urls", "dev/check/broken-urls.bash"),
		},
	},
	{
		Name:    "go",
		Help:    "Check go code for linting errors, forbidden imports, generated files...",
		FlagSet: lintGoFlagSet,
		Linters: []lint.Runner{
			lint.RunScript("Go format", "dev/check/gofmt.sh"),
			lintGoGenerate,
			lint.RunScript("Go lint", "dev/check/go-lint.sh"),
			lint.RunScript("Go pkg/database/dbconn", "dev/check/go-dbconn-import.sh"),
			lint.RunScript("Go enterprise imports in OSS", "dev/check/go-enterprise-import.sh"),
		},
	},
	{
		Name:    "docsite",
		Help:    "Check the code powering docs.sourcegraph.com for broken links and linting errors.",
		FlagSet: lintDocsiteFlagSet,
		Linters: []lint.Runner{
			lint.RunScript("Docsite lint", "dev/check/docsite.sh"),
		},
	},
	{
		Name:    "docker",
		Help:    "Check Dockerfiles for Sourcegraph best practices",
		FlagSet: lintDockerFlagSet,
		Linters: []lint.Runner{
			lint.RunScript("Docker lint", "dev/check/docker-lint.sh"),
			lintDockerfiles(),
		},
	},
	{
		Name:    "client",
		Help:    "Check client code for linting errors, forbidden imports, ...",
		FlagSet: lintClientFlagSet,
		Linters: []lint.Runner{
			lint.RunScript("Typescript imports in OSS", "dev/check/ts-enterprise-import.sh"),
			lint.RunScript("Inline templates", "dev/check/template-inlines.sh"),
			lint.RunScript("Yarn duplicate", "dev/check/yarn-deduplicate.sh"),
			lint.RunScript("SVG Compression", "dev/check/svgo.sh"),
		},
	},
	{
		Name:    "shell",
		Help:    "Check shell code for linting errors, formatting, ...",
		FlagSet: lintShellFlagSet,
		Linters: []lint.Runner{
			lint.RunScript("Shell formatting", "dev/check/shfmt.sh"),
			lint.RunScript("Shell lint", "dev/check/shellcheck.sh"),
		},
	},
}

// lintDockerfiles runs custom Sourcegraph Dockerfile linters
func lintDockerfiles() lint.Runner {
	return func(ctx context.Context) *lint.Report {
		start := time.Now()
		var combinedErrors error
		for _, dir := range []string{
			"docker-images",
			// cmd dirs
			"cmd",
			"enterprise/cmd",
			"internal/cmd",
			// dev dirs
			"dev",
			"enterprise/dev",
		} {
			if err := filepath.Walk(dir,
				func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !strings.Contains(filepath.Base(path), "Dockerfile") {
						return nil
					}
					data, err := os.ReadFile(path)
					if err != nil {
						return err
					}

					if err := docker.ProcessDockerfile(data, docker.LintDockerfile(path)); err != nil {
						// track error but don't exit
						combinedErrors = errors.Append(combinedErrors, err)
					}

					return nil
				},
			); err != nil {
				combinedErrors = errors.Append(combinedErrors, err)
			}
		}
		return &lint.Report{
			Duration: time.Since(start),
			Header:   "Sourcegraph Dockerfile linters",
			Output: func() string {
				if combinedErrors != nil {
					return strings.TrimSpace(combinedErrors.Error())
				}
				return ""
			}(),
			Err: combinedErrors,
		}
	}
}

func lintGoGenerate(ctx context.Context) *lint.Report {
	start := time.Now()
	err := generateDo(ctx, nil, generateQuiet)
	if err != nil {
		return &lint.Report{
			Header:   "Go generate check",
			Duration: time.Since(start),
			Err:      err,
		}
	}

	cmd := exec.CommandContext(ctx, "git", "diff", "--exit-code", "--", ".", ":!go.sum")
	out, err := cmd.CombinedOutput()
	r := lint.Report{
		Header:   "Go generate check",
		Duration: time.Since(start),
	}
	if err != nil {
		r.Err = err
		r.Output = string(out)
		return &r
	}

	return &r
}
