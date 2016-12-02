package cli

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"sourcegraph.com/sourcegraph/go-flags"

	"sourcegraph.com/sourcegraph/srclib/graph"
	"sourcegraph.com/sourcegraph/srclib/grapher"
	"sourcegraph.com/sourcegraph/srclib/plan"
	"sourcegraph.com/sourcegraph/srclib/unit"
)

func init() {
	cliInit = append(cliInit, func(cli *flags.Command) {
		c, err := cli.AddCommand("internal", "(internal subcommands - do not use)", "Internal subcommands. Do not use.", &struct{}{})
		if err != nil {
			log.Fatal(err)
		}

		_, err = c.AddCommand("emit-unit-data", "", "", &emitUnitDataCmd)
		if err != nil {
			log.Fatal(err)
		}

		_, err = c.AddCommand("normalize-graph-data", "", "", &normalizeGraphDataCmd)
		if err != nil {
			log.Fatal(err)
		}
	})
}

type EmitUnitDataCmd struct {
	Args struct {
		Units []string `name:"units" description:"Paths to source units."`
	} `positional-args:"yes"`
}

var emitUnitDataCmd EmitUnitDataCmd

func (c *EmitUnitDataCmd) Execute(args []string) error {
	var units unit.SourceUnits

	for _, path := range c.Args.Units {
		unitFile, err := os.Open(path)
		if err != nil {
			return err
		}
		var u *unit.SourceUnit
		if err := json.NewDecoder(unitFile).Decode(&u); err != nil {
			return err
		}
		units = append(units, u)
	}

	if err := json.NewEncoder(os.Stdout).Encode(units); err != nil {
		return err
	}

	return nil
}

type NormalizeGraphDataCmd struct {
	UnitType string `long:"unit-type" description:"source unit type (e.g., GoPackage)"`
	Dir      string `long:"dir" description:"directory of source unit (SourceUnit.Dir field)"`
	Multi    bool   `long:"multi" description:"the input contains graph data for multiple units; output will be split into different files per source unit"`
	DataDir  string `long:"data-dir" description:"output data dir"`
}

var normalizeGraphDataCmd NormalizeGraphDataCmd

func (c *NormalizeGraphDataCmd) Execute(args []string) error {
	in := os.Stdin

	var o *graph.Output
	if err := json.NewDecoder(in).Decode(&o); err != nil {
		return err
	}

	if !c.Multi {
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

	// If `graph` emits multiple source units, in this case, don't
	// write to stdout (the rule will not direct it to a file), but
	// instead write to multiple .graph.json files (one for each
	// source unit). This is a HACK.

	// graphPerUnit maps source unit names to the graph data of
	// that unit.
	graphPerUnit := make(map[string]*graph.Output)
	initUnitGraph := func(unitName string) {
		if _, ok := graphPerUnit[unitName]; !ok {
			graphPerUnit[unitName] = &graph.Output{}
		}
	}

	// Split the graph data per source unit.
	for _, d := range o.Defs {
		if d.Unit == "" {
			log.Printf("skip def with empty unit: %v", d)
			continue
		}
		initUnitGraph(d.Unit)
		graphPerUnit[d.Unit].Defs = append(graphPerUnit[d.Unit].Defs, d)
	}
	for _, r := range o.Refs {
		if r.Unit == "" {
			log.Printf("skip ref with empty unit: %v", r)
			continue
		}
		initUnitGraph(r.Unit)
		graphPerUnit[r.Unit].Refs = append(graphPerUnit[r.Unit].Refs, r)
	}
	for _, d := range o.Docs {
		if d.DocUnit == "" {
			log.Printf("skip doc with empty unit: %v", d)
			continue
		}
		initUnitGraph(d.DocUnit)
		graphPerUnit[d.DocUnit].Docs = append(graphPerUnit[d.DocUnit].Docs, d)
	}
	for _, a := range o.Anns {
		if a.Unit == "" {
			log.Printf("skip ann with empty unit: %v", a)
			continue
		}
		initUnitGraph(a.Unit)
		graphPerUnit[a.Unit].Anns = append(graphPerUnit[a.Unit].Anns, a)
	}

	// Write the graph data to a separate file for each source unit.
	for unitName, graphData := range graphPerUnit {
		if err := grapher.NormalizeData(c.UnitType, c.Dir, graphData); err != nil {
			log.Printf("skipping unit %s because failed to normalize data: %s", unitName, err)
			continue
		}

		path := filepath.ToSlash(filepath.Join(c.DataDir, plan.SourceUnitDataFilename(&graph.Output{}, &unit.SourceUnit{Key: unit.Key{Name: unitName, Type: c.UnitType}})))
		graphFile, err := os.Create(path)
		if err != nil {
			return err
		}

		data, err := json.MarshalIndent(graphData, "", "  ")
		if err != nil {
			return err
		}

		if _, err := graphFile.Write(data); err != nil {
			return err
		}
	}

	return nil
}
