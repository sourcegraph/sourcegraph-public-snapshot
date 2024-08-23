// This file was auto-generated by Fern from our API Definition.

package api

type CreateConnectorRequest struct {
	// A human-readable name for the connector.
	Name string `json:"name" url:"-"`
	// A description of the connector.
	Description *string `json:"description,omitempty" url:"-"`
	// The URL of the connector that will be used to search for documents.
	Url string `json:"url" url:"-"`
	// A list of fields to exclude from the prompt (fields remain in the document).
	Excludes []string `json:"excludes,omitempty" url:"-"`
	// The OAuth 2.0 configuration for the connector. Cannot be specified if service_auth is specified.
	Oauth *CreateConnectorOAuth `json:"oauth,omitempty" url:"-"`
	// Whether the connector is active or not.
	Active *bool `json:"active,omitempty" url:"-"`
	// Whether a chat request should continue or not if the request to this connector fails.
	ContinueOnFailure *bool `json:"continue_on_failure,omitempty" url:"-"`
	// The service to service authentication configuration for the connector. Cannot be specified if oauth is specified.
	ServiceAuth *CreateConnectorServiceAuth `json:"service_auth,omitempty" url:"-"`
}

type ConnectorsListRequest struct {
	// Maximum number of connectors to return [0, 100].
	Limit *float64 `json:"-" url:"limit,omitempty"`
	// Number of connectors to skip before returning results [0, inf].
	Offset *float64 `json:"-" url:"offset,omitempty"`
}

type ConnectorsOAuthAuthorizeRequest struct {
	// The URL to redirect to after the connector has been authorized.
	AfterTokenRedirect *string `json:"-" url:"after_token_redirect,omitempty"`
}

type UpdateConnectorRequest struct {
	// A human-readable name for the connector.
	Name *string `json:"name,omitempty" url:"-"`
	// The URL of the connector that will be used to search for documents.
	Url *string `json:"url,omitempty" url:"-"`
	// A list of fields to exclude from the prompt (fields remain in the document).
	Excludes []string `json:"excludes,omitempty" url:"-"`
	// The OAuth 2.0 configuration for the connector. Cannot be specified if service_auth is specified.
	Oauth             *CreateConnectorOAuth `json:"oauth,omitempty" url:"-"`
	Active            *bool                 `json:"active,omitempty" url:"-"`
	ContinueOnFailure *bool                 `json:"continue_on_failure,omitempty" url:"-"`
	// The service to service authentication configuration for the connector. Cannot be specified if oauth is specified.
	ServiceAuth *CreateConnectorServiceAuth `json:"service_auth,omitempty" url:"-"`
}
