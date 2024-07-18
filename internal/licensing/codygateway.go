package licensing

import (
	"time"
)

// CodyGatewayRateLimit indicates rate limits for Sourcegraph's managed Cody Gateway service.
//
// Zero values in either field indicates no access.
type CodyGatewayRateLimit struct {
	// AllowedModels is a list of allowed models for the given feature in the
	// format "$PROVIDER/$MODEL_NAME", for example "anthropic/claude-2".
	// A single-item slice with value '*' means that all models in the Cody
	// Gateway allowlist are allowed.
	AllowedModels []string

	Limit           int64
	IntervalSeconds int32
}

func (r CodyGatewayRateLimit) IntervalDuration() time.Duration {
	return time.Duration(r.IntervalSeconds) * time.Second
}

// NewCodyGatewayChatRateLimit applies default Cody Gateway access based on the plan.
func NewCodyGatewayChatRateLimit(plan Plan, userCount *int) CodyGatewayRateLimit {
	uc := 0
	if userCount != nil {
		uc = *userCount
	}
	if uc < 1 {
		uc = 1
	}
	models := []string{"*"} // allow all models that are allowlisted by Cody Gateway
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
func NewCodyGatewayCodeRateLimit(plan Plan, userCount *int) CodyGatewayRateLimit {
	uc := 0
	if userCount != nil {
		uc = *userCount
	}
	if uc < 1 {
		uc = 1
	}
	models := []string{"*"} // allow all models allowlisted by Cody Gateway
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
			Limit:           int64(1000 * uc),
			IntervalSeconds: 60 * 60 * 24, // day
		}
	}
}

// tokensPerDollar is the number of tokens that will cost us roughly $1. It's used
// below for some better illustration of math.
const tokensPerDollar = int(1 / (0.0001 / 1_000))

// NewCodyGatewayEmbeddingsRateLimit applies default Cody Gateway access based on the plan.
func NewCodyGatewayEmbeddingsRateLimit(plan Plan, userCount *int) CodyGatewayRateLimit {
	uc := 0
	if userCount != nil {
		uc = *userCount
	}
	if uc < 1 {
		uc = 1
	}

	models := []string{"*"} // allow all models allowlisted by Cody Gateway
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
