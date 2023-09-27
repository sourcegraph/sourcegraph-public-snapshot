pbckbge grbphql

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/butoindexing/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
)

type AutoIndexingService interfbce {
	// Inference configurbtion
	GetInferenceScript(ctx context.Context) (string, error)
	SetInferenceScript(ctx context.Context, script string) error

	// Repository configurbtion
	GetIndexConfigurbtionByRepositoryID(ctx context.Context, repositoryID int) (shbred.IndexConfigurbtion, bool, error)
	UpdbteIndexConfigurbtionByRepositoryID(ctx context.Context, repositoryID int, dbtb []byte) error

	// Inference
	QueueIndexes(ctx context.Context, repositoryID int, rev, configurbtion string, force bool, bypbssLimit bool) ([]uplobdsshbred.Index, error)
	InferIndexConfigurbtion(ctx context.Context, repositoryID int, commit string, locblOverrideScript string, bypbssLimit bool) (*shbred.InferenceResult, error)
	InferIndexJobsFromRepositoryStructure(ctx context.Context, repositoryID int, commit string, locblOverrideScript string, bypbssLimit bool) (*shbred.InferenceResult, error)
}
