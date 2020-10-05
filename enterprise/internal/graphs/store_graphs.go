package graphs

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/graphs"
)

// graphColumns are used by the graph-related Store methods to query graphs.
var graphColumns = []*sqlf.Query{
	sqlf.Sprintf("graphs.id"),
	sqlf.Sprintf("graphs.owner_user_id"),
	sqlf.Sprintf("graphs.owner_org_id"),
	sqlf.Sprintf("graphs.name"),
	sqlf.Sprintf("graphs.description"),
	sqlf.Sprintf("graphs.spec"),
	sqlf.Sprintf("graphs.created_at"),
	sqlf.Sprintf("graphs.updated_at"),
}

// graphInsertColumns is the list of graph columns that are modified in the CreateGraph and
// UpdateGraph operations.
var graphInsertColumns = []*sqlf.Query{
	sqlf.Sprintf("owner_user_id"),
	sqlf.Sprintf("owner_org_id"),
	sqlf.Sprintf("name"),
	sqlf.Sprintf("description"),
	sqlf.Sprintf("spec"),
	sqlf.Sprintf("created_at"),
	sqlf.Sprintf("updated_at"),
}

// CreatGraph creates the graph.
func (s *Store) CreateGraph(ctx context.Context, g *graphs.Graph) error {
	q, err := s.createGraphQuery(g)
	if err != nil {
		return err
	}

	return s.query(ctx, q, func(sc scanner) error { return scanGraph(g, sc) })
}

var createGraphQueryFmtstr = `
-- source: enterprise/internal/graphs/store_graphs.go:CreateGraph
INSERT INTO graphs (%s)
VALUES (%s, %s, %s, %s, %s, %s, %s)
RETURNING %s
`

func (s *Store) createGraphQuery(g *graphs.Graph) (*sqlf.Query, error) {
	if g.CreatedAt.IsZero() {
		g.CreatedAt = s.now()
	}

	if g.UpdatedAt.IsZero() {
		g.UpdatedAt = g.CreatedAt
	}

	return sqlf.Sprintf(
		createGraphQueryFmtstr,
		sqlf.Join(graphInsertColumns, ", "),
		nullInt32Column(g.OwnerUserID),
		nullInt32Column(g.OwnerOrgID),
		g.Name,
		g.Description,
		g.Spec,
		g.CreatedAt,
		g.UpdatedAt,
		sqlf.Join(graphColumns, ", "),
	), nil
}

// UpdateGraph updates the graph.
func (s *Store) UpdateGraph(ctx context.Context, g *graphs.Graph) error {
	q, err := s.updateGraphQuery(g)
	if err != nil {
		return err
	}

	return s.query(ctx, q, func(sc scanner) error { return scanGraph(g, sc) })
}

var updateGraphQueryFmtstr = `
-- source: enterprise/internal/graphs/store_graphs.go:UpdateGraph
UPDATE graphs
SET (%s) = (%v, %v, %s, %s, %s, %s, %s)
WHERE id = %s
RETURNING %s
`

func (s *Store) updateGraphQuery(g *graphs.Graph) (*sqlf.Query, error) {
	g.UpdatedAt = s.now()

	return sqlf.Sprintf(
		updateGraphQueryFmtstr,
		sqlf.Join(graphInsertColumns, ", "),
		sqlf.Sprintf("graphs.owner_user_id"), // don't update this column's value
		sqlf.Sprintf("graphs.owner_org_id"),  // don't update this column's value
		g.Name,
		g.Description,
		g.Spec,
		g.CreatedAt,
		g.UpdatedAt,
		g.ID,
		sqlf.Join(graphColumns, ", "),
	), nil
}

// DeleteGraph deletes the graph with the given ID.
func (s *Store) DeleteGraph(ctx context.Context, id int64) error {
	return s.Store.Exec(ctx, sqlf.Sprintf(deleteGraphQueryFmtstr, id))
}

var deleteGraphQueryFmtstr = `
-- source: enterprise/internal/graphs/store_graphs.go:DeleteGraph
DELETE FROM graphs WHERE id = %s
`

// CountGraphsOpts captures the query options needed for counting graphs.
type CountGraphsOpts struct {
	OwnerUserID int32
	OwnerOrgID  int32
}

// CountGraphs returns the number of graphs in the database.
func (s *Store) CountGraphs(ctx context.Context, opts CountGraphsOpts) (int, error) {
	return s.queryCount(ctx, countGraphsQuery(&opts))
}

var countGraphsQueryFmtstr = `
-- source: enterprise/internal/graphs/store_graphs.go:CountGraphs
SELECT COUNT(id)
FROM graphs
WHERE %s
`

func countGraphsQuery(opts *CountGraphsOpts) *sqlf.Query {
	var preds []*sqlf.Query

	if opts.OwnerUserID != 0 {
		preds = append(preds, sqlf.Sprintf("owner_user_id = %s", opts.OwnerUserID))
	}

	if opts.OwnerOrgID != 0 {
		preds = append(preds, sqlf.Sprintf("owner_org_id = %s", opts.OwnerOrgID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(countGraphsQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

// GetGraphOpts captures the query options needed for getting a graph.
type GetGraphOpts struct {
	ID int64

	OwnerUserID int32
	OwnerOrgID  int32
	Name        string
}

// GetGraph gets a graph matching the given options.
func (s *Store) GetGraph(ctx context.Context, opts GetGraphOpts) (*graphs.Graph, error) {
	q := getGraphQuery(opts)

	var g graphs.Graph
	err := s.query(ctx, q, func(sc scanner) error {
		return scanGraph(&g, sc)
	})
	if err != nil {
		return nil, err
	}

	if g.ID == 0 {
		return nil, ErrNoResults
	}

	return &g, nil
}

var getGraphQueryFmtstr = `
-- source: enterprise/internal/graphs/store_graphs.go:GetGraph
SELECT %s FROM graphs
WHERE %s
LIMIT 1
`

func getGraphQuery(opts GetGraphOpts) *sqlf.Query {
	// TODO(sqs): validate, with eg `if opts.ID == 0 && (opts.OwnerUserID == 0 && opts.OwnerOrgID == 0) || (owner but no name) ... { ... }`

	var preds []*sqlf.Query

	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("graphs.id = %s", opts.ID))
	}

	if opts.OwnerUserID != 0 {
		preds = append(preds, sqlf.Sprintf("graphs.owner_user_id = %s", opts.OwnerUserID))
	}

	if opts.OwnerOrgID != 0 {
		preds = append(preds, sqlf.Sprintf("graphs.owner_org_id = %s", opts.OwnerOrgID))
	}

	if opts.Name != "" {
		preds = append(preds, sqlf.Sprintf("graphs.name = %s", opts.Name))

	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		getGraphQueryFmtstr,
		sqlf.Join(graphColumns, ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

// ListGraphsOpts is the query options needed for listing graphs.
type ListGraphsOpts struct {
	LimitOpts
	Cursor int64

	OwnerUserID int32
	OwnerOrgID  int32

	AffiliatedWithUserID int32
}

// ListGraphs lists graphs with the given filters.
func (s *Store) ListGraphs(ctx context.Context, opts ListGraphsOpts) (gs []*graphs.Graph, next int64, err error) {
	q := listGraphsQuery(opts)

	gs = make([]*graphs.Graph, 0, opts.DBLimit())
	err = s.query(ctx, q, func(sc scanner) error {
		var c graphs.Graph
		if err := scanGraph(&c, sc); err != nil {
			return err
		}
		gs = append(gs, &c)
		return nil
	})

	if opts.Limit != 0 && len(gs) == opts.DBLimit() {
		next = gs[len(gs)-1].ID
		gs = gs[:len(gs)-1]
	}

	return gs, next, err
}

var listGraphsQueryFmtstr = `
-- source: enterprise/internal/graphs/store.go:ListGraphs
SELECT %s FROM graphs
WHERE %s
ORDER BY id DESC
`

func listGraphsQuery(opts ListGraphsOpts) *sqlf.Query {
	preds := []*sqlf.Query{}

	if opts.Cursor != 0 {
		preds = append(preds, sqlf.Sprintf("id <= %s", opts.Cursor))
	}

	if opts.OwnerUserID != 0 {
		preds = append(preds, sqlf.Sprintf("graphs.owner_user_id = %s", opts.OwnerUserID))
	}

	if opts.OwnerOrgID != 0 {
		preds = append(preds, sqlf.Sprintf("graphs.owner_org_id = %s", opts.OwnerOrgID))
	}

	if opts.AffiliatedWithUserID != 0 {
		preds = append(preds, sqlf.Sprintf("(graphs.owner_user_id = %s OR graphs.owner_org_id IN (SELECT org_members.org_id FROM org_members JOIN orgs ON orgs.id = org_members.org_id WHERE org_members.user_id = %s AND orgs.deleted_at IS NULL))", opts.AffiliatedWithUserID, opts.AffiliatedWithUserID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(
		listGraphsQueryFmtstr+opts.LimitOpts.ToDB(),
		sqlf.Join(graphColumns, ", "),
		sqlf.Join(preds, "\n AND "),
	)
}

func scanGraph(g *graphs.Graph, sc scanner) error {
	return sc.Scan(
		&g.ID,
		&dbutil.NullInt32{N: &g.OwnerUserID},
		&dbutil.NullInt32{N: &g.OwnerOrgID},
		&g.Name,
		&g.Description,
		&g.Spec,
		&g.CreatedAt,
		&g.UpdatedAt,
	)
}
