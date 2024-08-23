package responses

import "github.com/uber/gonduit/entities"

// ConduitCapabilitiesResponse represents a response from calling
// conduit.capabilities.
type ConduitCapabilitiesResponse struct {
	Authentication []string `json:"authentication"`
	Signatures     []string `json:"signatures"`
	Input          []string `json:"input"`
	Output         []string `json:"output"`
}

// ConduitQueryResponse is the response of calling conduit.query.
type ConduitQueryResponse map[string]*entities.ConduitMethod

// ConduitConnectResponse represents the response from calling conduit.connect.
type ConduitConnectResponse struct {
	SessionKey   string `json:"sessionKey"`
	ConnectionID int64  `json:"connectionID"`
}
