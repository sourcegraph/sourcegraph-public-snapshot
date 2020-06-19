// Command replacer is an interface to replace and rewrite code. It passes a zipped repo
// to external tools and streams back JSON lines results.
package main

import (
	"github.com/sourcegraph/sourcegraph/cmd/replacer/shared"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
)

func main() {
	go debugserver.Start()

	shared.Main()
}
