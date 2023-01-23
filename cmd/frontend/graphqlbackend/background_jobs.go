package graphqlbackend

import (
	"context"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/goroutine/recorder"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

// dayCountForStats is hard-coded for now. This signifies the number of days to use for generating the stats for each routine.
const dayCountForStats = 7

// defaultRecentRunCount signifies the default number of recent runs to return for each routine.
const defaultRecentRunCount = 5

type backgroundJobsArgs struct {
	First          *int32
	After          *string
	RecentRunCount *int32
}

type BackgroundJobResolver struct {
	jobInfo recorder.JobInfo
}

type RoutineResolver struct {
	routine recorder.RoutineInfo
}

type RoutineInstanceResolver struct {
	instance recorder.RoutineInstanceInfo
}

type RoutineRecentRunResolver struct {
	recentRun recorder.RoutineRun
}

type RoutineStatsResolver struct {
	stats recorder.RoutineRunStats
}

// backgroundJobConnectionResolver resolves a list of access tokens.
//
// 🚨 SECURITY: When instantiating a backgroundJobConnectionResolver value, the caller MUST check
// permissions.
type backgroundJobConnectionResolver struct {
	first          *int32
	after          string
	recentRunCount *int32

	// cache results because they are used by multiple fields
	once      sync.Once
	resolvers []*BackgroundJobResolver
	err       error
}

func (r *schemaResolver) BackgroundJobs(ctx context.Context, args *backgroundJobsArgs) (*backgroundJobConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins may list background jobs.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	// Parse `after` argument
	var after string
	if args.After != nil {
		err := relay.UnmarshalSpec(graphql.ID(*args.After), &after)
		if err != nil {
			return nil, err
		}
	}

	return &backgroundJobConnectionResolver{
		first:          args.First,
		after:          after,
		recentRunCount: args.RecentRunCount,
	}, nil
}

func (r *schemaResolver) backgroundJobByID(ctx context.Context, id graphql.ID) (*BackgroundJobResolver, error) {
	// 🚨 SECURITY: Only site admins may view background jobs.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	var jobName string
	err := relay.UnmarshalSpec(id, &jobName)
	if err != nil {
		return nil, err
	}
	item, err := recorder.GetBackgroundJobInfo(recorder.GetCache(), jobName, defaultRecentRunCount, dayCountForStats)
	if err != nil {
		return nil, err
	}
	return &BackgroundJobResolver{jobInfo: item}, nil
}

func (r *backgroundJobConnectionResolver) Nodes(context.Context) ([]*BackgroundJobResolver, error) {
	resolvers, err := r.compute()
	if err != nil {
		return nil, err
	}

	if r.first != nil && *r.first > -1 && len(resolvers) > int(*r.first) {
		resolvers = resolvers[:*r.first]
	}

	return resolvers, nil
}

func (r *backgroundJobConnectionResolver) TotalCount(context.Context) (int32, error) {
	resolvers, err := r.compute()
	if err != nil {
		return 0, err
	}
	return int32(len(resolvers)), nil
}

func (r *backgroundJobConnectionResolver) PageInfo(context.Context) (*graphqlutil.PageInfo, error) {
	resolvers, err := r.compute()
	if err != nil {
		return nil, err
	}

	if r.first != nil && *r.first > -1 && len(resolvers) > int(*r.first) {
		return graphqlutil.NextPageCursor(string(resolvers[*r.first-1].ID())), nil
	}
	return graphqlutil.HasNextPage(false), nil
}

func (r *backgroundJobConnectionResolver) compute() ([]*BackgroundJobResolver, error) {
	recentRunCount := defaultRecentRunCount
	if r.recentRunCount != nil {
		recentRunCount = int(*r.recentRunCount)
	}
	r.once.Do(func() {
		jobInfos, err := recorder.GetBackgroundJobInfos(recorder.GetCache(), r.after, recentRunCount, dayCountForStats)
		if err != nil {
			r.resolvers, r.err = nil, err
			return
		}

		resolvers := make([]*BackgroundJobResolver, 0, len(jobInfos))
		for _, jobInfo := range jobInfos {
			resolvers = append(resolvers, &BackgroundJobResolver{jobInfo: jobInfo})
		}

		r.resolvers, r.err = resolvers, nil
	})
	return r.resolvers, r.err
}

func (r *BackgroundJobResolver) ID() graphql.ID {
	return relay.MarshalID("BackgroundJob", r.jobInfo.ID)
}

func (r *BackgroundJobResolver) Name() string { return r.jobInfo.Name }

func (r *BackgroundJobResolver) Routines() []*RoutineResolver {
	resolvers := make([]*RoutineResolver, 0, len(r.jobInfo.Routines))
	for _, routine := range r.jobInfo.Routines {
		resolvers = append(resolvers, &RoutineResolver{routine: routine})
	}
	return resolvers
}

func (r *RoutineResolver) Name() string { return r.routine.Name }

func (r *RoutineResolver) Type() recorder.RoutineType { return r.routine.Type }

func (r *RoutineResolver) Description() string { return r.routine.Description }

func (r *RoutineResolver) IntervalMs() *int32 {
	if r.routine.IntervalMs == 0 {
		return nil
	}
	return &r.routine.IntervalMs
}

func (r *RoutineResolver) Instances() []*RoutineInstanceResolver {
	resolvers := make([]*RoutineInstanceResolver, 0, len(r.routine.Instances))
	for _, routineInstance := range r.routine.Instances {
		resolvers = append(resolvers, &RoutineInstanceResolver{instance: routineInstance})
	}
	return resolvers
}

func (r *RoutineInstanceResolver) HostName() string { return r.instance.HostName }

func (r *RoutineInstanceResolver) LastStartedAt() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(r.instance.LastStartedAt)
}

func (r *RoutineInstanceResolver) LastStoppedAt() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(r.instance.LastStoppedAt)
}

func (r *RoutineResolver) RecentRuns() []*RoutineRecentRunResolver {
	resolvers := make([]*RoutineRecentRunResolver, 0, len(r.routine.RecentRuns))
	for _, recentRun := range r.routine.RecentRuns {
		resolvers = append(resolvers, &RoutineRecentRunResolver{recentRun: recentRun})
	}
	return resolvers
}

func (r *RoutineRecentRunResolver) At() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.recentRun.At}
}

func (r *RoutineRecentRunResolver) HostName() string { return r.recentRun.HostName }

func (r *RoutineRecentRunResolver) DurationMs() int32 { return r.recentRun.DurationMs }

func (r *RoutineRecentRunResolver) ErrorMessage() *string {
	if r.recentRun.ErrorMessage == "" {
		return nil
	}
	return &r.recentRun.ErrorMessage
}

func (r *RoutineResolver) Stats() *RoutineStatsResolver {
	return &RoutineStatsResolver{stats: r.routine.Stats}
}

func (r *RoutineStatsResolver) Since() *gqlutil.DateTime {
	if r.stats.Since.IsZero() {
		return nil
	}
	return gqlutil.DateTimeOrNil(&r.stats.Since)
}

func (r *RoutineStatsResolver) RunCount() int32 { return r.stats.RunCount }

func (r *RoutineStatsResolver) ErrorCount() int32 { return r.stats.ErrorCount }

func (r *RoutineStatsResolver) MinDurationMs() int32 { return r.stats.MinDurationMs }

func (r *RoutineStatsResolver) AvgDurationMs() int32 { return r.stats.AvgDurationMs }

func (r *RoutineStatsResolver) MaxDurationMs() int32 { return r.stats.MaxDurationMs }
