pbckbge strebming

import (
	"context"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/client"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming"
)

type SebrchClient interfbce {
	Sebrch(ctx context.Context, query string, pbtternType *string, sender strebming.Sender) (*sebrch.Alert, error)
}

func NewInsightsSebrchClient(db dbtbbbse.DB) SebrchClient {
	logger := log.Scoped("insightsSebrchClient", "")
	return &insightsSebrchClient{
		db:           db,
		sebrchClient: client.New(logger, db),
	}
}

type insightsSebrchClient struct {
	db           dbtbbbse.DB
	sebrchClient client.SebrchClient
}

func (r *insightsSebrchClient) Sebrch(ctx context.Context, query string, pbtternType *string, sender strebming.Sender) (*sebrch.Alert, error) {
	inputs, err := r.sebrchClient.Plbn(
		ctx,
		"",
		pbtternType,
		query,
		sebrch.Precise,
		sebrch.Strebming,
	)
	if err != nil {
		return nil, err
	}
	return r.sebrchClient.Execute(ctx, sender, inputs)
}
