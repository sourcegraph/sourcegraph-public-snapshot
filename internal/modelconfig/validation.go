package modelconfig

import (
	"net/url"
	"strings"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// IMPORTANT: Validation MUST be backwards-compatible.
//
// We can NEVER "relax" any validation checks we perform. Because that would lead to older
// Sourcegraph instances failing to accept newer versions of the configuration data.

const (
	maxDisplayNameLength = 128

	// validResourceCharREString is the regular expression for validating a single character
	// and if it is valid according to the "naming rules for resource IDs". This is to ensure
	// that resource IDs are constrainted to not contain gibberish, etc.
	//
	// Ideally we would have a very strict bar to ensure names are clear, e.g.
	// ^[a-z0-9_][a-z0-9_\-\.]*[a-z0-9_]$  However, we don't control what names are supplied
	// by an LLM provider. And enforcing our users to differentate the "Model ID" from "Model
	// Name" and "Display Name" is not great.
	//
	// So the following abomination of a regex is the allowed set of characters from RFC 3986
	// "Uniform Resource Identifier: Generic Syntax". Which may  still cause failures for more
	// exotic model names, but broad enough that a user wouldn't be surprised if the ID was
	// rejected.
	//
	// Pedant's note: We added '%' which is NOT included in the RFC, but is instead called out
	// for escape encoding. (e.g. "%20".) Also, colons were removed, since we use those to
	// delimit the sections of a ModelReference.
	validResourceCharREString = `^[a-zA-Z0-9-._~/?#[\]@!$&'()*+,;=%]+$`
)

// resourceIDRE verifies if the entire string matches our naming rules for IDs.
var resourceIDRE = regexp.MustCompile(validResourceCharREString)

func validateProvider(p types.Provider) error {
	// Display name is optional, but if it is set ensure it is under 128 chars.
	if l := len(p.DisplayName); l > 0 && l > maxDisplayNameLength {
		return errors.Errorf("display name length: %d", l)
	}
	if !resourceIDRE.MatchString(string(p.ID)) {
		return errors.New("id format")
	}

	// We intentionally don't validate any provider configuration
	// data here, as it isn't clear if that is even a good idea
	// in the first place. (e.g. we may not know the exact format
	// of some 3rd party provider's configuration knob.)

	// But there are some cases which are just too error prone to let slide...
	if ssCfg := p.ServerSideConfig; ssCfg != nil {
		// If using the "GenericProvider", we need to know the actual shape of the HTTP request
		// to send!
		if genCfg := ssCfg.GenericProvider; genCfg != nil {
			if genCfg.ServiceName == "" {
				return errors.New("no service name set for generic provider")
			}
		}
	}

	// BUG: We should verify the at the provider contains _some_
	// server-side config. Since otherwise the provider cannot actually
	// be used. However, we don't expect the Provider data exposed from
	// the embedded binary or Cody Gateway to actually contain that
	// server-side config because it doesn't make sense. So part of the
	// rendering of data needs to fall back to using the Sourcegraph
	// instance and its access token, etc.

	return nil
}

func ValidateModelRef(ref types.ModelRef) error {
	if ref == "" {
		return errors.New("modelRef is blank")
	}

	parts := strings.Split(string(ref), "::")
	if len(parts) != 3 {
		return errors.New("modelRef syntax error")
	}
	if !resourceIDRE.MatchString(parts[0]) {
		return errors.New("invalid ProviderID")
	}
	if !resourceIDRE.MatchString(parts[1]) {
		return errors.New("invalid APIVersionID")
	}
	if !resourceIDRE.MatchString(parts[2]) {
		return errors.New("invalid ModelID")
	}
	return nil
}

func validateModel(m types.Model) error {
	if l := len(m.DisplayName); l > 0 && l > maxDisplayNameLength {
		return errors.Errorf("display name length: %d", l)
	}
	if err := ValidateModelRef(m.ModelRef); err != nil {
		return errors.Wrap(err, "modelref")
	}

	ctxWin := m.ContextWindow
	if ctxWin.MaxInputTokens <= 0 {
		return errors.New("invalid max input tokens")
	}
	if ctxWin.MaxOutputTokens <= 0 {
		return errors.New("invalid max output tokens")
	}

	// We intentionally do not validate any of the enum metadata fields, because
	// older Sourcegraph instances wouldn't be able to recognize any newer values.

	return nil
}

// ValidateModelConfig validates that the model configuration data expressed is valid.
func ValidateModelConfig(cfg *types.ModelConfiguration) error {
	if cfg == nil {
		return errors.New("no config provided")
	}

	for _, provider := range cfg.Providers {
		if err := validateProvider(provider); err != nil {
			return errors.Wrapf(err, "validating provider %q", provider.ID)
		}
	}

	for _, model := range cfg.Models {
		if err := validateModel(model); err != nil {
			return errors.Wrapf(err, "validating model %q", model.ModelRef)
		}

		// Verify the model is referencing a known provider.
		var forKnownProvider bool
		for _, provider := range cfg.Providers {
			if strings.HasPrefix(string(model.ModelRef), string(provider.ID)+"::") {
				forKnownProvider = true
				break
			}
		}
		if !forKnownProvider {
			return errors.Errorf("model %q does not match a known provider", model.ModelRef)
		}
	}

	isKnownModel := func(mref types.ModelRef) bool {
		for _, model := range cfg.Models {
			if model.ModelRef == mref {
				return true
			}
		}
		return false
	}

	if !isKnownModel(cfg.DefaultModels.Chat) {
		return errors.Errorf("unknown chat model %q", cfg.DefaultModels.Chat)
	}
	if !isKnownModel(cfg.DefaultModels.CodeCompletion) {
		return errors.Errorf("unknown code completion model %q", cfg.DefaultModels.CodeCompletion)
	}
	if !isKnownModel(cfg.DefaultModels.FastChat) {
		return errors.Errorf("unknown fast chat model %q", cfg.DefaultModels.FastChat)
	}

	return nil
}

// isValidRule returns whether the ModelRef allow/deny list rule is well-formed.
func isValidRule(rule string) bool {
	if rule == "" {
		return false
	}
	// Aserisks can only be the first or last character of the rule.
	for i := 1; i < len(rule)-1; i++ {
		if rule[i] == '*' {
			return false
		}
	}
	return true
}

func verifySourcegraphSiteConfig(sgConfig *types.SourcegraphModelConfig) error {
	if sgConfig == nil {
		return nil
	}

	if endpoint := sgConfig.Endpoint; endpoint != nil {
		u, err := url.Parse(*endpoint)
		if err != nil || u.Scheme == "" {
			return errors.New("invalid endpoint URL")
		}
	}

	if modelFilters := sgConfig.ModelFilters; modelFilters != nil {
		for _, allowRule := range modelFilters.Allow {
			if !isValidRule(allowRule) {
				return errors.Errorf("invalid allow list rule: %q", allowRule)
			}
		}
		for _, denyRule := range modelFilters.Deny {
			if !isValidRule(denyRule) {
				return errors.Errorf("invalid deny list rule: %q", denyRule)
			}
		}
	}

	return nil
}

func validateProviderOverrides(overrides []types.ProviderOverride) error {
	seenProviderIDs := map[types.ProviderID]bool{}
	for _, override := range overrides {
		// All provider IDs are unique.
		if seenProviderIDs[override.ID] {
			return errors.Errorf("provider %q specified twice", override.ID)
		}
		seenProviderIDs[override.ID] = true

		// All provider IDs are valid.
		if !resourceIDRE.MatchString(string(override.ID)) {
			return errors.Errorf("invalid provider ID %q", override.ID)
		}
	}
	return nil
}

func validateModelOverrides(overrides []types.ModelOverride) error {
	seenModelRefs := map[types.ModelRef]bool{}
	for _, override := range overrides {
		// All models have a valid ModelRef.
		if err := ValidateModelRef(override.ModelRef); err != nil {
			return errors.Wrapf(err, "validating model ref %q", override.ModelRef)
		}

		// All model ID overrides are unique.
		if seenModelRefs[override.ModelRef] {
			return errors.Errorf("model %q specified twice", override.ModelRef)
		}
		seenModelRefs[override.ModelRef] = true
	}
	return nil
}

// ValidateSiteConfig validates that the site configuration data expressed is valid.
func ValidateSiteConfig(doc *types.SiteModelConfiguration) error {
	if err := verifySourcegraphSiteConfig(doc.SourcegraphModelConfig); err != nil {
		return errors.Wrap(err, "sourcegraph config")
	}
	if err := validateProviderOverrides(doc.ProviderOverrides); err != nil {
		return errors.Wrap(err, "provider overrides")
	}
	if err := validateModelOverrides(doc.ModelOverrides); err != nil {
		return errors.Wrap(err, "model overrides")
	}

	// When verifying default models, we expect it to be OK to default to
	// a model NOT explicitly defined in the site config. e.g. using something
	// that we expect to be supplied by Sourcegraph. So we just check if they
	// are valid ModelRefs.
	if defModels := doc.DefaultModels; defModels != nil {
		if err := ValidateModelRef(defModels.Chat); defModels.Chat != "" && err != nil {
			return errors.Wrap(err, "default chat model")
		}
		if err := ValidateModelRef(defModels.CodeCompletion); defModels.CodeCompletion != "" && err != nil {
			return errors.Wrap(err, "default completion model")
		}
		if err := ValidateModelRef(defModels.FastChat); defModels.FastChat != "" && err != nil {
			return errors.Wrap(err, "default fast chat model")
		}
	}

	return nil
}
