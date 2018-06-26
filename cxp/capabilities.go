package cxp

import (
	"encoding/json"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/sourcegraph/xlang/lspext"
)

// ParseExperimentalClientCapabilities parses the "clientCapabilities.experimental" object from the
// client's initialize message.
func ParseExperimentalClientCapabilities(initializeParams []byte) (*lspext.ExperimentalClientCapabilities, error) {
	var params struct {
		lspext.InitializeParams
		Capabilities struct {
			lsp.ClientCapabilities
			Experimental lspext.ExperimentalClientCapabilities `json:"experimental"`
		} `json:"capabilities"`
	}
	if err := json.Unmarshal(initializeParams, &params); err != nil {
		return nil, err
	}
	return &params.Capabilities.Experimental, nil
}
