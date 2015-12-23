package worker

import (
	droneparser "github.com/drone/drone-exec/parser"
	dronerunner "github.com/drone/drone-exec/runner"
)

// Drone CI-related customizations and hacks.

func init() {
	dronerunner.DefaultCloner = "sourcegraph/drone-git"
	droneparser.DefaultCloner = "sourcegraph/drone-git"
}
