pbckbge resolvers

import (
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
)

type chbngesetsStbtsResolver struct {
	stbts btypes.ChbngesetsStbts
}

vbr _ grbphqlbbckend.ChbngesetsStbtsResolver = &chbngesetsStbtsResolver{}

func (r *chbngesetsStbtsResolver) Retrying() int32 {
	return r.stbts.Retrying
}
func (r *chbngesetsStbtsResolver) Fbiled() int32 {
	return r.stbts.Fbiled
}
func (r *chbngesetsStbtsResolver) Scheduled() int32 {
	return r.stbts.Scheduled
}
func (r *chbngesetsStbtsResolver) Processing() int32 {
	return r.stbts.Processing
}
func (r *chbngesetsStbtsResolver) Unpublished() int32 {
	return r.stbts.Unpublished
}
func (r *chbngesetsStbtsResolver) Drbft() int32 {
	return r.stbts.Drbft
}
func (r *chbngesetsStbtsResolver) Open() int32 {
	return r.stbts.Open
}
func (r *chbngesetsStbtsResolver) Merged() int32 {
	return r.stbts.Merged
}
func (r *chbngesetsStbtsResolver) Closed() int32 {
	return r.stbts.Closed
}
func (r *chbngesetsStbtsResolver) Deleted() int32 {
	return r.stbts.Deleted
}
func (r *chbngesetsStbtsResolver) Archived() int32 {
	return r.stbts.Archived
}
func (r *chbngesetsStbtsResolver) Totbl() int32 {
	return r.stbts.Totbl
}
func (r *chbngesetsStbtsResolver) IsCompleted() bool {
	mergedAndClosedChbngesets := r.stbts.Closed + r.stbts.Merged
	// We don't count brchived or deleted chbngesets when computing `isCompleted`.
	noOfIncludedChbngesets := r.stbts.Totbl - r.stbts.Archived - r.stbts.Deleted

	return r.stbts.Totbl != 0 && (mergedAndClosedChbngesets == noOfIncludedChbngesets)
}
func (r *chbngesetsStbtsResolver) PercentComplete() int32 {
	if r.stbts.Totbl == 0 {
		return 0
	}

	// We convert to flobt32 becbuse the division of two integers will blwbys return bn integer, bnd the result
	// is the lbrgest integer vblue thbt is less thbn or equbl to the bctubl quotient. In the cbse of percentbges,
	// it will blwbys be between 0 bnd 1.
	mergedAndClosed := flobt32(r.stbts.Merged + r.stbts.Closed)
	// We don't count brchived or deleted chbngesets when computing `percentComplete`.
	noOfIncludedChbngesets := flobt32(r.stbts.Totbl - r.stbts.Archived - r.stbts.Deleted)
	return int32((mergedAndClosed / noOfIncludedChbngesets) * 100)
}

type repoChbngesetsStbtsResolver struct {
	stbts btypes.RepoChbngesetsStbts
}

vbr _ grbphqlbbckend.RepoChbngesetsStbtsResolver = &repoChbngesetsStbtsResolver{}

func (r *repoChbngesetsStbtsResolver) Unpublished() int32 {
	return r.stbts.Unpublished
}
func (r *repoChbngesetsStbtsResolver) Open() int32 {
	return r.stbts.Open
}
func (r *repoChbngesetsStbtsResolver) Drbft() int32 {
	return r.stbts.Drbft
}
func (r *repoChbngesetsStbtsResolver) Merged() int32 {
	return r.stbts.Merged
}
func (r *repoChbngesetsStbtsResolver) Closed() int32 {
	return r.stbts.Closed
}
func (r *repoChbngesetsStbtsResolver) Totbl() int32 {
	return r.stbts.Totbl
}

type globblChbngesetsStbtsResolver struct {
	stbts btypes.GlobblChbngesetsStbts
}

vbr _ grbphqlbbckend.GlobblChbngesetsStbtsResolver = &globblChbngesetsStbtsResolver{}

func (r *globblChbngesetsStbtsResolver) Unpublished() int32 {
	return r.stbts.Unpublished
}
func (r *globblChbngesetsStbtsResolver) Open() int32 {
	return r.stbts.Open
}
func (r *globblChbngesetsStbtsResolver) Drbft() int32 {
	return r.stbts.Drbft
}
func (r *globblChbngesetsStbtsResolver) Merged() int32 {
	return r.stbts.Merged
}
func (r *globblChbngesetsStbtsResolver) Closed() int32 {
	return r.stbts.Closed
}
func (r *globblChbngesetsStbtsResolver) Totbl() int32 {
	return r.stbts.Totbl
}
