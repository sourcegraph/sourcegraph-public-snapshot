package db

import (
	"context"

	qdrant "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func NewQdrantDBFromConn(conn *grpc.ClientConn) VectorDB {
	return NewQdrantDB(qdrant.NewPointsClient(conn), qdrant.NewCollectionsClient(conn))
}

func NewQdrantDB(pointsClient qdrant.PointsClient, collectionsClient qdrant.CollectionsClient) VectorDB {
	return &qdrantDB{
		pointsClient:      pointsClient,
		collectionsClient: collectionsClient,
	}
}

type qdrantDB struct {
	pointsClient      qdrant.PointsClient
	collectionsClient qdrant.CollectionsClient
}

var _ VectorDB = (*qdrantDB)(nil)

type SearchParams struct {
	// RepoIDs is the set of repos to search.
	// If empty, all repos are searched.
	RepoIDs []api.RepoID

	// The ID of the model that the query was embedded with.
	// Embeddings for other models will not be searched.
	ModelID string

	// Query is the embedding for the search query.
	// Its dimensions must match the model dimensions.
	Query []float32

	// The maximum number of code results to return
	CodeLimit int

	// The maximum number of text results to return
	TextLimit int
}

func (db *qdrantDB) Search(ctx context.Context, params SearchParams) ([]ChunkResult, error) {
	collectionName := CollectionName(params.ModelID)

	getSearchPoints := func(isCode bool) *qdrant.SearchPoints {
		var limit uint64
		if isCode {
			limit = uint64(params.CodeLimit)
		} else {
			limit = uint64(params.TextLimit)
		}
		return &qdrant.SearchPoints{
			CollectionName: collectionName,
			Vector:         params.Query,
			WithPayload:    fullPayloadSelector,
			Filter: &qdrant.Filter{
				Should: repoIDsConditions(params.RepoIDs),
				Must:   []*qdrant.Condition{isCodeCondition(isCode)},
			},
			Limit: limit,
		}
	}
	codeSearch := getSearchPoints(true)
	textSearch := getSearchPoints(false)

	resp, err := db.pointsClient.SearchBatch(ctx, &qdrant.SearchBatchPoints{
		CollectionName: collectionName,
		SearchPoints:   []*qdrant.SearchPoints{codeSearch, textSearch},
	})
	if err != nil {
		return nil, err
	}

	results := make([]ChunkResult, 0, params.CodeLimit+params.TextLimit)
	for _, group := range resp.GetResult() {
		for _, res := range group.GetResult() {
			var cr ChunkResult
			if err := cr.FromQdrantResult(res); err != nil {
				return nil, err
			}
			results = append(results, cr)
		}
	}
	return results, nil
}

func (db *qdrantDB) PrepareUpdate(ctx context.Context, modelID string, modelDims uint64) error {
	return ensureModelCollection(ctx, db, modelID, modelDims)
}

func (db *qdrantDB) HasIndex(ctx context.Context, modelID string, repoID api.RepoID, revision api.CommitID) (bool, error) {
	resp, err := db.pointsClient.Scroll(ctx, &qdrant.ScrollPoints{
		CollectionName: CollectionName(modelID),
		Filter: &qdrant.Filter{
			Must: []*qdrant.Condition{
				repoIDCondition(repoID),
				revisionCondition(revision),
			},
		},
		Limit: pointers.Ptr(uint32(1)),
	})
	if err != nil {
		return false, err
	}

	return len(resp.GetResult()) > 0, nil
}

type InsertParams struct {
	ModelID     string
	ChunkPoints ChunkPoints
}

func (db *qdrantDB) InsertChunks(ctx context.Context, params InsertParams) error {
	_, err := db.pointsClient.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: CollectionName(params.ModelID),
		// Wait to avoid overloading the server
		Wait:     pointers.Ptr(true),
		Points:   params.ChunkPoints.ToQdrantPoints(),
		Ordering: nil,
	})
	return err
}

type FinalizeUpdateParams struct {
	ModelID       string
	RepoID        api.RepoID
	Revision      api.CommitID
	FilesToRemove []string
}

// TODO: document that this is idempotent and why it's important
func (db *qdrantDB) FinalizeUpdate(ctx context.Context, params FinalizeUpdateParams) error {
	// First, delete the old files
	err := db.deleteFiles(ctx, params)
	if err != nil {
		return err
	}

	// Then, update all the unchanged chunks to use the latest revision
	err = db.updateRevisions(ctx, params)
	if err != nil {
		return err
	}

	return nil
}

func (db *qdrantDB) deleteFiles(ctx context.Context, params FinalizeUpdateParams) error {
	// TODO: batch the deletes in case the file list is extremely large
	filePathConditions := make([]*qdrant.Condition, len(params.FilesToRemove))
	for i, path := range params.FilesToRemove {
		filePathConditions[i] = filePathCondition(path)
	}

	_, err := db.pointsClient.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: CollectionName(params.ModelID),
		Wait:           pointers.Ptr(true), // wait until deleted before sending update
		Ordering:       &qdrant.WriteOrdering{Type: qdrant.WriteOrderingType_Strong},
		Points: &qdrant.PointsSelector{
			PointsSelectorOneOf: &qdrant.PointsSelector_Filter{
				Filter: &qdrant.Filter{
					// Only chunks for this repo
					Must: []*qdrant.Condition{repoIDCondition(params.RepoID)},
					// No chunks that are from the newest revision
					MustNot: []*qdrant.Condition{revisionCondition(params.Revision)},
					// Chunks that match at least one of the "to remove" filenames
					Should: filePathConditions,
				},
			},
		},
	})
	return err
}

func (db *qdrantDB) updateRevisions(ctx context.Context, params FinalizeUpdateParams) error {
	_, err := db.pointsClient.SetPayload(ctx, &qdrant.SetPayloadPoints{
		CollectionName: CollectionName(params.ModelID),
		Wait:           pointers.Ptr(true), // wait until deleted before sending update
		Ordering:       &qdrant.WriteOrdering{Type: qdrant.WriteOrderingType_Strong},
		Payload: map[string]*qdrant.Value{
			fieldRevision: {
				Kind: &qdrant.Value_StringValue{
					StringValue: string(params.Revision),
				},
			},
		},
		PointsSelector: &qdrant.PointsSelector{
			PointsSelectorOneOf: &qdrant.PointsSelector_Filter{
				Filter: &qdrant.Filter{
					// Only chunks in this repo
					Must: []*qdrant.Condition{repoIDCondition(params.RepoID)},
					// Only chunks that are not already marked as part of this revision
					MustNot: []*qdrant.Condition{revisionCondition(params.Revision)},
				},
			},
		},
	})
	return err
}

func filePathCondition(path string) *qdrant.Condition {
	return &qdrant.Condition{
		ConditionOneOf: &qdrant.Condition_Field{
			Field: &qdrant.FieldCondition{
				Key: fieldFilePath,
				Match: &qdrant.Match{
					MatchValue: &qdrant.Match_Keyword{
						Keyword: string(path),
					},
				},
			},
		},
	}
}

func revisionCondition(revision api.CommitID) *qdrant.Condition {
	return &qdrant.Condition{
		ConditionOneOf: &qdrant.Condition_Field{
			Field: &qdrant.FieldCondition{
				Key: fieldRevision,
				Match: &qdrant.Match{
					MatchValue: &qdrant.Match_Keyword{
						Keyword: string(revision),
					},
				},
			},
		},
	}
}

func isCodeCondition(isCode bool) *qdrant.Condition {
	return &qdrant.Condition{
		ConditionOneOf: &qdrant.Condition_Field{
			Field: &qdrant.FieldCondition{
				Key: fieldIsCode,
				Match: &qdrant.Match{
					MatchValue: &qdrant.Match_Boolean{
						Boolean: isCode,
					},
				},
			},
		},
	}
}

func repoIDsConditions(ids []api.RepoID) []*qdrant.Condition {
	conds := make([]*qdrant.Condition, len(ids))
	for i, id := range ids {
		conds[i] = repoIDCondition(id)
	}
	return conds
}

func repoIDCondition(repoID api.RepoID) *qdrant.Condition {
	return &qdrant.Condition{
		ConditionOneOf: &qdrant.Condition_Field{
			Field: &qdrant.FieldCondition{
				Key: fieldRepoID,
				Match: &qdrant.Match{
					MatchValue: &qdrant.Match_Integer{
						Integer: int64(repoID),
					},
				},
			},
		},
	}
}

// Select the full payload
var fullPayloadSelector = &qdrant.WithPayloadSelector{
	SelectorOptions: &qdrant.WithPayloadSelector_Enable{
		Enable: true,
	},
}
