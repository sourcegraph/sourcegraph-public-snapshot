package types

import (
	"github.com/sourcegraph/sourcegraph/schema"
)

type AzureDevOpsConnection struct {
	URN string
	*schema.AzureDevOpsConnection
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

type GerritConnection struct {
	// The unique resource identifier of the external service.
	URN string
	*schema.GerritConnection
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

type PerforceConnection struct {
	// The unique resource identifier of the external service.
	URN string
	*schema.PerforceConnection
}
