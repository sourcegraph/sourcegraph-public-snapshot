package graph

import (
	"os/exec"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/depgraph/root"
)

const RootPackage = "github.com/sourcegraph/sourcegraph"

// trimPackage remvoes leading RootPackage from the given value.
func trimPackage(pkg string) string {
	return strings.TrimPrefix(strings.TrimPrefix(pkg, RootPackage), "/")
}

// runGo invokes a go command on the host.
func runGo(commands ...string) (string, error) {
	root, err := root.RepositoryRoot()
	if err != nil {
		return "", err
	}

	cmd := exec.Command("go", commands...)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}
