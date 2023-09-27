package db

import (
	"context"

	qdrant "github.com/qdrant/go-client/qdrant"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func ensureModelCollection(ctx context.Context, db *qdrantDB, modelID string, modelDims uint64) error {
	// Make the actual collection end with `.default` so we can switch between
	// configurations with aliases.
	name := CollectionName(modelID)
	realName := name + ".default"

	err := ensureCollectionWithConfig(ctx, db.collectionsClient, realName, modelDims, conf.GetEmbeddingsConfig(conf.Get().SiteConfiguration).Qdrant)
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

func ensureCollectionWithConfig(ctx context.Context, cc qdrant.CollectionsClient, name string, dims uint64, qc conftypes.QdrantConfig) error {
	resp, err := cc.List(ctx, &qdrant.ListCollectionsRequest{})
	if err != nil {
		return err
	}

	for _, collection := range resp.GetCollections() {
		if collection.GetName() == name {
			_, err = cc.Update(ctx, updateCollectionParams(name, qc))
			return err
		}
	}

	_, err = cc.Create(ctx, createCollectionParams(name, dims, qc))
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
