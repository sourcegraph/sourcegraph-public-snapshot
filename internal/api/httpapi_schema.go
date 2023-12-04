package api

type ExternalServiceConfigsRequest struct {
	Kind    string `json:"kind"`
	Limit   int    `json:"limit"`
	AfterID int    `json:"after_id"`
}
