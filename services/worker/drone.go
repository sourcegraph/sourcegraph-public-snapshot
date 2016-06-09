package worker

import (
	droneparser "github.com/drone/drone-exec/parser"
	dronerunner "github.com/drone/drone-exec/runner"
)

// Drone CI-related customizations and hacks.

func init() {
	// This image adds multiple netrc support and the "complete" param.
	//
	// We can switch back to the upstream plugin/drone-git when these
	// PRs are merged:
	// https://github.com/drone-plugins/drone-git/pull/14 and the
	// not-yet-submitted PR based on the github.com/sqs/drone-git
	// multiple-netrc-entries branch.
	dronerunner.DefaultCloner = "sourcegraph/drone-git@sha256:09ddfc452b92657fcef5da70a71687f149ef2c4f0b4501fe3f4e0b9c1e91821c"
	droneparser.DefaultCloner = dronerunner.DefaultCloner
}
