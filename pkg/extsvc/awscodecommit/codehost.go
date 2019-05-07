package awscodecommit

import (
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// AWSCodeCommitServiceType is the (api.ExternalRepoSpec).ServiceType value for AWS CodeCommit
// repositories. The ServiceID value is the ARN (Amazon Resource Name) omitting the repository name
// suffix (e.g., "arn:aws:codecommit:us-west-1:123456789:").
const ServiceType = "awscodecommit"

// AWSCodeCommitExternalRepoSpec returns an api.ExternalRepoSpec that refers to the specified AWS
// CodeCommit repository.
func ExternalRepoSpec(repo *Repository, serviceID string) *api.ExternalRepoSpec {
	return &api.ExternalRepoSpec{
		ID:          repo.ID,
		ServiceType: ServiceType,
		ServiceID:   serviceID,
	}
}

// ServiceID creates the repository external service ID. See AWSCodeCommitServiceType for
// documentation on the format of this value.
//
// This value uniquely identifies the most specific namespace in which AWS CodeCommit repositories
// are defined.
func ServiceID(awsPartition endpoints.Partition, awsRegion endpoints.Region, awsAccountID string) string {
	return "arn:" + awsPartition.ID() + ":codecommit:" + awsRegion.ID() + ":" + awsAccountID + ":"
}
