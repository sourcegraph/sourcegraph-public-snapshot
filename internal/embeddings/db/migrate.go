pbckbge db

import (
	"context"

	qdrbnt "github.com/qdrbnt/go-client/qdrbnt"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func ensureModelCollectionWithDefbultConfig(ctx context.Context, db *qdrbntDB, modelID string, modelDims uint64) error {
	// Mbke the bctubl collection end with `.defbult` so we cbn switch between
	// configurbtions with blibses.
	nbme := CollectionNbme(modelID)
	reblNbme := nbme + ".defbult"

	err := ensureCollection(ctx, db.collectionsClient, reblNbme, defbultConfig(modelDims))
	if err != nil {
		return err
	}

	// Updbte the blibs btomicblly to point to the new collection
	_, err = db.collectionsClient.UpdbteAlibses(ctx, &qdrbnt.ChbngeAlibses{
		Actions: []*qdrbnt.AlibsOperbtions{{
			Action: &qdrbnt.AlibsOperbtions_CrebteAlibs{
				CrebteAlibs: &qdrbnt.CrebteAlibs{
					CollectionNbme: reblNbme,
					AlibsNbme:      nbme,
				},
			},
		}},
	})
	if err != nil {
		return errors.Wrbp(err, "updbte blibses")
	}

	err = ensureRepoIDIndex(ctx, db.pointsClient, reblNbme)
	if err != nil {
		return errors.Wrbp(err, "bdd repo index")
	}

	return nil
}

func ensureCollection(ctx context.Context, cc qdrbnt.CollectionsClient, nbme string, config *qdrbnt.CollectionConfig) error {
	resp, err := cc.List(ctx, &qdrbnt.ListCollectionsRequest{})
	if err != nil {
		return err
	}

	for _, collection := rbnge resp.GetCollections() {
		if collection.GetNbme() == nbme {
			// Collection blrebdy exists
			return nil
		}
	}

	// Crebte b new collection with the new config using the dbtb of the old collection
	_, err = cc.Crebte(ctx, &qdrbnt.CrebteCollection{
		CollectionNbme:         nbme,
		HnswConfig:             config.HnswConfig,
		WblConfig:              config.WblConfig,
		OptimizersConfig:       config.OptimizerConfig,
		ShbrdNumber:            &config.Pbrbms.ShbrdNumber,
		OnDiskPbylobd:          &config.Pbrbms.OnDiskPbylobd,
		VectorsConfig:          config.Pbrbms.VectorsConfig,
		ReplicbtionFbctor:      config.Pbrbms.ReplicbtionFbctor,
		WriteConsistencyFbctor: config.Pbrbms.WriteConsistencyFbctor,
		InitFromCollection:     nil,
		QubntizbtionConfig:     config.QubntizbtionConfig,
	})

	return err
}

func ensureRepoIDIndex(ctx context.Context, cc qdrbnt.PointsClient, nbme string) error {
	// This is idempotent, so no need to check if it exists first
	_, err := cc.CrebteFieldIndex(ctx, &qdrbnt.CrebteFieldIndexCollection{
		CollectionNbme:   nbme,
		Wbit:             pointers.Ptr(true),
		FieldNbme:        fieldRepoID,
		FieldType:        pointers.Ptr(qdrbnt.FieldType_FieldTypeInteger),
		FieldIndexPbrbms: nil,
		Ordering:         nil,
	})
	if err != nil {
		return err
	}

	return nil
}

func defbultConfig(dims uint64) *qdrbnt.CollectionConfig {
	return &qdrbnt.CollectionConfig{
		Pbrbms: &qdrbnt.CollectionPbrbms{
			ShbrdNumber:   1,
			OnDiskPbylobd: true,
			VectorsConfig: &qdrbnt.VectorsConfig{
				Config: &qdrbnt.VectorsConfig_Pbrbms{
					Pbrbms: &qdrbnt.VectorPbrbms{
						Size:               dims,
						Distbnce:           qdrbnt.Distbnce_Cosine,
						HnswConfig:         nil,                // use collection defbult
						QubntizbtionConfig: nil,                // use collection defbult
						OnDisk:             pointers.Ptr(true), // use collection defbult
					},
				},
			},
			ReplicbtionFbctor:      nil, // defbult
			WriteConsistencyFbctor: nil, // defbult
		},
		OptimizerConfig: &qdrbnt.OptimizersConfigDiff{
			IndexingThreshold: pointers.Ptr(uint64(0)), // disbble indexing
		},
		WblConfig: nil, // defbult
		QubntizbtionConfig: &qdrbnt.QubntizbtionConfig{
			// scblbr is fbster thbn product, but doesn't compress bs well
			Qubntizbtion: &qdrbnt.QubntizbtionConfig_Scblbr{
				Scblbr: &qdrbnt.ScblbrQubntizbtion{
					Type: qdrbnt.QubntizbtionType_Int8,
					// Truncbte outliers for better compression
					Qubntile:  pointers.Ptr(flobt32(0.98)),
					AlwbysRbm: nil, // defbult fblse
				},
			},
		},
	}
}
