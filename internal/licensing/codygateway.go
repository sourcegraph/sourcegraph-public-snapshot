package licensing

import "golang.org/x/exp/slices"

// CodyGatewayRateLimit indicates rate limits for Sourcegraph's managed Cody Gateway service.
//
// Zero values in either field indicates no access.
type CodyGatewayRateLimit struct {
	// AllowedModels is a list of allowed models for the given feature in the
	// format "$PROVIDER/$MODEL_NAME", for example "anthropic/claude-2".
	AllowedModels []string

	Limit           int64
	IntervalSeconds int32
}

// NewCodyGatewayChatRateLimit applies default Cody Gateway access based on the plan.
func NewCodyGatewayChatRateLimit(plan Plan, userCount *int, licenseTags []string) CodyGatewayRateLimit {
	uc := 0
	if userCount != nil {
		uc = *userCount
	}
	if uc < 1 {
		uc = 1
	}
	// Switch on GPT models by default if the customer license has the GPT tag.
	models := []string{"anthropic/claude-v1", "anthropic/claude-2", "anthropic/claude-instant-v1", "anthropic/claude-instant-1"}
	if slices.Contains(licenseTags, GPTLLMAccessTag) {
		models = append(models, "openai/gpt-4", "openai/gpt-3.5-turbo")
	}
	switch plan {
	// TODO: This is just an example for now.
	case PlanEnterprise1,
		PlanEnterprise0:
		return CodyGatewayRateLimit{
			AllowedModels:   models,
			Limit:           int64(50 * uc),
			IntervalSeconds: 60 * 60 * 24, // day
		}

	// TODO: Defaults for other plans
	default:
		return CodyGatewayRateLimit{
			AllowedModels:   models,
			Limit:           int64(10 * uc),
			IntervalSeconds: 60 * 60 * 24, // day
		}
	}
}

// NewCodyGatewayCodeRateLimit applies default Cody Gateway access based on the plan.
func NewCodyGatewayCodeRateLimit(plan Plan, userCount *int, licenseTags []string) CodyGatewayRateLimit {
	uc := 0
	if userCount != nil {
		uc = *userCount
	}
	if uc < 1 {
		uc = 1
	}
	// Switch on GPT models by default if the customer license has the GPT tag.
	models := []string{"anthropic/claude-instant-v1", "anthropic/claude-instant-1"}
	if slices.Contains(licenseTags, GPTLLMAccessTag) {
		models = append(models, "openai/gpt-3.5-turbo")
	}
	switch plan {
	// TODO: This is just an example for now.
	case PlanEnterprise1,
		PlanEnterprise0:
		return CodyGatewayRateLimit{
			AllowedModels:   models,
			Limit:           int64(1000 * uc),
			IntervalSeconds: 60 * 60 * 24, // day
		}

	// TODO: Defaults for other plans
	default:
		return CodyGatewayRateLimit{
			AllowedModels:   models,
			Limit:           int64(100 * uc),
			IntervalSeconds: 60 * 60 * 24, // day
		}
	}
}

// tokensPerDollar is the number of tokens that will cost us roughly $1. It's used
// below for some better illustration of math.
const tokensPerDollar = int(1 / (0.0001 / 1_000))

// NewCodyGatewayEmbeddingsRateLimit applies default Cody Gateway access based on the plan.
func NewCodyGatewayEmbeddingsRateLimit(plan Plan, userCount *int, licenseTags []string) CodyGatewayRateLimit {
	uc := 0
	if userCount != nil {
		uc = *userCount
	}
	if uc < 1 {
		uc = 1
	}

	models := []string{"openai/text-embedding-ada-002"}
	switch plan {
	// TODO: This is just an example for now.
	case PlanEnterprise1,
		PlanEnterprise0:
		return CodyGatewayRateLimit{
			AllowedModels:   models,
			Limit:           int64(20 * uc * tokensPerDollar / 30),
			IntervalSeconds: 60 * 60 * 24, // day
		}

	// TODO: Defaults for other plans
	default:
		return CodyGatewayRateLimit{
			AllowedModels:   models,
			Limit:           int64(10 * uc * tokensPerDollar / 30),
			IntervalSeconds: 60 * 60 * 24, // day
		}
	}
}
