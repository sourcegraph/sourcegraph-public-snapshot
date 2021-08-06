package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
)

const defaultDocDir = "doc"

var docCommand = &ffcli.Command{
	Name:       "doc",
	ShortUsage: "sg doc <command>...",
	ShortHelp:  "Run the given doc manipulation",
	Subcommands: []*ffcli.Command{
		docMirrorCommand,
		docDotsCommand,
	},
}

var (
	docMirrorFlagset = flag.NewFlagSet("sg doc mirror", flag.ExitOnError)
	docMirrorDirFlag = docMirrorFlagset.String("dir", defaultDocDir, "Directory to traverse for .dot diagrams")
	docMirrorCommand = &ffcli.Command{
		Name:      "mirror",
		ShortHelp: "Update 'sg-doc-mirror' directives",
		FlagSet:   docMirrorFlagset,
		Exec: func(ctx context.Context, args []string) error {
			matchCmd := exec.CommandContext(ctx, "comby",
				"-f", ".md", "-d", *docMirrorDirFlag, "-match-only",
				"<!-- sg-doc-mirror begin --> :[mirror] <!-- sg-doc-mirror end -->", ":[mirror]")
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			matchCmd.Dir = cwd
			out, err := matchCmd.CombinedOutput()
			if err != nil {
				return err
			}
			matches := strings.Split(string(out), "\n")
			for _, match := range matches {
				if match == "" {
					continue
				}
				// doc/admin/install/kubernetes/overlays.md:274:<...>
				parts := strings.SplitN(match, ":", 3)
				path := parts[0]
				mirrorBlock := strings.ReplaceAll(parts[2], `\n`, "\n")
				if err := updateMirrorBlock(path, mirrorBlock); err != nil {
					return err
				}
			}

			return nil
		},
	}

	docDotsFlagset = flag.NewFlagSet("sg doc dots", flag.ExitOnError)
	docDotsDirFlag = docDotsFlagset.String("dir", defaultDocDir, "Directory to traverse for .dot diagrams")
	docDotsCommand = &ffcli.Command{
		Name:      "dots",
		ShortHelp: "Render SVGs from .dot files",
		FlagSet:   docDotsFlagset,
		Exec: func(ctx context.Context, args []string) error {
			return filepath.WalkDir(*docDotsDirFlag, func(path string, d fs.DirEntry, err error) error {
				if filepath.Ext(path) == ".dot" {
					base := filepath.Base(path)
					// dot architecture.dot -Tsvg >architecture.svg
					svgPath := strings.TrimSuffix(base, ".dot") + ".svg"
					dotCmd := exec.CommandContext(ctx, "dot",
						base,
						"-Tsvg",
						"-o"+svgPath)
					cwd, err := os.Getwd()
					if err != nil {
						return err
					}
					dotCmd.Dir = filepath.Join(cwd, filepath.Dir(path))
					dotCmd.Stderr = os.Stderr
					return dotCmd.Run()
				}
				return nil
			})
		},
	}
)

/**
<!-- sg-doc-mirror begin -->
<!-- sg-doc-mirror header-level=3 https://github.com/sourcegraph/deploy-sourcegraph/tree/master/overlays -->

hello world

<!-- sg-doc-mirror end -->
**/
var headerLevelMatch = regexp.MustCompile("(header-level=\\d)")
var urlMatch = regexp.MustCompile("https:\\/\\/github\\.com.* ")

func updateMirrorBlock(filepath, mirrorBlock string) error {
	headerLevelArg := headerLevelMatch.FindString(mirrorBlock)
	headerLevel, err := strconv.Atoi(strings.Split(headerLevelArg, "=")[1])
	if err != nil {
		return err
	}

	repoDirArg := strings.TrimSpace(urlMatch.FindString(mirrorBlock))

	fmt.Printf("%s %d %s\n", filepath, headerLevel, repoDirArg)

	repoParts := strings.Split(repoDirArg, "/tree/")
	repo := path.Base(repoParts[0])
	repoDir := strings.SplitN(repoParts[1], "/", 2)[1]

	fmt.Printf("%s %s\n", repo, repoDir)

	return nil
}
