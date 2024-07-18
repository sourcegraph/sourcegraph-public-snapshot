package licensing

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func TestNewCodyGatewayChatRateLimit(t *testing.T) {
	tests := []struct {
		name      string
		plan      Plan
		userCount *int
		want      CodyGatewayRateLimit
	}{
		{
			name:      "Enterprise plan with user count",
			plan:      PlanEnterprise1,
			userCount: pointers.Ptr(50),
			want: CodyGatewayRateLimit{
				AllowedModels:   []string{"*"},
				Limit:           2500,
				IntervalSeconds: 60 * 60 * 24,
			},
		},
		{
			name: "Enterprise plan with no user count",
			plan: PlanEnterprise1,
			want: CodyGatewayRateLimit{
				AllowedModels:   []string{"*"},
				Limit:           50,
				IntervalSeconds: 60 * 60 * 24,
			},
		},
		{
			name: "Non-enterprise plan with no user count",
			plan: "unknown",
			want: CodyGatewayRateLimit{
				AllowedModels:   []string{"*"},
				Limit:           10,
				IntervalSeconds: 60 * 60 * 24,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCodyGatewayChatRateLimit(tt.plan, tt.userCount)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Fatalf("incorrect rate limit computed: %s", diff)
			}
		})
	}
}

func TestCodyGatewayCodeRateLimit(t *testing.T) {
	tests := []struct {
		name        string
		plan        Plan
		userCount   *int
		licenseTags []string
		want        CodyGatewayRateLimit
	}{
		{
			name:        "Enterprise plan with legacy GPT tag and user count",
			plan:        PlanEnterprise1,
			userCount:   pointers.Ptr(50),
			licenseTags: []string{"gpt"},
			want: CodyGatewayRateLimit{
				AllowedModels:   []string{"*"},
				Limit:           50000,
				IntervalSeconds: 60 * 60 * 24,
			},
		},
		{
			name:      "Enterprise plan with no GPT tag",
			plan:      PlanEnterprise1,
			userCount: pointers.Ptr(50),
			want: CodyGatewayRateLimit{
				AllowedModels:   []string{"*"},
				Limit:           50000,
				IntervalSeconds: 60 * 60 * 24,
			},
		},
		{
			name: "Enterprise plan with no user count",
			plan: PlanEnterprise1,
			want: CodyGatewayRateLimit{
				AllowedModels:   []string{"*"},
				Limit:           1000,
				IntervalSeconds: 60 * 60 * 24,
			},
		},
		{
			name: "Non-enterprise plan with no GPT tag and no user count",
			plan: "unknown",
			want: CodyGatewayRateLimit{
				AllowedModels:   []string{"*"},
				Limit:           1000,
				IntervalSeconds: 60 * 60 * 24,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCodyGatewayCodeRateLimit(tt.plan, tt.userCount)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Fatalf("incorrect rate limit computed: %s", diff)
			}
		})
	}
}

func TestCodyGatewayEmbeddingsRateLimit(t *testing.T) {
	tests := []struct {
		name        string
		plan        Plan
		userCount   *int
		licenseTags []string
		want        CodyGatewayRateLimit
	}{
		{
			name:      "Enterprise plan",
			plan:      PlanEnterprise1,
			userCount: pointers.Ptr(50),
			want: CodyGatewayRateLimit{
				AllowedModels:   []string{"*"},
				Limit:           20 * 50 * 10_000_000 / 30,
				IntervalSeconds: 60 * 60 * 24,
			},
		},
		{
			name: "Enterprise plan with no user count",
			plan: PlanEnterprise1,
			want: CodyGatewayRateLimit{
				AllowedModels:   []string{"*"},
				Limit:           1 * 20 * 10_000_000 / 30,
				IntervalSeconds: 60 * 60 * 24,
			},
		},
		{
			name: "Non-enterprise plan with no user count",
			plan: "unknown",
			want: CodyGatewayRateLimit{
				AllowedModels:   []string{"*"},
				Limit:           1 * 10 * 10_000_000 / 30,
				IntervalSeconds: 60 * 60 * 24,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewCodyGatewayEmbeddingsRateLimit(tt.plan, tt.userCount)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Fatalf("incorrect rate limit computed: %s", diff)
			}
		})
	}
}
