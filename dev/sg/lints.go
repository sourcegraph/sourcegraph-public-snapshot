package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/docker"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/generate/golang"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var allLintTargets = lintTargets{
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
		},
	},
	{
		Name: "logger-migration",
		Help: "Run linter that enforces the new logger library",
		Linters: []lint.Runner{
			lintLoggingLibraries(),
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
			lint.RunScript("Docker lint", "dev/check/docker-lint.sh"),
			lintDockerfiles(),
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

// lintLoggingLibraries enforces that only usages of lib/log are added
func lintLoggingLibraries() lint.Runner {
	const header = "Logging library linter"

	var (
		bannedImports = []string{
			// No standard log library
			`"log"`,
			// No log15
			`"github.com/inconshreveable/log15"`,
			`log15.`,
			// No zap - we re-rexport everything via lib/log
			`"go.uber.org/zap"`,
			`"go.uber.org/zap/zapcore"`,
		}

		allowedFiles = map[string]struct{}{
			// Banned imports will match on the linter here
			"dev/sg/lints.go": {},
			// We re-export things here
			"lib/log": {},
			// We allow one usage of a direct zap import here
			"internal/observation/fields.go": {},
		}
	)

	// checkHunk returns an error if a banned library is used
	checkHunk := func(file string, hunk repo.DiffHunk) error {
		if _, allowed := allowedFiles[file]; allowed {
			return nil
		}

		for _, l := range hunk.AddedLines {
			for _, banned := range bannedImports {
				if strings.Contains(l, banned) {
					return errors.Newf(`%s:%d: banned usage of '%s': use "github.com/sourcegraph/sourcegraph/lib/log" instead`,
						file, hunk.StartLine, banned)
				}
			}
		}
		return nil
	}

	return func(ctx context.Context, state *repo.State) *lint.Report {
		start := time.Now()

		diffs, err := state.GetDiff("**/*.go")
		if err != nil {
			return &lint.Report{
				Header: header,
				Err:    err,
			}
		}

		var errs error
		for file, hunks := range diffs {
			for _, hunk := range hunks {
				if err := checkHunk(file, hunk); err != nil {
					errs = errors.Append(errs, err)
				}
			}
		}

		return &lint.Report{
			Duration: time.Since(start),
			Header:   header,
			Output: func() string {
				if errs != nil {
					return strings.TrimSpace(errs.Error()) +
						"\n\nLearn more about logging and why some libraries are banned: https://docs.sourcegraph.com/dev/how-to/add_logging"
				}
				return ""
			}(),
			Err: errs,
		}
	}
}

// lintDockerfiles runs custom Sourcegraph Dockerfile linters
func lintDockerfiles() lint.Runner {
	return func(ctx context.Context, _ *repo.State) *lint.Report {
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

func lintGoGenerate(ctx context.Context, _ *repo.State) *lint.Report {
	start := time.Now()
	report := golang.Generate(ctx, nil, golang.QuietOutput)
	if report.Err != nil {
		return &lint.Report{
			Header:   "Go generate check",
			Duration: report.Duration + time.Since(start),
			Err:      report.Err,
		}
	}

	cmd := exec.CommandContext(ctx, "git", "diff", "--exit-code", "--", ".", ":!go.sum")
	out, err := cmd.CombinedOutput()
	r := lint.Report{
		Header:   "Go generate check",
		Duration: report.Duration + time.Since(start),
	}
	if err != nil {
		var sb strings.Builder
		reportOut := output.NewOutput(&sb, output.OutputOpts{
			ForceColor: true,
			ForceTTY:   true,
		})
		reportOut.WriteLine(output.Line(output.EmojiFailure, output.StyleWarning, "Uncommitted changes found after running go generate:"))
		sb.WriteString("\n")
		sb.WriteString(string(out))
		r.Err = err
		r.Output = sb.String()
		return &r
	}

	return &r
}
