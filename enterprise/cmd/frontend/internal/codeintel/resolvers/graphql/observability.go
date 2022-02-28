package graphql

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	lsifUploadByID            *observation.Operation
	lsifUploads               *observation.Operation
	lsifUploadsByRepo         *observation.Operation
	deleteLsifUpload          *observation.Operation
	lsifIndexByID             *observation.Operation
	lsifIndexes               *observation.Operation
	lsifIndexesByRepo         *observation.Operation
	deleteLsifIndexes         *observation.Operation
	commitGraph               *observation.Operation
	queueAutoIndexJobsForRepo *observation.Operation
	gitBlobLsifData           *observation.Operation
	gitBlobCodeIntelInfo      *observation.Operation
	configurationPolicyByID   *observation.Operation
	configurationPolicies     *observation.Operation
	createConfigurationPolicy *observation.Operation
	updateConfigurationPolicy *observation.Operation
	deleteConfigurationPolicy *observation.Operation
	indexConfiguration        *observation.Operation
	updateIndexConfiguration  *observation.Operation
	previewRepoFilter         *observation.Operation
	previewGitObjectFilter    *observation.Operation
}

func newOperations(observationContext *observation.Context) *operations {
	op := func(name string) *observation.Operation {
		return observationContext.Operation(observation.Op{
			Name: fmt.Sprintf("codeintel.resolver.%s", name),
			ErrorFilter: func(err error) observation.ErrorFilterBehaviour {
				if err == ErrIllegalBounds || err == ErrIllegalLimit {
					return observation.EmitForNone
				}
				return observation.EmitForSentry
			},
		})
	}

	return &operations{
		lsifUploadByID:            op("LSIFUploadByID"),
		lsifUploads:               op("LSIFUploads"),
		lsifUploadsByRepo:         op("LSIFUploadsByRepo"),
		deleteLsifUpload:          op("DeleteLSIFUpload"),
		lsifIndexByID:             op("LSIFIndexByID"),
		lsifIndexes:               op("LSIFIndexes"),
		lsifIndexesByRepo:         op("LSIFIndexesByRepo"),
		deleteLsifIndexes:         op("DeleteLSIFIndexes"),
		commitGraph:               op("CommitGraph"),
		queueAutoIndexJobsForRepo: op("QueueAutoIndexJobsForRepo"),
		gitBlobLsifData:           op("GitBlobLSIFData"),
		gitBlobCodeIntelInfo:      op("GitBlobCodeIntelInfo"),
		configurationPolicyByID:   op("ConfigurationPolicyByID"),
		configurationPolicies:     op("ConfigurationPolicies"),
		createConfigurationPolicy: op("CreateConfigurationPolicy"),
		updateConfigurationPolicy: op("UpdateConfigurationPolicy"),
		deleteConfigurationPolicy: op("DeleteConfigurationPolicy"),
		indexConfiguration:        op("IndexConfiguration"),
		updateIndexConfiguration:  op("UpdateIndexConfiguration"),
		previewRepoFilter:         op("PreviewRepoFilter"),
		previewGitObjectFilter:    op("PreviewGitObjectFilter"),
	}
}
