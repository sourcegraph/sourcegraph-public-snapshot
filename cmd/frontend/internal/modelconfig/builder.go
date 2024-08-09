package modelconfig

import (
	"fmt"
	"slices"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/modelconfig"
	"github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// builder implements the logic for constructing the Sourcegraph instance's
// LLM model configuration data, based on various configuration settings and available
// data.
type builder struct {
	// staticData is what is embedded into this binary, known at build-time.
	staticData *types.ModelConfiguration

	// codyGatewayData is what we have recently obtained by checking Cody Gateway
	// for any recent updates.
	//
	// TODO(PRIME-290): Implement this capability. Currently NYI.
	codyGatewayData *types.ModelConfiguration

	// siteConfigData is the data that is defined in the site configuration.
	// This is in a slightly different format to be more expressive than what
	// is provided by Cody Gateway or embedded in the binary.
	siteConfigData *types.SiteModelConfiguration
}

// build merges all of the model configuration data together, presenting it in
// its final form to be consumed by the Sourcegraph instance and passed onto its
// clients.
func (b *builder) build() (*types.ModelConfiguration, error) {
	baseConfig := &types.ModelConfiguration{}

	// Static data is by definition always available, but won't be provided in
	// some unit tests.
	if b.staticData != nil {
		baseConfig = b.staticData
	}

	// If we have newer data from Cody Gateway, use that instead of what is
	// baked into our codebase.
	if b.codyGatewayData != nil {
		baseConfig = b.codyGatewayData
	}

	// Interpret site configuration.

	// If no site configuration data is supplied, then just use Sourcegraph
	// supplied data.
	if b.siteConfigData == nil {
		routeUnconfiguredProvidersToCodyGateway(baseConfig)
		return baseConfig, nil
	}
	// But if we are using site config data, ensure it is valid before appying.
	if vErr := modelconfig.ValidateSiteConfig(b.siteConfigData); vErr != nil {
		return nil, errors.Wrap(vErr, "invalid site configuration")
	}

	outConfig, err := applySiteConfig(baseConfig, b.siteConfigData)
	if err != nil {
		return nil, errors.Wrap(err, "applying site config")
	}

	return outConfig, nil
}

// applySiteConfig returns the LLM Model configuration after applying the Sourcegraph admin supplied site config overrides.
func applySiteConfig(baseConfig *types.ModelConfiguration, siteConfig *types.SiteModelConfiguration) (*types.ModelConfiguration, error) {
	if baseConfig == nil || siteConfig == nil {
		return nil, errors.New("baseConfig or siteConfig nil")
	}

	// We initialize the merged configuration data in-place.
	var (
		mergedConfig *types.ModelConfiguration
		err          error
	)

	// If the admin has explicitly disabled the Sourcegraph-supplied data, zero out the base config.
	if sgModelConfig := siteConfig.SourcegraphModelConfig; sgModelConfig == nil {
		mergedConfig = &types.ModelConfiguration{
			Revision:      "-",
			SchemaVersion: types.CurrentModelSchemaVersion,

			// No Models or Providers.
			Providers: nil,
			Models:    nil,

			// Don't provide any DefaultModels either.
			//
			// These instead need to come from the siteConfig, or be inferred from
			// the models available.
			DefaultModels: types.DefaultModels{},
		}
	} else {
		// Apply any model filters from the base configuration. We start with copying the base config
		// so we can just mutate it in-memory.
		mergedConfig, err = deepCopy(baseConfig)
		if err != nil {
			return nil, errors.New("copying base config")
		}

		// Apply any admin-defined model filters.
		if modelFilters := sgModelConfig.ModelFilters; modelFilters != nil {
			var filteredModels []types.Model
			for _, baseConfigModel := range mergedConfig.Models {
				if isModelAllowed(baseConfigModel, *modelFilters) {
					filteredModels = append(filteredModels, baseConfigModel)
				}
			}

			// Replace the base config models with the filtered set.
			mergedConfig.Models = filteredModels
		}
	}

	// Filter out deprecated models from the merged configuration.
	if models := mergedConfig.Models; models != nil {
		var filteredModels []types.Model
		for _, model := range models {
			if model.Status != types.ModelStatusDeprecated {
				filteredModels = append(filteredModels, model)
			}
		}
		mergedConfig.Models = filteredModels
	}

	// Apply any ProviderOverrides from the site configuration to the mergedConfig object.
	providerOverrideLookup := map[types.ProviderID]*types.ProviderOverride{}
	for i := range siteConfig.ProviderOverrides {
		providerOverride := &siteConfig.ProviderOverrides[i]
		providerOverrideLookup[providerOverride.ID] = providerOverride

		// Lookup the provider this configuration is for.
		var providerToOverride *types.Provider
		for mergedProvIdx := range mergedConfig.Providers {
			if mergedConfig.Providers[mergedProvIdx].ID == providerOverride.ID {
				providerToOverride = &mergedConfig.Providers[mergedProvIdx]
				break
			}
		}

		// The site configuration has an override for a provider that
		// isn't in the base config. So it is "new" and defined exclusively
		// in the site configuration.
		if providerToOverride == nil {

			displayName := providerOverride.DisplayName
			if displayName == "" {
				displayName = fmt.Sprintf("Provider %q", providerOverride.ID)
			}
			providerToOverride = &types.Provider{
				ID:               providerOverride.ID,
				DisplayName:      displayName,
				ClientSideConfig: providerOverride.ClientSideConfig,
				ServerSideConfig: providerOverride.ServerSideConfig,
			}
			mergedConfig.Providers = append(mergedConfig.Providers, *providerToOverride)
		}

		// Blow away the provider*TO*override's configuration with the
		// provider override defined in the site config.
		providerToOverride.ClientSideConfig = providerOverride.ClientSideConfig
		providerToOverride.ServerSideConfig = providerOverride.ServerSideConfig
	}

	// Apply Model Overrides. Since we need to apply any ProviderOverride.DefaultModelConfig,
	// we just build a lookup and add any entries to the mergedConfig.Models. We can actually
	// set their fields later.
	modelOverrideLookup := map[types.ModelRef]*types.ModelOverride{}
	for i := range siteConfig.ModelOverrides {
		modelOverride := &siteConfig.ModelOverrides[i]
		modelOverrideLookup[modelOverride.ModelRef] = modelOverride
	}

	// Now loop through all baseConfig models, and apply the override or provider defaults.
	for i := range mergedConfig.Models {
		mod := &mergedConfig.Models[i]

		// If this model is associated with one of the ProviderOverrides, then fetch
		// its DefaultModelConfig.
		var providerDefaultModelConfig *types.DefaultModelConfig
		modelProviderID := mod.ModelRef.ProviderID()
		if providerOverride := providerOverrideLookup[modelProviderID]; providerOverride != nil {
			providerDefaultModelConfig = providerOverride.DefaultModelConfig
		}

		// Apply the Provider's DefaultModelConfig, if applicable. (Is no-op if DefaultModelConfig is nil.)
		if err := modelconfig.ApplyDefaultModelConfiguration(mod, providerDefaultModelConfig); err != nil {
			return nil, errors.Wrapf(err, "applying provider default model config (%q)", modelProviderID)
		}

		// Next, apply any specific ModelOverride data for that model.
		if modelOverride := modelOverrideLookup[mod.ModelRef]; modelOverride != nil {
			if err := modelconfig.ApplyModelOverride(mod, *modelOverride); err != nil {
				return nil, errors.Wrapf(err, "applying model override (%q)", mod.ModelRef)
			}

			// Remove the key from the modelOverrideLookup, see below.
			delete(modelOverrideLookup, mod.ModelRef)
		}
	}

	// For any providers still missing server-side config, wire them to send
	// requests to Cody Gateway. We do this BEFORE merging in model/provider
	// overrides so that new providers defined by the admin don't get wired
	// to Cody Gateway. (And instead will fail at runtime because no server-side
	// config is available.)
	routeUnconfiguredProvidersToCodyGateway(mergedConfig)

	// If there are remaining keys in `modelOverrideLookup` means that the are for a ModelRef that
	// was NOT found in the base configuration. So in that case we add those as "entirely new" models
	// that were only defined in the site config, and wasn't referenced in the base config.
	for _, modelOverride := range modelOverrideLookup {
		newModelRef := modelOverride.ModelRef
		newModel := &types.Model{
			ModelRef: newModelRef,
			// This isn't to provide a "default" so much as it is just to
			// ensure the model will work.
			ContextWindow: types.ContextWindow{
				MaxInputTokens:  4_000,
				MaxOutputTokens: 4_000,
			},
		}

		// Lookup and apply the model's provider's DefaultModelConfig, if applicable.
		modelProviderID := newModelRef.ProviderID()
		if providerOverride := providerOverrideLookup[modelProviderID]; providerOverride != nil {
			providerDefaultModelConfig := providerOverride.DefaultModelConfig

			err := modelconfig.ApplyDefaultModelConfiguration(newModel, providerDefaultModelConfig)
			if err != nil {
				return nil, errors.Wrapf(err, "applying default provider config (%q)", modelProviderID)
			}
		}

		// Apply the ModelOverride from the site config to the new Model object we are building.
		if err := modelconfig.ApplyModelOverride(newModel, *modelOverride); err != nil {
			return nil, errors.Wrapf(err, "applying model override (%q)", newModelRef)
		}

		mergedConfig.Models = append(mergedConfig.Models, *newModel)
	}

	// Have any admin-supplied default models overwrite what may have been set by
	// the base config.
	if siteConfig.DefaultModels != nil {
		mergedConfig.DefaultModels.Chat = siteConfig.DefaultModels.Chat
		mergedConfig.DefaultModels.CodeCompletion = siteConfig.DefaultModels.CodeCompletion
		mergedConfig.DefaultModels.FastChat = siteConfig.DefaultModels.FastChat
	}

	// But we still need to confirm that all the default models are actually valid.
	// e.g. not filtered out because of the model filter, or have an invalid ModelRef.
	if err := maybeFixDefaultModels(&mergedConfig.DefaultModels, mergedConfig.Models); err != nil {
		return nil, err
	}

	// Validate the resulting configuration.
	if err := modelconfig.ValidateModelConfig(mergedConfig); err != nil {
		return nil, errors.Wrap(err, "result of application was invalid configuration")
	}
	return mergedConfig, nil
}

// isModelAllowed returns whether or not the model should be supported as per the
// supplied model filter configuration.
func isModelAllowed(m types.Model, filter types.ModelFilters) bool {
	// Status filter. Exclude any models whose status doesn't match what is required.
	if filter.StatusFilter != nil {
		if !slices.Contains(filter.StatusFilter, string(m.Status)) {
			return false
		}
	}

	// Allow list. If not specified, include all models. Otherwise, ONLY include the model
	// IFF it matches one of the allow rules.
	if len(filter.Allow) > 0 {
		if !filterListMatches(m.ModelRef, filter.Allow) {
			return false
		}
	}
	// Deny list. Filter the model if it matches any deny rules.
	if len(filter.Deny) > 0 {
		if filterListMatches(m.ModelRef, filter.Deny) {
			return false
		}
	}

	return true
}

// maybeFixDefaultModels will verify that the supplied DefaultModels set is valid, possibly
// modifying a model in case the existing value was incorrect.
func maybeFixDefaultModels(defaultModels *types.DefaultModels, allModels []types.Model) error {
	// getModelWithRequirements returns the the first model available with the specific capability and a matching
	// category. Returns nil if no such model is found.
	getModelWithRequirements := func(
		wantCapability types.ModelCapability, wantCategories ...types.ModelCategory) *types.ModelRef {
		for _, model := range allModels {
			// Check if the model can be used for that purpose.
			var hasCapability bool
			for _, gotCapability := range model.Capabilities {
				if gotCapability == wantCapability {
					hasCapability = true
					break
				}
			}
			if !hasCapability {
				continue
			}

			// Check if the model has a matching category.
			for _, wantCategory := range wantCategories {
				if model.Category == wantCategory {
					return &model.ModelRef
				}
			}
		}
		return nil
	}

	modelFound := func(mref types.ModelRef) bool {
		for _, model := range allModels {
			if model.ModelRef == mref {
				return true
			}
		}
		return false
	}

	const (
		accuracy = types.ModelCategoryAccuracy
		balanced = types.ModelCategoryBalanced
		speed    = types.ModelCategorySpeed
	)

	// Infer the default models to used based on category. This is probably not going to lead to great
	// results. But :shrug: it's better than just crash looping because the config is under-specified.
	if defaultModels.Chat == "" || !modelFound(defaultModels.Chat) {
		validModelRef := getModelWithRequirements(types.ModelCapabilityChat, accuracy, balanced)
		if validModelRef == nil {
			return errors.Errorf("no suitable model found for Chat (%d candidates)", len(allModels))
		}
		defaultModels.Chat = *validModelRef
	}
	if defaultModels.FastChat == "" || !modelFound(defaultModels.FastChat) {
		validModelRef := getModelWithRequirements(types.ModelCapabilityChat, speed, balanced)
		if validModelRef == nil {
			return errors.Errorf("no suitable model found for FastChat (%d candidates)", len(allModels))
		}
		defaultModels.FastChat = *validModelRef
	}
	if defaultModels.CodeCompletion == "" || !modelFound(defaultModels.CodeCompletion) {
		validModelRef := getModelWithRequirements(types.ModelCapabilityAutocomplete, speed, balanced)
		if validModelRef == nil {
			return errors.Errorf("no suitable model found for CodeCompletion (%d candidates)", len(allModels))
		}
		defaultModels.CodeCompletion = *validModelRef
	}

	return nil
}

// routeUnconfiguredProvidersToCodyGateway verifies that each provider has server-side configuration
// data present. (So that it can actually serve LLM traffic.) However, for any providers missing
// server-side configuration, they will be updated to send traffic to Cody Gateway.
func routeUnconfiguredProvidersToCodyGateway(config *types.ModelConfiguration) {
	siteConfig := conf.Get()

	endpoint := conf.CodyGatewayProdEndpoint
	accessToken := license.GenerateLicenseKeyBasedAccessToken(siteConfig.LicenseKey)

	// Apply any overrides if present in the site config.
	if siteConfig.ModelConfiguration != nil && siteConfig.ModelConfiguration.Sourcegraph != nil {
		sgConfig := siteConfig.ModelConfiguration.Sourcegraph
		if sgConfig.AccessToken != nil {
			accessToken = *sgConfig.AccessToken
		}
		if sgConfig.Endpoint != nil {
			endpoint = *sgConfig.Endpoint
		}
	}

	for i := range config.Providers {
		provider := &config.Providers[i]

		if provider.ServerSideConfig == nil {
			provider.ServerSideConfig = &types.ServerSideProviderConfig{
				SourcegraphProvider: &types.SourcegraphProviderConfig{
					AccessToken: accessToken,
					Endpoint:    endpoint,
				},
			}
		}
	}
}
