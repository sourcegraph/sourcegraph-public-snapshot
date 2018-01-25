package api

type DefsRefreshIndexRequest struct {
	RepoURI  `json:"uri"`
	CommitID `json:"revision"`
}

type PkgsRefreshIndexRequest struct {
	RepoURI  `json:"uri"`
	CommitID `json:"revision"`
}

type RepoCreateOrUpdateRequest struct {
	RepoURI     `json:"uri"`
	Description string `json:"description"`
	Fork        bool   `json:"fork"`
	Enabled     bool   `json:"enabled"` // only used when creating the repository, does not update an existing repository's enablement
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
