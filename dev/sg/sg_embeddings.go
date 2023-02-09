package main

import (
	"bufio"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/grafana/regexp"
	"github.com/urfave/cli/v2"
)

// sg_embeddings.go parses our doc files and sends them to the
// OpenAI embeddings API. It then stores the results into a vector db

var embeddings = &cli.Command{
	Name:     "embeddings",
	Usage:    "Generate embeddings for our docs",
	Category: CategoryUtil,
	Subcommands: []*cli.Command{
		{
			Name:  "generate",
			Usage: "Generate embeddings for our docs",
			Action: func(context *cli.Context) error {
				err := parseDocs("/Users/dax/work/sourcegraph/doc")
				return err
			},
		},
	},
}

func parseDocs(docDir string) error {
	return VistDocPages(docDir)
}

var (
	markdownFileNameRegExp = regexp.MustCompile(`.+\.md`)
	// https://regex101.com/r/oFN6vu/1
	markdownHeadersRegexp = regexp.MustCompile(`^(#{1,6}[\s]+.*)\n((?:.*\n)*)#{1,6}[\s]+.*`)
)

func VistDocPages(docDir string) error {
	return filepath.WalkDir(docDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return nil
		}
		// regex that matches markdown files
		filenameMatch := markdownFileNameRegExp.FindAllStringSubmatch(entry.Name(), 1)
		if filenameMatch == nil {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// TODO: We need to either write to file or something inside each function call
		s := bufio.NewScanner(file)
		for s.Scan() {
			// match each section demarcated by the markdown header
			sectionMatches := markdownHeadersRegexp.FindAllStringSubmatch(s.Text(), 0)

			if len(sectionMatches) == 0 {
				println("strange file: ", path)
				return nil
			}
			// send each to embeddings API and store the response in the vector db

		}
		return nil
	})
}

func parseHandbook() {

}
