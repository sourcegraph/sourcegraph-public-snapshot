package cli

import (
	"fmt"
	"log"
)

func init() {
	_, err := CLI.AddCommand("repo",
		"display current repo info",
		"The repo subcommand displays autodetected info about the current repo.",
		&repoCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type repoCmd struct{}

func (c *repoCmd) Execute(args []string) error {
	repo, err := OpenRepo(".")
	if err != nil {
		return err
	}
	fmt.Println("# Current repository:")
	fmt.Println("URI:", repo.URI())
	fmt.Println("Clone URL:", repo.CloneURL)
	fmt.Println("VCS:", repo.VCSType)
	fmt.Println("Root dir:", repo.RootDir)
	fmt.Println("Commit ID:", repo.CommitID)
	return nil
}
