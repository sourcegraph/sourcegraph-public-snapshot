package cxp

import (
	"encoding/json"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonx"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
)

// ClientProxyInitializeParams contains the parameters for the client's "initialize" request to the
// CXP proxy.
//
// It is lspext.ClientProxyInitializeParams with an added nested initializationOptions.settings
// field.
type ClientProxyInitializeParams struct {
	InitializationOptions ClientProxyInitializationOptions `json:"initializationOptions"`
	Capabilities          ClientCapabilities               `json:"capabilities"`
	lspext.ClientProxyInitializeParams
}

// ClientProxyInitializeParams contains the initialization options for the client's "initialize"
// request to the CXP proxy.
type ClientProxyInitializationOptions struct {
	lspext.ClientProxyInitializationOptions
	Settings ExtensionSettings `json:"settings"`
}

// InitializeParams contains the parameters for the client's (or proxy's) "initialize" request to
// the extension.
//
// It is lspext.InitializeParams with an added nested initializationOptions.settings field.
type InitializeParams struct {
	InitializationOptions *InitializationOptions `json:"initializationOptions,omitempty"`
	Capabilities          ClientCapabilities     `json:"capabilities"`
	lspext.InitializeParams
}

// InitializationOptions contains arbitrary initialization options at the top level, plus extension
// settings.
type InitializationOptions struct {
	Other    map[string]interface{} `json:"-"`
	Settings ExtensionSettings      `json:"settings"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (o *InitializationOptions) UnmarshalJSON(data []byte) error {
	var s struct {
		Settings ExtensionSettings `json:"settings"`
	}
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*o = InitializationOptions{Settings: s.Settings}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	delete(m, "settings")

	if len(m) > 0 {
		(*o).Other = make(map[string]interface{}, len(m))
	}
	for k, v := range m {
		(*o).Other[k] = v
	}
	return nil
}

// MarshalJSON implements json.Marshaler.
func (o InitializationOptions) MarshalJSON() ([]byte, error) {
	m := make(map[string]interface{}, len(o.Other)+1)
	for k, v := range o.Other {
		m[k] = v
	}
	m["settings"] = o.Settings
	return json.Marshal(m)
}

// ExtensionSettings contains the global/organization/user settings for an extension.
type ExtensionSettings struct {
	Merged *json.RawMessage `json:"merged,omitempty"`
}

type ClientCapabilities struct {
	lsp.ClientCapabilities

	Decorations *DecorationsCapabilityOptions `json:"decorations,omitempty"`

	// TODO(sqs): add this to cxp-js
	Exec bool `json:"exec"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	lsp.InitializeResult
}

type ServerCapabilities struct {
	lsp.ServerCapabilities

	DecorationsProvider *DecorationsProviderServerCapabilities `json:"decorationsProvider,omitempty"`

	Contributions *Contributions `json:"contributions,omitempty"`
}

type DecorationsCapabilityOptions struct {
	Static  bool `json:"static,omitempty"`
	Dynamic bool `json:"dynamic,omitempty"`
}

type DecorationsProviderServerCapabilities struct {
	DecorationsCapabilityOptions
}

type TextDocumentPublishDecorationsParams struct {
	TextDocument lsp.TextDocumentIdentifier      `json:"textDocument"`
	Decorations  []lspext.TextDocumentDecoration `json:"decorations"`
}

// ParseClientCapabilities parses the client capabilities from the client's initialize message.
func ParseClientCapabilities(initializeParams []byte) (*ClientCapabilities, error) {
	var params struct {
		Capabilities ClientCapabilities `json:"capabilities"`
	}
	if err := json.Unmarshal(initializeParams, &params); err != nil {
		return nil, err
	}
	return &params.Capabilities, nil
}

type ConfigurationUpdateParams struct {
	Path  jsonx.Path  `json:"path"`
	Value interface{} `json:"value"`
}
