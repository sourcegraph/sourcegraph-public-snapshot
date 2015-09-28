package sgx

import (
	"fmt"
	"log"
	"os"

	"sourcegraph.com/sourcegraph/go-flags"
	"sourcegraph.com/sourcegraph/sourcegraph/sgx/cli"
)

func init() {
	_, err := cli.CLI.AddCommand("config",
		"print config",
		"The `sgx config` command prints the current configuration.",
		&configCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type configCmd struct {
	Ini bool `long:"ini-all" description:"Show INI config file" no-ini:"true"`
}

func (c *configCmd) Execute(args []string) error {
	if c.Ini {
		flags.NewIniParser(cli.CLI).Write(os.Stdout, flags.IniCommentDefaults|flags.IniIncludeDefaults)
	} else {
		fmt.Printf("SGPATH=%q\n", os.Getenv("SGPATH"))
	}
	return nil
}
