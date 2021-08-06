package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

var docMirrorredMatch = regexp.MustCompile("(?s)(<!-- sg-doc-mirror begin).*(<!-- sg-doc-mirror end -->)")

var (
	docMirrorFlagset = flag.NewFlagSet("sg doc mirror", flag.ExitOnError)
	docMirrorDirFlag = docMirrorFlagset.String("dir", defaultDocDir, "Directory to traverse for .dot diagrams")
	docMirrorCommand = &ffcli.Command{
		Name:      "mirror",
		ShortHelp: "Update 'sg-doc-mirror' directives",
		FlagSet:   docMirrorFlagset,
		Exec: func(ctx context.Context, args []string) error {
			return filepath.WalkDir(*docDotsDirFlag, func(path string, d fs.DirEntry, err error) error {
				if filepath.Ext(path) != ".md" {
					return nil
				}
				b, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}
				contents := string(b)
				if !strings.Contains(contents, "<!-- sg-doc-mirror begin") {
					return nil
				}
				captured := docMirrorredMatch.FindStringSubmatch(contents)
				fmt.Printf("%v\n", captured)

				return nil
			})
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
					dotCmd := exec.Command("dot",
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
