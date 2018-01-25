package api

type DefsRefreshIndexRequest struct {
	RepoURI  `json:"uri"`
	CommitID `json:"revision"`
}

type PkgsRefreshIndexRequest struct {
	RepoURI  `json:"uri"`
	CommitID `json:"revision"`
}

// RepoCreateOrUpdateRequest is a request to create or update a repository.
//
// NOTE: Some fields are only used during creation (and are not used to update an existing repository).
type RepoCreateOrUpdateRequest struct {
	// RepoURI is the repository's URI.
	//
	// TODO(sqs): Add a way for callers to request that this repository's URI be renamed.
	RepoURI `json:"uri"`

	// Enabled is whether the repository should be enabled when initially created.
	//
	// NOTE: If the repository already exists when this request is received, its enablement is not updated. This
	// field is used only when creating the repository.
	Enabled bool `json:"enabled"`

	// Description is the repository's description on its external origin.
	Description string `json:"description"`

	// Fork is whether this repository is a fork (according to its external origin).
	Fork bool `json:"fork"`
}

type RepoUpdateIndexRequest struct {
	RepoID   `json:"repoID"`
	CommitID `json:"revision"`
	Language string `json:"language"`
}

type RepoUnindexedDependenciesRequest struct {
	RepoID   `json:"repoID"`
	Language string `json:"language"`
}

type ReposGetInventoryUncachedRequest struct {
	Repo RepoID
	CommitID
}

type PhabricatorRepoCreateRequest struct {
	RepoURI  `json:"uri"`
	Callsign string `json:"callsign"`
	URL      string `json:"url"`
}
