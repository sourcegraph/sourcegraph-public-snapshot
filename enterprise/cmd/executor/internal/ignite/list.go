package ignite

import (
	"context"
	"os/exec"
	"strings"
)

// CurrentlyRunningVMs returns the set of VMs existant on the host as a map from VM names
// to VM identifiers. VMs starting with a prefix distinct from the given prefix are ignored.
func CurrentlyRunningVMs(ctx context.Context, prefix string) (map[string]string, error) {
	cmd := exec.CommandContext(ctx, "ignite", "ps", "-a", "-t", "{{ .Name }}:{{ .UID }}")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return parseIgniteList(prefix, string(out)), nil
}

// parseIgniteList parses the output from the `ignite ps` invocation in CurrentlyRunningVMs.
// VMs starting with a prefix distinct from the given prefix are ignored.
func parseIgniteList(prefix, out string) map[string]string {
	activeVMsMap := map[string]string{}
	for _, line := range strings.Split(out, "\n") {
		if parts := strings.Split(line, ":"); len(parts) == 2 && strings.HasPrefix(parts[0], prefix) {
			activeVMsMap[parts[0]] = parts[1]
		}
	}

	return activeVMsMap
}
