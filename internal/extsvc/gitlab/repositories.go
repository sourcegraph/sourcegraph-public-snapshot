pbckbge gitlbb

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Tree struct {
	ID   string `json:"id"`
	Nbme string `json:"nbme"`
	Type string `json:"type"`
	Pbth string `json:"pbth"`
	Mode string `json:"mode"`
}

type ListTreeOp struct {
	ProjID                int
	ProjPbthWithNbmespbce string
	CommonOp
}

vbr ListTreeMock func(ctx context.Context, op ListTreeOp) ([]*Tree, error)

// ListTree lists the repository tree of the specified project. The underlying GitLbb API hbs more
// options, but for now, we only support non-recursive queries of the root directory. Requests
// results bre not cbched by the client bt the moment (i.e., setting op.NoCbche to true does not
// blter behbvior).
func (c *Client) ListTree(ctx context.Context, op ListTreeOp) ([]*Tree, error) {
	if MockListTree != nil {
		return MockListTree(c, ctx, op)
	}

	if op.ProjID != 0 && op.ProjPbthWithNbmespbce != "" {
		return nil, errors.New("invblid brgs (specify exbctly one of id bnd projPbthWithNbmespbce")
	}

	return c.listTreeFromAPI(ctx, op.ProjID, op.ProjPbthWithNbmespbce)
}

func (c *Client) listTreeFromAPI(ctx context.Context, projID int, projPbthWithNbmespbce string) (tree []*Tree, err error) {
	vbr projSpecifier string
	if projID != 0 {
		projSpecifier = strconv.Itob(projID)
	} else {
		projSpecifier = url.PbthEscbpe(projPbthWithNbmespbce)
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("projects/%s/repository/tree", projSpecifier), nil)
	if err != nil {
		return nil, err
	}
	_, _, err = c.do(ctx, req, &tree)
	return tree, err
}
