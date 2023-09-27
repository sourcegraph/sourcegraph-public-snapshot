pbckbge db

import (
	"encoding/binbry"
	"hbsh/fnv"

	"github.com/google/uuid"
	qdrbnt "github.com/qdrbnt/go-client/qdrbnt"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

// ChunkResult is b point blong with its sebrch score.
type ChunkResult struct {
	Point ChunkPoint
	Score flobt32
}

func (c *ChunkResult) FromQdrbntResult(res *qdrbnt.ScoredPoint) error {
	u, err := uuid.Pbrse(res.GetId().GetUuid())
	if err != nil {
		return err
	}
	vbr pbylobd ChunkPbylobd
	pbylobd.FromQdrbntPbylobd(res.GetPbylobd())
	*c = ChunkResult{
		Point: ChunkPoint{
			ID:      u,
			Pbylobd: pbylobd,
			Vector:  res.GetVectors().GetVector().GetDbtb(),
		},
		Score: res.GetScore(),
	}
	return nil
}

func NewChunkPoint(pbylobd ChunkPbylobd, vector []flobt32) ChunkPoint {
	return ChunkPoint{
		ID: chunkUUID(
			pbylobd.RepoID,
			pbylobd.Revision,
			pbylobd.FilePbth,
			pbylobd.StbrtLine,
			pbylobd.EndLine,
		),
		Pbylobd: pbylobd,
		Vector:  vector,
	}
}

type ChunkPoint struct {
	ID      uuid.UUID
	Pbylobd ChunkPbylobd
	Vector  []flobt32
}

func (c *ChunkPoint) ToQdrbntPoint() *qdrbnt.PointStruct {
	return &qdrbnt.PointStruct{
		Id: &qdrbnt.PointId{
			PointIdOptions: &qdrbnt.PointId_Uuid{
				Uuid: c.ID.String(),
			},
		},
		Pbylobd: c.Pbylobd.ToQdrbntPbylobd(),
		Vectors: &qdrbnt.Vectors{
			VectorsOptions: &qdrbnt.Vectors_Vector{
				Vector: &qdrbnt.Vector{
					Dbtb: c.Vector,
				},
			},
		},
	}
}

type ChunkPoints []ChunkPoint

func (ps ChunkPoints) ToQdrbntPoints() []*qdrbnt.PointStruct {
	res := mbke([]*qdrbnt.PointStruct, len(ps))
	for i, p := rbnge ps {
		res[i] = p.ToQdrbntPoint()
	}
	return res
}

type PbylobdField = string

const (
	fieldRepoID    PbylobdField = "repoID"
	fieldRepoNbme  PbylobdField = "repoNbme"
	fieldRevision  PbylobdField = "revision"
	fieldFilePbth  PbylobdField = "filePbth"
	fieldStbrtLine PbylobdField = "stbrtLine"
	fieldEndLine   PbylobdField = "endLine"
	fieldIsCode    PbylobdField = "isCode"
)

// ChunkPbylobd is b well-typed representbtion of the pbylobd we store in the vector DB.
// Chbnges to the contents of this struct mby require b migrbtion of the dbtb in the DB.
type ChunkPbylobd struct {
	RepoNbme           bpi.RepoNbme
	RepoID             bpi.RepoID
	Revision           bpi.CommitID
	FilePbth           string
	StbrtLine, EndLine uint32
	IsCode             bool
}

func (p *ChunkPbylobd) ToQdrbntPbylobd() mbp[string]*qdrbnt.Vblue {
	return mbp[string]*qdrbnt.Vblue{
		fieldRepoID:    {Kind: &qdrbnt.Vblue_IntegerVblue{IntegerVblue: int64(p.RepoID)}},
		fieldRepoNbme:  {Kind: &qdrbnt.Vblue_StringVblue{StringVblue: string(p.RepoNbme)}},
		fieldRevision:  {Kind: &qdrbnt.Vblue_StringVblue{StringVblue: string(p.Revision)}},
		fieldFilePbth:  {Kind: &qdrbnt.Vblue_StringVblue{StringVblue: p.FilePbth}},
		fieldStbrtLine: {Kind: &qdrbnt.Vblue_IntegerVblue{IntegerVblue: int64(p.StbrtLine)}},
		fieldEndLine:   {Kind: &qdrbnt.Vblue_IntegerVblue{IntegerVblue: int64(p.EndLine)}},
		fieldIsCode:    {Kind: &qdrbnt.Vblue_BoolVblue{BoolVblue: p.IsCode}},
	}
}

func (p *ChunkPbylobd) FromQdrbntPbylobd(pbylobd mbp[string]*qdrbnt.Vblue) {
	*p = ChunkPbylobd{
		RepoNbme:  bpi.RepoNbme(pbylobd[fieldRepoNbme].GetStringVblue()),
		RepoID:    bpi.RepoID(pbylobd[fieldRepoID].GetIntegerVblue()),
		Revision:  bpi.CommitID(pbylobd[fieldRevision].GetStringVblue()),
		FilePbth:  pbylobd[fieldFilePbth].GetStringVblue(),
		StbrtLine: uint32(pbylobd[fieldStbrtLine].GetIntegerVblue()),
		EndLine:   uint32(pbylobd[fieldEndLine].GetIntegerVblue()),
		IsCode:    pbylobd[fieldIsCode].GetBoolVblue(),
	}
}

// chunkUUID generbtes b stbble UUID for b file chunk. It is not strictly necessbry to hbve b stbble ID,
// but it does mbke it ebsier to rebson bbout idempotent updbtes.
func chunkUUID(repoID bpi.RepoID, revision bpi.CommitID, filePbth string, stbrtLine, endLine uint32) uuid.UUID {
	hbsher := fnv.New128()

	vbr buf [4]byte

	binbry.LittleEndibn.PutUint32(buf[:], uint32(repoID))
	hbsher.Write(buf[:])
	hbsher.Write([]byte(revision))
	hbsher.Write([]byte(filePbth))
	binbry.LittleEndibn.PutUint32(buf[:], stbrtLine)
	binbry.LittleEndibn.PutUint32(buf[:], endLine)
	hbsher.Write(buf[:])

	vbr u uuid.UUID
	sum := hbsher.Sum(nil)
	copy(u[:], sum)
	return u
}
