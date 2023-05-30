package licensing

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewLLMProxyChatRateLimit(t *testing.T) {
	tests := []struct {
		name        string
		plan        Plan
		userCount   *int
		licenseTags []string
		want        LLMProxyRateLimit
	}{
		{
			name:        "Enterprise plan with GPT tag and user count",
			plan:        PlanEnterprise1,
			userCount:   intPtr(50),
			licenseTags: []string{GPTLLMAccessTag},
			want: LLMProxyRateLimit{
				AllowedModels:   []string{"gpt-4", "gpt-3.5-turbo"},
				Limit:           2500,
				IntervalSeconds: 60 * 60 * 24,
			},
		},
		{
			name:      "Enterprise plan with no GPT tag",
			plan:      PlanEnterprise1,
			userCount: intPtr(50),
			want: LLMProxyRateLimit{
				AllowedModels:   []string{"claude-v1", "claude-instant-v1"},
				Limit:           2500,
				IntervalSeconds: 60 * 60 * 24,
			},
		},
		{
			name: "Enterprise plan with no user count",
			plan: PlanEnterprise1,
			want: LLMProxyRateLimit{
				AllowedModels:   []string{"claude-v1", "claude-instant-v1"},
				Limit:           50,
				IntervalSeconds: 60 * 60 * 24,
			},
		},
		{
			name: "Non-enterprise plan with no GPT tag and no user count",
			plan: "unknown",
			want: LLMProxyRateLimit{
				AllowedModels:   []string{"claude-v1", "claude-instant-v1"},
				Limit:           10,
				IntervalSeconds: 60 * 60 * 24,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewLLMProxyChatRateLimit(tt.plan, tt.userCount, tt.licenseTags)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Fatalf("incorrect rate limit computed: %s", diff)
			}
		})
	}
}

func TestNewLLMProxyCodeRateLimit(t *testing.T) {
	tests := []struct {
		name        string
		plan        Plan
		userCount   *int
		licenseTags []string
		want        LLMProxyRateLimit
	}{
		{
			name:        "Enterprise plan with GPT tag and user count",
			plan:        PlanEnterprise1,
			userCount:   intPtr(50),
			licenseTags: []string{GPTLLMAccessTag},
			want: LLMProxyRateLimit{
				AllowedModels:   []string{"gpt-3.5-turbo"},
				Limit:           25000,
				IntervalSeconds: 60 * 60 * 24,
			},
		},
		{
			name:      "Enterprise plan with no GPT tag",
			plan:      PlanEnterprise1,
			userCount: intPtr(50),
			want: LLMProxyRateLimit{
				AllowedModels:   []string{"claude-instant-v1"},
				Limit:           25000,
				IntervalSeconds: 60 * 60 * 24,
			},
		},
		{
			name: "Enterprise plan with no user count",
			plan: PlanEnterprise1,
			want: LLMProxyRateLimit{
				AllowedModels:   []string{"claude-instant-v1"},
				Limit:           500,
				IntervalSeconds: 60 * 60 * 24,
			},
		},
		{
			name: "Non-enterprise plan with no GPT tag and no user count",
			plan: "unknown",
			want: LLMProxyRateLimit{
				AllowedModels:   []string{"claude-instant-v1"},
				Limit:           100,
				IntervalSeconds: 60 * 60 * 24,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewLLMProxyCodeRateLimit(tt.plan, tt.userCount, tt.licenseTags)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Fatalf("incorrect rate limit computed: %s", diff)
			}
		})
	}
}

func intPtr(i int) *int { return &i }
