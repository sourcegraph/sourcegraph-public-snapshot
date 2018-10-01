package repos

import (
	"container/heap"
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	gitserverprotocol "github.com/sourcegraph/sourcegraph/pkg/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/pkg/honey"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// This file contains the repo-updater scheduler and "repos.list" config
// handling. The repo-updater scheduler is a scheduler for running git fetch
// which scales to tens of thousands of repositories.
//
// The best way to understand the scheduler is to start by reading the
// updateLoop function. The scheduler is designed to not run more than
// maxConcurrent fetches at once, and to order fetches so as to minimise the
// syncing lag between gitserver and a remote repository.
//
// The way it achieves this is by setting a deadline for when to run fetch for
// a repository (repoData.due). This deadline is based on the last time a
// fetch ran and the last time the repository changed. Repositories that are
// recently changed will be checked more often.
//
// We store all repositories in a heap to efficiently find the next due
// repository as well as efficiently mutate the due time for a repository.
//
// The scheduler also takes into account user traffic and will update
// repositories users are browsing. This is done via a separate queue called
// bump. Additionally if we do not know the last time a repository was fetched
// we do not have a good deadline to check it. So those items are placed in a
// third queue called newQueue.
//
// updateLoop will always try to schedule repositories in bump immediately. If
// we are not at maxConcurrent fetches yet, then it will also schedule items
// past due in the heap. If we are still not at maxConcurrent fetches then it
// will schedule as many repositories from newQueue as possible. The reason
// newQueue is processed last is we do not want a large number of new
// repositories starving existing repositories from being updated.
//
// See the original design document at
// https://github.com/sourcegraph/docs-private/blob/master/201806/repo.md
//
// TODO: Separate "repos.list" code and the scheduler.

const (
	minDelay = 45 * time.Second
	maxDelay = 8 * time.Hour
)

var (
	envNewScheduler      = env.Get("SRC_UPDATE_SCHEDULER", "enabled", "Use updated repo-update scheduler.")
	useNewScheduler      bool
	useNewSchedulerMutex sync.Mutex
)

// NewScheduler indicates whether the new scheduler is active.
func NewScheduler() bool {
	useNewSchedulerMutex.Lock()
	defer useNewSchedulerMutex.Unlock()
	return useNewScheduler
}

// setNewScheduler enables or disables the new scheduler, returning the
// previous state.
func setNewScheduler(v bool) bool {
	useNewSchedulerMutex.Lock()
	defer useNewSchedulerMutex.Unlock()
	// Swap states so we can return the old one.
	useNewScheduler, v = v, useNewScheduler
	return v
}

// repo represents a repository we're tracking.
type repoData struct {
	name      string        // name used as the unique key, also sometimes api.RepoURI
	url       string        // origin URL
	heapIndex int           // the location of this repo in our heap
	due       time.Time     // next time this repo should be updated
	started   time.Time     // timestamp from starting work, used to compute runtime
	auto      bool          // schedule for automatic updates
	manual    bool          // has a manual update request pending
	working   bool          // currently being processed
	interval  time.Duration // how often to check it
	fetchTime time.Duration // duration of last fetch
}

func (r *repoData) scatterDelay() time.Duration {
	seconds := r.interval.Seconds()
	if seconds < minDelay.Seconds() {
		seconds = minDelay.Seconds()
	}
	return time.Duration(rand.Int()%int(seconds)) * time.Second
}

// Loosely based on the Priority Queue example in godocs.
type repoHeap []*repoData

func (rh repoHeap) Len() int {
	return len(rh)
}

// The item Pop() yields will be the one with the lowest time.
func (rh repoHeap) Less(i, j int) bool {
	return rh[i].due.Before(rh[j].due)
}

func (rh repoHeap) Swap(i, j int) {
	rh[i], rh[j] = rh[j], rh[i]
	rh[i].heapIndex = i
	rh[j].heapIndex = j
}

func (rh *repoHeap) Push(x interface{}) {
	n := len(*rh)
	repo := x.(*repoData)
	repo.heapIndex = n
	*rh = append(*rh, repo)
}

func (rh *repoHeap) Pop() interface{} {
	old := *rh
	n := len(old)
	repo := old[n-1]
	repo.heapIndex = -1
	*rh = old[0 : n-1]
	return repo
}

// peek() peeks ahead to see the next item on the heap without popping it
func (rh *repoHeap) peek() *repoData {
	h := *rh
	if len(h) > 0 {
		return h[0]
	}
	return nil
}

// repoListStats is a set of stats bundled up for easy reporting. The "scale"
// value is just a copy of something from the parent repoList, present here
// so the String() method can print it, but present there because it's
// not really contingent on the stat reporting.
type repoListStats struct {
	manualFetches int     // number of manual fetch operations
	autoFetches   int     // number of auto/queued fetch operations
	autoQueue     int     // length of queue
	knownRepos    int     // total number of repos known
	loops         int     // number of times through main loop
	errors        int     // errors encountered trying to do things
	scale         float64 // the interval scale (repoList.intervalScale)
}

// Honey uses Honeycomb, if configured, to report the stats.
func (s repoListStats) Honey() {
	if !honey.Enabled() {
		return
	}
	ev := honey.Event("repo-updater")
	ev.AddField("source", "new-scheduler-stats")
	ev.AddField("fetches", s.manualFetches+s.autoFetches)
	ev.AddField("errors", s.errors)
	ev.AddField("manual_fetches", s.manualFetches)
	ev.AddField("auto_fetches", s.autoFetches)
	ev.AddField("auto_queue", s.autoQueue)
	ev.AddField("known_repos", s.knownRepos)
	ev.AddField("loops", s.loops)
	ev.AddField("scale", s.scale)
	ev.Send()
}

// String allows log15 to display the stats in an intelligble format.
func (s repoListStats) String() string {
	return fmt.Sprintf("fetches: %d manual/%d auto, repos: %d queued/%d seen, loops: %d, timescale: %.2f",
		s.manualFetches, s.autoFetches, s.autoQueue, s.knownRepos, s.loops, s.scale)
}

// repoList is a list of repositories we're tracking. we keep them indexed
// both as a map, so we can look them up by name, and as a heap sorted by
// next-due timestamp.
type repoList struct {
	heap                repoHeap                  // a priority queue, actually sorted on timestamps
	autoUpdatesDisabled bool                      // should we do auto-updates?
	repos               map[string]*repoData      // reponame lookup
	bumped              []*repoData               // manual updates that get priority
	newQueue            []*repoData               // we only want to process new repos when there is spare capacity
	mu                  sync.Mutex                // locking to avoid races
	pingChan            chan string               // send reason-for-ping as a string here to ping the update worker
	confRepos           map[string]sourceRepoList // list of configured repos from each source
	intervalScale       float64                   // how much to scale intervals by
	nextDue             time.Time                 // next time we expect a thing to be ready
	stats               repoListStats             // usage stats so we can observe usage
	activeRequests      int                       // current active requests
	maxRequests         int                       // max requests we should attempt at once
}

// A configuredRepo represents the configuration data for a given repo from
// a configuration source, such as information retrieved from GitHub for a
// given GitHubConnection. The URI isn't present because it's the key used
// to look the repo up in a sourceRepoList.
type configuredRepo struct {
	url     string
	enabled bool
}

// a sourceRepoList represents the set of repositories associated with a
// specific source, such as a given GitHubConnection, or the main sourcegraph
// config file.
type sourceRepoList map[string]configuredRepo

// This list is the common point between the sync worker and incoming
// requests.
var repos = repoList{
	repos:               make(map[string]*repoData),
	autoUpdatesDisabled: false,
	pingChan:            make(chan string, 1), // buffered to prevent deadlocks
	maxRequests:         2,                    // intentionally low, should get updated by live data
}

func conservativeMaxRequests(max int) int {
	if max > 3 {
		return max - 2
	}
	return 1
}

// recomputeScale determines how long it'd take to do all the repo
// processing we would like to do over a given time interval, and if the
// answer is "too much", we set a scale factor to slow this down.
//
// Call this only when you hold r's mutex.
func (r *repoList) recomputeScale() {
	var scale float64
	queued := 0
	for _, repo := range r.repos {
		if repo.auto {
			queued++
			scale += float64(repo.fetchTime) / float64(repo.interval)
		}
	}
	r.stats.autoQueue = queued

	log15.Debug("computed scale", "scale", scale)
	r.intervalScale = scale
}

// baseInterval computes a reasonable update interval to use with a
// repository of a given age.
func baseInterval(age time.Duration) time.Duration {
	minimum := minDelay
	maximum := maxDelay
	hours := age.Hours()
	var interval time.Duration
	if hours > 48 {
		// 48 hours => 5 minutes
		// each additional day => one more minute
		interval = time.Duration(3+hours/24) * time.Minute
	} else {
		// roughly hours/12 minutes, so 48 ~= 4 minutes, which is to
		// say 5 seconds per hour, plus the 45-second minimum
		interval = time.Duration(45+hours*5) * time.Second
	}
	if interval < minimum {
		interval = minimum
	}
	if interval > maximum {
		interval = maximum
	}
	return interval
}

// interval computes a scaled update interval; intervals will be increased
// if there's enough repositories that we'd start flooding if we tried to keep
// them all updated.
func (r *repoList) interval(age time.Duration) time.Duration {
	if r.intervalScale == 0 {
		r.recomputeScale()
	}
	interval := baseInterval(age)
	// Scale intervals up (slow down) if we need to. We never scale
	// the other way; the goal is to avoid oversaturating, not to
	// ensure saturation.
	if r.intervalScale > 1 {
		interval = time.Duration(float64(interval) * r.intervalScale)
	}
	return interval
}

// ping attempts to wake up the main update loop if it would not already
// be waking up at or before the 'due' time.
//
// Call this only when you hold r's mutex.
func (r *repoList) ping(due time.Time, s string) {
	// only ping if our next wake up time needs to be adjusted.
	if !due.Before(r.nextDue) {
		return
	}

	// We do a non-blocking send. pingChan has a buffer size of 1, so if the
	// buffer is full, someone else already requested the ping. This is fine,
	// since it has the same desired effect.
	select {
	case r.pingChan <- s:
	default:
	}
}

// queue marks the named repository for automatic scheduling.
// if the repository does not exist, it is created by calling add().
//
// call only when you hold the mutex.
func (r *repoList) queue(name, url string) {
	repo, ok := r.repos[name]
	if !ok {
		// add() automatically queues it (to ensure initial clone)
		r.add(name, url, true)
		return
	}
	// Possibly update the URL for future updates.
	repo.url = url
	if repo.auto && repo.heapIndex >= 0 {
		// already done
		return
	}
	repo.auto = true
	// it's being processed manually, so it will get requeued later
	if repo.manual {
		return
	}
	// don't schedule an update when auto-updates are off.
	if r.autoUpdatesDisabled {
		return
	}
	repo.due = time.Now().Add(repo.scatterDelay())
	if repo.heapIndex >= 0 {
		heap.Fix(&r.heap, repo.heapIndex)
	} else {
		heap.Push(&r.heap, repo)
	}
	r.ping(repo.due, repo.name)
}

// dequeue unmarks the named repository for automatic scheduling. it does
// not create the repository if the repository doesn't already exist.
//
// call only when you hold the mutex.
func (r *repoList) dequeue(name, url string) {
	repo, ok := r.repos[name]
	if !ok {
		// nothing to do
		return
	}
	// Possibly update the URL for future updates.
	repo.url = url
	// remove from automatic schedule
	if repo.heapIndex >= 0 {
		heap.Remove(&r.heap, repo.heapIndex)
	}
	repo.auto = false
}

// update marks the named repository for a manual update.
// if the repository does not exist, it is created by calling add().
//
// call only when you hold the mutex.
func (r *repoList) update(name, url string) {
	repo, ok := r.repos[name]
	if !ok {
		r.add(name, url, false)
		return
	}
	// Possibly update the URL for future updates.
	repo.url = url
	if repo.manual || repo.working {
		// already scheduled
		return
	}
	repo.manual = true
	repo.due = time.Now()
	r.bumped = append(r.bumped, repo)
	// cancel any automatic scheduled processing; we'll still requeue
	// later if set for auto.
	if repo.heapIndex >= 0 {
		heap.Remove(&r.heap, repo.heapIndex)
	}
	r.ping(repo.due, repo.name)
}

// add creates the repository described, and schedules it for
// an initial clone sync. do not call add unless you hold the mutex
// for repoList.
func (r *repoList) add(name, url string, queue bool) {
	_, ok := r.repos[name]
	if ok {
		return
	}
	// create an entry
	log15.Debug("repoList add new", "repo", name)
	repo := &repoData{
		name:      string(name),
		url:       url,
		heapIndex: -1,
		due:       time.Now(),
		auto:      queue,
		interval:  r.interval(24 * time.Hour),
		fetchTime: 1 * time.Second,
	}
	r.repos[name] = repo
	if queue {
		r.newQueue = append(r.newQueue, repo)
	} else {
		r.bumped = append(r.bumped, repo)
	}
	r.ping(repo.due, repo.name)
}

// startUpdate is a helper function for initiating an update and setting
// flags. call only when you hold the mutex.
func (r *repoList) startUpdate(ctx context.Context, nextUp *repoData, auto bool) {
	r.activeRequests++
	nextUp.working = true
	nextUp.started = time.Now()
	if auto {
		r.stats.autoFetches++
		schedAutoFetch.Inc()
	} else {
		r.stats.manualFetches++
		schedManualFetch.Inc()
	}
	go r.doUpdate(ctx, nextUp)
}

// doUpdate attempts the actual update for a repo, calling r.requeue()
// when done. The URL is provided as an explicit argument so it's safe
// to modify the repo's URL while this function is running; it will
// use the URL configured when it was called.
//
// Only run if not holding the mutex.
func (r *repoList) doUpdate(ctx context.Context, repo *repoData) {
	// We need to hold the lock to read values from repo and r. So we read
	// everything we need at the very start.
	r.mu.Lock()
	name := repo.name
	url := repo.url
	interval := repo.interval
	manual := repo.manual
	autoUpdatesDisabled := r.autoUpdatesDisabled
	r.mu.Unlock()

	uri := api.RepoURI(name)
	var resp *gitserverprotocol.RepoUpdateResponse
	var err error

	// We do this even if we don't think we'll want to requeue it, because
	// someone could request that this become queued *while it is in process*?
	defer r.requeue(repo, &resp, &err)

	log15.Debug("doUpdate", "repo", name, "interval", interval)

	// Check whether it's cloned.
	cloned, err := gitserver.DefaultClient.IsRepoCloned(ctx, api.RepoURI(uri))
	if err != nil {
		log15.Warn("error checking if repo cloned", "repo", uri, "err", err)
		return
	}
	// We request an update if auto updates are enabled, or if the repo isn't
	// cloned, or the manual flag is set.
	if !cloned || manual || !autoUpdatesDisabled {
		// manual updates should happen even if the repo is usually rarely-updated.
		if manual {
			interval = 5 * time.Second
		}
		resp, err = gitserver.DefaultClient.RequestRepoUpdate(ctx, gitserver.Repo{Name: api.RepoURI(name), URL: url}, interval)
		if err != nil {
			log15.Warn("error requesting repo update", "repo", name, "err", err)
			return
		}
	}
}

// requeue() picks a reasonable time to next update a repository which is
// marked for periodic updates.
func (r *repoList) requeue(repo *repoData, respP **gitserverprotocol.RepoUpdateResponse, errP *error) {
	resp := *respP
	err := *errP
	r.mu.Lock()
	defer r.mu.Unlock()
	r.activeRequests--
	if r.activeRequests < 0 {
		log15.Error("activeRequests went under zero", "repo", repo.name)
		r.activeRequests = 0
	}
	now := time.Now()
	// Clear inProcess flag while holding the lock on the repolist; see interactions
	// with repoList.add()
	repo.working = false
	// whether or not this was actually manual, any manual request is cleared by completing
	// a fetch attempt.
	repo.manual = false
	repo.fetchTime = now.Sub(repo.started)
	if err != nil {
		r.stats.errors++
		schedError.Inc()
	}
	if resp != nil {
		if resp.QueueCap > 0 {
			newMax := conservativeMaxRequests(resp.QueueCap)
			if newMax != r.maxRequests {
				log15.Info("changing max requests to avoid flooding gitserver", "old", r.maxRequests, "new", newMax, "gitserver", resp.QueueCap)
				r.maxRequests = newMax
			}
		}
		if resp.Finished != nil && resp.Received != nil {
			altTime := resp.Finished.Sub(*resp.Received)
			log15.Debug("time taken/reported", "repo", repo.name, "fetchTime", repo.fetchTime, "altTime", altTime)
		}
		switch {
		case resp.Error != "":
			// A failed fetch could indicate a problem like a bad auth token, so we want to be
			// conservative.
			repo.interval = repo.interval * 2
			// cap at a one-day loop.
			if repo.interval > 24*time.Hour {
				repo.interval = 24 * time.Hour
			}
			log15.Info("interval backoff due to error", "repo", repo.name, "error", resp.Error, "interval", repo.interval)
		case resp.LastChanged != nil:
			sinceLast := now.Sub(*resp.LastChanged)
			repo.interval = r.interval(sinceLast)
			log15.Debug("interval set", "repo", repo.name, "sinceLast", sinceLast, "interval", repo.interval)
		default:
			// If we don't have data on how old the repo is, we'll be aggressive,
			// partially because we'll probably get that data "soon"; usually this
			// would only happen during initial cloning. Note, this won't happen
			// if we get an actual error back from gitserver, and shouldn't happen
			// in any case where gitserver didn't have an error to report.
			repo.interval = r.interval(1 * time.Hour)
		}
	} else {
		log15.Debug("no response data from gitserver", "repo", repo.name)
		// No response at all, we try again relatively soon.
		repo.interval = r.interval(1 * time.Hour)
	}
	// if this repo is set for auto updates, and auto-updates are not disabled,
	// add it back to the queue.
	if repo.auto && !r.autoUpdatesDisabled && NewScheduler() {
		// Stagger retries to reduce flooding.
		repo.due = now.Add(repo.interval + time.Duration(rand.Int()%10)*time.Second)
		if repo.heapIndex >= 0 {
			heap.Fix(&r.heap, repo.heapIndex)
		} else {
			heap.Push(&r.heap, repo)
		}
		r.ping(repo.due, "requeue: "+repo.name)
	}
}

// updateLoop() does the actual periodic updates.
//
// Each time through the loop, we fire off any items which are "bumped",
// then fire off any scheduled items which are currently due, then wait
// until the next item is due, or until something else wakes us up.
func (r *repoList) updateLoop(ctx context.Context, shutdown chan struct{}) {
	log15.Debug("starting repo update loop")
	// We don't want to do the whole scale recomputation super often, so we
	// do a counter and run that computation every so often; currently
	// trying once per ten loops, which will mean once per ten times that
	// a thing is added *or* we wake up to send things.
	loopCounter := 0
	// For periodic things, such as "reporting on repo-updater activity"
	statTime := time.Now()
	// It's nice to get any feedback at all shortly after startup; about
	// a minute should be long enough to get the repo list somewhat populated.
	nextStatTime := statTime.Add(1 * time.Minute)
	for {
		r.mu.Lock()
		log15.Debug("updateLoop", "repos", len(r.repos), "queue", len(r.heap))
		now := time.Now()
		// Every ten loops or so, recompute scaling factor for time.
		loopCounter = (loopCounter + 1) % 10
		if loopCounter == 0 {
			r.recomputeScale()
		}
		// Update counters (cheap to do)
		{
			schedKnownRepos.Set(float64(len(r.repos)))
			schedScale.Set(r.intervalScale)
			schedLoops.Inc()
		}

		if now.After(nextStatTime) {
			// Report some convenient stats.
			r.stats.knownRepos = len(r.repos)
			r.stats.scale = r.intervalScale
			r.stats.Honey() // report also to Honeycomb
			log15.Info("update loop", "last", now.Sub(statTime), "stats", r.stats)
			r.stats.manualFetches = 0
			r.stats.autoFetches = 0
			r.stats.loops = 0
			r.stats.errors = 0
			// Hourly updates.
			nextStatTime = now.Add(60 * time.Minute)
			statTime = now
		}
		r.stats.loops++

		var nextUp *repoData
		var newBumped []*repoData
		for _, nextUp := range r.bumped {
			if r.activeRequests >= r.maxRequests {
				newBumped = append(newBumped, nextUp)
			} else {
				r.startUpdate(ctx, nextUp, false)
			}
		}
		r.bumped = newBumped
		for nextUp = r.heap.peek(); nextUp != nil && nextUp.due.Before(now) && r.activeRequests < r.maxRequests; nextUp = r.heap.peek() {
			// We didn't use Pop() above because popping and immediately pushing again
			// would be much more expensive, in the case where we woke up from a ping
			// rather than because the next item was due.
			nextUp = heap.Pop(&r.heap).(*repoData)
			log15.Debug("repo ready", "repo", nextUp.name)
			// process this entry, if it's not already running
			if !nextUp.working {
				r.startUpdate(ctx, nextUp, true)
			} else {
				// skip this update, maybe try again in the normal update interval.
				nextUp.due = now.Add(nextUp.interval)
				heap.Push(&r.heap, nextUp)
			}
		}

		// We have spare capacity to process new repos
		for r.activeRequests < r.maxRequests && len(r.newQueue) > 0 {
			newEntry := r.newQueue[0]
			r.newQueue = r.newQueue[1:]
			// A new entry is due when it is created. If something else, such as a
			// manual bump, caused it to get processed already, it might actually be
			// in process now, and we should skip it. Or it might have been processed,
			// and had a new due time picked, and we should skip it. In either case,
			// it's already where it needs to be in the regular queue(s).
			if now.After(newEntry.due) && !newEntry.working {
				// This is always an "automatic" update; it was not triggered by user
				// interaction with the repo, but was done automatically from config.
				r.startUpdate(ctx, newEntry, true)
			}
		}

		// Default time in the unlikely event that we have no repos at all,
		// in which case this loop waking up fairly often won't be a problem.
		waitTime := 10 * time.Second
		if nextUp != nil {
			waitTime = time.Until(nextUp.due) + 50*time.Millisecond
			if waitTime < 0 {
				waitTime = 1 * time.Second
			}
			log15.Debug("nextUp due", "interval", nextUp.interval, "due", nextUp.due, "repo", nextUp.name)
		}
		// If something is bumped, but in-process, we'll try again soon
		// note, if auto updates are on, that will be redundant, but that's why
		// we have the debouncing...
		if (len(r.bumped) > 0 || len(r.newQueue) > 0) && waitTime > 1*time.Second {
			waitTime = 1 * time.Second
		}
		r.nextDue = now.Add(waitTime)
		r.mu.Unlock()

		// DO NOT lock r.mu around this select; that would prevent ping from
		// working.
		select {
		case <-time.After(waitTime):
			log15.Debug("woke up after time", "interval", waitTime)
		case s := <-r.pingChan:
			log15.Debug("woken by ping", "s", s)
		case <-ctx.Done():
			log15.Info("context complete, terminating update loop.")
			return
		case <-shutdown:
			log15.Info("shutdown received. scheduler should be restarted soon.")
			// Drop any existing lists; they'll get recreated by periodic updates later
			// if the scheduler gets reenabled.
			r.mu.Lock()
			r.repos = make(map[string]*repoData)
			r.heap = make([]*repoData, 0)
			r.bumped = make([]*repoData, 0)
			r.mu.Unlock()
			return
		}
	}
}

// updateSource updates the list of configured repos associated with the given
// source.
func (r *repoList) updateSource(source string, newList sourceRepoList) (enqueued, dequeued int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.confRepos == nil {
		r.confRepos = make(map[string]sourceRepoList)
	}
	if r.confRepos[source] == nil {
		r.confRepos[source] = make(sourceRepoList)
	}
	oldList := r.confRepos[source]
	for name, value := range oldList {
		_, ok := newList[name]
		if !ok {
			dequeued++
			r.dequeue(name, value.url)
			delete(oldList, name)
		}
	}
	for name, value := range newList {
		if name == "" {
			log15.Warn("ignoring repo with empty name", "source", source, "url", value.url)
			continue
		}
		old, ok := oldList[name]
		switch {
		case !ok:
			log15.Info("adding repo", "source", source, "name", name, "url", value.url)
		case value.url != old.url, value.enabled != old.enabled:
			log15.Debug("updating repo", "source", source, "name", name, "url", value.url)
		default:
			// No change in whether or not it's enabled, no change in URL, we
			// can ignore this one.
			continue
		}
		oldList[name] = value
		if value.enabled {
			enqueued++
			r.queue(name, value.url)
		} else {
			dequeued++
			r.dequeue(name, value.url)
		}
	}
	return enqueued, dequeued
}

// RunRepositorySyncWorker runs the worker that syncs repositories from external code hosts to Sourcegraph
func RunRepositorySyncWorker(ctx context.Context) {
	shutdown := make(chan struct{})
	conf.Watch(func() {
		// Determine which scheduler to run.
		c := conf.Get()
		ef := c.ExperimentalFeatures
		sched := true
		if ef != nil {
			sched = ef.UpdateScheduler != "disabled"
		}
		// Allow direct environment override.
		if envNewScheduler == "disabled" {
			sched = false
		}
		prevSched := setNewScheduler(sched)

		// For any state other than "was using new scheduler, still are",
		// we need to shut down the previous scheduler.
		//
		if !sched || !prevSched {
			close(shutdown)
			shutdown = make(chan struct{})
		}
		// The new scheduler has to be started on transitions to it only,
		// the old scheduler gets restarted on every config change.
		if sched {
			repos.mu.Lock()
			repos.autoUpdatesDisabled = c.DisableAutoGitUpdates
			repos.mu.Unlock()
			if !prevSched {
				// Actually start the scheduler.
				go repos.updateLoop(ctx, shutdown)
			}
			repos.updateConfig(ctx, c.ReposList)
		} else {
			go startRepositorySyncWorker(ctx, shutdown)
		}
	})
}

// updateConfig responds to changes in the configured list of repositories;
// this is specifically the list of repositories directly configured, as opposed
// to repositories found by looking up keys from various services.
func (r *repoList) updateConfig(ctx context.Context, configs []*schema.Repository) {
	log15.Debug("repolist updateConfig")
	newList := make(sourceRepoList)
	for _, cfg := range configs {
		if cfg.Type == "" {
			cfg.Type = "git"
		}
		if cfg.Type != "git" {
			continue
		}
		if cfg.Path == "" {
			log15.Warn("ignoring repo with empty path in repos.list", "url", cfg.Url)
			continue
		}
		// Check whether repo already exists, if not create an entry for it.
		newRepo, err := api.InternalClient.ReposCreateIfNotExists(ctx, api.RepoCreateOrUpdateRequest{RepoURI: api.RepoURI(cfg.Path), Enabled: true})
		if err != nil {
			log15.Warn("error creating or checking for repo", "repo", cfg.Path)
			continue
		}
		newList[cfg.Path] = configuredRepo{url: cfg.Url, enabled: newRepo.Enabled}
	}
	r.updateSource("internalConfig", newList)
}

func startRepositorySyncWorker(ctx context.Context, shutdown chan struct{}) {
	configs := conf.Get().ReposList
	if len(configs) == 0 {
		return
	}

	for _, cfg := range configs {
		if cfg.Type == "" {
			cfg.Type = "git"
		}
		// We only support git repos at the moment.
		if cfg.Type != "git" {
			log15.Error("Error syncing repos, VCS type not supported", "type", cfg.Type, "repo", cfg.Path)
		}
	}
	for {
		fetches := 0
		errors := 0
		for i, cfg := range configs {
			log15.Debug("RunRepositorySyncWorker:updateRepo", "repoURL", cfg.Url, "ith", i, "total", len(configs))
			err := updateRepo(ctx, cfg)
			fetches++
			if err != nil {
				log15.Warn("error updating repo", "path", cfg.Path, "error", err)
				errors++
				continue
			}
		}
		if honey.Enabled() {
			ev := honey.Event("repo-updater")
			ev.AddField("source", "repo-sync-worker")
			ev.AddField("fetches", fetches)
			ev.AddField("errors", errors)
		}
		repoListUpdateTime.Set(float64(time.Now().Unix()))
		select {
		case <-shutdown:
			return
		case <-time.After(getUpdateInterval()):
		}
	}
}

func updateRepo(ctx context.Context, repoConf *schema.Repository) error {
	uri := api.RepoURI(repoConf.Path)
	repo, err := api.InternalClient.ReposCreateIfNotExists(ctx, api.RepoCreateOrUpdateRequest{RepoURI: uri, Enabled: true})
	if err != nil {
		return err
	}

	if !repo.Enabled {
		// The repo is not enabled.
		return nil
	}

	// Run a git fetch to kick-off an update or a clone if the repo doesn't already exist.
	cloned, err := gitserver.DefaultClient.IsRepoCloned(ctx, uri)
	if err != nil {
		return errors.Wrap(err, "error checking if repo cloned")
	}
	if !conf.Get().DisableAutoGitUpdates || !cloned {
		log15.Debug("fetching repos.list repo", "repo", uri, "url", repoConf.Url, "cloned", cloned)
		err := gitserver.DefaultClient.EnqueueRepoUpdateDeprecated(ctx, gitserver.Repo{Name: repo.URI, URL: repoConf.Url})
		if err != nil {
			return errors.Wrap(err, "error cloning repo")
		}
	}
	return nil
}

// UpdateOnce causes a single update of the given repository.
func UpdateOnce(ctx context.Context, name api.RepoURI, url string) {
	repos.mu.Lock()
	defer repos.mu.Unlock()
	repos.update(string(name), url)
}

// Queue requests periodic automatic updates of the given repository, which
// will happen only if automatic updates are enabled. It will also perform
// a one-time fetch/clone.
//
// When not using the new scheduler, we just perform an immediate update
// attempt. This "works" because the code host interfaces will be sending
// this request to us on their update cycle.
func Queue(ctx context.Context, name api.RepoURI, url string) {
	repos.mu.Lock()
	defer repos.mu.Unlock()
	repos.queue(string(name), url)
}

// Dequeue cancels periodic automatic updates of the given repository.
//
// When the scheduler isn't running, this does almost nothing, but it
// could unset the queue flag, which could matter if the configuration
// later re-enables the scheduler.
func Dequeue(ctx context.Context, name api.RepoURI, url string) {
	repos.mu.Lock()
	defer repos.mu.Unlock()
	repos.dequeue(string(name), url)
}

// GetExplicitlyConfiguredRepository reports information about a repository configured explicitly with "repos.list".
func GetExplicitlyConfiguredRepository(ctx context.Context, args protocol.RepoLookupArgs) (repo *protocol.RepoInfo, authoritative bool, err error) {
	if args.Repo == "" {
		return nil, false, nil
	}

	repoNameLower := api.RepoURI(strings.ToLower(string(args.Repo)))
	for _, repo := range conf.Get().ReposList {
		if api.RepoURI(strings.ToLower(string(repo.Path))) == repoNameLower {
			repoInfo := &protocol.RepoInfo{
				URI:          api.RepoURI(repo.Path),
				ExternalRepo: nil,
				VCS:          protocol.VCSInfo{URL: repo.Url},
			}
			if repo.Links != nil {
				repoInfo.Links = &protocol.RepoLinks{
					Root:   repo.Links.Repository,
					Blob:   repo.Links.Blob,
					Tree:   repo.Links.Tree,
					Commit: repo.Links.Commit,
				}
			}
			return repoInfo, true, nil
		}
	}

	return nil, false, nil // not found
}
