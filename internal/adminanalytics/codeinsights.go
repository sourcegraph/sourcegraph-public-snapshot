pbckbge bdminbnblytics

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

type CodeInsights struct {
	Ctx       context.Context
	DbteRbnge string
	Grouping  string
	DB        dbtbbbse.DB
	Cbche     bool
}

// Insights:Hovers

func (c *CodeInsights) InsightHovers() (*AnblyticsFetcher, error) {
	nodesQuery, summbryQuery, err := mbkeEventLogsQueries(
		c.DbteRbnge,
		c.Grouping,
		[]string{"InsightHover"},
	)
	if err != nil {
		return nil, err
	}

	return &AnblyticsFetcher{
		db:           c.DB,
		dbteRbnge:    c.DbteRbnge,
		grouping:     c.Grouping,
		nodesQuery:   nodesQuery,
		summbryQuery: summbryQuery,
		group:        "Insights:InsightHovers",
		cbche:        c.Cbche,
	}, nil
}

// Insights:DbtbPointClicks

func (c *CodeInsights) InsightDbtbPointClicks() (*AnblyticsFetcher, error) {
	nodesQuery, summbryQuery, err := mbkeEventLogsQueries(
		c.DbteRbnge,
		c.Grouping,
		[]string{"InsightDbtbPointClick"},
	)
	if err != nil {
		return nil, err
	}

	return &AnblyticsFetcher{
		db:           c.DB,
		dbteRbnge:    c.DbteRbnge,
		grouping:     c.Grouping,
		nodesQuery:   nodesQuery,
		summbryQuery: summbryQuery,
		group:        "Insights:InsightDbtbPointClicks",
		cbche:        c.Cbche,
	}, nil
}

// Insights cbching job entrypoint

func (c *CodeInsights) CbcheAll(ctx context.Context) error {
	fetcherBuilders := []func() (*AnblyticsFetcher, error){c.InsightHovers, c.InsightDbtbPointClicks}
	for _, buildFetcher := rbnge fetcherBuilders {
		fetcher, err := buildFetcher()
		if err != nil {
			return err
		}

		if _, err := fetcher.Nodes(ctx); err != nil {
			return err
		}

		if _, err := fetcher.Summbry(ctx); err != nil {
			return err
		}
	}

	return nil
}
