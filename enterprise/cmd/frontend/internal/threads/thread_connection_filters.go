package threads

import (
	"context"
	"sort"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func newThreadConnectionFiltersFromDB(ctx context.Context, opt dbThreadsListOptions) (graphqlbackend.ThreadConnectionFilters, error) {
	// For the purpose of determining filters, ignore most existing filters. (E.g., we still want to
	// show the open threads count when the user has "is:closed" in the thread query.)
	opt.LimitOffset = nil
	opt.States = nil
	// TODO!(sqs): probably also zero out repositories...we need to run multiple queries to determine the right counts here
	//
	// TODO!(sqs) security respect repo perms
	allThreads, err := dbThreads{}.List(ctx, opt)
	if err != nil {
		return nil, err
	}
	return &threadConnectionFilters{ToThreadOrThreadPreviews(toThreads(allThreads), nil)}, nil
}

func newThreadConnectionFiltersFromConst(allThreads []graphqlbackend.ToThreadOrThreadPreview) graphqlbackend.ThreadConnectionFilters {
	return &threadConnectionFilters{allThreads: allThreads}
}

type threadConnectionFilters struct {
	allThreads []graphqlbackend.ToThreadOrThreadPreview
}

func (f *threadConnectionFilters) Repository(ctx context.Context) ([]graphqlbackend.RepositoryFilter, error) {
	// TODO!(sqs) security respect repo perms
	repos := map[api.RepoID]int32{}
	for _, t := range f.allThreads {
		repos[t.Common().Internal_RepositoryID()]++
	}

	filters := make([]graphqlbackend.RepositoryFilter, 0, len(repos))
	for repoID, count := range repos {
		repo, err := graphqlbackend.RepositoryByDBID(ctx, repoID)
		if err != nil {
			return nil, err
		}
		filters = append(filters, graphqlbackend.RepositoryFilter{Repository_: repo, Count_: count})
	}
	sort.Slice(filters, func(i, j int) bool { return filters[i].Count_ > filters[j].Count_ })
	return filters, nil
}

func (f *threadConnectionFilters) Label(ctx context.Context) ([]graphqlbackend.LabelFilter, error) {
	// TODO!(sqs) security respect label perms
	var (
		labelForName       = map[string]graphqlbackend.Label{}
		labelNameConflicts = map[string]struct{}{}
		labelCounts        = map[string]int32{}
	)
	for _, t := range f.allThreads {
		labelConnection, err := t.Common().Labels(ctx, &graphqlutil.ConnectionArgs{})
		if err != nil {
			return nil, err
		}
		labels, err := labelConnection.Nodes(ctx)
		if err != nil {
			return nil, err
		}
		for _, label := range labels {
			name := label.Name()
			if _, conflict := labelForName[name]; conflict {
				labelNameConflicts[name] = struct{}{}
			} else if _, conflict := labelNameConflicts[name]; !conflict {
				labelForName[name] = label
			}
			labelCounts[name]++
		}
	}

	filters := make([]graphqlbackend.LabelFilter, 0, len(labelCounts))
	for labelName, count := range labelCounts {
		filters = append(filters, graphqlbackend.LabelFilter{
			Label_:     labelForName[labelName],
			LabelName_: labelName,
			Count_:     count,
		})
	}
	sort.Slice(filters, func(i, j int) bool { return filters[i].Count_ > filters[j].Count_ })
	return filters, nil
}

func (f *threadConnectionFilters) count(filter func(graphqlbackend.ToThreadOrThreadPreview) bool) int32 {
	var count int32
	for _, thread := range f.allThreads {
		if filter(thread) {
			count++
		}
	}
	return count
}

func (f *threadConnectionFilters) OpenCount(ctx context.Context) (int32, error) {
	return f.count(func(thread graphqlbackend.ToThreadOrThreadPreview) bool {
		return thread.ThreadPreview != nil || thread.Thread.State() == graphqlbackend.ThreadStateOpen
	}), nil
}

func (f *threadConnectionFilters) ClosedCount(ctx context.Context) (int32, error) {
	return f.count(func(thread graphqlbackend.ToThreadOrThreadPreview) bool {
		return thread.Thread != nil &&
			(thread.Thread.State() == graphqlbackend.ThreadStateMerged || thread.Thread.State() == graphqlbackend.ThreadStateClosed)
	}), nil
}
