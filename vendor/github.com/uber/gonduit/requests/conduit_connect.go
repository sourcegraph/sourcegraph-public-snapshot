package requests

// ConduitConnectRequest represents a request to conduit.connect.
type ConduitConnectRequest struct {
	Client            string `json:"client"`
	ClientVersion     string `json:"clientVersion"`
	ClientDescription string `json:"clientDescription"`
	Host              string `json:"host"`
	User              string `json:"user"`
	AuthToken         string `json:"authToken"`
	AuthSignature     string `json:"authSignature"`
}
