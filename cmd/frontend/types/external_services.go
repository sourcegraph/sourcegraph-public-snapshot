package types

import (
	"github.com/sourcegraph/sourcegraph/schema"
)

type AWSCodeCommitConnection struct {
	// The unique resource identifier of the external service.
	URN string
	*schema.AWSCodeCommitConnection
}

type BitbucketCloudConnection struct {
	// The unique resource identifier of the external service.
	URN string
	*schema.BitbucketCloudConnection
}

type BitbucketServerConnection struct {
	// The unique resource identifier of the external service.
	URN string
	*schema.BitbucketServerConnection
}

type GitHubConnection struct {
	// The unique resource identifier of the external service.
	URN string
	*schema.GitHubConnection
}

type GitLabConnection struct {
	// The unique resource identifier of the external service.
	URN string
	*schema.GitLabConnection
}

type GitoliteConnection struct {
	// The unique resource identifier of the external service.
	URN string
	*schema.GitoliteConnection
}

type OtherExternalServiceConnection struct {
	// The unique resource identifier of the external service.
	URN string
	*schema.OtherExternalServiceConnection
}

type PhabricatorConnection struct {
	// The unique resource identifier of the external service.
	URN string
	*schema.PhabricatorConnection
}
