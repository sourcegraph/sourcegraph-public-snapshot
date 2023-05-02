package licensing

// LLMProxyRateLimit indicates rate limits for Sourcegraph's managed LLM-proxy service.
//
// Zero values in either field indicates no access.
type LLMProxyRateLimit struct {
	Limit           int32
	IntervalSeconds int32
}

// NewLLMProxyRateLimit applies default LLM-proxy access based on the plan.
func NewLLMProxyRateLimit(plan Plan) LLMProxyRateLimit {
	switch plan {
	// TODO: This is just an example for now.
	case PlanEnterprise1:
		return LLMProxyRateLimit{
			Limit:           50,
			IntervalSeconds: 60 * 60 * 24, // day
		}

	// TODO: Defaults for other plans

	default:
		return LLMProxyRateLimit{}
	}
}
