package ignite

import (
	"context"
	"os/exec"
	"strings"
)

// ActiveVMsByName returns the set of VMs existant on the host as a map from VM names
// to VM identifiers. VMs starting with a prefix distinct from the given prefix are
// ignored.
func ActiveVMsByName(ctx context.Context, prefix string, all bool) (map[string]string, error) {
	args := []string{"ps", "-t", "{{ .Name }}:{{ .UID }}"}
	if all {
		args = append(args, "-a")
	}

	cmd := exec.CommandContext(ctx, "ignite", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	return parseIgniteList(prefix, string(out)), nil
}

// parseIgniteList parses the output from the `ignite ps` invocation in ActiveVMsByName.
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
