package main

import (
	"fmt"
	"os"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

func main() {
	env.Lock()
	err := cli.Main()
	if err != nil {
		fmt.Fprintln(os.Stderr, "fatal:", err)
		os.Exit(1)
	}
}
