package api

type DefsRefreshIndexRequest struct {
	URI      string `json:"uri"`
	Revision string `json:"revision"`
}

type PkgsRefreshIndexRequest struct {
	URI      string `json:"uri"`
	Revision string `json:"revision"`
}

type RepoCreateOrUpdateRequest struct {
	URI         string `json:"uri"`
	Description string `json:"description"`
	Fork        bool   `json:"fork"`
	Enabled     bool   `json:"enabled"`
}

type RepoUpdateIndexRequest struct {
	RepoID   int32  `json:"repoID"`
	Revision string `json:"revision"`
	Language string `json:"language"`
}

type RepoUnindexedDependenciesRequest struct {
	RepoID   int32  `json:"repoID"`
	Language string `json:"language"`
}

type ReposGetInventoryUncachedRequest struct {
	Repo     int32
	CommitID string
}

type PhabricatorRepoCreateRequest struct {
	URI      string `json:"uri"`
	Callsign string `json:"callsign"`
	URL      string `json:"url"`
}
