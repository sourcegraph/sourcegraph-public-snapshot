// Package ratelimit implements rate limiting for discussions.
package ratelimit

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/discussions/mentions"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

type limit struct {
	// maxActions describes the maximum number of actions that can be performed
	// within the sliding window period.
	maxActions int

	// window describes how large the sliding window of opportunity is.
	window time.Duration

	// uniquePenaltyFactor describes how many actions to subtract from
	// maxActions when a unique action is encountered. The exact number of
	// actions removed is calculated via:
	//
	// 	penalty = maxActions - min(maxActions / (uniquePenaltyFactor * uniqueActions-1), maxUniquePenalty)
	//
	// For example, this is used to allow creating many comments/threads within
	// the same file (e.g. for soft real-time chat), but penalize spammers from
	// creating many comments/threads in random files.
	uniquePenaltyFactor float64

	// maxUniquePenalty describes the maximum number of actions that can be
	// subtracted from maxActions as a penalty when unique actions are
	// encountered. See uniquePenaltyFactor for more information.
	maxUniquePenalty int
}

// calculateWaitTime tells how long a user must wait to perform a single action
// given the actions that have previously occurred in the window of oppertunity.
func (l limit) calculateWaitTime(actions []action) time.Duration {
	// Calculate the penalty (see uniquePenaltyFactor's documentation for
	// information about this).
	penalty := 0
	uniqueActions := max(countUniqueActions(actions), 1) - 1
	if l.uniquePenaltyFactor != 0 && uniqueActions != 0 {
		penalty = l.maxActions - int(float64(l.maxActions)/(l.uniquePenaltyFactor*float64(uniqueActions)))
	}
	if penalty > l.maxUniquePenalty {
		penalty = l.maxUniquePenalty
	}

	maxActions := max(l.maxActions-penalty, 0)
	consumedActions := len(actions)
	remainingActions := maxActions - consumedActions
	if remainingActions >= 1 {
		return 0 // they do not need to wait.
	}

	sortActionsOldestFirst(actions)
	return l.window - time.Since(actions[0].time)
}

func countUniqueActions(actions []action) int {
	set := map[string]struct{}{}
	for _, a := range actions {
		if _, ok := set[a.key]; !ok {
			set[a.key] = struct{}{}
		}
	}
	return len(set)
}

func sortActionsOldestFirst(actions []action) {
	sort.Slice(actions, func(i, j int) bool {
		return actions[i].time.Before(actions[j].time)
	})
}

type action struct {
	time time.Time // the time the action occurred at, e.g. when a thread or comment was created
	key  string    // a unique key representing the action, e.g. the thread ID of a comment, or the repo+rev+file of a thread.
}

func (a action) String() string { return time.Since(a.time).String() }

var (
	// createThreadLimit defines the rate limit for thread creation, it is
	// chosen based on:
	//
	// - We assume threads themselves are created less often than comments on
	//   threads.
	// - We assume each thread will take at least 30s to write and submit
	//   end-to-end.
	// - Some users may wish to create threads very quickly in rapid succession
	//   _within the same file_, e.g. to quickly call out lines of code that
	//   are not needed with "also here" comments.
	//
	// Examples of the limits imposed below (based on the number of unique
	// files a thread was created on in the past 1m):
	//
	// 	1 unique files == 12 threads / 1m (1 every 5s)
	// 	2 unique files == 6 threads  / 1m (1 every 10s)
	// 	3 unique files == 3 threads  / 1m (1 every 20s)
	// 	4 unique files == 2 threads  / 1m (1 every 30s)
	//
	createThreadLimit = limit{
		maxActions:          12,
		window:              1 * time.Minute,
		uniquePenaltyFactor: 2,
		maxUniquePenalty:    10,
	}

	// addCommentLimit defines the rate limit for adding comments to a thread,
	// it is chosen based on:
	//
	// - The fact that comments are created more often than threads.
	// - We want to support soft-realtime chat where e.g. you quickly respond
	//   to someone as you would in e.g. IRC or Slack.
	// - Comment spam is less harmful when it spans only a few threads, e.g.
	//   60 comments in a single thread is better than 60 comments each in
	//   random threads.
	//
	// Examples of the limits imposed below (based on the number of unique
	// threads a comment was added to in the past 5m):
	//
	// 	1 unique threads == 20 comments / 1m (1 every 3s)
	// 	2 unique threads == 6  comments / 1m (1 every 9s)
	// 	3 unique threads == 3  comments / 1m (1 every 18s)
	// 	4 unique threads == 2  comments / 1m (1 every 30s)
	//
	addCommentLimit = limit{
		maxActions:          20,
		window:              1 * time.Minute,
		uniquePenaltyFactor: 3,
		maxUniquePenalty:    18,
	}

	// Mentions defines how many @mention notifications a user can send,
	// globally across all threads, during the window period.
	//
	// When the limit is hit, thread or comment creation will fail. This limit
	// has no effect on subsequent messages which notify people previously
	// @mentioned.
	//
	// The primary purpose of this limit is just to prevent someone from
	// @mentioning hundreds of users and spamming them.
	mentionsLimit = limit{
		maxActions: 15,
		window:     1 * time.Minute,
	}
)

var one800DBError = time.Duration(18003237767) // "18.003237767s" (or "1-800-DBERROR")

// TimeUntilUserCanCreateThread tells how long the user must wait until they
// may create one new discussion thread according to rate limiting.
//
// This ONLY considers rate limiting, it does NOT verify the user otherwise has
// permission to create discussion threads.
func TimeUntilUserCanCreateThread(ctx context.Context, userID int32, newThreadTitle, newThreadContent string) (mustWait time.Duration) {
	// Determine how many comments the user has created in the last window
	// period.
	createdAfter := time.Now().Add(-createThreadLimit.window)
	threads, err := db.DiscussionThreads.List(ctx, &db.DiscussionThreadsListOptions{
		AuthorUserIDs: []int32{userID},
		CreatedAfter:  &createdAfter,
	})
	if err != nil {
		// It's OK to swallow the error here because showing this to the user
		// would never be useful, and if there is an error here it is likely to
		// be e.g. an outright database failure.
		log15.Error("discussions: failed to determine ratelimit for thread creation", "error", err)
		return one800DBError
	}
	var actions []action
	for _, t := range threads {
		targets, err := db.DiscussionThreads.ListTargets(ctx, t.ID)
		if err != nil {
			log15.Error("discussions: failed to determine ratelimit for thread creation", "error", err)
			return one800DBError
		}

		// For rate-limiting purposes, take only the first target of each thread (if any). Treating
		// a thread action as a unique action on each target would be unduly harsh, and taking only
		// the first target is not subject to gaming (i.e., the user altering the order of a
		// thread's targets would only yield a stricter rate limit, not a more lenient one).
		var key string
		if len(targets) == 1 {
			key = fmt.Sprint(targets[0].RepoID, orEmpty(targets[0].Path), orEmpty(targets[0].Branch), orEmpty(targets[0].Revision))
		} else {
			// We don't know what this type of target is, so we assume it does
			// not need to be rate limited harshly.
			key = "not-unique"
		}
		actions = append(actions, action{
			time: t.CreatedAt,
			key:  key,
		})
	}
	t1 := createThreadLimit.calculateWaitTime(actions)
	t2 := timeUntilUserCanMention(ctx, userID, newThreadTitle+" "+newThreadContent)
	if t1 > t2 {
		return t1
	}
	return t2
}

// TimeUntilUserCanAddCommentToThread tells how long the user must wait until
// they may add a new comment to a thread according to rate limiting.
//
// This ONLY considers rate limiting, it does NOT verify the user otherwise has
// permission to create discussion threads.
func TimeUntilUserCanAddCommentToThread(ctx context.Context, userID int32, newCommentContent string) (mustWait time.Duration) {
	// Determine how many comments the user has created in the last window
	// period.
	createdAfter := time.Now().Add(-addCommentLimit.window)
	comments, err := db.DiscussionComments.List(ctx, &db.DiscussionCommentsListOptions{
		AuthorUserID: &userID,
		CreatedAfter: &createdAfter,
	})
	if err != nil {
		// It's OK to swallow the error here because showing this to the user
		// would never be useful, and if there is an error here it is likely to
		// be e.g. an outright database failure.
		log15.Error("discussions: failed to determine ratelimit for adding comment to thread", "error", err)
		return one800DBError
	}
	var actions []action
	for _, c := range comments {
		actions = append(actions, action{
			time: c.CreatedAt,
			key:  fmt.Sprint(c.ThreadID),
		})
	}
	t1 := addCommentLimit.calculateWaitTime(actions)
	t2 := timeUntilUserCanMention(ctx, userID, newCommentContent)
	if t1 > t2 {
		return t1
	}
	return t2
}

// timeUntilUserCanMention tells how long the user must wait until they may
// add a comment or create a thread which mentions the user specified in
// newContents.
//
// IMPORTANT: We must outright block creation of the comment or thread (or
// actually remove the @mentions from their message) when the mention rate
// limit is hit. If we did not and instead just did not send the notification,
// those mentioned users would become implicitly 'subscribed' to the thread and
// the rate limit would be bypassed by simply creating a new reply to the
// thread after posting the rate-limited one mentioning users. To solve this,
// we would likely have to create a proper subscription database and not just
// use the thread comment history for determining subscriptions.
func timeUntilUserCanMention(ctx context.Context, userID int32, newContents string) (mustWait time.Duration) {
	// Determine how many comments the user has created in the last window
	// period.
	createdAfter := time.Now().Add(-mentionsLimit.window)
	comments, err := db.DiscussionComments.List(ctx, &db.DiscussionCommentsListOptions{
		CreatedAfter: &createdAfter,
		AuthorUserID: &userID,
	})
	if err != nil {
		// It's OK to swallow the error here because showing this to the user
		// would never be useful, and if there is an error here it is likely to
		// be e.g. an outright database failure.
		log15.Error("discussions: failed to determine ratelimit for mentions in thread", "error", err)
		return one800DBError
	}
	set := map[string]struct{}{}
	var actions []action
	for _, c := range comments {
		for _, mention := range mentions.Parse(c.Contents) {
			if _, ok := set[mention]; ok {
				continue
			}
			set[mention] = struct{}{}

			// Verify the user actually exists so that someone e.g. pasting a
			// poorly formatted list starting with @ constantly doesn't consume
			// their entire notification ratelimit.
			_, err := db.Users.GetByUsername(ctx, mention)
			if err != nil {
				continue
			}

			actions = append(actions, action{
				key:  mention,
				time: c.CreatedAt,
			})
		}
	}
	return mentionsLimit.calculateWaitTime(actions)
}

func orEmpty(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
