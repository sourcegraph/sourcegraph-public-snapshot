package store

import (
	"time"

	"sort"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
)

type Builds interface {
	Get(ctx context.Context, build sourcegraph.BuildSpec) (*sourcegraph.Build, error)
	List(ctx context.Context, opt *sourcegraph.BuildListOptions) ([]*sourcegraph.Build, error)
	Create(context.Context, *sourcegraph.Build) (*sourcegraph.Build, error)
	Update(ctx context.Context, build sourcegraph.BuildSpec, info sourcegraph.BuildUpdate) error
	ListBuildTasks(ctx context.Context, build sourcegraph.BuildSpec, opt *sourcegraph.BuildTaskListOptions) ([]*sourcegraph.BuildTask, error)
	CreateTasks(ctx context.Context, tasks []*sourcegraph.BuildTask) ([]*sourcegraph.BuildTask, error)
	UpdateTask(ctx context.Context, task sourcegraph.TaskSpec, info sourcegraph.TaskUpdate) error
	DequeueNext(ctx context.Context) (*sourcegraph.BuildJob, error)
	GetTask(ctx context.Context, task sourcegraph.TaskSpec) (*sourcegraph.BuildTask, error)
}

type BuildLogs interface {
	Get(ctx context.Context, task sourcegraph.TaskSpec, minID string, minTime, maxTime time.Time) (*sourcegraph.LogEntries, error)
}

// SortAndPaginateBuilds sorts and paginates a list of builds
// according to opt, returning the result. It is allowed to modify the "builds" argument.
func SortAndPaginateBuilds(builds []*sourcegraph.Build, opt *sourcegraph.BuildListOptions) []*sourcegraph.Build {
	if opt != nil {
		// Sort.
		v := buildsSorter{builds: builds}
		switch opt.Sort {
		case "updated_at":
			v.less = func(a, b *sourcegraph.Build) bool {
				return buildUpdatedAt(a).Before(buildUpdatedAt(b))
			}
		case "started_at":
			v.less = func(a, b *sourcegraph.Build) bool {
				return a.StartedAt != nil && (b.StartedAt == nil || a.StartedAt.Time().Before(b.StartedAt.Time()))
			}
		case "priority":
			v.less = func(a, b *sourcegraph.Build) bool {
				return a.Priority < b.Priority
			}
		default:
			v.less = func(a, b *sourcegraph.Build) bool {
				return a.CreatedAt.Time().Before(b.CreatedAt.Time())
			}
		}
		if opt.Direction == "desc" {
			sort.Sort(sort.Reverse(v))
		} else {
			sort.Sort(v)
		}

		// Paginate.
		offset, limit := opt.ListOptions.Offset(), opt.ListOptions.Limit()
		if offset > len(builds) {
			offset = len(builds)
		}
		builds = builds[offset:]
		if len(builds) > limit {
			builds = builds[:limit]
		}
	}

	return builds
}

type buildsSorter struct {
	builds []*sourcegraph.Build
	less   func(a, b *sourcegraph.Build) bool
}

func (bs buildsSorter) Len() int           { return len(bs.builds) }
func (bs buildsSorter) Swap(i, j int)      { bs.builds[i], bs.builds[j] = bs.builds[j], bs.builds[i] }
func (bs buildsSorter) Less(i, j int) bool { return bs.less(bs.builds[i], bs.builds[j]) }

// buildUpdatedAt returns the most recent time that the build was updated.
func buildUpdatedAt(b *sourcegraph.Build) time.Time {
	return newestTime(&b.CreatedAt, b.StartedAt, b.EndedAt)
}

func SortAndPaginateTasks(tasks []*sourcegraph.BuildTask, opt *sourcegraph.BuildTaskListOptions) []*sourcegraph.BuildTask {
	if opt == nil {
		opt = &sourcegraph.BuildTaskListOptions{}
	}

	// Sort.
	v := tasksSorter{tasks: tasks}
	v.less = func(a, b *sourcegraph.BuildTask) bool {
		return a.ID < b.ID
	}
	sort.Sort(v)

	// Paginate.
	offset, limit := opt.ListOptions.Offset(), opt.ListOptions.Limit()
	if offset > len(tasks) {
		offset = len(tasks)
	}
	tasks = tasks[offset:]
	if len(tasks) > limit {
		tasks = tasks[:limit]
	}

	return tasks
}

type tasksSorter struct {
	tasks []*sourcegraph.BuildTask
	less  func(a, b *sourcegraph.BuildTask) bool
}

func (bs tasksSorter) Len() int           { return len(bs.tasks) }
func (bs tasksSorter) Swap(i, j int)      { bs.tasks[i], bs.tasks[j] = bs.tasks[j], bs.tasks[i] }
func (bs tasksSorter) Less(i, j int) bool { return bs.less(bs.tasks[i], bs.tasks[j]) }

// newestTime returns the newest time among all of the times
// specified. If all times are zero or nil, time.Time{} is returned.
func newestTime(pbtimes ...*pbtypes.Timestamp) time.Time {
	var newest time.Time
	for _, t := range pbtimes {
		if t != nil && t.Time().After(newest) {
			newest = t.Time()
		}
	}
	return newest
}
