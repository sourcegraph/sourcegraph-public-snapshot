package sourcegraph

type DefsRefreshIndexRequest struct {
	URI      string `json:"uri"`
	Revision string `json:"revision"`
}

type RepoCreateOrUpdateRequest struct {
	URI         string `json:"uri"`
	Description string `json:"description"`
	Fork        bool   `json:"fork"`
	Private     bool   `json:"private"`
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

type PhabricatorRepoCreateRequest struct {
	URI      string `json:"uri"`
	Callsign string `json:"callsign"`
	URL      string `json:"url"`
}
