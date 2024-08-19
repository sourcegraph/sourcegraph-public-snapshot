package entities

// Session is the conduit session state
// that will be sent in the JSON params as __conduit__.
type Session struct {
	SessionKey   string `json:"sessionKey"`
	ConnectionID int64  `json:"connectionID"`
}
