package protocol

import (
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

// RepoLookupArgs is a request for information about a repository on repoupdater.
//
// Exactly one of Repo and ExternalRepo should be set.
type RepoLookupArgs struct {
	// Repo is the repository URI to look up. If the ExternalRepo information is available to the
	// caller, it is preferred to use that (because it is robust to renames).
	Repo api.RepoURI `json:",omitempty"`

	// ExternalRepo specifies the repository to look up by its external repository identity.
	ExternalRepo *api.ExternalRepoSpec
}

// RepoLookupResult is the response to a repository information request (RepoLookupArgs).
type RepoLookupResult struct {
	// Repo contains information about the repository, if it is found. If it's not found, it is nil.
	Repo *RepoInfo
}

// RepoInfo is information about a repository that lives on an external service (such as GitHub or GitLab).
type RepoInfo struct {
	// URI is the canonical URI of the repository. Its case (uppercase/lowercase) may differ from the URI arg used
	// in the lookup. If the repository was renamed on the external service, this URI will be the new name.
	URI api.RepoURI

	Description string // repository description (from the external service)
	Fork        bool   // whether this repository is a fork of another repository (from the external service)

	// ExternalRepo specifies this repository's ID on the external service where it resides (and the external
	// service itself).
	//
	// TODO(sqs): make this required (non-pointer) when both sides have been upgraded to use it. It is only
	// optional during the transition period.
	ExternalRepo *api.ExternalRepoSpec
}
