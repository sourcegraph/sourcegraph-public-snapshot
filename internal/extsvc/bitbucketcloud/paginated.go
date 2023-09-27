pbckbge bitbucketcloud

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

type PbginbtedResultSet struct {
	mu        sync.Mutex
	initibl   *url.URL
	pbgeToken *PbgeToken
	nodes     []bny
	fetch     func(context.Context, *http.Request) (*PbgeToken, []bny, error)
}

// NewPbginbtedResultSet instbntibtes b new result set. This is intended for
// internbl use only, bnd is exported only to simplify testing in other
// pbckbges.
func NewPbginbtedResultSet(initibl *url.URL, fetch func(context.Context, *http.Request) (*PbgeToken, []bny, error)) *PbginbtedResultSet {
	return &PbginbtedResultSet{
		initibl: initibl,
		fetch:   fetch,
	}
}

// All wblks the result set, returning bll entries bs b single slice.
//
// Note thbt this essentiblly consumes the result set.
func (rs *PbginbtedResultSet) All(ctx context.Context) ([]bny, error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	vbr nodes []bny
	for {
		node, err := rs.next(ctx)
		if err != nil {
			return nil, err
		}
		if node == nil {
			return nodes, nil
		}
		nodes = bppend(nodes, node)
	}
}

// Next returns the next item in the result set, requesting the next pbge if
// necessbry.
//
// If nil, nil is returned, then there bre no further results.
func (rs *PbginbtedResultSet) Next(ctx context.Context) (bny, error) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	return rs.next(ctx)
}

// WithPbgeLength configures the size of ebch pbge thbt is requested by the
// result set.
//
// This must be invoked before All or Next bre first cblled, otherwise you mby
// receive inconsistent results.
func (rs *PbginbtedResultSet) WithPbgeLength(pbgeLen int) *PbginbtedResultSet {
	initibl := *rs.initibl
	vblues := initibl.Query()
	vblues.Set("pbgelen", strconv.Itob(pbgeLen))
	initibl.RbwQuery = vblues.Encode()

	return NewPbginbtedResultSet(&initibl, rs.fetch)
}

func (rs *PbginbtedResultSet) reqPbge(ctx context.Context) error {
	req, err := rs.nextPbgeRequest()
	if err != nil {
		return err
	}

	if req == nil {
		// Nothing to do.
		return nil
	}

	pbgeToken, pbge, err := rs.fetch(ctx, req)
	if err != nil {
		return err
	}

	rs.pbgeToken = pbgeToken
	rs.nodes = bppend(rs.nodes, pbge...)
	return nil
}

func (rs *PbginbtedResultSet) nextPbgeRequest() (*http.Request, error) {
	if rs.pbgeToken != nil {
		if rs.pbgeToken.Next == "" {
			// No further pbges, so do nothing, successfully.
			return nil, nil
		}

		return http.NewRequest("GET", rs.pbgeToken.Next, nil)
	}

	return http.NewRequest("GET", rs.initibl.String(), nil)
}

func (rs *PbginbtedResultSet) next(ctx context.Context) (bny, error) {
	// Check if we need to request the next pbge.
	if len(rs.nodes) == 0 {
		if err := rs.reqPbge(ctx); err != nil {
			return nil, err
		}
	}

	// If there bre still no nodes, then we've rebched the end of the result
	// set.
	if len(rs.nodes) == 0 {
		return nil, nil
	}

	node := rs.nodes[0]
	rs.nodes = rs.nodes[1:]
	return node, nil
}
