package db

import (
	"context"

	qdrant "github.com/qdrant/go-client/qdrant"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"
)

func ensureModelCollection(ctx context.Context, db *qdrantDB, modelID string, modelDims uint64) error {
	// Make the actual collection end with `.default` so we can switch between
	// configurations with aliases.
	name := CollectionName(modelID)
	realName := name + ".default"

	err := ensureCollectionWithConfig(ctx, db.collectionsClient, realName, modelDims, conf.Get().Embeddings.QdrantConfig)
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

func ensureCollectionWithConfig(ctx context.Context, cc qdrant.CollectionsClient, name string, dims uint64, conf *schema.QdrantConfig) error {
	resp, err := cc.List(ctx, &qdrant.ListCollectionsRequest{})
	if err != nil {
		return err
	}

	for _, collection := range resp.GetCollections() {
		if collection.GetName() == name {
			return updateCollectionConfig(ctx, cc, name, conf)
		}
	}

	return createCollection(ctx, cc, name, dims, conf)
}

func updateCollectionConfig(ctx context.Context, cc qdrant.CollectionsClient, name string, conf *schema.QdrantConfig) error {
	_, err := cc.Update(ctx, &qdrant.UpdateCollection{
		CollectionName:     name,
		HnswConfig:         getHnswConfigDiff(conf),
		OptimizersConfig:   getOptimizersConfigDiff(conf),
		QuantizationConfig: getQuantizationConfigDiff(conf),
		Params:             nil, // do not update collection params
		VectorsConfig:      nil, // do not update vectors config
	})
	return err
}

func createCollection(ctx context.Context, cc qdrant.CollectionsClient, name string, dims uint64, conf *schema.QdrantConfig) error {
	// Create a new collection with the new config using the data of the old collection
	_, err := cc.Create(ctx, &qdrant.CreateCollection{
		CollectionName:     name,
		HnswConfig:         getHnswConfigDiff(conf),
		OptimizersConfig:   getOptimizersConfigDiff(conf),
		QuantizationConfig: getQuantizationConfig(conf),
		ShardNumber:        pointers.Ptr(uint32(1)),
		OnDiskPayload:      pointers.Ptr(true),
		VectorsConfig: &qdrant.VectorsConfig{
			Config: &qdrant.VectorsConfig_Params{
				Params: &qdrant.VectorParams{
					Size:               dims,
					Distance:           qdrant.Distance_Cosine,
					HnswConfig:         nil, // use collection default
					QuantizationConfig: nil, // use collection default
					OnDisk:             pointers.Ptr(true),
				},
			},
		},
		WalConfig:              nil, // default
		ReplicationFactor:      nil, // default
		WriteConsistencyFactor: nil, // default
		InitFromCollection:     nil, // default
	})
	return err
}

func ensureRepoIDIndex(ctx context.Context, cc qdrant.PointsClient, name string) error {
	// This is idempotent, so no need to check if it exists first
	_, err := cc.CreateFieldIndex(ctx, &qdrant.CreateFieldIndexCollection{
		CollectionName:   name,
		Wait:             pointers.Ptr(true),
		FieldName:        fieldRepoID,
		FieldType:        pointers.Ptr(qdrant.FieldType_FieldTypeInteger),
		FieldIndexParams: nil,
		Ordering:         nil,
	})
	if err != nil {
		return err
	}

	return nil
}

func getHnswConfigDiff(conf *schema.QdrantConfig) *qdrant.HnswConfigDiff {
	return &qdrant.HnswConfigDiff{
		M:                 pointers.Ptr(uint64(conf.Hnsw.M)),
		EfConstruct:       pointers.Ptr(uint64(conf.Hnsw.EfConstruct)),
		FullScanThreshold: pointers.Ptr(uint64(conf.Hnsw.FullScanThreshold)),
		OnDisk:            pointers.Ptr(conf.Hnsw.OnDisk),
		PayloadM:          pointers.Ptr(uint64(conf.Hnsw.PayloadM)),
	}
}

func getOptimizersConfigDiff(conf *schema.QdrantConfig) *qdrant.OptimizersConfigDiff {
	return &qdrant.OptimizersConfigDiff{
		IndexingThreshold: pointers.Ptr(uint64(conf.Optimizers.IndexingThreshold)),
		MemmapThreshold:   pointers.Ptr(uint64(conf.Optimizers.MemmapThreshold)),
	}
}

func getQuantizationConfigDiff(conf *schema.QdrantConfig) *qdrant.QuantizationConfigDiff {
	if !conf.Quantization.Enabled {
		return &qdrant.QuantizationConfigDiff{
			Quantization: &qdrant.QuantizationConfigDiff_Disabled{},
		}
	}
	return &qdrant.QuantizationConfigDiff{
		Quantization: &qdrant.QuantizationConfigDiff_Scalar{
			Scalar: &qdrant.ScalarQuantization{
				Type:      qdrant.QuantizationType_Int8,
				Quantile:  pointers.Ptr(float32(conf.Quantization.Quantile)),
				AlwaysRam: pointers.Ptr(false),
			},
		},
	}
}

func getQuantizationConfig(conf *schema.QdrantConfig) *qdrant.QuantizationConfig {
	if !conf.Quantization.Enabled {
		return nil
	}
	return &qdrant.QuantizationConfig{
		Quantization: &qdrant.QuantizationConfig_Scalar{
			Scalar: &qdrant.ScalarQuantization{
				Type:      qdrant.QuantizationType_Int8,
				Quantile:  pointers.Ptr(float32(conf.Quantization.Quantile)),
				AlwaysRam: pointers.Ptr(false),
			},
		},
	}
}
