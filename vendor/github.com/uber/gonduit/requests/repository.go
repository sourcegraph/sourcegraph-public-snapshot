package requests

// RepositoryQueryRequest represents a request to the repository.query call.
type RepositoryQueryRequest struct {
	IDs        []uint64 `json:"ids"`
	PHIDs      []string `json:"phids"`
	Callsigns  []string `json:"callsigns"`
	VCSTypes   []string `json:"vcsTypes"`
	RemoteURIs []string `json:"remoteURIs"`
	UUIDs      []string `json:"uuids"`
	Order      string   `json:"order"`
	Before     string   `json:"before"`
	After      string   `json:"after"`
	Limit      uint64   `json:"limit"`
	Request
}
