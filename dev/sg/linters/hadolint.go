package linters

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/internal/download"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func hadolint() *linter {
	const hadolintVersion = "v2.10.0"
	hadolintBinary := fmt.Sprintf("./.bin/hadolint-%s", hadolintVersion)

	runHadolint := func(ctx context.Context, out *std.Output, files []string) error {
		return run.Cmd(ctx, "xargs "+hadolintBinary).
			Input(strings.NewReader(strings.Join(files, "\n"))).
			Run().
			StreamLines(out.Verbose)
	}

	return runCheck("Hadolint", func(ctx context.Context, out *std.Output, state *repo.State) error {
		diff, err := state.GetDiff("**/*Dockerfile*")
		if err != nil {
			return err
		}
		var dockerfiles []string
		for f := range diff {
			dockerfiles = append(dockerfiles, f)
		}
		if len(dockerfiles) == 0 {
			out.Verbose("no dockerfiles changed")
			return nil
		}

		// If our binary is already here, just go!
		if _, err := os.Stat(hadolintBinary); err == nil {
			return runHadolint(ctx, out, dockerfiles)
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
		if _, err := download.Executable(ctx, url, hadolintBinary, false); err != nil {
			return errors.Wrap(err, "downloading hadolint")
		}

		return runHadolint(ctx, out, dockerfiles)
	})
}
