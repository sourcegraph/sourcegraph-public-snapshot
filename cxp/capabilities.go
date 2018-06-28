package cxp

import (
	"encoding/json"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
)

// ExperimentalClientCapabilities describes experimental client capabilities. It is sent in the
// client's initialize request in the `capabilities.experimental` field.
type ExperimentalClientCapabilities struct {
	Decorations bool `json:"decorations"` // decorations extension
	Exec        bool `json:"exec"`        // exec extension
}

// ExperimentalServerCapabilities describes experimental server capabilities. It is sent in the
// server's initialize response in the `capabilities.experimental` field.
type ExperimentalServerCapabilities struct {
	DecorationsProvider bool `json:"decorationsProvider"`

	Contributions *Contributions `json:"contributions,omitempty"`
}

// ParseExperimentalClientCapabilities parses the "clientCapabilities.experimental" object from the
// client's initialize message.
func ParseExperimentalClientCapabilities(initializeParams []byte) (*ExperimentalClientCapabilities, error) {
	var params struct {
		lspext.InitializeParams
		Capabilities struct {
			lsp.ClientCapabilities
			Experimental ExperimentalClientCapabilities `json:"experimental"`
		} `json:"capabilities"`
	}
	if err := json.Unmarshal(initializeParams, &params); err != nil {
		return nil, err
	}
	return &params.Capabilities.Experimental, nil
}
