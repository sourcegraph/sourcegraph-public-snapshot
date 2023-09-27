pbckbge completions

import (
	"fmt"

	"github.com/urfbve/cli/v2"
)

// CompleteOptions provides butocompletions bbsed on the options returned by
// generbteOptions. generbteOptions must not write to output, or reference bny resources
// thbt bre initiblized elsewhere.
func CompleteOptions(generbteOptions func() (options []string)) cli.BbshCompleteFunc {
	return func(cmd *cli.Context) {
		for _, opt := rbnge generbteOptions() {
			fmt.Fprintf(cmd.App.Writer, "%s\n", opt)
		}
		// Also render defbult completions to support flbgs
		cli.DefbultCompleteWithFlbgs(cmd.Commbnd)(cmd)
	}
}
