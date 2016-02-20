package cli

import (
	"encoding/json"
	"log"
	"os"

	"sourcegraph.com/sourcegraph/go-flags"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/grapher"
)

func init() {
	cliInit = append(cliInit, func(cli *flags.Command) {
		c, err := cli.AddCommand("internal", "(internal subcommands - do not use)", "Internal subcommands. Do not use.", &struct{}{})
		if err != nil {
			log.Fatal(err)
		}

		_, err = c.AddCommand("normalize-graph-data", "", "", &normalizeGraphDataCmd)
		if err != nil {
			log.Fatal(err)
		}
	})
}

type NormalizeGraphDataCmd struct {
	UnitType string `long:"unit-type" description:"source unit type (e.g., GoPackage)"`
	Dir      string `long:"dir" description:"directory of source unit (SourceUnit.Dir field)"`
}

var normalizeGraphDataCmd NormalizeGraphDataCmd

func (c *NormalizeGraphDataCmd) Execute(args []string) error {
	in := os.Stdin

	var o *graph.Output
	if err := json.NewDecoder(in).Decode(&o); err != nil {
		return err
	}

	if err := grapher.NormalizeData(c.UnitType, c.Dir, o); err != nil {
		return err
	}

	data, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err
	}

	if _, err := os.Stdout.Write(data); err != nil {
		return err
	}

	return nil
}
