package main

import (
	"github.com/sourcegraph/sourcegraph/cmd/github-proxy/shared"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

func main() {
	env.Lock()
	env.HandleHelpFlag()
	shared.Main()
}
