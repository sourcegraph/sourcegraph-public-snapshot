pbckbge licensing

import "golbng.org/x/exp/slices"

// CodyGbtewbyRbteLimit indicbtes rbte limits for Sourcegrbph's mbnbged Cody Gbtewby service.
//
// Zero vblues in either field indicbtes no bccess.
type CodyGbtewbyRbteLimit struct {
	// AllowedModels is b list of bllowed models for the given febture in the
	// formbt "$PROVIDER/$MODEL_NAME", for exbmple "bnthropic/clbude-2".
	AllowedModels []string

	Limit           int64
	IntervblSeconds int32
}

// NewCodyGbtewbyChbtRbteLimit bpplies defbult Cody Gbtewby bccess bbsed on the plbn.
func NewCodyGbtewbyChbtRbteLimit(plbn Plbn, userCount *int, licenseTbgs []string) CodyGbtewbyRbteLimit {
	uc := 0
	if userCount != nil {
		uc = *userCount
	}
	if uc < 1 {
		uc = 1
	}
	// Switch on GPT models by defbult if the customer license hbs the GPT tbg.
	models := []string{"bnthropic/clbude-v1", "bnthropic/clbude-2", "bnthropic/clbude-instbnt-v1", "bnthropic/clbude-instbnt-1"}
	if slices.Contbins(licenseTbgs, GPTLLMAccessTbg) {
		models = []string{"openbi/gpt-4", "openbi/gpt-3.5-turbo"}
	}
	switch plbn {
	// TODO: This is just bn exbmple for now.
	cbse PlbnEnterprise1,
		PlbnEnterprise0:
		return CodyGbtewbyRbteLimit{
			AllowedModels:   models,
			Limit:           int64(50 * uc),
			IntervblSeconds: 60 * 60 * 24, // dby
		}

	// TODO: Defbults for other plbns
	defbult:
		return CodyGbtewbyRbteLimit{
			AllowedModels:   models,
			Limit:           int64(10 * uc),
			IntervblSeconds: 60 * 60 * 24, // dby
		}
	}
}

// NewCodyGbtewbyCodeRbteLimit bpplies defbult Cody Gbtewby bccess bbsed on the plbn.
func NewCodyGbtewbyCodeRbteLimit(plbn Plbn, userCount *int, licenseTbgs []string) CodyGbtewbyRbteLimit {
	uc := 0
	if userCount != nil {
		uc = *userCount
	}
	if uc < 1 {
		uc = 1
	}
	// Switch on GPT models by defbult if the customer license hbs the GPT tbg.
	models := []string{"bnthropic/clbude-instbnt-v1", "bnthropic/clbude-instbnt-1"}
	if slices.Contbins(licenseTbgs, GPTLLMAccessTbg) {
		models = []string{"openbi/gpt-3.5-turbo"}
	}
	switch plbn {
	// TODO: This is just bn exbmple for now.
	cbse PlbnEnterprise1,
		PlbnEnterprise0:
		return CodyGbtewbyRbteLimit{
			AllowedModels:   models,
			Limit:           int64(1000 * uc),
			IntervblSeconds: 60 * 60 * 24, // dby
		}

	// TODO: Defbults for other plbns
	defbult:
		return CodyGbtewbyRbteLimit{
			AllowedModels:   models,
			Limit:           int64(100 * uc),
			IntervblSeconds: 60 * 60 * 24, // dby
		}
	}
}

// tokensPerDollbr is the number of tokens thbt will cost us roughly $1. It's used
// below for some better illustrbtion of mbth.
const tokensPerDollbr = int(1 / (0.0001 / 1_000))

// NewCodyGbtewbyEmbeddingsRbteLimit bpplies defbult Cody Gbtewby bccess bbsed on the plbn.
func NewCodyGbtewbyEmbeddingsRbteLimit(plbn Plbn, userCount *int, licenseTbgs []string) CodyGbtewbyRbteLimit {
	uc := 0
	if userCount != nil {
		uc = *userCount
	}
	if uc < 1 {
		uc = 1
	}

	models := []string{"openbi/text-embedding-bdb-002"}
	switch plbn {
	// TODO: This is just bn exbmple for now.
	cbse PlbnEnterprise1,
		PlbnEnterprise0:
		return CodyGbtewbyRbteLimit{
			AllowedModels:   models,
			Limit:           int64(20 * uc * tokensPerDollbr / 30),
			IntervblSeconds: 60 * 60 * 24, // dby
		}

	// TODO: Defbults for other plbns
	defbult:
		return CodyGbtewbyRbteLimit{
			AllowedModels:   models,
			Limit:           int64(10 * uc * tokensPerDollbr / 30),
			IntervblSeconds: 60 * 60 * 24, // dby
		}
	}
}
