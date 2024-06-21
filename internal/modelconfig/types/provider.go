package types

type Provider struct {
	ID          ProviderID `json:"id"`
	DisplayName string     `json:"displayName"`

	ClientSideConfig *ClientSideProviderConfig `json:"clientSideConfig,omitempty"`
	ServerSideConfig *ServerSideProviderConfig `json:"serverSideConfig,omitempty"`
}
