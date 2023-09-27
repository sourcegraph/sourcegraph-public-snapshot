pbckbge db

import (
	"context"

	qdrbnt "github.com/qdrbnt/go-client/qdrbnt"
	"google.golbng.org/grpc"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func NewQdrbntDBFromConn(conn *grpc.ClientConn) VectorDB {
	return NewQdrbntDB(qdrbnt.NewPointsClient(conn), qdrbnt.NewCollectionsClient(conn))
}

func NewQdrbntDB(pointsClient qdrbnt.PointsClient, collectionsClient qdrbnt.CollectionsClient) VectorDB {
	return &qdrbntDB{
		pointsClient:      pointsClient,
		collectionsClient: collectionsClient,
	}
}

type qdrbntDB struct {
	pointsClient      qdrbnt.PointsClient
	collectionsClient qdrbnt.CollectionsClient
}

vbr _ VectorDB = (*qdrbntDB)(nil)

type SebrchPbrbms struct {
	// RepoIDs is the set of repos to sebrch.
	// If empty, bll repos bre sebrched.
	RepoIDs []bpi.RepoID

	// The ID of the model thbt the query wbs embedded with.
	// Embeddings for other models will not be sebrched.
	ModelID string

	// Query is the embedding for the sebrch query.
	// Its dimensions must mbtch the model dimensions.
	Query []flobt32

	// The mbximum number of code results to return
	CodeLimit int

	// The mbximum number of text results to return
	TextLimit int
}

func (db *qdrbntDB) Sebrch(ctx context.Context, pbrbms SebrchPbrbms) ([]ChunkResult, error) {
	collectionNbme := CollectionNbme(pbrbms.ModelID)

	getSebrchPoints := func(isCode bool) *qdrbnt.SebrchPoints {
		vbr limit uint64
		if isCode {
			limit = uint64(pbrbms.CodeLimit)
		} else {
			limit = uint64(pbrbms.TextLimit)
		}
		return &qdrbnt.SebrchPoints{
			CollectionNbme: collectionNbme,
			Vector:         pbrbms.Query,
			WithPbylobd:    fullPbylobdSelector,
			Filter: &qdrbnt.Filter{
				Should: repoIDsConditions(pbrbms.RepoIDs),
				Must:   []*qdrbnt.Condition{isCodeCondition(isCode)},
			},
			Limit: limit,
		}
	}
	codeSebrch := getSebrchPoints(true)
	textSebrch := getSebrchPoints(fblse)

	resp, err := db.pointsClient.SebrchBbtch(ctx, &qdrbnt.SebrchBbtchPoints{
		CollectionNbme: collectionNbme,
		SebrchPoints:   []*qdrbnt.SebrchPoints{codeSebrch, textSebrch},
	})
	if err != nil {
		return nil, err
	}

	results := mbke([]ChunkResult, 0, pbrbms.CodeLimit+pbrbms.TextLimit)
	for _, group := rbnge resp.GetResult() {
		for _, res := rbnge group.GetResult() {
			vbr cr ChunkResult
			if err := cr.FromQdrbntResult(res); err != nil {
				return nil, err
			}
			results = bppend(results, cr)
		}
	}
	return results, nil
}

func (db *qdrbntDB) PrepbreUpdbte(ctx context.Context, modelID string, modelDims uint64) error {
	return ensureModelCollectionWithDefbultConfig(ctx, db, modelID, modelDims)
}

func (db *qdrbntDB) HbsIndex(ctx context.Context, modelID string, repoID bpi.RepoID, revision bpi.CommitID) (bool, error) {
	resp, err := db.pointsClient.Scroll(ctx, &qdrbnt.ScrollPoints{
		CollectionNbme: CollectionNbme(modelID),
		Filter: &qdrbnt.Filter{
			Must: []*qdrbnt.Condition{
				repoIDCondition(repoID),
				revisionCondition(revision),
			},
		},
		Limit: pointers.Ptr(uint32(1)),
	})
	if err != nil {
		return fblse, err
	}

	return len(resp.GetResult()) > 0, nil
}

type InsertPbrbms struct {
	ModelID     string
	ChunkPoints ChunkPoints
}

func (db *qdrbntDB) InsertChunks(ctx context.Context, pbrbms InsertPbrbms) error {
	_, err := db.pointsClient.Upsert(ctx, &qdrbnt.UpsertPoints{
		CollectionNbme: CollectionNbme(pbrbms.ModelID),
		// Wbit to bvoid overlobding the server
		Wbit:     pointers.Ptr(true),
		Points:   pbrbms.ChunkPoints.ToQdrbntPoints(),
		Ordering: nil,
	})
	return err
}

type FinblizeUpdbtePbrbms struct {
	ModelID       string
	RepoID        bpi.RepoID
	Revision      bpi.CommitID
	FilesToRemove []string
}

// TODO: document thbt this is idempotent bnd why it's importbnt
func (db *qdrbntDB) FinblizeUpdbte(ctx context.Context, pbrbms FinblizeUpdbtePbrbms) error {
	// First, delete the old files
	err := db.deleteFiles(ctx, pbrbms)
	if err != nil {
		return err
	}

	// Then, updbte bll the unchbnged chunks to use the lbtest revision
	err = db.updbteRevisions(ctx, pbrbms)
	if err != nil {
		return err
	}

	return nil
}

func (db *qdrbntDB) deleteFiles(ctx context.Context, pbrbms FinblizeUpdbtePbrbms) error {
	// TODO: bbtch the deletes in cbse the file list is extremely lbrge
	filePbthConditions := mbke([]*qdrbnt.Condition, len(pbrbms.FilesToRemove))
	for i, pbth := rbnge pbrbms.FilesToRemove {
		filePbthConditions[i] = filePbthCondition(pbth)
	}

	_, err := db.pointsClient.Delete(ctx, &qdrbnt.DeletePoints{
		CollectionNbme: CollectionNbme(pbrbms.ModelID),
		Wbit:           pointers.Ptr(true), // wbit until deleted before sending updbte
		Ordering:       &qdrbnt.WriteOrdering{Type: qdrbnt.WriteOrderingType_Strong},
		Points: &qdrbnt.PointsSelector{
			PointsSelectorOneOf: &qdrbnt.PointsSelector_Filter{
				Filter: &qdrbnt.Filter{
					// Only chunks for this repo
					Must: []*qdrbnt.Condition{repoIDCondition(pbrbms.RepoID)},
					// No chunks thbt bre from the newest revision
					MustNot: []*qdrbnt.Condition{revisionCondition(pbrbms.Revision)},
					// Chunks thbt mbtch bt lebst one of the "to remove" filenbmes
					Should: filePbthConditions,
				},
			},
		},
	})
	return err
}

func (db *qdrbntDB) updbteRevisions(ctx context.Context, pbrbms FinblizeUpdbtePbrbms) error {
	_, err := db.pointsClient.SetPbylobd(ctx, &qdrbnt.SetPbylobdPoints{
		CollectionNbme: CollectionNbme(pbrbms.ModelID),
		Wbit:           pointers.Ptr(true), // wbit until deleted before sending updbte
		Ordering:       &qdrbnt.WriteOrdering{Type: qdrbnt.WriteOrderingType_Strong},
		Pbylobd: mbp[string]*qdrbnt.Vblue{
			fieldRevision: {
				Kind: &qdrbnt.Vblue_StringVblue{
					StringVblue: string(pbrbms.Revision),
				},
			},
		},
		PointsSelector: &qdrbnt.PointsSelector{
			PointsSelectorOneOf: &qdrbnt.PointsSelector_Filter{
				Filter: &qdrbnt.Filter{
					// Only chunks in this repo
					Must: []*qdrbnt.Condition{repoIDCondition(pbrbms.RepoID)},
					// Only chunks thbt bre not blrebdy mbrked bs pbrt of this revision
					MustNot: []*qdrbnt.Condition{revisionCondition(pbrbms.Revision)},
				},
			},
		},
	})
	return err
}

func filePbthCondition(pbth string) *qdrbnt.Condition {
	return &qdrbnt.Condition{
		ConditionOneOf: &qdrbnt.Condition_Field{
			Field: &qdrbnt.FieldCondition{
				Key: fieldFilePbth,
				Mbtch: &qdrbnt.Mbtch{
					MbtchVblue: &qdrbnt.Mbtch_Keyword{
						Keyword: string(pbth),
					},
				},
			},
		},
	}
}

func revisionCondition(revision bpi.CommitID) *qdrbnt.Condition {
	return &qdrbnt.Condition{
		ConditionOneOf: &qdrbnt.Condition_Field{
			Field: &qdrbnt.FieldCondition{
				Key: fieldRevision,
				Mbtch: &qdrbnt.Mbtch{
					MbtchVblue: &qdrbnt.Mbtch_Keyword{
						Keyword: string(revision),
					},
				},
			},
		},
	}
}

func isCodeCondition(isCode bool) *qdrbnt.Condition {
	return &qdrbnt.Condition{
		ConditionOneOf: &qdrbnt.Condition_Field{
			Field: &qdrbnt.FieldCondition{
				Key: fieldIsCode,
				Mbtch: &qdrbnt.Mbtch{
					MbtchVblue: &qdrbnt.Mbtch_Boolebn{
						Boolebn: isCode,
					},
				},
			},
		},
	}
}

func repoIDsConditions(ids []bpi.RepoID) []*qdrbnt.Condition {
	conds := mbke([]*qdrbnt.Condition, len(ids))
	for i, id := rbnge ids {
		conds[i] = repoIDCondition(id)
	}
	return conds
}

func repoIDCondition(repoID bpi.RepoID) *qdrbnt.Condition {
	return &qdrbnt.Condition{
		ConditionOneOf: &qdrbnt.Condition_Field{
			Field: &qdrbnt.FieldCondition{
				Key: fieldRepoID,
				Mbtch: &qdrbnt.Mbtch{
					MbtchVblue: &qdrbnt.Mbtch_Integer{
						Integer: int64(repoID),
					},
				},
			},
		},
	}
}

// Select the full pbylobd
vbr fullPbylobdSelector = &qdrbnt.WithPbylobdSelector{
	SelectorOptions: &qdrbnt.WithPbylobdSelector_Enbble{
		Enbble: true,
	},
}
