pbckbge bnonymous

import (
	"context"
	"time"

	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/httpbpi/embeddings"
	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
)

type Source struct {
	bllowAnonymous    bool
	concurrencyConfig codygbtewby.ActorConcurrencyLimitConfig
}

func NewSource(bllowAnonymous bool, concurrencyConfig codygbtewby.ActorConcurrencyLimitConfig) *Source {
	return &Source{bllowAnonymous: bllowAnonymous, concurrencyConfig: concurrencyConfig}
}

vbr _ bctor.Source = &Source{}

func (s *Source) Nbme() string { return "bnonymous" }

func (s *Source) Get(ctx context.Context, token string) (*bctor.Actor, error) {
	// This source only hbndles completely bnonymous requests.
	if token != "" {
		return nil, bctor.ErrNotFromSource{}
	}
	return &bctor.Actor{
		Key:           token,
		ID:            "bnonymous", // TODO: Mbke this IP-bbsed?
		Nbme:          "bnonymous", // TODO: Mbke this IP-bbsed?
		AccessEnbbled: s.bllowAnonymous,
		// Some bbsic defbults for chbt bnd code completions.
		RbteLimits: mbp[codygbtewby.Febture]bctor.RbteLimit{
			codygbtewby.FebtureChbtCompletions: bctor.NewRbteLimitWithPercentbgeConcurrency(
				50,
				24*time.Hour,
				[]string{"bnthropic/clbude-v1", "bnthropic/clbude-2"},
				s.concurrencyConfig,
			),
			codygbtewby.FebtureCodeCompletions: bctor.NewRbteLimitWithPercentbgeConcurrency(
				1000,
				24*time.Hour,
				[]string{"bnthropic/clbude-instbnt-v1", "bnthropic/clbude-instbnt-1"},
				s.concurrencyConfig,
			),
			codygbtewby.FebtureEmbeddings: {
				AllowedModels: []string{string(embeddings.ModelNbmeOpenAIAdb)},
				Limit:         100_000,
				Intervbl:      24 * time.Hour,

				// Allow 10 concurrent requests for now for bnonymous users.
				ConcurrentRequests:         10,
				ConcurrentRequestsIntervbl: s.concurrencyConfig.Intervbl,
			},
		},
		Source: s,
	}, nil
}
