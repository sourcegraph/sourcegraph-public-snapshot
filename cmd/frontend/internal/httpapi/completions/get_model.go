package completions

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/cody"
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/modelconfig"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/internal/completions/client/anthropic"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/fireworks"
	"github.com/sourcegraph/sourcegraph/internal/completions/client/google"
	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/internal/database"

	sgactor "github.com/sourcegraph/sourcegraph/internal/actor"
	modelconfigSDK "github.com/sourcegraph/sourcegraph/internal/modelconfig/types"
)

// legacyModelRef is a model reference that is of the form "provider/model" or even
// just "model". (::sigh::)
//
// The type is just to catch conversion errors while we roll out modelconfigSDK.ModelRef
// as the "one, true model reference" mechanism.
type legacyModelRef string

func (lmr legacyModelRef) Parse() (string, string) {
	parts := strings.Split(string(lmr), "/")
	switch len(parts) {
	case 2:
		return parts[0], parts[1]

	case 0, 1:
		// If there was no slash, then we just have a model name.
		// and have no idea about the provider.
		return "", string(lmr)

	default:
		// If there as _more_ than 2 slashes, then we are probably dealing
		// with an unfortunate case like "fireworks/accounts/fireworks/models/mixtral-8x7b-instruct".
		//
		// We cannot just return ("fireworks", "mixtral-8x7b-instruct") because
		// Cody Gateway actually expects the model name to be "accounts/fireworks/models/...".
		provider := parts[0]
		model := strings.TrimPrefix(string(lmr), provider+"/")
		return provider, model
	}
}

// EqualToIgnoringAPIVersion compares the legacyModelRef to the actual ModelRef,
// but ignoring the ModelRef's APIVersionID. (Because that's information the legacy
// format doesn't have.)
func (lmr legacyModelRef) EqualToIgnoringAPIVersion(mref modelconfigSDK.ModelRef) bool {
	if lmr == "" {
		return false
	}

	lmrProvider, lmrModel := lmr.Parse()
	mrefProvider := string(mref.ProviderID())
	mrefModel := string(mref.ModelID())

	// Only check the provider matches if the legacyModelRef had one. In other words,
	// if we only have "gpt-3.5-turbo" then we will consider it matching "anthropic::unknown::gpt-3.5-turbo".
	//
	// This will get more risky as backends support more models, with potentially overlapping
	// ModelIDs. But for now is safe enough.
	if lmrProvider != "" && lmrProvider != mrefProvider {
		return false
	}
	return lmrModel == mrefModel
}

// ToModelRef returns a BEST GUESS at the actual ModelRef for the
// LLM model. We don't know the APIVersionID, and there is no guarantee
// that the configured providers match the referenced name. So there will
// be bugs if various things don't align perfectly.
func (lmr legacyModelRef) ToModelRef() modelconfigSDK.ModelRef {
	provider, model := lmr.Parse()
	converted := fmt.Sprintf("%s::unknown::%s", provider, model)
	return modelconfigSDK.ModelRef(converted)
}

// toLegacyMRef converts a newer ModelRef to the older format.
// Required until we can remove all of the hard-coded model references
// in the codebase.
func toLegacyMRef(mref modelconfigSDK.ModelRef) legacyModelRef {
	olderFormat := fmt.Sprintf("%s/%s", mref.ProviderID(), mref.ModelID())
	return legacyModelRef(olderFormat)
}

// getModelFn is the thunk used to return the LLM model we should use for processing
// the supplied completion request. Depending on the incomming request, site config,
// feature used, etc. it could be any number of things.
//
// IMPORTANT: This returns a `ModelRef“, of the form "provider-id::api-verison-id::model-id",
// but for backwards compatibility currently expects to receive models in the older style
// `legacyModelRef“, of the form "provider-id/model-id".
type getModelFn func(
	ctx context.Context, requestParams types.CodyCompletionRequestParameters, c *modelconfigSDK.ModelConfiguration) (
	modelconfigSDK.ModelRef, error)

func getCodeCompletionModelFn() getModelFn {
	return func(
		_ context.Context, requestParams types.CodyCompletionRequestParameters, cfg *modelconfigSDK.ModelConfiguration) (
		modelconfigSDK.ModelRef, error) {
		if cfg == nil {
			return "", errors.New("no configuration data supplied")
		}

		// Default to model to use to the site's configuration if unspecified.
		initialRequestedModel := requestParams.RequestedModel
		if requestParams.RequestedModel == "" {
			requestParams.RequestedModel = types.TaintedModelRef(cfg.DefaultModels.CodeCompletion)
		}

		// We want to support newer clients sending fully-qualified ModelRefs, as well as older
		// clients using the legacy format. So we check if the incomming model reference is in either
		// format.
		var (
			mref       modelconfigSDK.ModelRef = modelconfigSDK.ModelRef(requestParams.RequestedModel)
			legacyMRef legacyModelRef          = legacyModelRef(requestParams.RequestedModel)
		)

		// BUG: When we have the ability within the site config to rely on Sourcegraph
		// supplied models, we can remove this step and rely on the data embedded in
		// the binary. (This requires us updating the site configuration for Sourcegraph.com,
		// and specifying the "modelconfig.sourcegraph" section, so the static siteconfig
		// will be used.)
		//
		// BUG: A side effect of this, until we make that change, Cody Pro users _cannot_ specify models
		// using the newer MRef syntax. As this check only looks for older "provider/model" names.
		if dotcom.SourcegraphDotComMode() {
			if isAllowedCodyProCompletionModel(toLegacyMRef(mref)) {
				return mref, nil
			}
			if isAllowedCodyProCompletionModel(legacyMRef) {
				return legacyMRef.ToModelRef(), nil
			}
			return "", errors.Errorf("unsupported Cody Pro completion model %q", legacyMRef)
		}

		// Now, for Cody Enterprise, if the caller requested a specific model we simply look
		// it up in the site config. By definition, if it is found then the model is allowed.
		for _, supportedModel := range cfg.Models {
			// Requested model was in the newer format.
			if supportedModel.ModelRef == mref {
				return mref, nil
			}
			// If the request is using the older format, then we are relying on the
			// assumption that the user-supplied legacyMRef's provider and model names
			// match the ProviderID and ModelID. (Likely, but may not be the case.)
			//
			// e.g. in order to support "fireworks/star-coder", we require that the
			// Sourcegraph instance has ap rovider named "fireworks" and a model with ID
			// "star-coder".
			if legacyMRef.EqualToIgnoringAPIVersion(supportedModel.ModelRef) {
				return supportedModel.ModelRef, nil
			}
		}

		err := errors.Errorf(
			"unsupported code completion model %q (default %q)",
			initialRequestedModel, cfg.DefaultModels.CodeCompletion)
		return "", err
	}
}

func getChatModelFn(db database.DB) getModelFn {
	return func(
		ctx context.Context, requestParams types.CodyCompletionRequestParameters, cfg *modelconfigSDK.ModelConfiguration) (
		modelconfigSDK.ModelRef, error) {
		// We want to support newer clients sending fully-qualified ModelRefs, as well as older
		// clients using the legacy format. So we check if the incomming model reference is in either
		// format.
		initialRequestedModel := requestParams.RequestedModel
		if requestParams.RequestedModel == "" {
			requestParams.RequestedModel = types.TaintedModelRef(cfg.DefaultModels.Chat)
		}
		var (
			mref       modelconfigSDK.ModelRef = modelconfigSDK.ModelRef(requestParams.RequestedModel)
			legacyMRef legacyModelRef          = legacyModelRef(requestParams.RequestedModel)
		)

		// If running on dotcom, i.e. using Cody Free/Cody Pro, then the available
		// models depend on the caller's subscription status.
		//
		// Like mentioned in a comment above, this logic is required until we update
		// the Sourcegraph.com site configuration to rely on staticly embedded data.
		if dotcom.SourcegraphDotComMode() {
			actor := sgactor.FromContext(ctx)
			user, err := actor.User(ctx, db.Users())
			if err != nil {
				return "", err
			}

			subscription, err := cody.SubscriptionForUser(ctx, db, *user)
			if err != nil {
				return "", err
			}

			// Note that Cody Pro users MUST specify the model to use on all requests.
			legacyMRef := legacyModelRef(requestParams.RequestedModel)
			if isAllowedCodyProChatModel(toLegacyMRef(mref), subscription.ApplyProRateLimits) {
				return mref, nil
			}
			if isAllowedCodyProChatModel(legacyMRef, subscription.ApplyProRateLimits) {
				return legacyMRef.ToModelRef(), nil
			}
			errModelNotAllowed := errors.Errorf(
				"the requested chat model is not available (%q, onProTier=%v)",
				requestParams.RequestedModel, subscription.ApplyProRateLimits)
			return "", errModelNotAllowed
		}

		// If FastChat is specified, we just use whatever the designated "fast" model is.
		// Otherwise, we try to find whatever model matches based on the default.
		if requestParams.Fast {
			return cfg.DefaultModels.FastChat, nil
		}

		// Now, for Cody Enterprise, if the caller requested a specific model we simply look
		// it up in the site config. By definition, if it is found then the model is allowed.
		for _, supportedModel := range cfg.Models {
			// Requested model was in the newer format.
			if supportedModel.ModelRef == mref {
				return mref, nil
			}
			// If the request is using the older format, then we are relying on the
			// assumption that the user-supplied legacyMRef's provider and model names
			// match the ProviderID and ModelID. (Likely, but may not be the case.)
			//
			// e.g. in order to support "fireworks/star-coder", we require that the
			// Sourcegraph instance has ap rovider named "fireworks" and a model with ID
			// "star-coder".
			if legacyMRef.EqualToIgnoringAPIVersion(supportedModel.ModelRef) {
				return supportedModel.ModelRef, nil
			}
		}

		err := errors.Errorf(
			"unsupported chat model %q (default %q)",
			initialRequestedModel, cfg.DefaultModels.Chat)
		return "", err
	}
}

// Returns whether or not Cody Pro users have access to the given model.
// See the comment on `isAllowedCodyProModelChatModel` why this function
// is required as we transition to using server-side LLM model configuration.
//
// BUG: This is temporary, and will be replaced when we support the site configuration
// allowing a Sourcegraph instance to fall back to "Sourcegraph supplied" LLM models.
//
// For now, because that isn't possible, all a Sourcegraph instance can use to determine
// which models are supported are the 3x that are put into the site's "completions config".
func isAllowedCodyProCompletionModel(model legacyModelRef) bool {
	switch model {
	case "fireworks/starcoder",
		"fireworks/starcoder-16b",
		"fireworks/starcoder-7b",
		"fireworks/starcoder2-15b",
		"fireworks/starcoder2-7b",
		"fireworks/" + fireworks.Starcoder16b,
		"fireworks/" + fireworks.Starcoder7b,
		"fireworks/" + fireworks.Llama27bCode,
		"fireworks/" + fireworks.Llama213bCode,
		"fireworks/" + fireworks.Llama213bCodeInstruct,
		"fireworks/" + fireworks.Llama234bCodeInstruct,
		"fireworks/" + fireworks.Mistral7bInstruct,
		"fireworks/" + fireworks.FineTunedFIMVariant1,
		"fireworks/" + fireworks.FineTunedFIMVariant2,
		"fireworks/" + fireworks.FineTunedFIMVariant3,
		"fireworks/" + fireworks.FineTunedFIMVariant4,
		"fireworks/" + fireworks.FineTunedFIMLangSpecificMixtral,
		"fireworks/" + fireworks.DeepseekCoder1p3b,
		"fireworks/" + fireworks.DeepseekCoder7b,
		"fireworks/" + fireworks.DeepseekCoderV2LiteBase,
		"fireworks/" + fireworks.CodeQwen7B,
		"anthropic/claude-instant-1.2",
		"anthropic/claude-3-haiku-20240307",
		// Deprecated model identifiers
		"anthropic/claude-instant-v1",
		"anthropic/claude-instant-1",
		"anthropic/claude-instant-1.2-cyan",
		"google/" + google.Gemini15Flash,
		"google/" + google.Gemini15FlashLatest,
		"google/" + google.Gemini15Flash001,
		"google/" + google.GeminiPro,
		"google/" + google.GeminiProLatest,
		"fireworks/accounts/sourcegraph/models/starcoder-7b",
		"fireworks/accounts/sourcegraph/models/starcoder-16b",
		"fireworks/accounts/fireworks/models/starcoder-3b-w8a16",
		"fireworks/accounts/fireworks/models/starcoder-1b-w8a16":
		return true
	}

	return false
}

// Returns whether or not the supplied model is available to Cody Pro users.
//
// BUG: This is temporary, and will be replaced when we support the site configuration
// allowing a Sourcegraph instance to fall back to "Sourcegraph supplied" LLM models.
//
// For now, because that isn't possible, all a Sourcegraph instance can use to determine
// which models are supported are the 3x that are put into the site's "completions config".
func isAllowedCodyProChatModel(model legacyModelRef, isProUser bool) bool {
	// When updating these two lists, make sure you also update `allowedModels` in codygateway_dotcom_user.go.
	if isProUser {
		switch model {
		case
			// Virtual model names used in the modelconfig data embedded into the binary.
			"anthropic/claude-3-haiku",
			"anthropic/claude-3-opus",
			"anthropic/claude-3-sonnet",
			"anthropic/claude-3.5-sonnet",
			// Models that have an even more confusing story.
			"mistral/mixtral-8x7b-instruct",
			"mistral/mixtral-8x22b-instruct",

			"anthropic/" + anthropic.Claude3Haiku,
			"anthropic/" + anthropic.Claude3Sonnet,
			"anthropic/" + anthropic.Claude35Sonnet,
			"anthropic/" + anthropic.Claude3Opus,
			"fireworks/" + fireworks.Mixtral8x7bInstruct,
			"fireworks/" + fireworks.Mixtral8x22Instruct,
			"openai/gpt-3.5-turbo",
			"openai/gpt-4o",
			"openai/gpt-4-turbo",
			"openai/gpt-4-turbo-preview",
			"google/" + google.Gemini15FlashLatest,
			"google/" + google.Gemini15ProLatest,
			"google/" + google.GeminiProLatest,
			"google/" + google.Gemini15Flash001,
			"google/" + google.Gemini15Pro001,
			"google/" + google.Gemini15Flash,
			"google/" + google.Gemini15Pro,
			"google/" + google.GeminiPro,

			// Remove after the Claude 3 rollout is complete
			"anthropic/claude-2",
			"anthropic/claude-2.0",
			"anthropic/claude-2.1",
			"anthropic/claude-instant-1.2-cyan",
			"anthropic/claude-instant-1.2",
			"anthropic/claude-instant-v1",
			"anthropic/claude-instant-1":
			return true
		}
	} else {
		// Models available to Cody Free users.
		switch model {
		case
			// Virtual model names used in the modelconfig data embedded into the binary.
			// (Opus is omitted, because it isn't available to Cody Free.)
			"anthropic/claude-3-haiku",
			"anthropic/claude-3-sonnet",
			"anthropic/claude-3.5-sonnet",
			// Models that have an even more confusing story.
			"mistral/mixtral-8x7b-instruct",
			"mistral/mixtral-8x22b-instruct",

			"anthropic/" + anthropic.Claude3Haiku,
			"anthropic/" + anthropic.Claude3Sonnet,
			"anthropic/" + anthropic.Claude35Sonnet,
			"fireworks/" + fireworks.Mixtral8x7bInstruct,
			"fireworks/" + fireworks.Mixtral8x22Instruct,
			"openai/gpt-3.5-turbo",
			"google/" + google.Gemini15FlashLatest,
			"google/" + google.Gemini15ProLatest,
			"google/" + google.GeminiProLatest,
			"google/" + google.Gemini15Flash,
			"google/" + google.Gemini15Pro,
			"google/" + google.GeminiPro,
			// Remove after the Claude 3 rollout is complete
			"anthropic/claude-2",
			"anthropic/claude-2.0",
			"anthropic/claude-instant-v1",
			"anthropic/claude-instant-1":
			return true
		}
	}

	return false
}

// virutalizedModelRefLookup is a super-lame hack that we can remove once Sourcegraph.com is only serving
// Sourcegraph-supplied LLM models. And we no longer have the hard-coded lists like `isAllowedCodyProChatModel`.
// See dotcom_models.go.
//
// Until then, we have a problem: the Sourcegraph-supplied models have a few instances where the ModelID and
// ModelName do not match. This is a good thing. But makes it a little tricky when for in the dotcom case the
// exact list of supported models is a bit murkier.
var virutalizedModelRefLookup = map[string]string{
	"claude-3-sonnet":        "claude-3-sonnet-20240229",
	"claude-3.5-sonnet":      "claude-3-5-sonnet-20240620",
	"claude-3-opus":          "claude-3-opus-20240229",
	"claude-3-haiku":         "claude-3-haiku-20240307",
	"mixtral-8x7b-instruct":  "accounts/fireworks/models/mixtral-8x7b-instruct",
	"mixtral-8x22b-instruct": "accounts/fireworks/models/mixtral-8x22b-instruct",
}

// resolveRequestedModel loads the provider and model configuration data for whatever model the user is requesting.
// Any errors returned are assumed to be user-facing, such as "you don't have access to model X", etc.
func resolveRequestedModel(
	ctx context.Context, logger log.Logger,
	cfg *modelconfigSDK.ModelConfiguration, request types.CodyCompletionRequestParameters, getModelFn getModelFn) (
	*modelconfigSDK.Provider, *modelconfigSDK.Model, error) {

	// Resolve the requested model.
	mref, err := getModelFn(ctx, request, cfg)
	if err != nil {
		return nil, nil, err
	}
	logger.Info(
		"resolved completion model",
		log.String("requestedModel", string(request.RequestedModel)),
		log.String("resolvedMRef", string(mref)))

	// SUPER SHADY HACK: Because right now we do NOT restrict Cody Pro models to be ONLY those defined in the
	// configuration data, it's very likely that the model and provider simply won't be found. So for the dotcom
	// case, the configuration data is kinda useless. dotcom sets the "completions.provider" to "sourcegraph",
	// and leaves the "chatModel" and related config to their defaults.
	//
	// So unfortunately we have to syntesize the Provider and Model objects dynamically. (And rely on the
	// Cody Gateway completion provider to not get fancy and look for any client-side configuration data.)
	if dotcom.SourcegraphDotComMode() {

		modelName := string(mref.ModelID())
		// Hack around Sourcegraph supplied models using "claude-3-sonnet" instead of "claude-3-sonnet-20240229".
		if devirtualizedModelName, ok := virutalizedModelRefLookup[modelName]; ok {
			modelName = devirtualizedModelName
		}

		fauxProvider := modelconfigSDK.Provider{
			ID: mref.ProviderID(),
			// If the instance is configured to be in dotcom mode, we assume the "completions.provider"
			// is "sourcegraph", and therefore the ServerSideConfig will be set correctly.
			// See `frontend/internal/modelconfig/siteconfig_completions_test.go`.
			ServerSideConfig: cfg.Providers[0].ServerSideConfig,
		}
		fauxModel := modelconfigSDK.Model{
			ModelRef:  mref,
			ModelName: modelName,
			// Leave everything invalid, even ContextWindow.
			// Which will for the time being be set within the
			// completion provider.
		}
		return &fauxProvider, &fauxModel, nil
	}

	// Look up the provider and model config from the configuration data available.
	if err := modelconfig.ValidateModelRef(mref); err != nil {
		// This shouldn't happen in-practice outside of unit tests, and is more to
		// catch bugs on our end.
		return nil, nil, errors.Wrapf(err, "getModelFn(%q) returned invalid mref", mref)
	}

	var (
		gotProvider *modelconfigSDK.Provider
		gotModel    *modelconfigSDK.Model
	)
	wantProviderID := mref.ProviderID()
	for i := range cfg.Providers {
		if cfg.Providers[i].ID == wantProviderID {
			gotProvider = &cfg.Providers[i]
			break
		}
	}
	if gotProvider == nil {
		return nil, nil, errors.Errorf("unable to find provider for mref %q", mref)
	}

	for i := range cfg.Models {
		if cfg.Models[i].ModelRef == mref {
			gotModel = &cfg.Models[i]
			break
		}
	}
	if gotModel == nil {
		return nil, nil, errors.Errorf("unable to find model %q", mref)
	}

	return gotProvider, gotModel, nil
}
