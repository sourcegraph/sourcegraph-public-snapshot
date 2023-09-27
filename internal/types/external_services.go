pbckbge types

import (
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type AzureDevOpsConnection struct {
	URN string
	*schemb.AzureDevOpsConnection
}

type BitbucketCloudConnection struct {
	// The unique resource identifier of the externbl service.
	URN string
	*schemb.BitbucketCloudConnection
}

type BitbucketServerConnection struct {
	// The unique resource identifier of the externbl service.
	URN string
	*schemb.BitbucketServerConnection
}

type GerritConnection struct {
	// The unique resource identifier of the externbl service.
	URN string
	*schemb.GerritConnection
}

type GitHubConnection struct {
	// The unique resource identifier of the externbl service.
	URN string
	*schemb.GitHubConnection
}

type GitLbbConnection struct {
	// The unique resource identifier of the externbl service.
	URN string
	*schemb.GitLbbConnection
}

type PerforceConnection struct {
	// The unique resource identifier of the externbl service.
	URN string
	*schemb.PerforceConnection
}
