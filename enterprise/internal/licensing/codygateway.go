package licensing

import "golang.org/x/exp/slices"

// CodyGatewayRateLimit indicates rate limits for Sourcegraph's managed Cody Gateway service.
//
// Zero values in either field indicates no access.
type CodyGatewayRateLimit struct {
	AllowedModels   []string
	Limit           int32
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
	models := []string{"claude-v1", "claude-instant-v1"}
	if slices.Contains(licenseTags, GPTLLMAccessTag) {
		models = []string{"gpt-4", "gpt-3.5-turbo"}
	}
	switch plan {
	// TODO: This is just an example for now.
	case PlanEnterprise1,
		PlanEnterprise0:
		return CodyGatewayRateLimit{
			AllowedModels:   models,
			Limit:           int32(50 * uc),
			IntervalSeconds: 60 * 60 * 24, // day
		}

	// TODO: Defaults for other plans
	default:
		return CodyGatewayRateLimit{
			AllowedModels:   models,
			Limit:           int32(10 * uc),
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
	models := []string{"claude-instant-v1"}
	if slices.Contains(licenseTags, GPTLLMAccessTag) {
		models = []string{"gpt-3.5-turbo"}
	}
	switch plan {
	// TODO: This is just an example for now.
	case PlanEnterprise1,
		PlanEnterprise0:
		return CodyGatewayRateLimit{
			AllowedModels:   models,
			Limit:           int32(500 * uc),
			IntervalSeconds: 60 * 60 * 24, // day
		}

	// TODO: Defaults for other plans
	default:
		return CodyGatewayRateLimit{
			AllowedModels:   models,
			Limit:           int32(100 * uc),
			IntervalSeconds: 60 * 60 * 24, // day
		}
	}
}
