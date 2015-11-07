package main

import (
	"log"
	"os/exec"
)

func init() {
	_, err := CLI.AddCommand("deploy",
		"deploy to src.sourcegraph.com",
		"The deploy command deploys a Sourcegraph version to src.sourcegraph.com.",
		&deployCmd,
	)
	if err != nil {
		log.Fatal(err)
	}
}

type DeployCmd struct {
	Args struct {
		Version string `name:"version" description:"version number ('1.2.3') or identifier ('snapshot' is default)"`
	} `positional-args:"yes"`
}

var deployCmd DeployCmd

func (c *DeployCmd) Execute(args []string) error {
	// Check for dependencies before starting.
	if err := requireCmds("make"); err != nil {
		return err
	}

	if c.Args.Version == "" {
		c.Args.Version = "snapshot"
	}

	if err := execCmd(exec.Command("make", "deploy-dev", "V="+c.Args.Version)); err != nil {
		return err
	}

	return nil
}
