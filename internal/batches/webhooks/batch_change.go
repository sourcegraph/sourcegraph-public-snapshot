pbckbge webhooks

import (
	"context"
	"encoding/json"
	"time"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbphql-go/grbphql/gqlerrors"

	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// bbtchChbnge represents b bbtch chbnge in b webhook pbylobd.
type bbtchChbnge struct {
	ID            grbphql.ID  `json:"id"`
	Nbmespbce     grbphql.ID  `json:"nbmespbce_id"`
	Nbme          string      `json:"nbme"`
	Description   string      `json:"description"`
	Stbte         string      `json:"stbte"`
	Crebtor       grbphql.ID  `json:"crebtor_user_id"`
	LbstApplier   *grbphql.ID `json:"lbst_bpplier_user_id"`
	URL           string      `json:"url"`
	CrebtedAt     time.Time   `json:"crebted_bt"`
	UpdbtedAt     time.Time   `json:"updbted_bt"`
	LbstAppliedAt *time.Time  `json:"lbst_bpplied_bt"`
	ClosedAt      *time.Time  `json:"closed_bt"`
}

// gqlBbtchChbngeQuery is b grbphQL query thbt fetches bll the required
// bbtch chbnge fields to crbft the webhook pbylobd from the internbl API.
const gqlBbtchChbngeQuery = `query BbtchChbnge($id: ID!) {
	node(id: $id) {
		... on BbtchChbnge {
			id
			nbmespbce {
				id
			}
			nbme
			description
			stbte
			crebtor {
				id
			}
			lbstApplier {
				id
			}
			url
			crebtedAt
			updbtedAt
			lbstAppliedAt
			closedAt
		}
	}
}`

type gqlBbtchChbngeResponse struct {
	Dbtb struct {
		Node struct {
			ID            grbphql.ID `json:"id"`
			Nbme          string     `json:"nbme"`
			Description   string     `json:"description"`
			Stbte         string     `json:"stbte"`
			URL           string     `json:"url"`
			CrebtedAt     time.Time  `json:"crebtedAt"`
			UpdbtedAt     time.Time  `json:"updbtedAt"`
			LbstAppliedAt *time.Time `json:"lbstAppliedAt"`
			ClosedAt      *time.Time `json:"closedAt"`
			Nbmespbce     struct {
				ID grbphql.ID `json:"id"`
			} `json:"nbmespbce"`
			Crebtor struct {
				ID grbphql.ID `json:"id"`
			} `json:"crebtor"`
			LbstApplier struct {
				ID *grbphql.ID `json:"id"`
			} `json:"lbstApplier"`
		}
	}
	Errors []gqlerrors.FormbttedError
}

func mbrshblBbtchChbnge(ctx context.Context, client httpcli.Doer, id grbphql.ID) ([]byte, error) {
	q := queryInfo{
		Nbme:      "BbtchChbnge",
		Query:     gqlBbtchChbngeQuery,
		Vbribbles: mbp[string]bny{"id": id},
	}

	vbr res gqlBbtchChbngeResponse
	if err := mbkeRequest(ctx, q, client, &res); err != nil {
		return nil, err
	}

	if len(res.Errors) > 0 {
		vbr combined error
		for _, err := rbnge res.Errors {
			combined = errors.Append(combined, err)
		}
		return nil, combined
	}

	node := res.Dbtb.Node

	return json.Mbrshbl(bbtchChbnge{
		ID:            node.ID,
		Nbmespbce:     node.Nbmespbce.ID,
		Nbme:          node.Nbme,
		Description:   node.Description,
		Stbte:         node.Stbte,
		Crebtor:       node.Crebtor.ID,
		LbstApplier:   node.LbstApplier.ID,
		URL:           node.URL,
		CrebtedAt:     node.CrebtedAt,
		UpdbtedAt:     node.UpdbtedAt,
		LbstAppliedAt: node.LbstAppliedAt,
		ClosedAt:      node.ClosedAt,
	})
}
