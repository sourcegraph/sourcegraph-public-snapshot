package cli

import (
	"log"

	"sourcegraph.com/sourcegraph/go-flags"

	"sourcegraph.com/sourcegraph/srclib/gendata"
)

func init() {
	cliInit = append(cliInit, func(cli *flags.Command) {
		c, err := cli.AddCommand("gen-data",
			"generates fake data",
			`generates fake data for testing and benchmarking purposes. Run this command inside an empty or expendable directory.`,
			&gendata.GenDataCmd{},
		)
		if err != nil {
			log.Fatal(err)
		}
		c.Aliases = []string{"c"}

		_, err = c.AddCommand("simple",
			"generates a simple repository",
			"generates a simple repository with the specified source unit and file structure, with the given number of defs and refs to those defs in each file",
			&gendata.SimpleRepoCmd{},
		)
		if err != nil {
			log.Fatal(err)
		}

		_, err = c.AddCommand("urefs",
			"generates a repository with cross-unit references",
			"generates a repository with cross-unit references with the given unit and file structure, the number of defs in each file, and refs from each source unit to each def",
			&gendata.URefsRepoCmd{},
		)
		if err != nil {
			log.Fatal(err)
		}
	})
}
