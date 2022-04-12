package main

import (
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v41/github"
	"github.com/honeycombio/libhoney-go"

	"github.com/sourcegraph/sourcegraph/lib/errors"
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

type DeploymentTrace struct {
	Root *libhoney.Event
	ID   string

	Spans []*libhoney.Event
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
// Learn more about Honeycomb fields:
//
// - https://docs.honeycomb.io/working-with-your-data/home/#configuring-home
// - https://docs.honeycomb.io/getting-data-in/tracing/send-trace-data/#span-annotations
func GenerateDeploymentTrace(r *DeploymentReport) (*DeploymentTrace, error) {
	libhoney.UserAgentAddition = fmt.Sprintf("deployment-notifier/%s", traceVersion)

	rev := r.ManifestRevision
	if len(rev) > 12 {
		rev = rev[:12]
	}
	deploymentTraceID := newSpanID("deploy", r.Environment, rev)

	deployTime, err := time.Parse(time.RFC822Z, r.DeployedAt)
	if err != nil {
		return nil, errors.Wrap(err, "r.DeployedAt")
	}
	oldestPR := time.Now()

	prSet := map[int]*github.PullRequest{}
	for _, pr := range r.PullRequests {
		prSet[pr.GetNumber()] = pr
	}

	var spans []*libhoney.Event
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
				// Honeycomb fields
				"name":            service,
				"service.name":    "service",
				"trace.parent_id": prTraceID,
				"trace.span_id":   newSpanID("svc", strconv.Itoa(pr.GetNumber()), service),
				"duration_ms":     deployTime.Sub(pr.GetMergedAt()) / time.Millisecond,
				"user":            pr.GetUser().GetLogin(),

				// Extra metadata
				"service":               service,
				"pull_request.number":   pr.GetNumber(),
				"pull_request.revision": pr.GetMergeCommitSHA(),
			})
			spans = append(spans, prServiceEvent)
		}

		prEvent := newTraceEvent(deploymentTraceID)
		prEvent.Timestamp = pr.GetMergedAt()
		prEvent.Add(map[string]interface{}{
			// Honeycomb fields
			"name":            pr.GetNumber(),
			"service.name":    "pull_request",
			"trace.parent_id": deploymentTraceID,
			"trace.span_id":   prTraceID,
			"user":            pr.GetUser().GetLogin(),
			// Don't include a duration - PR might have other services not deployed yet

			// Extra metadata
			"pull_request.number":   pr.GetNumber(),
			"pull_request.title":    pr.GetTitle(),
			"pull_request.revision": pr.GetMergeCommitSHA(),
			"pull_request.url":      pr.GetHTMLURL(),
		})
		spans = append(spans, prEvent)
	}

	root := newTraceEvent(deploymentTraceID)
	root.Timestamp = oldestPR
	root.Add(map[string]interface{}{
		// Honeycomb fields
		"name":          fmt.Sprintf("%s (%s)", r.Environment, r.DeployedAt),
		"service.name":  fmt.Sprintf("deploy/%s", r.Environment),
		"trace.span_id": deploymentTraceID,
		"duration_ms":   deployTime.Sub(oldestPR) / time.Millisecond,

		// Extra metadata
		"environment":            r.Environment,
		"buildkite.build_url":    r.BuildkiteBuildURL,
		"manifest.revision":      r.ManifestRevision,
		"deployed.at":            r.DeployedAt,
		"deployed.pull_requests": len(r.PullRequests),
		"deployed.services":      len(r.Services),
	})

	return &DeploymentTrace{
		ID:    deploymentTraceID,
		Root:  root,
		Spans: spans,
	}, nil
}

// https://sourcegraph.com/search?q=context:global+repo:%5Egithub%5C.com/honeycombio/buildevents%24%40main+func+buildURL&patternType=literal
func buildTraceURL(cfg *libhoney.Config, traceID string, ts int64) (string, error) {
	teamName, err := libhoney.VerifyAPIKey(*cfg)
	if err != nil {
		return "", errors.Newf("unable to verify API key: %w", err)
	}
	uiHost := strings.Replace(cfg.APIHost, "api", "ui", 1)
	u, err := url.Parse(uiHost)
	if err != nil {
		return "", errors.Newf("unable to infer UI host: %s", uiHost)
	}
	u.Path = path.Join(teamName, "datasets", strings.ReplaceAll(cfg.Dataset, " ", "-"), "trace")
	endTime := time.Now().Add(10 * time.Minute).Unix()
	return fmt.Sprintf(
		"%s?trace_id=%s&trace_start_ts=%d&trace_end_ts=%d",
		u.String(), traceID, ts, endTime,
	), nil
}
