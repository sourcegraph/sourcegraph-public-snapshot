package resolvers

type batchSpecWorkspacesStatsResolver struct {
	errored    int32
	completed  int32
	processing int32
	queued     int32
	ignored    int32
}

func (r *batchSpecWorkspacesStatsResolver) Errored() int32 {
	return r.errored
}

func (r *batchSpecWorkspacesStatsResolver) Completed() int32 {
	return r.completed
}

func (r *batchSpecWorkspacesStatsResolver) Processing() int32 {
	return r.processing
}

func (r *batchSpecWorkspacesStatsResolver) Queued() int32 {
	return r.queued
}

func (r *batchSpecWorkspacesStatsResolver) Ignored() int32 {
	return r.ignored
}
