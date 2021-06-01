package main

import (
	"log"
	"os"
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/worker/shared"
)

var setupHooks = map[string]shared.SetupHook{
	// Empty for now
}

func main() {
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	if debug {
		log.Println("enterprise edition")
	}

	shared.Main(setupHooks)
}
