package protocol

// GitolitePhabricatorMetadataResponse is the response for a request
// for Phabricator metadata through the Gitolite API
type GitolitePhabricatorMetadataResponse struct {
	Callsign string `json:"callsign"`
}
