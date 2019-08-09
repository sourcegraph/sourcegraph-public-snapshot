package awscodecommit

import (
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// ServiceType is the (api.ExternalRepoSpec).ServiceType value for AWS CodeCommit
// repositories. The ServiceID value is the ARN (Amazon Resource Name) omitting the repository name
// suffix (e.g., "arn:aws:codecommit:us-west-1:123456789:").
const ServiceType = "awscodecommit"

// ExternalRepoSpec returns an api.ExternalRepoSpec that refers to the specified AWS
// CodeCommit repository.
func ExternalRepoSpec(repo *Repository, serviceID string) api.ExternalRepoSpec {
	return api.ExternalRepoSpec{
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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_780(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
