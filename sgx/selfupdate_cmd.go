package sgx

import (
	"log"

	"src.sourcegraph.com/sourcegraph/sgx/cli"

	"src.sourcegraph.com/sourcegraph/sgx/sgxcmd"
)

func init() {
	_, err := cli.CLI.AddCommand("selfupdate",
		"update this program",
		"The selfupdate command updates the "+sgxcmd.Name+" binary.",
		&selfUpdateCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type selfUpdateCmd struct {
	CheckOnly bool `short:"n" long:"check-only" description:"check for update but do not download and install it"`
}

func (c *selfUpdateCmd) Execute(_ []string) error {
	u := sgxcmd.SelfUpdater

	if c.CheckOnly {
		if err := u.Check(); err != nil {
			return err
		}
	} else {
		if err := u.Update(); err != nil {
			return err
		}
	}

	if u.CurrentVersion == u.Info.Version {
		log.Printf("# No updates available (current version is %s)", u.CurrentVersion)
	} else {
		var title string
		if c.CheckOnly {
			title = "An update is available"
		} else {
			title = "Updated"
		}
		log.Printf("# %s: %s -> %s", title, u.CurrentVersion, u.Info.Version)
	}

	return nil
}
