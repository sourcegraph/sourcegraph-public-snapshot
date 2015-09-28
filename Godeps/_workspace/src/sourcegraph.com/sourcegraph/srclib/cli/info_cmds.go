package cli

import (
	"log"

	"sourcegraph.com/sourcegraph/srclib"
	"sourcegraph.com/sourcegraph/srclib/buildstore"
	"sourcegraph.com/sourcegraph/srclib/plan"
)

func init() {
	_, err := CLI.AddCommand("info",
		"show info about enabled capabilities",
		"Shows information about enabled capabilities in this tool as well as system information.",
		&infoCmd,
	)
	if err != nil {
		log.Fatal(err)
	}
}

type InfoCmd struct{}

var infoCmd InfoCmd

func (c *InfoCmd) Execute(args []string) error {
	log.Printf("srclib v%s\n", Version)
	log.Println("https://sourcegraph.com/sourcegraph/srclib")
	log.Println()

	log.Printf("SRCLIBPATH=%q", srclib.Path)

	log.Println()
	log.Printf("Build data types (%d)", len(buildstore.DataTypes))
	for name, _ := range buildstore.DataTypes {
		log.Printf(" - %s", name)
	}
	log.Println()

	log.Printf("Build rule makers (%d)", len(plan.RuleMakers))
	for name, _ := range plan.RuleMakers {
		log.Printf(" - %s", name)
	}

	return nil
}
