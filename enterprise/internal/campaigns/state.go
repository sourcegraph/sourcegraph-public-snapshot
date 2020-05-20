package campaigns

import (
	"sort"
	"strings"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	cmpgn "github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

// SetDerivedState will update the external state fields on the Changeset based
// on the current  state of the changeset and associated events.
func SetDerivedState(c *cmpgn.Changeset, es []*cmpgn.ChangesetEvent) {
	// Copy so that we can sort without mutating the argument
	events := make(ChangesetEvents, len(es))
	copy(events, es)
	sort.Sort(events)

	c.ExternalCheckState = ComputeCheckState(c, events)

	history, err := computeHistory(c, events)
	if err != nil {
		log15.Warn("Computing changeset history", "err", err)
		return
	}

	if state, err := ComputeChangesetState(c, history); err != nil {
		log15.Warn("Computing changeset state", "err", err)
	} else {
		c.ExternalState = state
	}
	if state, err := ComputeReviewState(c, history); err != nil {
		log15.Warn("Computing changeset review state", "err", err)
	} else {
		c.ExternalReviewState = state
	}
}

// ComputeCheckState computes the overall check state based on the current synced check state
// and any webhook events that have arrived after the most recent sync
func ComputeCheckState(c *cmpgn.Changeset, events ChangesetEvents) cmpgn.ChangesetCheckState {
	switch m := c.Metadata.(type) {
	case *github.PullRequest:
		return computeGitHubCheckState(c.UpdatedAt, m, events)

	case *bitbucketserver.PullRequest:
		return computeBitbucketBuildStatus(c.UpdatedAt, m, events)
	}

	return cmpgn.ChangesetCheckStateUnknown
}

// ComputeChangesetState computes the overall state for the changeset and its
// associated events. The events should be presorted.
func ComputeChangesetState(c *cmpgn.Changeset, history []changesetStatesAtTime) (cmpgn.ChangesetState, error) {
	if len(history) == 0 {
		return computeSingleChangesetState(c)
	}
	newestDataPoint := history[len(history)-1]
	if c.UpdatedAt.After(newestDataPoint.t) {
		return computeSingleChangesetState(c)
	}
	return newestDataPoint.state, nil
}

// ComputeReviewState computes the review state for the changeset and its
// associated events. The events should be presorted.
func ComputeReviewState(c *cmpgn.Changeset, history []changesetStatesAtTime) (cmpgn.ChangesetReviewState, error) {
	if len(history) == 0 {
		return computeSingleChangesetReviewState(c)
	}

	newestDataPoint := history[len(history)-1]

	// GitHub only stores the ReviewState in events, we can't look at the
	// Changeset.
	if c.ExternalServiceType == github.ServiceType {
		return newestDataPoint.reviewState, nil
	}

	// For other codehosts we check whether the Changeset is newer or the
	// events and use the newest entity to get the reviewstate.
	if c.UpdatedAt.After(newestDataPoint.t) {
		return computeSingleChangesetReviewState(c)
	}
	return newestDataPoint.reviewState, nil
}

func computeBitbucketBuildStatus(lastSynced time.Time, pr *bitbucketserver.PullRequest, events []*cmpgn.ChangesetEvent) cmpgn.ChangesetCheckState {
	var latestCommit bitbucketserver.Commit
	for _, c := range pr.Commits {
		if latestCommit.CommitterTimestamp <= c.CommitterTimestamp {
			latestCommit = *c
		}
	}

	stateMap := make(map[string]cmpgn.ChangesetCheckState)

	// States from last sync
	for _, status := range pr.CommitStatus {
		stateMap[status.Key()] = parseBitbucketBuildState(status.Status.State)
	}

	// Add any events we've received since our last sync
	for _, e := range events {
		switch m := e.Metadata.(type) {
		case *bitbucketserver.CommitStatus:
			if m.Commit != latestCommit.ID {
				continue
			}
			dateAdded := unixMilliToTime(m.Status.DateAdded)
			if dateAdded.Before(lastSynced) {
				continue
			}
			stateMap[m.Key()] = parseBitbucketBuildState(m.Status.State)
		}
	}

	states := make([]cmpgn.ChangesetCheckState, 0, len(stateMap))
	for _, v := range stateMap {
		states = append(states, v)
	}

	return combineCheckStates(states)
}

func parseBitbucketBuildState(s string) cmpgn.ChangesetCheckState {
	switch s {
	case "FAILED":
		return cmpgn.ChangesetCheckStateFailed
	case "INPROGRESS":
		return cmpgn.ChangesetCheckStatePending
	case "SUCCESSFUL":
		return cmpgn.ChangesetCheckStatePassed
	default:
		return cmpgn.ChangesetCheckStateUnknown
	}
}

func computeGitHubCheckState(lastSynced time.Time, pr *github.PullRequest, events []*cmpgn.ChangesetEvent) cmpgn.ChangesetCheckState {
	// We should only consider the latest commit. This could be from a sync or a webhook that
	// has occurred later
	var latestCommitTime time.Time
	var latestOID string
	statusPerContext := make(map[string]cmpgn.ChangesetCheckState)
	statusPerCheckSuite := make(map[string]cmpgn.ChangesetCheckState)
	statusPerCheckRun := make(map[string]cmpgn.ChangesetCheckState)

	if len(pr.Commits.Nodes) > 0 {
		// We only request the most recent commit
		commit := pr.Commits.Nodes[0]
		latestCommitTime = commit.Commit.CommittedDate
		latestOID = commit.Commit.OID
		// Calc status per context for the most recent synced commit
		for _, c := range commit.Commit.Status.Contexts {
			statusPerContext[c.Context] = parseGithubCheckState(c.State)
		}
		for _, c := range commit.Commit.CheckSuites.Nodes {
			if c.Status == "QUEUED" && len(c.CheckRuns.Nodes) == 0 {
				// Ignore queued suites with no runs.
				// It is common for suites to be created and then stay in the QUEUED state
				// forever with zero runs.
				continue
			}
			statusPerCheckSuite[c.ID] = parseGithubCheckSuiteState(c.Status, c.Conclusion)
			for _, r := range c.CheckRuns.Nodes {
				statusPerCheckRun[r.ID] = parseGithubCheckSuiteState(r.Status, r.Conclusion)
			}
		}
	}

	var statuses []*github.CommitStatus
	// Get all status updates that have happened since our last sync
	for _, e := range events {
		switch m := e.Metadata.(type) {
		case *github.CommitStatus:
			if m.ReceivedAt.After(lastSynced) {
				statuses = append(statuses, m)
			}
		case *github.PullRequestCommit:
			if m.Commit.CommittedDate.After(latestCommitTime) {
				latestCommitTime = m.Commit.CommittedDate
				latestOID = m.Commit.OID
				// statusPerContext is now out of date, reset it
				for k := range statusPerContext {
					delete(statusPerContext, k)
				}
			}
		case *github.CheckSuite:
			if m.Status == "QUEUED" && len(m.CheckRuns.Nodes) == 0 {
				// Ignore suites with no runs.
				// See previous comment.
				continue
			}
			if m.ReceivedAt.After(lastSynced) {
				statusPerCheckSuite[m.ID] = parseGithubCheckSuiteState(m.Status, m.Conclusion)
			}
		case *github.CheckRun:
			if m.ReceivedAt.After(lastSynced) {
				statusPerCheckRun[m.ID] = parseGithubCheckSuiteState(m.Status, m.Conclusion)
			}
		}
	}

	if len(statuses) > 0 {
		// Update the statuses using any new webhook events for the latest commit
		sort.Slice(statuses, func(i, j int) bool {
			return statuses[i].ReceivedAt.Before(statuses[j].ReceivedAt)
		})
		for _, s := range statuses {
			if s.SHA != latestOID {
				continue
			}
			statusPerContext[s.Context] = parseGithubCheckState(s.State)
		}
	}
	finalStates := make([]cmpgn.ChangesetCheckState, 0, len(statusPerContext))
	for k := range statusPerContext {
		finalStates = append(finalStates, statusPerContext[k])
	}
	for k := range statusPerCheckSuite {
		finalStates = append(finalStates, statusPerCheckSuite[k])
	}
	for k := range statusPerCheckRun {
		finalStates = append(finalStates, statusPerCheckRun[k])
	}
	return combineCheckStates(finalStates)
}

// combineCheckStates combines multiple check states into an overall state
// pending takes highest priority
// followed by error
// success return only if all successful
func combineCheckStates(states []cmpgn.ChangesetCheckState) cmpgn.ChangesetCheckState {
	if len(states) == 0 {
		return cmpgn.ChangesetCheckStateUnknown
	}
	stateMap := make(map[cmpgn.ChangesetCheckState]bool)
	for _, s := range states {
		stateMap[s] = true
	}

	switch {
	case stateMap[cmpgn.ChangesetCheckStateUnknown]:
		// If are pending, overall is Pending
		return cmpgn.ChangesetCheckStateUnknown
	case stateMap[cmpgn.ChangesetCheckStatePending]:
		// If are pending, overall is Pending
		return cmpgn.ChangesetCheckStatePending
	case stateMap[cmpgn.ChangesetCheckStateFailed]:
		// If no pending, but have errors then overall is Failed
		return cmpgn.ChangesetCheckStateFailed
	case stateMap[cmpgn.ChangesetCheckStatePassed]:
		// No pending or errors then overall is Passed
		return cmpgn.ChangesetCheckStatePassed
	}

	return cmpgn.ChangesetCheckStateUnknown
}

func parseGithubCheckState(s string) cmpgn.ChangesetCheckState {
	s = strings.ToUpper(s)
	switch s {
	case "ERROR", "FAILURE":
		return cmpgn.ChangesetCheckStateFailed
	case "EXPECTED", "PENDING":
		return cmpgn.ChangesetCheckStatePending
	case "SUCCESS":
		return cmpgn.ChangesetCheckStatePassed
	default:
		return cmpgn.ChangesetCheckStateUnknown
	}
}

func parseGithubCheckSuiteState(status, conclusion string) cmpgn.ChangesetCheckState {
	status = strings.ToUpper(status)
	conclusion = strings.ToUpper(conclusion)
	switch status {
	case "IN_PROGRESS", "QUEUED", "REQUESTED":
		return cmpgn.ChangesetCheckStatePending
	}
	if status != "COMPLETED" {
		return cmpgn.ChangesetCheckStateUnknown
	}
	switch conclusion {
	case "SUCCESS", "NEUTRAL":
		return cmpgn.ChangesetCheckStatePassed
	case "ACTION_REQUIRED":
		return cmpgn.ChangesetCheckStatePending
	case "CANCELLED", "FAILURE", "TIMED_OUT":
		return cmpgn.ChangesetCheckStateFailed
	}
	return cmpgn.ChangesetCheckStateUnknown
}

// computeSingleChangesetState of a Changeset based on the metadata.
// It does NOT reflect the final calculated state, use `ExternalState` instead.
func computeSingleChangesetState(c *cmpgn.Changeset) (s cmpgn.ChangesetState, err error) {
	if !c.ExternalDeletedAt.IsZero() {
		return cmpgn.ChangesetStateDeleted, nil
	}

	switch m := c.Metadata.(type) {
	case *github.PullRequest:
		s = cmpgn.ChangesetState(m.State)
	case *bitbucketserver.PullRequest:
		if m.State == "DECLINED" {
			s = cmpgn.ChangesetStateClosed
		} else {
			s = cmpgn.ChangesetState(m.State)
		}
	default:
		return "", errors.New("unknown changeset type")
	}

	if !s.Valid() {
		return "", errors.Errorf("changeset state %q invalid", s)
	}

	return s, nil
}

// computeSingleChangesetReviewState computes the review state of a Changeset.
// GitHub doesn't keep the review state on a changeset, so a GitHub Changeset
// will always return ChangesetReviewStatePending.
//
// This method should NOT be called directly. Use ComputeReviewState instead.
func computeSingleChangesetReviewState(c *cmpgn.Changeset) (s cmpgn.ChangesetReviewState, err error) {
	states := map[cmpgn.ChangesetReviewState]bool{}

	switch m := c.Metadata.(type) {
	case *github.PullRequest:
		// For GitHub we need to use `ChangesetEvents.ReviewState`
		log15.Warn("Changeset.ReviewState() called, but GitHub review state is calculated through ChangesetEvents.ReviewState", "changeset", c)
		return cmpgn.ChangesetReviewStatePending, nil

	case *bitbucketserver.PullRequest:
		for _, r := range m.Reviewers {
			switch r.Status {
			case "UNAPPROVED":
				states[cmpgn.ChangesetReviewStatePending] = true
			case "NEEDS_WORK":
				states[cmpgn.ChangesetReviewStateChangesRequested] = true
			case "APPROVED":
				states[cmpgn.ChangesetReviewStateApproved] = true
			}
		}
	default:
		return "", errors.New("unknown changeset type")
	}

	return selectReviewState(states), nil
}

// selectReviewState computes the single review state for a given set of
// ChangesetReviewStates. Since a pull request, for example, can have multiple
// reviews with different states, we need a function to determine what the
// state for the pull request is.
func selectReviewState(states map[cmpgn.ChangesetReviewState]bool) cmpgn.ChangesetReviewState {
	// If any review requested changes, that state takes precedence over all
	// other review states, followed by explicit approval. Everything else is
	// considered pending.
	for _, state := range [...]cmpgn.ChangesetReviewState{
		cmpgn.ChangesetReviewStateChangesRequested,
		cmpgn.ChangesetReviewStateApproved,
	} {
		if states[state] {
			return state
		}
	}

	return cmpgn.ChangesetReviewStatePending
}

// computeOverallReviewState returns the overall review state given a map of
// reviews per author.
func computeReviewState(statesByAuthor map[string]campaigns.ChangesetReviewState) campaigns.ChangesetReviewState {
	states := make(map[campaigns.ChangesetReviewState]bool)
	for _, s := range statesByAuthor {
		states[s] = true
	}
	return selectReviewState(states)
}

func unixMilliToTime(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond))
}
