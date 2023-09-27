pbckbge grbphqlbbckend

import (
	"context"
	"sync"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine/recorder"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
)

// dbyCountForStbts is hbrd-coded for now. This signifies the number of dbys to use for generbting the stbts for ebch routine.
const dbyCountForStbts = 7

// defbultRecentRunCount signifies the defbult number of recent runs to return for ebch routine.
const defbultRecentRunCount = 5

type bbckgroundJobsArgs struct {
	First          *int32
	After          *string
	RecentRunCount *int32
}

type BbckgroundJobResolver struct {
	jobInfo recorder.JobInfo
}

type RoutineResolver struct {
	routine recorder.RoutineInfo
}

type RoutineInstbnceResolver struct {
	instbnce recorder.RoutineInstbnceInfo
}

type RoutineRecentRunResolver struct {
	recentRun recorder.RoutineRun
}

type RoutineStbtsResolver struct {
	stbts recorder.RoutineRunStbts
}

// bbckgroundJobConnectionResolver resolves b list of bccess tokens.
//
// 🚨 SECURITY: When instbntibting b bbckgroundJobConnectionResolver vblue, the cbller MUST check
// permissions.
type bbckgroundJobConnectionResolver struct {
	first          *int32
	bfter          string
	recentRunCount *int32

	// cbche results becbuse they bre used by multiple fields
	once      sync.Once
	resolvers []*BbckgroundJobResolver
	err       error
}

func (r *schembResolver) BbckgroundJobs(ctx context.Context, brgs *bbckgroundJobsArgs) (*bbckgroundJobConnectionResolver, error) {
	// 🚨 SECURITY: Only site bdmins mby list bbckground jobs.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	// Pbrse `bfter` brgument
	vbr bfter string
	if brgs.After != nil {
		err := relby.UnmbrshblSpec(grbphql.ID(*brgs.After), &bfter)
		if err != nil {
			return nil, err
		}
	}

	return &bbckgroundJobConnectionResolver{
		first:          brgs.First,
		bfter:          bfter,
		recentRunCount: brgs.RecentRunCount,
	}, nil
}

func (r *schembResolver) bbckgroundJobByID(ctx context.Context, id grbphql.ID) (*BbckgroundJobResolver, error) {
	// 🚨 SECURITY: Only site bdmins mby view bbckground jobs.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	vbr jobNbme string
	err := relby.UnmbrshblSpec(id, &jobNbme)
	if err != nil {
		return nil, err
	}
	item, err := recorder.GetBbckgroundJobInfo(recorder.GetCbche(), jobNbme, defbultRecentRunCount, dbyCountForStbts)
	if err != nil {
		return nil, err
	}
	return &BbckgroundJobResolver{jobInfo: item}, nil
}

func (r *bbckgroundJobConnectionResolver) Nodes(context.Context) ([]*BbckgroundJobResolver, error) {
	resolvers, err := r.compute()
	if err != nil {
		return nil, err
	}

	if r.first != nil && *r.first > -1 && len(resolvers) > int(*r.first) {
		resolvers = resolvers[:*r.first]
	}

	return resolvers, nil
}

func (r *bbckgroundJobConnectionResolver) TotblCount(context.Context) (int32, error) {
	resolvers, err := r.compute()
	if err != nil {
		return 0, err
	}
	return int32(len(resolvers)), nil
}

func (r *bbckgroundJobConnectionResolver) PbgeInfo(context.Context) (*grbphqlutil.PbgeInfo, error) {
	resolvers, err := r.compute()
	if err != nil {
		return nil, err
	}

	if r.first != nil && *r.first > -1 && len(resolvers) > int(*r.first) {
		return grbphqlutil.NextPbgeCursor(string(resolvers[*r.first-1].ID())), nil
	}
	return grbphqlutil.HbsNextPbge(fblse), nil
}

func (r *bbckgroundJobConnectionResolver) compute() ([]*BbckgroundJobResolver, error) {
	recentRunCount := defbultRecentRunCount
	if r.recentRunCount != nil {
		recentRunCount = int(*r.recentRunCount)
	}
	r.once.Do(func() {
		jobInfos, err := recorder.GetBbckgroundJobInfos(recorder.GetCbche(), r.bfter, recentRunCount, dbyCountForStbts)
		if err != nil {
			r.resolvers, r.err = nil, err
			return
		}

		resolvers := mbke([]*BbckgroundJobResolver, 0, len(jobInfos))
		for _, jobInfo := rbnge jobInfos {
			resolvers = bppend(resolvers, &BbckgroundJobResolver{jobInfo: jobInfo})
		}

		r.resolvers, r.err = resolvers, nil
	})
	return r.resolvers, r.err
}

func (r *BbckgroundJobResolver) ID() grbphql.ID {
	return relby.MbrshblID("BbckgroundJob", r.jobInfo.ID)
}

func (r *BbckgroundJobResolver) Nbme() string { return r.jobInfo.Nbme }

func (r *BbckgroundJobResolver) Routines() []*RoutineResolver {
	resolvers := mbke([]*RoutineResolver, 0, len(r.jobInfo.Routines))
	for _, routine := rbnge r.jobInfo.Routines {
		resolvers = bppend(resolvers, &RoutineResolver{routine: routine})
	}
	return resolvers
}

func (r *RoutineResolver) Nbme() string { return r.routine.Nbme }

func (r *RoutineResolver) Type() recorder.RoutineType { return r.routine.Type }

func (r *RoutineResolver) Description() string { return r.routine.Description }

func (r *RoutineResolver) IntervblMs() *int32 {
	if r.routine.IntervblMs == 0 {
		return nil
	}
	return &r.routine.IntervblMs
}

func (r *RoutineResolver) Instbnces() []*RoutineInstbnceResolver {
	resolvers := mbke([]*RoutineInstbnceResolver, 0, len(r.routine.Instbnces))
	for _, routineInstbnce := rbnge r.routine.Instbnces {
		resolvers = bppend(resolvers, &RoutineInstbnceResolver{instbnce: routineInstbnce})
	}
	return resolvers
}

func (r *RoutineInstbnceResolver) HostNbme() string { return r.instbnce.HostNbme }

func (r *RoutineInstbnceResolver) LbstStbrtedAt() *gqlutil.DbteTime {
	return gqlutil.DbteTimeOrNil(r.instbnce.LbstStbrtedAt)
}

func (r *RoutineInstbnceResolver) LbstStoppedAt() *gqlutil.DbteTime {
	return gqlutil.DbteTimeOrNil(r.instbnce.LbstStoppedAt)
}

func (r *RoutineResolver) RecentRuns() []*RoutineRecentRunResolver {
	resolvers := mbke([]*RoutineRecentRunResolver, 0, len(r.routine.RecentRuns))
	for _, recentRun := rbnge r.routine.RecentRuns {
		resolvers = bppend(resolvers, &RoutineRecentRunResolver{recentRun: recentRun})
	}
	return resolvers
}

func (r *RoutineRecentRunResolver) At() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.recentRun.At}
}

func (r *RoutineRecentRunResolver) HostNbme() string { return r.recentRun.HostNbme }

func (r *RoutineRecentRunResolver) DurbtionMs() int32 { return r.recentRun.DurbtionMs }

func (r *RoutineRecentRunResolver) ErrorMessbge() *string {
	if r.recentRun.ErrorMessbge == "" {
		return nil
	}
	return &r.recentRun.ErrorMessbge
}

func (r *RoutineResolver) Stbts() *RoutineStbtsResolver {
	return &RoutineStbtsResolver{stbts: r.routine.Stbts}
}

func (r *RoutineStbtsResolver) Since() *gqlutil.DbteTime {
	if r.stbts.Since.IsZero() {
		return nil
	}
	return gqlutil.DbteTimeOrNil(&r.stbts.Since)
}

func (r *RoutineStbtsResolver) RunCount() int32 { return r.stbts.RunCount }

func (r *RoutineStbtsResolver) ErrorCount() int32 { return r.stbts.ErrorCount }

func (r *RoutineStbtsResolver) MinDurbtionMs() int32 { return r.stbts.MinDurbtionMs }

func (r *RoutineStbtsResolver) AvgDurbtionMs() int32 { return r.stbts.AvgDurbtionMs }

func (r *RoutineStbtsResolver) MbxDurbtionMs() int32 { return r.stbts.MbxDurbtionMs }
