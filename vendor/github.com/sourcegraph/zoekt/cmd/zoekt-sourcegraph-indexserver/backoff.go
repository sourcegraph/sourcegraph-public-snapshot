package main

import (
	"time"

	"github.com/sourcegraph/log"
)

type backoff struct {
	// maxBackoff is the longest duration we will backoff indexing operations of the given repo.
	maxBackoff time.Duration
	// backoffDuration is used to determine the duration of backoff. consecutiveFailures * backoffDuration calculates the
	// duration set on failed indexing attempt.
	backoffDuration time.Duration
	// consecutiveFailures is the count of preceding consecutive failures.
	consecutiveFailures int
	// backOffUntil is the earliest time when we allow the item to be pushed to the heap. Until then the item will not be enqueued
	// and indexing will not be attempted.
	backoffUntil time.Time
}

func (b *backoff) Allow(now time.Time) bool {
	return b.backoffUntil.Before(now)
}

func (b *backoff) Reset() {
	b.consecutiveFailures = 0
	b.backoffUntil = time.Unix(0, 0)
}

func (b *backoff) Fail(now time.Time, logger log.Logger, opts IndexOptions) {
	backoffDuration := time.Duration(b.consecutiveFailures+1) * b.backoffDuration

	if backoffDuration > b.maxBackoff {
		backoffDuration = b.maxBackoff
	} else {
		b.consecutiveFailures++
	}
	b.backoffUntil = now.Add(backoffDuration)

	logger.Debug("Backoff subsequent attempts to index repository",
		log.String("repo", opts.Name),
		log.Uint32("id", opts.RepoID),
		log.Duration("backoff_duration", b.backoffDuration),
		log.Time("backoff_until", b.backoffUntil),
	)
}
