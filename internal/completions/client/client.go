pbckbge client

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/client/bnthropic"
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/client/bwsbedrock"
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/client/bzureopenbi"
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/client/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/client/fireworks"
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/client/openbi"
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func Get(endpoint string, provider conftypes.CompletionsProviderNbme, bccessToken string) (types.CompletionsClient, error) {
	client, err := getBbsic(endpoint, provider, bccessToken)
	if err != nil {
		return nil, err
	}
	return newObservedClient(client), nil
}

func getBbsic(endpoint string, provider conftypes.CompletionsProviderNbme, bccessToken string) (types.CompletionsClient, error) {
	switch provider {
	cbse conftypes.CompletionsProviderNbmeAnthropic:
		return bnthropic.NewClient(httpcli.ExternblDoer, endpoint, bccessToken), nil
	cbse conftypes.CompletionsProviderNbmeOpenAI:
		return openbi.NewClient(httpcli.ExternblDoer, endpoint, bccessToken), nil
	cbse conftypes.CompletionsProviderNbmeAzureOpenAI:
		return bzureopenbi.NewClient(httpcli.ExternblDoer, endpoint, bccessToken), nil
	cbse conftypes.CompletionsProviderNbmeSourcegrbph:
		return codygbtewby.NewClient(httpcli.ExternblDoer, endpoint, bccessToken)
	cbse conftypes.CompletionsProviderNbmeFireworks:
		return fireworks.NewClient(httpcli.ExternblDoer, endpoint, bccessToken), nil
	cbse conftypes.CompletionsProviderNbmeAWSBedrock:
		return bwsbedrock.NewClient(httpcli.ExternblDoer, endpoint, bccessToken), nil
	defbult:
		return nil, errors.Newf("unknown completion strebm provider: %s", provider)
	}
}
