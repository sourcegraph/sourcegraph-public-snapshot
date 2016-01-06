package sgx

import (
	"errors"
	"fmt"
	"log"
	"time"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/pkg/sysreq"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
)

func init() {
	_, err := cli.CLI.AddCommand("info",
		"show info",
		"The info subcommand displays system information.",
		&infoCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type infoCmd struct{}

func (c *infoCmd) Execute(_ []string) error {
	fmt.Println("# System requirements")

	ctx, cancel := context.WithTimeout(cli.Ctx, 5*time.Second)
	defer cancel()
	hasErr := false
	for _, st := range sysreq.Check(ctx, skippedSysReqs()) {
		fmt.Printf("%s: ", st.Name)
		if st.OK() {
			fmt.Println("OK")
			continue
		}
		if st.Skipped {
			fmt.Println("Skipped")
			continue
		}
		if st.Problem != "" {
			fmt.Println(st.Problem)
		}
		if st.Err != nil {
			if st.Problem != "" {
				fmt.Print("\t")
			}
			fmt.Printf("Error: %s\n", st.Err)
			hasErr = true
		}
		if st.Fix != "" {
			fmt.Printf("\tPossible fix: %s\n", st.Fix)
		}
	}

	if hasErr {
		return errors.New("system requirement checks failed (see above)")
	}
	return nil
}
