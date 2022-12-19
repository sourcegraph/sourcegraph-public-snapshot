package graphqlbackend

import (
	"context"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// This is hard-coded for now. This signifies the number of days to use for generating the stats for each routine.
const dayCountForStats = 7

type backgroundJobsArgs struct {
	First          *int32
	After          *string
	RecentRunCount *int32
}

type BackgroundJobResolver struct {
	jobInfo *types.BackgroundJobInfo
}

type RoutineResolver struct {
	routine *types.BackgroundRoutineInfo
}

type RoutineInstanceResolver struct {
	instance *types.BackgroundRoutineInstanceInfo
}

type RoutineRecentRunResolver struct {
	recentRun *types.BackgroundRoutineRun
}

type RoutineRecentRunErrorResolver struct {
	recentRun *types.BackgroundRoutineRun
}

type RoutineStatsResolver struct {
	stats *types.BackgroundRoutineRunStats
}

// backgroundJobConnectionResolver resolves a list of access tokens.
//
// 🚨 SECURITY: When instantiating a backgroundJobConnectionResolver value, the caller MUST check
// permissions.
type backgroundJobConnectionResolver struct {
	first            *int32
	after            string
	recentRunCount   *int32
	dayCountForStats *int32

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
	} else {
		after = ""
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

	recentRunCount := 5 // Magic value for now

	var jobName string
	err := relay.UnmarshalSpec(id, jobName)
	if err != nil {
		return nil, err
	}
	item, err := goroutine.GetBackgroundJobInfo(goroutine.GetMonitorCache(), jobName, int32(recentRunCount), dayCountForStats)
	if err != nil {
		return nil, err
	}
	return &BackgroundJobResolver{jobInfo: &item}, nil
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
	r.once.Do(func() {
		jobInfos, err := goroutine.GetBackgroundJobInfos(goroutine.GetMonitorCache(), r.after, *r.recentRunCount, dayCountForStats)
		if err != nil {
			r.resolvers, r.err = nil, err
		}

		resolvers := make([]*BackgroundJobResolver, 0, len(jobInfos))
		for _, jobInfo := range jobInfos {
			resolvers = append(resolvers, &BackgroundJobResolver{jobInfo: &jobInfo})
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
		resolvers = append(resolvers, &RoutineResolver{routine: &routine})
	}
	return resolvers
}

func (r *RoutineResolver) Name() string { return r.routine.Name }

func (r *RoutineResolver) Type() types.BackgroundRoutineType { return r.routine.Type }

func (r *RoutineResolver) Description() string { return r.routine.Description }

func (r *RoutineResolver) Instances() []*RoutineInstanceResolver {
	resolvers := make([]*RoutineInstanceResolver, 0, len(r.routine.Instances))
	for _, routineInstance := range r.routine.Instances {
		resolvers = append(resolvers, &RoutineInstanceResolver{instance: &routineInstance})
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
		resolvers = append(resolvers, &RoutineRecentRunResolver{recentRun: &recentRun})
	}
	return resolvers
}

func (r *RoutineRecentRunResolver) At() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.recentRun.At}
}

func (r *RoutineRecentRunResolver) HostName() string { return r.recentRun.HostName }

func (r *RoutineRecentRunResolver) DurationMs() int32 { return r.recentRun.DurationMs }

func (r *RoutineRecentRunResolver) Error() *RoutineRecentRunErrorResolver {
	if r.recentRun.ErrorMessage == "" {
		return nil
	}
	return &RoutineRecentRunErrorResolver{recentRun: r.recentRun}
}

func (r *RoutineRecentRunErrorResolver) Message() string { return r.recentRun.ErrorMessage }

func (r *RoutineRecentRunErrorResolver) StackTrace() string { return r.recentRun.StackTrace }

func (r *RoutineResolver) Stats() *RoutineStatsResolver {
	return &RoutineStatsResolver{stats: &r.routine.Stats}
}

func (r *RoutineStatsResolver) RunCount() int32 { return r.stats.Count }

func (r *RoutineStatsResolver) ErrorCount() int32 { return r.stats.ErrorCount }

func (r *RoutineStatsResolver) MinDurationMs() int32 { return r.stats.MinDurationMs }

func (r *RoutineStatsResolver) AvgDurationMs() int32 { return r.stats.AvgDurationMs }

func (r *RoutineStatsResolver) MaxDurationMs() int32 { return r.stats.MaxDurationMs }
