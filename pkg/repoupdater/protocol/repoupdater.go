package protocol

import (
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

// RepoLookupArgs is a request for information about a repository on repoupdater.
type RepoLookupArgs struct {
	// Repo is the repository to get information about.
	Repo api.RepoURI
}

// RepoLookupResult is the response to a repository information request (RepoLookupArgs).
type RepoLookupResult struct {
	// Repo contains information about the repository, if it is found. If it's not found, it is nil.
	Repo *RepoInfo
}

// RepoInfo is information about a repository.
type RepoInfo struct {
	// URI is the canonical URI of the repository. Its case (uppercase/lowercase) may differ from the URI arg used
	// in the lookup. If the repository was renamed on the external service, this URI will be the new name.
	URI api.RepoURI

	Description string // repository description (from the external service)
	Fork        bool   // whether this repository is a fork of another repository (from the external service)
}
