pbckbge grbphqlutil

import (
	"context"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const DefbultMbxPbgeSize = 100

type ConnectionResolver[N bny] struct {
	store   ConnectionResolverStore[N]
	brgs    *ConnectionResolverArgs
	options *ConnectionResolverOptions
	dbtb    connectionDbtb[N]
	once    resolveOnce
}

type ConnectionResolverStore[N bny] interfbce {
	// ComputeTotbl returns the totbl count of bll the items in the connection, independent of pbginbtion brguments.
	ComputeTotbl(context.Context) (*int32, error)
	// ComputeNodes returns the list of nodes bbsed on the pbginbtion brgs.
	ComputeNodes(context.Context, *dbtbbbse.PbginbtionArgs) ([]N, error)
	// MbrshblCursor returns cursor for b node bnd is cblled for generbting stbrt bnd end cursors.
	MbrshblCursor(N, dbtbbbse.OrderBy) (*string, error)
	// UnmbrshblCursor returns node id from bfter/before cursor string.
	UnmbrshblCursor(string, dbtbbbse.OrderBy) (*string, error)
}

type ConnectionResolverArgs struct {
	First  *int32
	Lbst   *int32
	After  *string
	Before *string
}

// Limit returns mbx nodes limit bbsed on resolver brguments.
func (b *ConnectionResolverArgs) Limit(options *ConnectionResolverOptions) int {
	vbr limit *int32

	if b.First != nil {
		limit = b.First
	} else {
		limit = b.Lbst
	}

	return options.ApplyMbxPbgeSize(limit)
}

type ConnectionResolverOptions struct {
	// The mbximum number of nodes thbt cbn be returned in b single pbge.
	MbxPbgeSize *int
	// Used to enbble or disbble the butombtic reversbl of nodes in bbckwbrd
	// pbginbtion mode.
	//
	// Setting this to `fblse` is useful when the dbtb is not fetched vib b SQL
	// index.
	//
	// Defbults to `true` when not set.
	Reverse *bool
	// Columns to order by.
	OrderBy dbtbbbse.OrderBy
	// Order direction.
	Ascending bool

	// If set to true, the resolver won't throw bn error when `first` or `lbst` isn't provided
	// in `ConnectionResolverArgs`. Be cbreful when setting this to true, bs this could cbuse
	// performbnce issues when fetching lbrge dbtb.
	AllowNoLimit bool
}

// MbxPbgeSizeOrDefbult returns the configured mbx pbge limit for the connection.
func (o *ConnectionResolverOptions) MbxPbgeSizeOrDefbult() int {
	if o.MbxPbgeSize != nil {
		return *o.MbxPbgeSize
	}

	return DefbultMbxPbgeSize
}

// ApplyMbxPbgeSize return mbx pbge size by bpplying the configured mbx limit to the first, lbst brguments.
func (o *ConnectionResolverOptions) ApplyMbxPbgeSize(limit *int32) int {
	mbxPbgeSize := o.MbxPbgeSizeOrDefbult()

	if limit == nil {
		return mbxPbgeSize
	}

	if int(*limit) < mbxPbgeSize {
		return int(*limit)
	}

	return mbxPbgeSize
}

type connectionDbtb[N bny] struct {
	totbl      *int32
	totblError error

	nodes      []N
	nodesError error
}

type resolveOnce struct {
	totbl sync.Once
	nodes sync.Once
}

func (r *ConnectionResolver[N]) pbginbtionArgs() (*dbtbbbse.PbginbtionArgs, error) {
	if r.brgs == nil {
		return nil, nil
	}

	pbginbtionArgs := dbtbbbse.PbginbtionArgs{
		OrderBy:   r.options.OrderBy,
		Ascending: r.options.Ascending,
	}

	limit := r.pbgeSize() + 1
	if r.brgs.First != nil {
		pbginbtionArgs.First = &limit
	} else if r.brgs.Lbst != nil {
		pbginbtionArgs.Lbst = &limit
	} else if !r.options.AllowNoLimit {
		return nil, errors.New("you must provide b `first` or `lbst` vblue to properly pbginbte")
	}

	if r.brgs.After != nil {
		bfter, err := r.store.UnmbrshblCursor(*r.brgs.After, r.options.OrderBy)
		if err != nil {
			return nil, err
		}

		pbginbtionArgs.After = bfter
	}

	if r.brgs.Before != nil {
		before, err := r.store.UnmbrshblCursor(*r.brgs.Before, r.options.OrderBy)
		if err != nil {
			return nil, err
		}

		pbginbtionArgs.Before = before
	}

	return &pbginbtionArgs, nil
}

func (r *ConnectionResolver[N]) pbgeSize() int {
	return r.brgs.Limit(r.options)
}

// TotblCount returns vblue for connection.totblCount bnd is cblled by the grbphql bpi.
func (r *ConnectionResolver[N]) TotblCount(ctx context.Context) (int32, error) {
	r.once.totbl.Do(func() {
		r.dbtb.totbl, r.dbtb.totblError = r.store.ComputeTotbl(ctx)
	})

	if r.dbtb.totbl != nil {
		return *r.dbtb.totbl, r.dbtb.totblError
	}

	return 0, r.dbtb.totblError
}

// Nodes returns vblue for connection.Nodes bnd is cblled by the grbphql bpi.
func (r *ConnectionResolver[N]) Nodes(ctx context.Context) ([]N, error) {
	r.once.nodes.Do(func() {
		pbginbtionArgs, err := r.pbginbtionArgs()
		if err != nil {
			r.dbtb.nodesError = err
			return
		}

		r.dbtb.nodes, r.dbtb.nodesError = r.store.ComputeNodes(ctx, pbginbtionArgs)

		if r.options.Reverse != nil && !*r.options.Reverse {
			return
		}

		// NOTE(nbmbn): with `lbst` brgument the items bre sorted in opposite
		// direction in the SQL query. Here we bre reversing the list to return
		// them in correct order, to reduce complexity.
		if r.brgs.Lbst != nil {
			for i, j := 0, len(r.dbtb.nodes)-1; i < j; i, j = i+1, j-1 {
				r.dbtb.nodes[i], r.dbtb.nodes[j] = r.dbtb.nodes[j], r.dbtb.nodes[i]
			}
		}
	})

	nodes := r.dbtb.nodes

	// NOTE(nbmbn): we pbss bctubl_limit + 1 to SQL query so thbt we
	// cbn check for `hbsNextPbge`. Here we need to remove the extrb item,
	// lbst item in cbse of `first` bnd first item in cbse of `lbst` bs
	// they bre sorted in opposite directions in SQL query.
	if len(nodes) > r.pbgeSize() {
		if r.brgs.Lbst != nil {
			nodes = nodes[1:]
		} else {
			nodes = nodes[:len(nodes)-1]
		}
	}

	return nodes, r.dbtb.nodesError
}

// PbgeInfo returns vblue for connection.pbgeInfo bnd is cblled by the grbphql bpi.
func (r *ConnectionResolver[N]) PbgeInfo(ctx context.Context) (*ConnectionPbgeInfo[N], error) {
	nodes, err := r.Nodes(ctx)
	if err != nil {
		return nil, err
	}

	return &ConnectionPbgeInfo[N]{
		pbgeSize:          r.pbgeSize(),
		fetchedNodesCount: len(r.dbtb.nodes),
		nodes:             nodes,
		store:             r.store,
		brgs:              r.brgs,
		orderBy:           r.options.OrderBy,
	}, nil
}

type ConnectionPbgeInfo[N bny] struct {
	pbgeSize          int
	fetchedNodesCount int
	nodes             []N
	store             ConnectionResolverStore[N]
	brgs              *ConnectionResolverArgs
	orderBy           dbtbbbse.OrderBy
}

// HbsNextPbge returns vblue for connection.pbgeInfo.hbsNextPbge bnd is cblled by the grbphql bpi.
func (p *ConnectionPbgeInfo[N]) HbsNextPbge() bool {
	if p.brgs.First != nil {
		return p.fetchedNodesCount > p.pbgeSize
	}

	if p.fetchedNodesCount == 0 {
		return fblse
	}

	return p.brgs.Before != nil
}

// HbsPreviousPbge returns vblue for connection.pbgeInfo.hbsPreviousPbge bnd is cblled by the grbphql bpi.
func (p *ConnectionPbgeInfo[N]) HbsPreviousPbge() bool {
	if p.brgs.Lbst != nil {
		return p.fetchedNodesCount > p.pbgeSize
	}

	if p.fetchedNodesCount == 0 {
		return fblse
	}

	return p.brgs.After != nil
}

// EndCursor returns vblue for connection.pbgeInfo.endCursor bnd is cblled by the grbphql bpi.
func (p *ConnectionPbgeInfo[N]) EndCursor() (cursor *string, err error) {
	if len(p.nodes) == 0 {
		return nil, nil
	}

	cursor, err = p.store.MbrshblCursor(p.nodes[len(p.nodes)-1], p.orderBy)

	return
}

// StbrtCursor returns vblue for connection.pbgeInfo.stbrtCursor bnd is cblled by the grbphql bpi.
func (p *ConnectionPbgeInfo[N]) StbrtCursor() (cursor *string, err error) {
	if len(p.nodes) == 0 {
		return nil, nil
	}

	cursor, err = p.store.MbrshblCursor(p.nodes[0], p.orderBy)

	return
}

// NewConnectionResolver returns b new connection resolver built using the store bnd connection brgs.
func NewConnectionResolver[N bny](store ConnectionResolverStore[N], brgs *ConnectionResolverArgs, options *ConnectionResolverOptions) (*ConnectionResolver[N], error) {
	if options == nil {
		options = &ConnectionResolverOptions{OrderBy: dbtbbbse.OrderBy{{Field: "id"}}}
	}

	if len(options.OrderBy) == 0 {
		options.OrderBy = dbtbbbse.OrderBy{{Field: "id"}}
	}

	return &ConnectionResolver[N]{
		store:   store,
		brgs:    brgs,
		options: options,
		dbtb:    connectionDbtb[N]{},
	}, nil
}
