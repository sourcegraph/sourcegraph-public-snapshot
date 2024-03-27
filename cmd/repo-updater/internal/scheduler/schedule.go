package scheduler

import (
	"container/heap"
	"sync"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// schedule is the schedule of when repos get enqueued into the updateQueue.
type schedule struct {
	mu sync.Mutex

	heap  []*scheduledRepoUpdate // min heap of scheduledRepoUpdates based on their due time.
	index map[api.RepoID]*scheduledRepoUpdate

	// timer sends a value on the wakeup channel when it is time
	timer  *time.Timer
	wakeup chan struct{}
	logger log.Logger

	// random source used to add jitter to repo update intervals.
	randGenerator interface {
		Int63n(n int64) int64
	}
}

// upsert inserts or updates a repo in the schedule.
func (s *schedule) upsert(repo configuredRepo) (updated bool) {
	if repo.ID == 0 {
		panic("repo.id is zero")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if update := s.index[repo.ID]; update != nil {
		update.Repo = repo
		return true
	}

	heap.Push(s, &scheduledRepoUpdate{
		Repo:     repo,
		Interval: minDelay,
		Due:      timeNow().Add(minDelay),
	})

	s.rescheduleTimer()

	return false
}

func (s *schedule) prioritiseUncloned(uncloned []types.MinimalRepo) {
	// All non-cloned repos will be due for cloning as if they are newly added
	// repos.
	notClonedDue := timeNow().Add(minDelay)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Iterate over all repos in the scheduler. If it isn't in cloned bump it
	// up the queue. Note: we iterate over index because we will be mutating
	// heap.
	rescheduleTimer := false
	for _, repo := range uncloned {
		if repoUpdate := s.index[repo.ID]; repoUpdate == nil {
			heap.Push(s, &scheduledRepoUpdate{
				Repo:     configuredRepo{ID: repo.ID, Name: repo.Name},
				Interval: minDelay,
				Due:      notClonedDue,
			})
			rescheduleTimer = true
		} else if repoUpdate.Due.After(notClonedDue) {
			repoUpdate.Due = notClonedDue
			heap.Fix(s, repoUpdate.Index)
			rescheduleTimer = true
		}
	}

	// We updated the queue, inform the scheduler loop.
	if rescheduleTimer {
		s.rescheduleTimer()
	}
}

// insertNew will insert repos only if they are not known to the scheduler
func (s *schedule) insertNew(repos []types.MinimalRepo) {
	configuredRepos := make([]configuredRepo, len(repos))
	for i := range repos {
		configuredRepos[i] = configuredRepo{
			ID:   repos[i].ID,
			Name: repos[i].Name,
		}
	}

	due := timeNow().Add(minDelay)
	rescheduleTimer := false

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, repo := range configuredRepos {
		if update := s.index[repo.ID]; update != nil {
			continue
		}
		heap.Push(s, &scheduledRepoUpdate{
			Repo:     repo,
			Interval: minDelay,
			Due:      due,
		})
		rescheduleTimer = true
	}

	if rescheduleTimer {
		s.rescheduleTimer()
	}
}

// updateInterval updates the update interval of a repo in the schedule.
// It does nothing if the repo is not in the schedule.
func (s *schedule) updateInterval(repo configuredRepo, interval time.Duration) {
	if repo.ID == 0 {
		panic("repo.id is zero")
	}

	s.mu.Lock()
	if update := s.index[repo.ID]; update != nil {
		switch {
		case interval > maxDelay:
			update.Interval = maxDelay
		case interval < minDelay:
			update.Interval = minDelay
		default:
			update.Interval = interval
		}

		// Add a jitter of 5% on either side of the interval to avoid
		// repos getting updated at the same time.
		delta := int64(update.Interval) / 20
		update.Interval = update.Interval + time.Duration(s.randGenerator.Int63n(2*delta)-delta)

		update.Due = timeNow().Add(update.Interval)
		s.logger.Debug("updated repo",
			log.Object("repo", log.String("name", string(repo.Name)), log.Duration("due", update.Due.Sub(timeNow()))),
		)
		heap.Fix(s, update.Index)
		s.rescheduleTimer()
	}
	s.mu.Unlock()
}

// getCurrentInterval gets the current interval for the supplied repo and a bool
// indicating whether it was found.
func (s *schedule) getCurrentInterval(repo configuredRepo) (time.Duration, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	update, ok := s.index[repo.ID]
	if !ok || update == nil {
		return 0, false
	}
	return update.Interval, true
}

// remove removes a repo from the schedule.
func (s *schedule) remove(repo configuredRepo) (removed bool) {
	if repo.ID == 0 {
		panic("repo.id is zero")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	update := s.index[repo.ID]
	if update == nil {
		return false
	}

	reschedule := update.Index == 0
	if heap.Remove(s, update.Index); reschedule {
		s.rescheduleTimer()
	}

	return true
}

// rescheduleTimer schedules the scheduler to wakeup
// at the time that the next repo is due for an update.
// The caller must hold the lock on s.mu.
func (s *schedule) rescheduleTimer() {
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
	if len(s.heap) > 0 {
		delay := s.heap[0].Due.Sub(timeNow())
		s.timer = timeAfterFunc(delay, func() {
			notify(s.wakeup)
		})
	}
}

func (s *schedule) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.heap = s.heap[:0]
	s.index = map[api.RepoID]*scheduledRepoUpdate{}
	s.wakeup = make(chan struct{}, notifyChanBuffer)
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}

	s.logger.Debug("schedKnownRepos reset")
	schedKnownRepos.Set(0)
}

// The following methods implement heap.Interface based on the priority queue example:
// https://golang.org/pkg/container/heap/#example__priorityQueue
// These methods are not safe for concurrent use. Therefore, it is the caller's
// responsibility to ensure they're being guarded by a mutex during any heap operation,
// i.e. heap.Fix, heap.Remove, heap.Push, heap.Pop.

func (s *schedule) Len() int { return len(s.heap) }

func (s *schedule) Less(i, j int) bool {
	return s.heap[i].Due.Before(s.heap[j].Due)
}

func (s *schedule) Swap(i, j int) {
	s.heap[i], s.heap[j] = s.heap[j], s.heap[i]
	s.heap[i].Index = i
	s.heap[j].Index = j
}

func (s *schedule) Push(x any) {
	n := len(s.heap)
	item := x.(*scheduledRepoUpdate)
	item.Index = n
	s.heap = append(s.heap, item)
	s.index[item.Repo.ID] = item
	schedKnownRepos.Inc()
}

func (s *schedule) Pop() any {
	n := len(s.heap)
	item := s.heap[n-1]
	item.Index = -1 // for safety
	s.heap = s.heap[0 : n-1]
	delete(s.index, item.Repo.ID)
	schedKnownRepos.Dec()
	return item
}
