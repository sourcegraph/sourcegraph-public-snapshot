package sgx

import (
	"encoding/json"
	"fmt"
	"log"

	"src.sourcegraph.com/sourcegraph/sgx/buildvar"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
)

func init() {
	_, err := cli.CLI.AddCommand("version",
		"show version",
		"The version subcommand displays the current version of this sgx program (and other information describing the environment in which the binary was built).",
		&versionCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type versionCmd struct {
	Quiet bool `short:"q" description:"only print version string"`
}

func (c *versionCmd) Execute(_ []string) error {
	fmt.Println(buildvar.Version)
	if !c.Quiet {
		log.Println()
		log.Println("# Build information")
		b, err := json.MarshalIndent(buildvar.All, "", "  ")
		if err != nil {
			return err
		}
		log.Println(string(b))
	}
	return nil
}
