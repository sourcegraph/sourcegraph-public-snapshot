package db

import (
	"context"

	qdrant "github.com/qdrant/go-client/qdrant"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func ensureModelCollectionWithDefaultConfig(ctx context.Context, db *qdrantDB, modelID string, modelDims uint64) error {
	// Make the actual collection end with `.default` so we can switch between
	// configurations with aliases.
	name := CollectionName(modelID)
	realName := name + ".default"

	err := ensureCollection(ctx, db.collectionsClient, realName, defaultConfig(modelDims))
	if err != nil {
		return err
	}

	// Update the alias atomically to point to the new collection
	_, err = db.collectionsClient.UpdateAliases(ctx, &qdrant.ChangeAliases{
		Actions: []*qdrant.AliasOperations{{
			Action: &qdrant.AliasOperations_CreateAlias{
				CreateAlias: &qdrant.CreateAlias{
					CollectionName: realName,
					AliasName:      name,
				},
			},
		}},
	})
	if err != nil {
		return errors.Wrap(err, "update aliases")
	}

	err = ensureRepoIDIndex(ctx, db.pointsClient, realName)
	if err != nil {
		return errors.Wrap(err, "add repo index")
	}

	return nil
}

func ensureCollection(ctx context.Context, cc qdrant.CollectionsClient, name string, config *qdrant.CollectionConfig) error {
	resp, err := cc.List(ctx, &qdrant.ListCollectionsRequest{})
	if err != nil {
		return err
	}

	for _, collection := range resp.GetCollections() {
		if collection.GetName() == name {
			// Collection already exists
			return nil
		}
	}

	// Create a new collection with the new config using the data of the old collection
	_, err = cc.Create(ctx, &qdrant.CreateCollection{
		CollectionName:         name,
		HnswConfig:             config.HnswConfig,
		WalConfig:              config.WalConfig,
		OptimizersConfig:       config.OptimizerConfig,
		ShardNumber:            &config.Params.ShardNumber,
		OnDiskPayload:          &config.Params.OnDiskPayload,
		VectorsConfig:          config.Params.VectorsConfig,
		ReplicationFactor:      config.Params.ReplicationFactor,
		WriteConsistencyFactor: config.Params.WriteConsistencyFactor,
		InitFromCollection:     nil,
		QuantizationConfig:     config.QuantizationConfig,
	})

	return err
}

func ensureRepoIDIndex(ctx context.Context, cc qdrant.PointsClient, name string) error {
	// This is idempotent, so no need to check if it exists first
	_, err := cc.CreateFieldIndex(ctx, &qdrant.CreateFieldIndexCollection{
		CollectionName:   name,
		Wait:             pointers.Ptr(true),
		FieldName:        fieldRepoID,
		FieldType:        pointers.Ptr(qdrant.FieldType_FieldTypeKeyword),
		FieldIndexParams: nil,
		Ordering:         nil,
	})
	if err != nil {
		return err
	}

	return nil
}

func defaultConfig(dims uint64) *qdrant.CollectionConfig {
	return &qdrant.CollectionConfig{
		Params: &qdrant.CollectionParams{
			ShardNumber:   1,
			OnDiskPayload: true,
			VectorsConfig: &qdrant.VectorsConfig{
				Config: &qdrant.VectorsConfig_Params{
					Params: &qdrant.VectorParams{
						Size:               dims,
						Distance:           qdrant.Distance_Cosine,
						HnswConfig:         nil,                // use collection default
						QuantizationConfig: nil,                // use collection default
						OnDisk:             pointers.Ptr(true), // use collection default
					},
				},
			},
			ReplicationFactor:      nil, // default
			WriteConsistencyFactor: nil, // default
		},
		OptimizerConfig: &qdrant.OptimizersConfigDiff{
			IndexingThreshold: pointers.Ptr(uint64(0)), // disable indexing
		},
		WalConfig: nil, // default
		QuantizationConfig: &qdrant.QuantizationConfig{
			// scalar is faster than product, but doesn't compress as well
			Quantization: &qdrant.QuantizationConfig_Scalar{
				Scalar: &qdrant.ScalarQuantization{
					Type: qdrant.QuantizationType_Int8,
					// Truncate outliers for better compression
					Quantile:  pointers.Ptr(float32(0.98)),
					AlwaysRam: nil, // default false
				},
			},
		},
	}
}
