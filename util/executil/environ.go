package executil

import (
	"os"
	"os/exec"
	"strings"
)

// OverrideEnv copies all of the current environment variables to cmd,
// except for the named variable. If present, it overwrites its value
// with the provided value; otherwise it sets to the provided value.
func OverrideEnv(cmd *exec.Cmd, name, value string) {
	for _, s := range os.Environ() {
		if !strings.HasPrefix(s, name+"=") {
			cmd.Env = append(cmd.Env, s)
		}
	}
	cmd.Env = append(cmd.Env, name+"="+value)
}
