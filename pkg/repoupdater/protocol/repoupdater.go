package protocol

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/pkg/api"
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

func (a *RepoLookupArgs) String() string {
	if a.ExternalRepo != nil {
		return fmt.Sprintf("RepoLookupArgs{%s}", a.ExternalRepo)
	}
	return fmt.Sprintf("RepoLookupArgs{%s}", a.Repo)
}

// RepoLookupResult is the response to a repository information request (RepoLookupArgs).
type RepoLookupResult struct {
	// Repo contains information about the repository, if it is found. If an error occurred, it is nil.
	Repo *RepoInfo

	ErrorNotFound               bool // the repository host reported that the repository was not found
	ErrorUnauthorized           bool // the repository host rejected the client's authorization
	ErrorTemporarilyUnavailable bool // the repository host was temporarily unavailable (e.g., rate limit exceeded)
}

func (r *RepoLookupResult) String() string {
	var parts []string
	if r.Repo != nil {
		parts = append(parts, "repo="+r.Repo.String())
	}
	if r.ErrorNotFound {
		parts = append(parts, "notfound")
	}
	if r.ErrorUnauthorized {
		parts = append(parts, "unauthorized")
	}
	return fmt.Sprintf("RepoLookupResult{%s}", strings.Join(parts, " "))
}

// RepoInfo is information about a repository that lives on an external service (such as GitHub or GitLab).
type RepoInfo struct {
	// URI is the canonical URI of the repository. Its case (uppercase/lowercase) may differ from the URI arg used
	// in the lookup. If the repository was renamed on the external service, this URI will be the new name.
	URI api.RepoURI

	Description string // repository description (from the external service)
	Fork        bool   // whether this repository is a fork of another repository (from the external service)
	Archived    bool   // whether this repository is archived (from the external service)

	VCS VCSInfo // VCS-related information (for cloning/updating)

	Links *RepoLinks // link URLs related to this repository

	// ExternalRepo specifies this repository's ID on the external service where it resides (and the external
	// service itself).
	//
	// TODO(sqs): make this required (non-pointer) when both sides have been upgraded to use it. It is only
	// optional during the transition period.
	ExternalRepo *api.ExternalRepoSpec
}

func (r *RepoInfo) String() string {
	return fmt.Sprintf("RepoInfo{%s}", r.URI)
}

// VCSInfo describes how to access an external repository's Git data (to clone or update it).
type VCSInfo struct {
	URL string // the Git remote URL
}

// RepoLinks contains URLs and URL patterns for objects in this repository.
type RepoLinks struct {
	Root   string // the repository's main (root) page URL
	Tree   string // the URL to a tree, with {rev} and {path} substitution variables
	Blob   string // the URL to a blob, with {rev} and {path} substitution variables
	Commit string // the URL to a commit, with {commit} substitution variable
}

// RepoUpdateRequest is a request to update the contents of a given repo, or clone it if it doesn't exist.
type RepoUpdateRequest struct {
	Repo api.RepoURI `json:"repo"`

	// URL is the repository's Git remote URL (from which to clone or update).
	URL string `json:"url"`
}
