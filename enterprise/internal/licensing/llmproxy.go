package licensing

import "golang.org/x/exp/slices"

// LLMProxyRateLimit indicates rate limits for Sourcegraph's managed LLM-proxy service.
//
// Zero values in either field indicates no access.
type LLMProxyRateLimit struct {
	AllowedModels   []string
	Limit           int32
	IntervalSeconds int32
}

// NewLLMProxyChatRateLimit applies default LLM-proxy access based on the plan.
func NewLLMProxyChatRateLimit(plan Plan, userCount *int, licenseTags []string) LLMProxyRateLimit {
	uc := 0
	if userCount != nil {
		uc = *userCount
	}
	if uc < 1 {
		uc = 1
	}
	// Switch on GPT models by default if the customer license has the GPT tag.
	models := []string{"claude-v1", "claude-instant-v1"}
	if slices.Contains(licenseTags, GPTLLMAccessTag) {
		models = []string{"gpt-4", "gpt-3.5-turbo"}
	}
	switch plan {
	// TODO: This is just an example for now.
	case PlanEnterprise1,
		PlanEnterprise0:
		return LLMProxyRateLimit{
			AllowedModels:   models,
			Limit:           int32(50 * uc),
			IntervalSeconds: 60 * 60 * 24, // day
		}

	// TODO: Defaults for other plans
	default:
		return LLMProxyRateLimit{
			AllowedModels:   models,
			Limit:           int32(10 * uc),
			IntervalSeconds: 60 * 60 * 24, // day
		}
	}
}

// NewLLMProxyCodeRateLimit applies default LLM-proxy access based on the plan.
func NewLLMProxyCodeRateLimit(plan Plan, userCount *int, licenseTags []string) LLMProxyRateLimit {
	uc := 0
	if userCount != nil {
		uc = *userCount
	}
	if uc < 1 {
		uc = 1
	}
	// Switch on GPT models by default if the customer license has the GPT tag.
	models := []string{"claude-instant-v1"}
	if slices.Contains(licenseTags, GPTLLMAccessTag) {
		models = []string{"gpt-3.5-turbo"}
	}
	switch plan {
	// TODO: This is just an example for now.
	case PlanEnterprise1,
		PlanEnterprise0:
		return LLMProxyRateLimit{
			AllowedModels:   models,
			Limit:           int32(500 * uc),
			IntervalSeconds: 60 * 60 * 24, // day
		}

	// TODO: Defaults for other plans
	default:
		return LLMProxyRateLimit{
			AllowedModels:   models,
			Limit:           int32(100 * uc),
			IntervalSeconds: 60 * 60 * 24, // day
		}
	}
}
