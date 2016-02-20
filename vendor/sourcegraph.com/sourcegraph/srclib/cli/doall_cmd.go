package cli

import (
	"log"
	"os"

	"sourcegraph.com/sourcegraph/go-flags"
)

func init() {
	cliInit = append(cliInit, func(cli *flags.Command) {
		// TODO(sqs): "do-all" is a stupid name
		_, err := cli.AddCommand("do-all",
			"fully process (config, plan, execute, and import)",
			`Fully processes a tree: configures it, plans the execution, executes all analysis steps, and imports the data.`,
			&doAllCmd,
		)
		if err != nil {
			log.Fatal(err)
		}
	})
}

type DoAllCmd struct {
	Dir Directory `short:"C" long:"directory" description:"change to DIR before doing anything" value-name:"DIR"`
}

var doAllCmd DoAllCmd

func (c *DoAllCmd) Execute(args []string) error {
	if c.Dir != "" {
		if err := os.Chdir(c.Dir.String()); err != nil {
			return err
		}
	}

	// config
	configCmd := &ConfigCmd{}
	if err := configCmd.Execute(nil); err != nil {
		return err
	}

	// make
	makeCmd := &MakeCmd{}
	if err := makeCmd.Execute(nil); err != nil {
		return err
	}

	return nil
}
