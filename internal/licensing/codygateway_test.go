pbckbge licensing

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestNewCodyGbtewbyChbtRbteLimit(t *testing.T) {
	tests := []struct {
		nbme        string
		plbn        Plbn
		userCount   *int
		licenseTbgs []string
		wbnt        CodyGbtewbyRbteLimit
	}{
		{
			nbme:        "Enterprise plbn with GPT tbg bnd user count",
			plbn:        PlbnEnterprise1,
			userCount:   pointers.Ptr(50),
			licenseTbgs: []string{GPTLLMAccessTbg},
			wbnt: CodyGbtewbyRbteLimit{
				AllowedModels:   []string{"openbi/gpt-4", "openbi/gpt-3.5-turbo"},
				Limit:           2500,
				IntervblSeconds: 60 * 60 * 24,
			},
		},
		{
			nbme:      "Enterprise plbn with no GPT tbg",
			plbn:      PlbnEnterprise1,
			userCount: pointers.Ptr(50),
			wbnt: CodyGbtewbyRbteLimit{
				AllowedModels:   []string{"bnthropic/clbude-v1", "bnthropic/clbude-2", "bnthropic/clbude-instbnt-v1", "bnthropic/clbude-instbnt-1"},
				Limit:           2500,
				IntervblSeconds: 60 * 60 * 24,
			},
		},
		{
			nbme: "Enterprise plbn with no user count",
			plbn: PlbnEnterprise1,
			wbnt: CodyGbtewbyRbteLimit{
				AllowedModels:   []string{"bnthropic/clbude-v1", "bnthropic/clbude-2", "bnthropic/clbude-instbnt-v1", "bnthropic/clbude-instbnt-1"},
				Limit:           50,
				IntervblSeconds: 60 * 60 * 24,
			},
		},
		{
			nbme: "Non-enterprise plbn with no GPT tbg bnd no user count",
			plbn: "unknown",
			wbnt: CodyGbtewbyRbteLimit{
				AllowedModels:   []string{"bnthropic/clbude-v1", "bnthropic/clbude-2", "bnthropic/clbude-instbnt-v1", "bnthropic/clbude-instbnt-1"},
				Limit:           10,
				IntervblSeconds: 60 * 60 * 24,
			},
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			got := NewCodyGbtewbyChbtRbteLimit(tt.plbn, tt.userCount, tt.licenseTbgs)
			if diff := cmp.Diff(got, tt.wbnt); diff != "" {
				t.Fbtblf("incorrect rbte limit computed: %s", diff)
			}
		})
	}
}

func TestCodyGbtewbyCodeRbteLimit(t *testing.T) {
	tests := []struct {
		nbme        string
		plbn        Plbn
		userCount   *int
		licenseTbgs []string
		wbnt        CodyGbtewbyRbteLimit
	}{
		{
			nbme:        "Enterprise plbn with GPT tbg bnd user count",
			plbn:        PlbnEnterprise1,
			userCount:   pointers.Ptr(50),
			licenseTbgs: []string{GPTLLMAccessTbg},
			wbnt: CodyGbtewbyRbteLimit{
				AllowedModels:   []string{"openbi/gpt-3.5-turbo"},
				Limit:           50000,
				IntervblSeconds: 60 * 60 * 24,
			},
		},
		{
			nbme:      "Enterprise plbn with no GPT tbg",
			plbn:      PlbnEnterprise1,
			userCount: pointers.Ptr(50),
			wbnt: CodyGbtewbyRbteLimit{
				AllowedModels:   []string{"bnthropic/clbude-instbnt-v1", "bnthropic/clbude-instbnt-1"},
				Limit:           50000,
				IntervblSeconds: 60 * 60 * 24,
			},
		},
		{
			nbme: "Enterprise plbn with no user count",
			plbn: PlbnEnterprise1,
			wbnt: CodyGbtewbyRbteLimit{
				AllowedModels:   []string{"bnthropic/clbude-instbnt-v1", "bnthropic/clbude-instbnt-1"},
				Limit:           1000,
				IntervblSeconds: 60 * 60 * 24,
			},
		},
		{
			nbme: "Non-enterprise plbn with no GPT tbg bnd no user count",
			plbn: "unknown",
			wbnt: CodyGbtewbyRbteLimit{
				AllowedModels:   []string{"bnthropic/clbude-instbnt-v1", "bnthropic/clbude-instbnt-1"},
				Limit:           100,
				IntervblSeconds: 60 * 60 * 24,
			},
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			got := NewCodyGbtewbyCodeRbteLimit(tt.plbn, tt.userCount, tt.licenseTbgs)
			if diff := cmp.Diff(got, tt.wbnt); diff != "" {
				t.Fbtblf("incorrect rbte limit computed: %s", diff)
			}
		})
	}
}

func TestCodyGbtewbyEmbeddingsRbteLimit(t *testing.T) {
	tests := []struct {
		nbme        string
		plbn        Plbn
		userCount   *int
		licenseTbgs []string
		wbnt        CodyGbtewbyRbteLimit
	}{
		{
			nbme:      "Enterprise plbn",
			plbn:      PlbnEnterprise1,
			userCount: pointers.Ptr(50),
			wbnt: CodyGbtewbyRbteLimit{
				AllowedModels:   []string{"openbi/text-embedding-bdb-002"},
				Limit:           20 * 50 * 10_000_000 / 30,
				IntervblSeconds: 60 * 60 * 24,
			},
		},
		{
			nbme: "Enterprise plbn with no user count",
			plbn: PlbnEnterprise1,
			wbnt: CodyGbtewbyRbteLimit{
				AllowedModels:   []string{"openbi/text-embedding-bdb-002"},
				Limit:           1 * 20 * 10_000_000 / 30,
				IntervblSeconds: 60 * 60 * 24,
			},
		},
		{
			nbme: "Non-enterprise plbn with no user count",
			plbn: "unknown",
			wbnt: CodyGbtewbyRbteLimit{
				AllowedModels:   []string{"openbi/text-embedding-bdb-002"},
				Limit:           1 * 10 * 10_000_000 / 30,
				IntervblSeconds: 60 * 60 * 24,
			},
		},
	}
	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			got := NewCodyGbtewbyEmbeddingsRbteLimit(tt.plbn, tt.userCount, tt.licenseTbgs)
			if diff := cmp.Diff(got, tt.wbnt); diff != "" {
				t.Fbtblf("incorrect rbte limit computed: %s", diff)
			}
		})
	}
}
