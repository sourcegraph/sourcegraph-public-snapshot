package resolvers

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
)

type changesetsStatsResolver struct {
	stats btypes.ChangesetsStats
}

var _ graphqlbackend.ChangesetsStatsResolver = &changesetsStatsResolver{}

func (r *changesetsStatsResolver) Retrying() int32 {
	return r.stats.Retrying
}
func (r *changesetsStatsResolver) Failed() int32 {
	return r.stats.Failed
}
func (r *changesetsStatsResolver) Scheduled() int32 {
	return r.stats.Scheduled
}
func (r *changesetsStatsResolver) Processing() int32 {
	return r.stats.Processing
}
func (r *changesetsStatsResolver) Unpublished() int32 {
	return r.stats.Unpublished
}
func (r *changesetsStatsResolver) Draft() int32 {
	return r.stats.Draft
}
func (r *changesetsStatsResolver) Open() int32 {
	return r.stats.Open
}
func (r *changesetsStatsResolver) Merged() int32 {
	return r.stats.Merged
}
func (r *changesetsStatsResolver) Closed() int32 {
	return r.stats.Closed
}
func (r *changesetsStatsResolver) Deleted() int32 {
	return r.stats.Deleted
}
func (r *changesetsStatsResolver) Archived() int32 {
	return r.stats.Archived
}
func (r *changesetsStatsResolver) Total() int32 {
	return r.stats.Total
}
func (r *changesetsStatsResolver) IsCompleted() bool {
	mergedAndClosedChangesets := r.stats.Closed + r.stats.Merged
	// We don't count archived or deleted changesets when computing `isCompleted`.
	noOfIncludedChangesets := r.stats.Total - r.stats.Archived - r.stats.Deleted

	return r.stats.Total != 0 && (mergedAndClosedChangesets == noOfIncludedChangesets)
}
func (r *changesetsStatsResolver) PercentComplete() int32 {
	if r.stats.Total == 0 {
		return 0
	}

	// We convert to float32 because the division of two integers will always return an integer, and the result
	// is the largest integer value that is less than or equal to the actual quotient. In the case of percentages,
	// it will always be between 0 and 1.
	mergedAndClosed := float32(r.stats.Merged + r.stats.Closed)
	// We don't count archived or deleted changesets when computing `percentComplete`.
	noOfIncludedChangesets := float32(r.stats.Total - r.stats.Archived - r.stats.Deleted)

	// To prevent an zero division, we check if the denominator is 0, if it is we return early.
	// Usually this happens when all the changesets belonging to a batch change are archived or deleted.
	if noOfIncludedChangesets == 0 {
		return 100
	}
	return int32((mergedAndClosed / noOfIncludedChangesets) * 100)
}

type repoChangesetsStatsResolver struct {
	stats btypes.RepoChangesetsStats
}

var _ graphqlbackend.RepoChangesetsStatsResolver = &repoChangesetsStatsResolver{}

func (r *repoChangesetsStatsResolver) Unpublished() int32 {
	return r.stats.Unpublished
}
func (r *repoChangesetsStatsResolver) Open() int32 {
	return r.stats.Open
}
func (r *repoChangesetsStatsResolver) Draft() int32 {
	return r.stats.Draft
}
func (r *repoChangesetsStatsResolver) Merged() int32 {
	return r.stats.Merged
}
func (r *repoChangesetsStatsResolver) Closed() int32 {
	return r.stats.Closed
}
func (r *repoChangesetsStatsResolver) Total() int32 {
	return r.stats.Total
}

type globalChangesetsStatsResolver struct {
	stats btypes.GlobalChangesetsStats
}

var _ graphqlbackend.GlobalChangesetsStatsResolver = &globalChangesetsStatsResolver{}

func (r *globalChangesetsStatsResolver) Unpublished() int32 {
	return r.stats.Unpublished
}
func (r *globalChangesetsStatsResolver) Open() int32 {
	return r.stats.Open
}
func (r *globalChangesetsStatsResolver) Draft() int32 {
	return r.stats.Draft
}
func (r *globalChangesetsStatsResolver) Merged() int32 {
	return r.stats.Merged
}
func (r *globalChangesetsStatsResolver) Closed() int32 {
	return r.stats.Closed
}
func (r *globalChangesetsStatsResolver) Total() int32 {
	return r.stats.Total
}
