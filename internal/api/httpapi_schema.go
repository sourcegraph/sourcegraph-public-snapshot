package api

type PhabricatorRepoCreateRequest struct {
	RepoName `json:"repo"`
	Callsign string `json:"callsign"`
	URL      string `json:"url"`
}

type ExternalServiceConfigsRequest struct {
	Kind    string `json:"kind"`
	Limit   int    `json:"limit"`
	AfterID int    `json:"after_id"`
}
