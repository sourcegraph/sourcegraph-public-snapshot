package cli

import (
	"log"

	"github.com/alexsaveliev/go-colorable-wrapper"
	"sourcegraph.com/sourcegraph/go-flags"
)

func init() {
	cliInit = append(cliInit, func(cli *flags.Command) {
		_, err := cli.AddCommand("repo",
			"display current repo info",
			"The repo subcommand displays autodetected info about the current repo.",
			&repoCmd{},
		)
		if err != nil {
			log.Fatal(err)
		}
	})
}

type repoCmd struct{}

func (c *repoCmd) Execute(args []string) error {
	repo, err := OpenRepo(".")
	if err != nil {
		return err
	}
	colorable.Println("# Current repository:")
	colorable.Println("VCS:", repo.VCSType)
	colorable.Println("Root dir:", repo.RootDir)
	colorable.Println("Commit ID:", repo.CommitID)
	return nil
}
