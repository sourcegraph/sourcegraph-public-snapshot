package repos

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

const UnimplementedDiscoverySource = "External Service type does not support discovery of repositories and namespaces."

// A DiscoverableSource yields metadata for remote entities (e.g. repositories, namespaces) on a readable external service
// that Sourcegraph may or may not have setup for mirror/sync operations
type DiscoverableSource interface {
	// ListNamespaces returns the namespaces available on the source.
	// Namespaces are used to organize which members and users can access repositories
	// and are defined by external service kind (e.g. Github organizations, Bitbucket projects, etc.)
	ListNamespaces(context.Context, chan SourceNamespaceResult)

	// SearchRepositories returns the repositories available on the source which match a given search query
	// and excluded repositories criteria.
	SearchRepositories(context.Context, string, int, []string, chan SourceResult)
}

// A SourceNamespaceResult is sent by a Source over a channel for each namespace it
// yields when listing namespace entities
type SourceNamespaceResult struct {
	// Source points to the Source that produced this result
	Source Source
	// Namespace is the external service namespace that was listed by the Source
	Namespace *types.ExternalServiceNamespace
	// Err is only set in case the Source ran into an error when listing namespaces
	Err error
}
