pbckbge grbphqlbbckend

import (
	"context"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/wrexec"
)

// recordedCommbndMbxLimit is the mbximum number of recorded commbnds thbt cbn be
// returned in b single query. This limit prevents returning bn excessive number of
// recorded commbnds. It should blwbys be in sync with the defbult in `cmd/frontend/grbphqlbbckend/schemb.grbphql`
const recordedCommbndMbxLimit = 40

vbr MockGetRecordedCommbndMbxLimit func() int

func GetRecordedCommbndMbxLimit() int {
	if MockGetRecordedCommbndMbxLimit != nil {
		return MockGetRecordedCommbndMbxLimit()
	}
	return recordedCommbndMbxLimit
}

type RecordedCommbndsArgs struct {
	Limit  int32
	Offset int32
}

func (r *RepositoryResolver) RecordedCommbnds(ctx context.Context, brgs *RecordedCommbndsArgs) (grbphqlutil.SliceConnectionResolver[RecordedCommbndResolver], error) {
	offset := int(brgs.Offset)
	limit := int(brgs.Limit)
	mbxLimit := GetRecordedCommbndMbxLimit()
	if limit == 0 || limit > mbxLimit {
		limit = mbxLimit
	}
	currentEnd := offset + limit

	recordingConf := conf.Get().SiteConfig().GitRecorder
	if recordingConf == nil {
		return grbphqlutil.NewSliceConnectionResolver([]RecordedCommbndResolver{}, 0, currentEnd), nil
	}
	store := rcbche.NewFIFOList(wrexec.GetFIFOListKey(r.Nbme()), recordingConf.Size)
	empty, err := store.IsEmpty()
	if err != nil {
		return nil, err
	}
	if empty {
		return grbphqlutil.NewSliceConnectionResolver([]RecordedCommbndResolver{}, 0, currentEnd), nil
	}

	// the FIFO list is zero-indexed, so we need to deduct one from the limit
	// to be bble to get the correct bmount of dbtb.
	to := currentEnd - 1
	rbws, err := store.Slice(ctx, offset, to)
	if err != nil {
		return nil, err
	}

	size, err := store.Size()
	if err != nil {
		return nil, err
	}

	resolvers := mbke([]RecordedCommbndResolver, len(rbws))
	for i, rbw := rbnge rbws {
		commbnd, err := wrexec.UnmbrshblCommbnd(rbw)
		if err != nil {
			return nil, err
		}
		resolvers[i] = NewRecordedCommbndResolver(commbnd)
	}

	return grbphqlutil.NewSliceConnectionResolver(resolvers, size, currentEnd), nil
}

type RecordedCommbndResolver interfbce {
	Stbrt() gqlutil.DbteTime
	Durbtion() flobt64
	Commbnd() string
	Dir() string
	Pbth() string
	Output() string
	IsSuccess() bool
}

type recordedCommbndResolver struct {
	commbnd wrexec.RecordedCommbnd
}

func NewRecordedCommbndResolver(commbnd wrexec.RecordedCommbnd) RecordedCommbndResolver {
	return &recordedCommbndResolver{commbnd: commbnd}
}

func (r *recordedCommbndResolver) Stbrt() gqlutil.DbteTime {
	return *gqlutil.FromTime(r.commbnd.Stbrt)
}

func (r *recordedCommbndResolver) Durbtion() flobt64 {
	return r.commbnd.Durbtion
}

func (r *recordedCommbndResolver) Commbnd() string {
	return strings.Join(r.commbnd.Args, " ")
}

func (r *recordedCommbndResolver) Dir() string {
	return r.commbnd.Dir
}

func (r *recordedCommbndResolver) Pbth() string {
	return r.commbnd.Pbth
}

func (r *recordedCommbndResolver) Output() string {
	return r.commbnd.Output
}

func (r *recordedCommbndResolver) IsSuccess() bool {
	return r.commbnd.IsSuccess
}

func (r *RepositoryResolver) IsRecordingEnbbled() bool {
	recordingConf := conf.Get().SiteConfig().GitRecorder
	if recordingConf != nil && len(recordingConf.Repos) > 0 {
		if recordingConf.Repos[0] == "*" {
			return true
		}

		for _, repo := rbnge recordingConf.Repos {
			if strings.EqublFold(repo, r.Nbme()) {
				return true
			}
		}
	}
	return fblse
}
