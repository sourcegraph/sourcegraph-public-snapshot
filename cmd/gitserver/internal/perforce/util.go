package perforce

import (
	"fmt"
	"os/exec"
	"strings"
)

func specifyCommandInErrorMessage(errorMsg string, command *exec.Cmd) string {
	if !strings.Contains(errorMsg, "this operation") {
		return errorMsg
	}
	if len(command.Args) == 0 {
		return errorMsg
	}
	return strings.Replace(errorMsg, "this operation", fmt.Sprintf("`%s`", strings.Join(command.Args, " ")), 1)
}
