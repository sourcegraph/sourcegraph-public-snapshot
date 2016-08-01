package backend

import (
	"fmt"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/prometheus/client_golang/prometheus"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/eventsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/repotrackutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/statsutil"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/store"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/accesscontrol"
	"sourcegraph.com/sourcegraph/sourcegraph/services/svc"
	"sourcegraph.com/sqs/pbtypes"
)

var Builds sourcegraph.BuildsServer = &builds{}

type builds struct{}

var _ sourcegraph.BuildsServer = (*builds)(nil)

func (s *builds) Get(ctx context.Context, build *sourcegraph.BuildSpec) (*sourcegraph.Build, error) {
	return store.BuildsFromContext(ctx).Get(ctx, *build)
}

func (s *builds) List(ctx context.Context, opt *sourcegraph.BuildListOptions) (*sourcegraph.BuildList, error) {
	builds, err := store.BuildsFromContext(ctx).List(ctx, opt)
	if err != nil {
		return nil, err
	}

	// Find out if there are more pages.
	// StreamResponse.HasMore is set to true if next page has non-zero entries.
	// TODO(shurcooL): This can be optimized by structuring how pagination works a little better.
	var streamResponse sourcegraph.StreamResponse
	if opt != nil {
		moreOpt := *opt
		moreOpt.ListOptions.Page = int32(moreOpt.ListOptions.PageOrDefault()) + 1
		moreBuilds, err := store.BuildsFromContext(ctx).List(ctx, &moreOpt)
		if err != nil {
			return nil, err
		}
		streamResponse = sourcegraph.StreamResponse{HasMore: len(moreBuilds) > 0}
	}

	return &sourcegraph.BuildList{Builds: builds, StreamResponse: streamResponse}, nil
}

func (s *builds) Create(ctx context.Context, op *sourcegraph.BuildsCreateOp) (*sourcegraph.Build, error) {
	if len(op.CommitID) != 40 {
		return nil, grpc.Errorf(codes.InvalidArgument, "Builds.Create requires full commit ID")
	}

	repo, err := svc.Repos(ctx).Get(ctx, &sourcegraph.RepoSpec{ID: op.Repo})
	if err != nil {
		return nil, err
	}

	if repo.Blocked {
		return nil, grpc.Errorf(codes.FailedPrecondition, "repo %s is blocked", repo.URI)
	}

	// If we don't have the commit synced yet, fail the Create but enqueue
	// a repo update so next attempt will work.
	if _, err := svc.Repos(ctx).ResolveRev(ctx, &sourcegraph.ReposResolveRevOp{Repo: op.Repo, Rev: op.CommitID}); err != nil {
		repoMaybeEnqueueUpdate(ctx, repo)
		return nil, err
	}

	// Only an admin can re-enqueue a successful build
	if err = accesscontrol.VerifyUserHasAdminAccess(ctx, "Builds.Create"); err != nil {
		successful, err := s.List(ctx, &sourcegraph.BuildListOptions{
			Repo:      repo.ID,
			CommitID:  op.CommitID,
			Succeeded: true,
			ListOptions: sourcegraph.ListOptions{
				PerPage: 1,
			},
		})

		if err == nil && len(successful.Builds) > 0 {
			return successful.Builds[0], nil
		}
	}

	if op.Branch != "" && op.Tag != "" {
		return nil, grpc.Errorf(codes.InvalidArgument, "at most one of Branch and Tag may be specified when creating a build (repo %s commit %q)", op.Repo, op.CommitID)
	}

	b := &sourcegraph.Build{
		Repo:        repo.ID,
		CommitID:    op.CommitID,
		Branch:      op.Branch,
		Tag:         op.Tag,
		CreatedAt:   pbtypes.NewTimestamp(time.Now()),
		BuildConfig: op.Config,
	}

	if repo.Private {
		b.Priority += 20
	}

	// If this is the first ever build for a repo, give it a high priority
	hasBuild, err := s.List(ctx, &sourcegraph.BuildListOptions{
		Repo: repo.ID,
		ListOptions: sourcegraph.ListOptions{
			PerPage: 1,
		},
	})
	if err == nil && len(hasBuild.Builds) == 0 {
		b.Priority += 20
	}

	b, err = store.BuildsFromContext(ctx).Create(ctx, b)
	if err != nil {
		return nil, err
	}

	observeNewBuild(repo.URI)

	return b, nil
}

func (s *builds) Update(ctx context.Context, op *sourcegraph.BuildsUpdateOp) (*sourcegraph.Build, error) {
	b, err := store.BuildsFromContext(ctx).Get(ctx, op.Build)
	if err != nil {
		return nil, err
	}

	var finished bool
	info := op.Info
	if info.StartedAt != nil {
		b.StartedAt = info.StartedAt
	}
	if info.EndedAt != nil {
		// TODO(keegancsmith) This is some temporary logging to see if
		// we are double updating finished builds.
		if b.EndedAt == nil {
			finished = true
		} else {
			log15.Debug("Builds.Update called on a finished build", "build", b, "op", op)
		}
		b.EndedAt = info.EndedAt
	}
	if info.HeartbeatAt != nil {
		b.HeartbeatAt = info.HeartbeatAt
	}
	if info.Host != "" {
		b.Host = info.Host
	}
	if info.Purged {
		b.Purged = info.Purged
	}
	if info.Success {
		b.Success = info.Success
	}
	if info.Failure {
		b.Failure = info.Failure
	}
	if info.Priority != 0 {
		b.Priority = info.Priority
	}
	if info.Killed {
		b.Killed = info.Killed
	}
	if info.BuilderConfig != "" {
		b.BuilderConfig = info.BuilderConfig
	}

	if err := store.BuildsFromContext(ctx).Update(ctx, b.Spec(), info); err != nil {
		return nil, err
	}

	if finished {
		repo, err := svc.Repos(ctx).Get(ctx, &sourcegraph.RepoSpec{ID: b.Repo})
		if err != nil {
			return nil, err
		}
		observeFinishedBuild(b, repo.URI)
	}

	var Result string
	if b.Success {
		Result = "success"
	} else if b.Failure {
		Result = "failed"
	}
	if Result != "" {
		eventsutil.LogBuildCompleted(ctx, b.Success)
	}

	return b, nil
}

func (s *builds) ListBuildTasks(ctx context.Context, op *sourcegraph.BuildsListBuildTasksOp) (*sourcegraph.BuildTaskList, error) {
	tasks, err := store.BuildsFromContext(ctx).ListBuildTasks(ctx, op.Build, op.Opt)
	if err != nil {
		return nil, err
	}
	return &sourcegraph.BuildTaskList{BuildTasks: tasks}, nil
}

func (s *builds) CreateTasks(ctx context.Context, op *sourcegraph.BuildsCreateTasksOp) (*sourcegraph.BuildTaskList, error) {
	// Validate.
	buildSpec := op.Build
	tasks := op.Tasks
	for _, task := range tasks {
		if task.Build != (sourcegraph.BuildSpec{}) && task.Build != buildSpec {
			return nil, fmt.Errorf("task build (%s) does not match build (%s)", task.Build.IDString(), buildSpec.IDString())
		}
	}

	tasks2 := make([]*sourcegraph.BuildTask, len(tasks)) // copy to avoid mutating
	for i, taskPtr := range tasks {
		task := *taskPtr
		task.CreatedAt = pbtypes.NewTimestamp(time.Now())
		task.Build = buildSpec
		tasks2[i] = &task
	}

	created, err := store.BuildsFromContext(ctx).CreateTasks(ctx, tasks2)
	if err != nil {
		return nil, err
	}
	return &sourcegraph.BuildTaskList{BuildTasks: created}, nil
}

func (s *builds) UpdateTask(ctx context.Context, op *sourcegraph.BuildsUpdateTaskOp) (*sourcegraph.BuildTask, error) {
	t, err := store.BuildsFromContext(ctx).GetTask(ctx, op.Task)
	if err != nil {
		return nil, err
	}

	info := op.Info
	if info.StartedAt != nil {
		t.StartedAt = info.StartedAt
	}
	if info.EndedAt != nil {
		t.EndedAt = info.EndedAt
	}
	if info.Success {
		t.Success = true
	}
	if info.Failure {
		t.Failure = true
	}

	if err := store.BuildsFromContext(ctx).UpdateTask(ctx, op.Task, info); err != nil {
		return nil, err
	}

	return t, nil
}

// GetTaskLog gets the logs for a task.
//
// The build is fetched using the task's key (IDString) and its
// StartedAt/EndedAt fields are used to set the start/end times for
// the log entry search, which speeds up the operation significantly
// for the Papertrail backend.
func (s *builds) GetTaskLog(ctx context.Context, op *sourcegraph.BuildsGetTaskLogOp) (*sourcegraph.LogEntries, error) {
	task := op.Task
	opt := op.Opt

	if opt == nil {
		opt = &sourcegraph.BuildGetLogOptions{}
	}

	var minID string
	var minTime, maxTime time.Time

	build, err := store.BuildsFromContext(ctx).Get(ctx, task.Build)
	if err != nil {
		return nil, err
	}

	if opt.MinID == "" {
		const timeBuffer = 120 * time.Second // in case clocks are off
		if build.StartedAt != nil {
			minTime = build.StartedAt.Time().Add(-1 * timeBuffer)
		}
		if build.EndedAt != nil {
			maxTime = build.EndedAt.Time().Add(timeBuffer)
		}
	} else {
		minID = opt.MinID
	}

	return store.BuildLogsFromContext(ctx).Get(ctx, task, minID, minTime, maxTime)
}

func (s *builds) DequeueNext(ctx context.Context, op *sourcegraph.BuildsDequeueNextOp) (*sourcegraph.BuildJob, error) {
	nextBuild, repoPath, err := store.BuildsFromContext(ctx).DequeueNext(ctx)
	if err != nil {
		return nil, err
	}
	if nextBuild == nil {
		return nil, grpc.Errorf(codes.NotFound, "build queue is empty")
	}
	observeDequeuedBuild(repoPath)
	return nextBuild, nil
}

func observeNewBuild(repo string) {
	labels := prometheus.Labels{"repo": repotrackutil.GetTrackedRepo(repo)}
	buildsCreate.With(labels).Inc()
}

func observeFinishedBuild(b *sourcegraph.Build, repo string) {
	// increment a counter labeled with status
	// "beat" a heartbeat
	var state string
	switch {
	case b.Success:
		state = "success"
	case b.Failure:
		state = "failure"
	case b.Killed:
		state = "killed"
	default:
		state = "unknown"
	}
	duration := b.EndedAt.Time().Sub(b.StartedAt.Time())
	labels := prometheus.Labels{
		"state": state,
		"repo":  repotrackutil.GetTrackedRepo(repo),
	}
	buildsDuration.With(labels).Observe(duration.Seconds())
	buildsHeartbeat.With(labels).Set(float64(time.Now().Unix()))
}

func observeDequeuedBuild(repo string) {
	labels := prometheus.Labels{"repo": repotrackutil.GetTrackedRepo(repo)}
	buildsDequeue.With(labels).Inc()
}

var metricLabels = []string{"state", "repo"}
var buildsDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "src",
	Subsystem: "builds",
	Name:      "duration_seconds",
	Help:      "The builds latencies in seconds.",
	Buckets:   statsutil.UserLatencyBuckets,
}, metricLabels)
var buildsHeartbeat = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "src",
	Subsystem: "builds",
	Name:      "last_timestamp_unixtime",
	Help:      "Last time a build finished.",
}, metricLabels)
var buildsCreate = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "builds",
	Name:      "create_total",
	Help:      "Number of builds created.",
}, []string{"repo"})
var buildsDequeue = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: "src",
	Subsystem: "builds",
	Name:      "dequeue_total",
	Help:      "Number of builds dequeued.",
}, []string{"repo"})

func init() {
	prometheus.MustRegister(buildsDequeue)
	prometheus.MustRegister(buildsCreate)
	prometheus.MustRegister(buildsDuration)
	prometheus.MustRegister(buildsHeartbeat)
}
