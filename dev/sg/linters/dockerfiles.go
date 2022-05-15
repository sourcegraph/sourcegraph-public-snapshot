package linters

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/docker"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// customDockerfileLinters runs custom Sourcegraph Dockerfile linters
func customDockerfileLinters() lint.Runner {
	return func(ctx context.Context, _ *repo.State) *lint.Report {
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

					// Define docker lints in a separate package because they are quite
					// involved.
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
			Header: "Sourcegraph Dockerfile linters",
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
