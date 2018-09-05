//docker:user sourcegraph

// Note: All frontend code should be added to shared.Main, not here. See that
// function for details.

package main

import "github.com/sourcegraph/sourcegraph/cmd/frontend/shared"

func main() {
	shared.Main()
}
