package clientconfig

// This is the JSON object which all clients request after authentication to determine how
// they should behave, e.g. if a site admin has restricted chat/autocomplete/other functionality,
// if experimental features are available, etc.
//
// The configuration is always specific to a single authenticated user.
//
// Adding new fields here is fine, but you cannot make backwards-incompatible changes (removing
// fields or change the meaning of fields in backwars-incompatible ways.) If you need to do that,
// then read up on https://github.com/sourcegraph/sourcegraph/pull/63591#discussion_r1663211601
//
// After adding a field here, you can implement it in cmd/frontend/internal/clientconfig/clientconfig.go
// GetForActor method.
type ClientConfig struct {
	// Whether the site admin allows this user to make use of Cody at all.
	CodyEnabled bool `json:"codyEnabled"`

	// Whether the site admin allows this user to make use of the Cody chat feature.
	ChatEnabled bool `json:"chatEnabled"`

	// Whether the site admin allows this user to make use of the Cody autocomplete feature.
	AutoCompleteEnabled bool `json:"autoCompleteEnabled"`

	// Whether the site admin allows the user to make use of the **custom** Cody commands feature.
	CustomCommandsEnabled bool `json:"customCommandsEnabled"`

	// Whether the site admin allows this user to make use of the Cody attribution feature.
	AttributionEnabled bool `json:"attributionEnabled"`

	// Whether the 'smart context window' feature should be enabled, and whether the Sourcegraph
	// instance supports various new GraphQL APIs needed to make it work.
	SmartContextWindowEnabled bool `json:"smartContextWindowEnabled"`

	// Whether the new Sourcegraph backend LLM models API endpoint should be used to query which
	// models are available.
	ModelsAPIEnabled bool `json:"modelsAPIEnabled"`
}
