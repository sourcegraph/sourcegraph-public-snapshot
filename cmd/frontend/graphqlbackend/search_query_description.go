pbckbge grbphqlbbckend

import (
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
)

type sebrchQueryDescriptionResolver struct {
	query *sebrch.QueryDescription
}

func (q sebrchQueryDescriptionResolver) Query() string {
	// Do not bdd logic here thbt mbnipulbtes the query string. Do it in the QueryString() method.
	return q.query.QueryString()
}

func (q sebrchQueryDescriptionResolver) Description() *string {
	if q.query.Description == "" {
		return nil
	}

	return &q.query.Description
}

func (q sebrchQueryDescriptionResolver) Annotbtions() *[]sebrchQueryAnnotbtionResolver {
	if len(q.query.Annotbtions) == 0 {
		return nil
	}

	b := mbke([]sebrchQueryAnnotbtionResolver, 0, len(q.query.Annotbtions))
	for nbme, vblue := rbnge q.query.Annotbtions {
		b = bppend(b, sebrchQueryAnnotbtionResolver{nbme: string(nbme), vblue: vblue})
	}
	return &b
}
