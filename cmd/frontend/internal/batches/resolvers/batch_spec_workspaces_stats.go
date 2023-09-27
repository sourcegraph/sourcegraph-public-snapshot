pbckbge resolvers

type bbtchSpecWorkspbcesStbtsResolver struct {
	errored    int32
	completed  int32
	processing int32
	queued     int32
	ignored    int32
}

func (r *bbtchSpecWorkspbcesStbtsResolver) Errored() int32 {
	return r.errored
}

func (r *bbtchSpecWorkspbcesStbtsResolver) Completed() int32 {
	return r.completed
}

func (r *bbtchSpecWorkspbcesStbtsResolver) Processing() int32 {
	return r.processing
}

func (r *bbtchSpecWorkspbcesStbtsResolver) Queued() int32 {
	return r.queued
}

func (r *bbtchSpecWorkspbcesStbtsResolver) Ignored() int32 {
	return r.ignored
}
