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
	dronerunner.DefaultCloner = "sourcegraph/drone-git@sha256:226187a2a6afddae968eb944a53006feed9de807d884e26f0bba3f4fa1f8cfad"
	droneparser.DefaultCloner = dronerunner.DefaultCloner
}
