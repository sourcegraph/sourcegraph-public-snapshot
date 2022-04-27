package graphql

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type operations struct {
	commitGraph               *observation.Operation
	configurationPolicies     *observation.Operation
	configurationPolicyByID   *observation.Operation
	createConfigurationPolicy *observation.Operation
	deleteConfigurationPolicy *observation.Operation
	deleteLsifIndexes         *observation.Operation
	deleteLsifUpload          *observation.Operation
	gitBlobCodeIntelInfo      *observation.Operation
	gitBlobLsifData           *observation.Operation
	gitTreeCodeIntelInfo      *observation.Operation
	indexConfiguration        *observation.Operation
	lsifIndexByID             *observation.Operation
	lsifIndexes               *observation.Operation
	lsifIndexesByRepo         *observation.Operation
	lsifUploadByID            *observation.Operation
	lsifUploads               *observation.Operation
	lsifUploadsByRepo         *observation.Operation
	previewGitObjectFilter    *observation.Operation
	previewRepoFilter         *observation.Operation
	queueAutoIndexJobsForRepo *observation.Operation
	repositorySummary         *observation.Operation
	requestedLanguageSupport  *observation.Operation
	requestLanguageSupport    *observation.Operation
	updateConfigurationPolicy *observation.Operation
	updateIndexConfiguration  *observation.Operation
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
		commitGraph:               op("CommitGraph"),
		configurationPolicies:     op("ConfigurationPolicies"),
		configurationPolicyByID:   op("ConfigurationPolicyByID"),
		createConfigurationPolicy: op("CreateConfigurationPolicy"),
		deleteConfigurationPolicy: op("DeleteConfigurationPolicy"),
		deleteLsifIndexes:         op("DeleteLSIFIndexes"),
		deleteLsifUpload:          op("DeleteLSIFUpload"),
		gitBlobCodeIntelInfo:      op("GitBlobCodeIntelInfo"),
		gitBlobLsifData:           op("GitBlobLSIFData"),
		gitTreeCodeIntelInfo:      op("GitTreeCodeIntelInfo"),
		indexConfiguration:        op("IndexConfiguration"),
		lsifIndexByID:             op("LSIFIndexByID"),
		lsifIndexes:               op("LSIFIndexes"),
		lsifIndexesByRepo:         op("LSIFIndexesByRepo"),
		lsifUploadByID:            op("LSIFUploadByID"),
		lsifUploads:               op("LSIFUploads"),
		lsifUploadsByRepo:         op("LSIFUploadsByRepo"),
		previewGitObjectFilter:    op("PreviewGitObjectFilter"),
		previewRepoFilter:         op("PreviewRepoFilter"),
		queueAutoIndexJobsForRepo: op("QueueAutoIndexJobsForRepo"),
		repositorySummary:         op("RepositorySummary"),
		requestedLanguageSupport:  op("RequestedLanguageSupport"),
		requestLanguageSupport:    op("RequestLanguageSupport"),
		updateConfigurationPolicy: op("UpdateConfigurationPolicy"),
		updateIndexConfiguration:  op("UpdateIndexConfiguration"),
	}
}
