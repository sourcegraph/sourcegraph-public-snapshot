package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/grafana/regexp"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var imagesCommand = &cli.Command{
	Name:     "images",
	Usage:    "Commands for interacting with containers",
	Category: category.Dev,
	Subcommands: []*cli.Command{
		{
			Name:  "list",
			Usage: "List container images available to build",
			Action: func(cctx *cli.Context) error {
				lines, err := listBazelOCITarballs(cctx.Context)
				if err != nil {
					return err
				}
				std.Out.WriteLine(output.Styledf(output.StylePending, "Found %d targets", len(lines)))
				std.Out.WriteLine(output.Styledf(output.StyleReset, strings.Join(lines, "\n")))
				std.Out.WriteLine(output.Styledf(output.StylePending, "💡 You can build and load the above targets using glob patterns with 'sg images build [pattern]..."))
				std.Out.WriteLine(output.Styledf(output.StylePending, "💡 Examples:"))
				std.Out.WriteLine(output.Styledf(output.StyleReset, "  'worker' to match //cmd/worker:image_tarball"))
				std.Out.WriteLine(output.Styledf(output.StyleReset, "  'cmd/*' to match all containers under //cmd/"))
				std.Out.WriteLine(output.Styledf(output.StyleReset, "  'postgres*' to match"))
				std.Out.WriteLine(output.Styledf(output.StyleReset, "    //docker-images/postgres-12-alpine:base_tarball//docker-images/postgres-12-alpine:image_tarball"))
				std.Out.WriteLine(output.Styledf(output.StyleReset, "    //docker-images/postgres-12-alpine:legacy_tarball"))
				std.Out.WriteLine(output.Styledf(output.StyleReset, "    //docker-images/postgres_exporter:base_tarball"))
				std.Out.WriteLine(output.Styledf(output.StyleReset, "    //docker-images/postgres_exporter:image_tarball"))

				return nil
			},
		},
		{
			Name:      "build",
			Usage:     "builds a container image by matching [pattern] using glob syntax to the target",
			ArgsUsage: "build [pattern1] ([pattern2] ...)",
			UsageText: `
sg images build worker
# Build everything under 'cmd/*'
sg images build cmd/*
# Build all postgres images
sg images build postgres*
`,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "load",
					Usage: "Load the image into the local Docker daemon",
					Value: true,
				},
			},
			Action: func(cctx *cli.Context) error {
				allTargets, err := listBazelOCITarballs(cctx.Context)
				if err != nil {
					return err
				}

				patterns := cctx.Args().Slice()
				selectedTargets := []string{}

				for _, name := range patterns {
					if strings.HasPrefix(name, "//") {
						std.Out.WriteLine(output.Styledf(output.StyleYellow, "Detected a Bazel target path (%q) aborting.", name))
						std.Out.WriteLine(output.Styledf(output.StyleBold, "Run the following command instead:"))
						std.Out.WriteLine(output.Styledf(output.StyleReset, "Building the image without loading it:\n  sg bazel build %s", name))
						std.Out.WriteLine(output.Styledf(output.StyleReset, "Building the image and loading it:\n  sg bazel run %s", name))
						return nil
					}
				}

				for _, target := range allTargets {
					for _, name := range patterns {
						var ok bool
						var err error
						if strings.Contains(name, "/") {
							ok, err = filepath.Match(name, trimImageTarballTarget(target))
						} else {
							ok, err = filepath.Match(fmt.Sprintf("*/%s", name), trimImageTarballTarget(target))
						}
						if err != nil {
							return err
						}
						if ok {
							selectedTargets = append(selectedTargets, target)
						}
					}
				}

				std.Out.WriteLine(output.Styledf(output.StylePending, "Found Bazel targets: \n%s", strings.Join(selectedTargets, "\n")))

				commandText := fmt.Sprintf("sg bazel build %s", strings.Join(selectedTargets, " "))
				std.Out.WriteLine(output.Styledf(output.StyleBold, "Running Bazel 'build' command for you"))
				std.Out.WriteLine(output.Styledf(output.StyleYellow, "  "+commandText))
				std.Out.WriteLine(output.Styledf(output.StyleReset, "--- 👇 Bazel output ---"))

				// Please note the added --color flag here, to ensure we keep the colors when streaming back the output.
				// And we run `bazel` directly, because we know that for build/run you don't need `sg bazel` but our users
				// don't have to know that.
				cmd := run.Bash(cctx.Context, fmt.Sprintf("bazel build --color=yes %s", strings.Join(selectedTargets, " ")))
				if err := cmd.Run().Stream(os.Stdout); err != nil {
					return err
				}
				std.Out.WriteLine(output.Styledf(output.StyleReset, "----------------------"))

				if cctx.Bool("load") {
					for _, target := range selectedTargets {
						commandText := fmt.Sprintf("bazel run %s", target)
						std.Out.WriteLine(output.Styledf(output.StyleBold, "Running Bazel command for you"))
						std.Out.WriteLine(output.Styledf(output.StyleYellow, "  "+commandText))
						std.Out.WriteLine(output.Styledf(output.StyleReset, "--- 👇 Bazel output ---"))

						// Please note the added --color flag here, to ensure we keep the colors when streaming back the output.
						cmd := run.Bash(cctx.Context, fmt.Sprintf("sg bazel run --color=yes %s", target))
						if err := cmd.Run().Stream(os.Stdout); err != nil {
							return err
						}
						std.Out.WriteLine(output.Styledf(output.StyleReset, "----------------------"))
					}
				}
				return nil
			},
		},
	},
}

func listBazelOCITarballs(ctx context.Context) ([]string, error) {
	cmd := run.Cmd(ctx, "bazel", "query", "kind('oci_tarball', //...)")
	lines, err := cmd.Run().Lines()

	if err != nil {
		return nil, errors.Wrap(err, "failed to list bazel images")
	}

	if len(lines) > 0 {
		if strings.HasPrefix(lines[0], "Loading") {
			// Trim the first line, which is just "Loading: (...)"
			lines = lines[1:]
		}
	}

	return lines, nil
}

// trimImageTarballTarget simplifies the Bazel target path for easier matching with glob syntax.
// For example, the target '//cmd/worker:image_tarball' becomes 'worker'.
func trimImageTarballTarget(target string) string {
	// Remove the leading '//'
	target = strings.TrimPrefix(target, "//")

	// Remove the trailing ':.*'
	reg := regexp.MustCompile(":.*")
	target = reg.ReplaceAllString(target, "")

	return target
}
