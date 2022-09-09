package api

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// This list is different from the builtinExtensions since it only includes code
// intel extensions.
var codeIntelExtensions = map[string]bool{
	"sourcegraph/apex":       true,
	"sourcegraph/clojure":    true,
	"sourcegraph/cobol":      true,
	"sourcegraph/cpp":        true,
	"sourcegraph/csharp":     true,
	"sourcegraph/cuda":       true,
	"sourcegraph/dart":       true,
	"sourcegraph/elixir":     true,
	"sourcegraph/erlang":     true,
	"sourcegraph/go":         true,
	"sourcegraph/graphql":    true,
	"sourcegraph/groovy":     true,
	"sourcegraph/haskell":    true,
	"sourcegraph/java":       true,
	"sourcegraph/jsonnet":    true,
	"sourcegraph/kotlin":     true,
	"sourcegraph/lisp":       true,
	"sourcegraph/lua":        true,
	"sourcegraph/ocaml":      true,
	"sourcegraph/pascal":     true,
	"sourcegraph/perl":       true,
	"sourcegraph/php":        true,
	"sourcegraph/powershell": true,
	"sourcegraph/protobuf":   true,
	"sourcegraph/python":     true,
	"sourcegraph/r":          true,
	"sourcegraph/ruby":       true,
	"sourcegraph/rust":       true,
	"sourcegraph/scala":      true,
	"sourcegraph/shell":      true,
	"sourcegraph/starlark":   true,
	"sourcegraph/swift":      true,
	"sourcegraph/tcl":        true,
	"sourcegraph/thrift":     true,
	"sourcegraph/typescript": true,
	"sourcegraph/verilog":    true,
	"sourcegraph/vhdl":       true,
}

func ExtensionRegistryReadEnabled() error {
	// We need to allow read access to the extension registry of sourcegraph.com to allow instances
	// on older versions to fetch extensions.
	if envvar.SourcegraphDotComMode() {
		return nil
	}

	return ExtensionRegistryWriteEnabled()
}

// The extensions list query (`extensionRegistry.extensions`) will be called by the native
// integration, a deployment method where we do not have control over updating the client.
//
// An example for this is the GitLab native integration that ships with GitLab releases.
// When enabled, it will make a GraphQL query to Sourcegraph to get a list of valid
// extensions.
//
// Since we want to support code navigation in the native integrations after 4.0, we are
// special-casing this API and make it return only code intel legacy extensions.
//
// Since the dotcom instance will be connected to from instances before 4.0, we'll need
// to keep the behavior of being able to list all extensions there.
//
// This method returns nil if all extensions are allowed to be listed.
func ExtensionRegistryListAllowedExtension() map[string]bool {
	err := ExtensionRegistryReadEnabled()

	if err != nil {
		return codeIntelExtensions
	} else {
		// nil means all extensions are allowed
		return nil
	}
}

func ExtensionRegistryWriteEnabled() error {
	cfg := conf.Get()
	if cfg.ExperimentalFeatures == nil || cfg.ExperimentalFeatures.EnableLegacyExtensions == nil || *cfg.ExperimentalFeatures.EnableLegacyExtensions == false {
		return errors.Errorf("Extensions are disabled. See https://docs.sourcegraph.com/extensions/deprecation")
	}

	return nil
}
