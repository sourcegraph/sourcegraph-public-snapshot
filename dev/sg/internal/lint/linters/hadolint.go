package linters

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/download"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/lint"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func hadolint() lint.Runner {
	const header = "Hadolint"
	const hadolintVersion = "v2.10.0"
	hadolintBinary := fmt.Sprintf("./.bin/hadolint-%s", hadolintVersion)

	runHadolint := func(ctx context.Context, files []string) *lint.Report {
		out, err := run.Cmd(ctx, "xargs "+hadolintBinary).
			Input(strings.NewReader(strings.Join(files, "\n"))).
			Run().
			Lines()

		return &lint.Report{
			Header: header,
			Output: strings.Join(out, "\n"),
			Err:    err,
		}
	}

	return func(ctx context.Context, s *repo.State) *lint.Report {
		diff, err := s.GetDiff("**/*Dockerfile*")
		if err != nil {
			return &lint.Report{Header: header, Err: err}
		}
		var dockerfiles []string
		for f := range diff {
			dockerfiles = append(dockerfiles, f)
		}
		if len(dockerfiles) == 0 {
			return &lint.Report{
				Header: header,
				Output: "No Dockerfiles changed",
			}
		}

		// If our binary is already here, just go!
		if _, err := os.Stat(hadolintBinary); err == nil {
			return runHadolint(ctx, dockerfiles)
		}

		// https://github.com/hadolint/hadolint/releases for downloads
		var distro, arch string
		switch runtime.GOARCH {
		case "arm64":
			arch = "arm64"
		default:
			arch = "x86_64"
		}
		switch runtime.GOOS {
		case "darwin":
			distro = "Darwin"
			arch = "x86_64"
		case "windows":
			distro = "Windows"
		default:
			distro = "Linux"
		}
		url := fmt.Sprintf("https://github.com/hadolint/hadolint/releases/download/%s/hadolint-%s-%s",
			hadolintVersion, distro, arch)

		// Download
		os.MkdirAll("./.bin", os.ModePerm)
		std.Out.WriteNoticef("Downloading hadolint from %s", url)
		if err := download.Exeuctable(url, hadolintBinary); err != nil {
			return &lint.Report{
				Header: header,
				Err:    errors.Wrap(err, "downloading hadolint"),
			}
		}

		return runHadolint(ctx, dockerfiles)
	}
}
