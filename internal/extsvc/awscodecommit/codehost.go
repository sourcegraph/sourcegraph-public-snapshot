package awscodecommit

import (
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

// ExternalRepoSpec returns an api.ExternalRepoSpec that refers to the specified AWS
// CodeCommit repository.
func ExternalRepoSpec(repo *Repository, serviceID string) api.ExternalRepoSpec {
	return api.ExternalRepoSpec{
		ID:          repo.ID,
		ServiceType: extsvc.TypeAWSCodeCommit,
		ServiceID:   serviceID,
	}
}

// ServiceID creates the repository external service ID. See AWSCodeCommitServiceType for
// documentation on the format of this value.
//
// This value uniquely identifies the most specific namespace in which AWS CodeCommit repositories
// are defined.
func ServiceID(awsPartition, awsRegion, awsAccountID string) string {
	return "arn:" + awsPartition + ":codecommit:" + awsRegion + ":" + awsAccountID + ":"
}
