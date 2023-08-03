package db

import (
	"encoding/binary"
	"hash/fnv"

	"github.com/google/uuid"
	qdrant "github.com/qdrant/go-client/qdrant"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

// ChunkResult is a point along with its search score.
type ChunkResult struct {
	Point ChunkPoint
	Score float32
}

func (c *ChunkResult) FromQdrantResult(res *qdrant.ScoredPoint) error {
	u, err := uuid.Parse(res.GetId().GetUuid())
	if err != nil {
		return err
	}
	var payload ChunkPayload
	payload.FromQdrantPayload(res.GetPayload())
	*c = ChunkResult{
		Point: ChunkPoint{
			ID:      u,
			Payload: payload,
			Vector:  res.GetVectors().GetVector().GetData(),
		},
		Score: res.GetScore(),
	}
	return nil
}

func NewChunkPoint(payload ChunkPayload, vector []float32) ChunkPoint {
	return ChunkPoint{
		ID: chunkUUID(
			payload.RepoID,
			payload.Revision,
			payload.FilePath,
			payload.StartLine,
			payload.EndLine,
		),
		Payload: payload,
		Vector:  vector,
	}
}

type ChunkPoint struct {
	ID      uuid.UUID
	Payload ChunkPayload
	Vector  []float32
}

func (c *ChunkPoint) ToQdrantPoint() *qdrant.PointStruct {
	return &qdrant.PointStruct{
		Id: &qdrant.PointId{
			PointIdOptions: &qdrant.PointId_Uuid{
				Uuid: c.ID.String(),
			},
		},
		Payload: c.Payload.ToQdrantPayload(),
		Vectors: &qdrant.Vectors{
			VectorsOptions: &qdrant.Vectors_Vector{
				Vector: &qdrant.Vector{
					Data: c.Vector,
				},
			},
		},
	}
}

type ChunkPoints []ChunkPoint

func (ps ChunkPoints) ToQdrantPoints() []*qdrant.PointStruct {
	res := make([]*qdrant.PointStruct, len(ps))
	for i, p := range ps {
		res[i] = p.ToQdrantPoint()
	}
	return res
}

type PayloadField = string

const (
	fieldRepoID    PayloadField = "repoID"
	fieldRepoName  PayloadField = "repoName"
	fieldRevision  PayloadField = "revision"
	fieldFilePath  PayloadField = "filePath"
	fieldStartLine PayloadField = "startLine"
	fieldEndLine   PayloadField = "endLine"
	fieldIsCode    PayloadField = "isCode"
)

// ChunkPayload is a well-typed representation of the payload we store in the vector DB.
// Changes to the contents of this struct may require a migration of the data in the DB.
type ChunkPayload struct {
	RepoName           api.RepoName
	RepoID             api.RepoID
	Revision           api.CommitID
	FilePath           string
	StartLine, EndLine uint32
	IsCode             bool
}

func (p *ChunkPayload) ToQdrantPayload() map[string]*qdrant.Value {
	return map[string]*qdrant.Value{
		fieldRepoID:    {Kind: &qdrant.Value_IntegerValue{IntegerValue: int64(p.RepoID)}},
		fieldRepoName:  {Kind: &qdrant.Value_StringValue{StringValue: string(p.RepoName)}},
		fieldRevision:  {Kind: &qdrant.Value_StringValue{StringValue: string(p.Revision)}},
		fieldFilePath:  {Kind: &qdrant.Value_StringValue{StringValue: p.FilePath}},
		fieldStartLine: {Kind: &qdrant.Value_IntegerValue{IntegerValue: int64(p.StartLine)}},
		fieldEndLine:   {Kind: &qdrant.Value_IntegerValue{IntegerValue: int64(p.EndLine)}},
		fieldIsCode:    {Kind: &qdrant.Value_BoolValue{BoolValue: p.IsCode}},
	}
}

func (p *ChunkPayload) FromQdrantPayload(payload map[string]*qdrant.Value) {
	*p = ChunkPayload{
		RepoName:  api.RepoName(payload[fieldRepoName].GetStringValue()),
		RepoID:    api.RepoID(payload[fieldRepoID].GetIntegerValue()),
		Revision:  api.CommitID(payload[fieldRevision].GetStringValue()),
		FilePath:  payload[fieldFilePath].GetStringValue(),
		StartLine: uint32(payload[fieldStartLine].GetIntegerValue()),
		EndLine:   uint32(payload[fieldEndLine].GetIntegerValue()),
		IsCode:    payload[fieldIsCode].GetBoolValue(),
	}
}

// chunkUUID generates a stable UUID for a file chunk. It is not strictly necessary to have a stable ID,
// but it does make it easier to reason about idempotent updates.
func chunkUUID(repoID api.RepoID, revision api.CommitID, filePath string, startLine, endLine uint32) uuid.UUID {
	hasher := fnv.New128()

	var buf [4]byte

	binary.LittleEndian.PutUint32(buf[:], uint32(repoID))
	hasher.Write(buf[:])
	hasher.Write([]byte(revision))
	hasher.Write([]byte(filePath))
	binary.LittleEndian.PutUint32(buf[:], startLine)
	binary.LittleEndian.PutUint32(buf[:], endLine)
	hasher.Write(buf[:])

	var u uuid.UUID
	sum := hasher.Sum(nil)
	copy(u[:], sum)
	return u
}
