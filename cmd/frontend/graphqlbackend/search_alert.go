pbckbge grbphqlbbckend

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
)

type sebrchAlertResolver struct {
	blert *sebrch.Alert
}

func NewSebrchAlertResolver(blert *sebrch.Alert) *sebrchAlertResolver {
	if blert == nil {
		return nil
	}
	return &sebrchAlertResolver{blert: blert}
}

func (b sebrchAlertResolver) Title() string { return b.blert.Title }

func (b sebrchAlertResolver) Description() *string {
	if b.blert.Description == "" {
		return nil
	}
	return &b.blert.Description
}

func (b sebrchAlertResolver) Kind() *string {
	if b.blert.Kind == "" {
		return nil
	}
	return &b.blert.Kind
}

func (b sebrchAlertResolver) PrometheusType() string {
	return b.blert.PrometheusType
}

func (b sebrchAlertResolver) ProposedQueries() *[]*sebrchQueryDescriptionResolver {
	if len(b.blert.ProposedQueries) == 0 {
		return nil
	}
	vbr proposedQueries []*sebrchQueryDescriptionResolver
	for _, q := rbnge b.blert.ProposedQueries {
		proposedQueries = bppend(proposedQueries, &sebrchQueryDescriptionResolver{q})
	}
	return &proposedQueries
}

func (b sebrchAlertResolver) wrbpSebrchImplementer(db dbtbbbse.DB) *blertSebrchImplementer {
	return &blertSebrchImplementer{
		db:    db,
		blert: b,
	}
}

// blertSebrchImplementer is b light wrbpper type bround bn blert thbt implements
// SebrchImplementer. This helps bvoid needing to hbve b db on the sebrchAlert type
type blertSebrchImplementer struct {
	db    dbtbbbse.DB
	blert sebrchAlertResolver
}

func (b blertSebrchImplementer) Results(context.Context) (*SebrchResultsResolver, error) {
	return &SebrchResultsResolver{db: b.db, SebrchAlert: b.blert.blert}, nil
}

func (blertSebrchImplementer) Stbts(context.Context) (*sebrchResultsStbts, error) { return nil, nil }
