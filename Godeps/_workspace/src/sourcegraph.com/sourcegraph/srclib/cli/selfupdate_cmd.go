package cli

import (
	"log"

	"sourcegraph.com/sourcegraph/srclib"

	"github.com/sqs/go-selfupdate/selfupdate"
)

func init() {
	_, err := CLI.AddCommand("selfupdate",
		"update this program",
		"The selfupdate command updates the srclib binary.",
		&selfUpdateCmd,
	)
	if err != nil {
		log.Fatal(err)
	}
}

var selfUpdateCmd SelfUpdateCmd

type SelfUpdateCmd struct {
	CheckOnly bool `short:"n" long:"check-only" description:"check for update but do not download and install it"`
}

func (c *SelfUpdateCmd) Execute(_ []string) error {
	url := "https://srclib-release.s3.amazonaws.com/"
	var u = &selfupdate.Updater{
		CurrentVersion: Version,
		ApiURL:         url,
		BinURL:         url,
		DiffURL:        url,
		Dir:            srclib.CommandName,
		CmdName:        srclib.CommandName,
	}

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
