package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/google/go-github/v41/github"
	"github.com/honeycombio/libhoney-go"
)

const traceVersion = "dev"

func newTraceEvent(traceID string) *libhoney.Event {
	event := libhoney.NewEvent()
	event.AddField("meta.version", traceVersion)
	event.AddField("trace.trace_id", traceID)
	return event
}

func newSpanID(root string, components ...string) string {
	for _, c := range components {
		root += fmt.Sprintf("/%s", c)
	}
	return root
}

// GenerateDeploymentTrace generates a set of events that trace PRs from merge to deploy.
//
// The generated trace is structured as follows:
//
//   deploy/env/rev -----
//     pr/1 -------------
//     ------------ svc/1
//     ------------ svc/2
//         pr/2 ---------
//         -------- svc/1
//         -------- svc/2
//                    ...
//
func GenerateDeploymentTrace(r *DeploymentReport) (events []*libhoney.Event, err error) {
	libhoney.UserAgentAddition = fmt.Sprintf("deployment-notifier/%s", traceVersion)

	rev := r.ManifestRevision
	if len(rev) > 12 {
		rev = rev[:12]
	}
	deploymentTraceID := newSpanID("deploy", r.Environment, rev)

	deployTime, err := time.Parse(time.RFC822Z, r.DeployedAt)
	if err != nil {
		return nil, err
	}
	oldestPR := time.Now()

	prSet := map[int]*github.PullRequest{}
	for _, pr := range r.PullRequests {
		prSet[pr.GetNumber()] = pr
	}

	for prNumber, prServices := range r.ServicesPerPullRequest {
		pr := prSet[prNumber]
		if pr.GetMergedAt().Before(oldestPR) {
			oldestPR = pr.GetMergedAt()
		}

		prTraceID := newSpanID("pr", strconv.Itoa(pr.GetNumber()))

		for _, service := range prServices {
			prServiceEvent := newTraceEvent(deploymentTraceID)
			prServiceEvent.Timestamp = pr.GetMergedAt()
			prServiceEvent.Add(map[string]interface{}{
				"trace.parent_id": prTraceID,
				"trace.span_id":   newSpanID("svc", strconv.Itoa(pr.GetNumber()), service),

				"service": service,

				"duration_ms": deployTime.Sub(pr.GetMergedAt()) / time.Millisecond,
			})
			events = append(events, prServiceEvent)
		}

		prEvent := newTraceEvent(deploymentTraceID)
		prEvent.Timestamp = pr.GetMergedAt()
		prEvent.Add(map[string]interface{}{
			"trace.parent_id": deploymentTraceID,
			"trace.span_id":   prTraceID,

			"pull_request.number":   pr.GetNumber(),
			"pull_request.title":    pr.GetTitle(),
			"pull_request.revision": pr.GetMergeCommitSHA(),
			"pull_request.url":      pr.GetHTMLURL(),

			// Don't include duration - PR might have other services not deployed yet
			// TODO does this work?
		})
		events = append(events, prEvent)
	}

	deployEvent := newTraceEvent(deploymentTraceID)
	deployEvent.Timestamp = oldestPR
	deployEvent.Add(map[string]interface{}{
		"trace.span_id": deploymentTraceID,
		"name":          r.DeployedAt,

		"environment":         r.Environment,
		"buildkite.build_url": r.BuildkiteBuildURL,
		"manifest.revision":   r.ManifestRevision,

		"deployed.at":            r.DeployedAt,
		"deployed.pull_requests": len(r.PullRequests),
		"deployed.services":      len(r.Services),

		"duration_ms": deployTime.Sub(oldestPR) / time.Millisecond,
	})
	events = append(events, deployEvent)

	return
}
