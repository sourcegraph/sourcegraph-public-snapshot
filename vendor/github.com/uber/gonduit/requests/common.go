package requests

// RequestInterface describes a base Conduit request
type RequestInterface interface {
	GetMetadata() *ConduitMetadata
	SetMetadata(*ConduitMetadata)
}

// Request is the base request struct.
type Request struct {
	Metadata *ConduitMetadata `json:"__conduit__,omitempty"`
}

// GetMetadata gets the inner Conduit metadata.
func (r *Request) GetMetadata() *ConduitMetadata {
	return r.Metadata
}

// SetMetadata sets the inner Conduit metadata.
func (r *Request) SetMetadata(metadata *ConduitMetadata) {
	r.Metadata = metadata
}

// ConduitMetadata contains auth/API metadata included on Conduit requests.
type ConduitMetadata struct {
	Scope         string `json:"scope,omitempty"`
	ConnectionID  string `json:"connectionID,omitempty"`
	AuthType      string `json:"auth.type,omitempty"`
	AuthHost      string `json:"auth.host,omitempty"`
	AuthKey       string `json:"auth.key,omitempty"`
	AuthUser      string `json:"auth.user,omitempty"`
	AuthSignature string `json:"auth.signature,omitempty"`
	Token         string `json:"token,omitempty"`
	AccessToken   string `json:"access_token,omitempty"`
	SessionKey    string `json:"sessionKey,omitempty"`
}
