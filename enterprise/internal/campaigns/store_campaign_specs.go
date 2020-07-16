package campaigns

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/dineshappavoo/basex"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
)

const campaignSpecInsertCols = `
  rand_id,
  raw_spec,
  spec,
  namespace_user_id,
  namespace_org_id,
  user_id,
  created_at,
  updated_at
`
const campaignSpecInsertColsFmt = `(%s, %s, %s, %s, %s, %s, %s, %s)`

const campaignSpecCols = `
  id,` + campaignSpecInsertCols

// CreateCampaignSpec creates the given CampaignSpec.
func (s *Store) CreateCampaignSpec(ctx context.Context, c *campaigns.CampaignSpec) error {
	q, err := s.createCampaignSpecQuery(c)
	if err != nil {
		return err
	}
	err = s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanCampaignSpec(c, sc)
		return c.ID, 1, err
	})

	if err, ok := err.(*pq.Error); ok {
		fmt.Printf("q: %s,\nargs: %s\n,pq error: %#v\n", q.Query(sqlf.PostgresBindVar), q.Args(), err)
	}

	return err
}

var createCampaignSpecQueryFmtstr = `
-- source: enterprise/internal/campaigns/store_campaign_specs.go:CreateCampaignSpec
INSERT INTO campaign_specs (` + campaignSpecInsertCols + `)
VALUES ` + campaignSpecInsertColsFmt + `
RETURNING` + campaignSpecCols + `;`

func (s *Store) createCampaignSpecQuery(c *campaigns.CampaignSpec) (*sqlf.Query, error) {
	spec, err := jsonbColumn(c.Spec)
	if err != nil {
		return nil, err
	}

	if c.CreatedAt.IsZero() {
		c.CreatedAt = s.now()
	}

	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = c.CreatedAt
	}

	if c.RandID == "" {
		if c.RandID, err = basex.Encode(strconv.Itoa(seededRand.Int())); err != nil {
			return nil, errors.Wrap(err, "creating RandID failed")
		}
	}

	return sqlf.Sprintf(
		createCampaignSpecQueryFmtstr,
		c.RandID,
		c.RawSpec,
		spec,
		nullInt32Column(c.NamespaceUserID),
		nullInt32Column(c.NamespaceOrgID),
		c.UserID,
		c.CreatedAt,
		c.UpdatedAt,
	), nil
}

// UpdateCampaignSpec updates the given CampaignSpec.
func (s *Store) UpdateCampaignSpec(ctx context.Context, c *campaigns.CampaignSpec) error {
	q, err := s.updateCampaignSpecQuery(c)
	if err != nil {
		return err
	}

	return s.exec(ctx, q, func(sc scanner) (last, count int64, err error) {
		err = scanCampaignSpec(c, sc)
		return c.ID, 1, err
	})
}

var updateCampaignSpecQueryFmtstr = `
-- source: enterprise/internal/campaigns/store_campaign_specs.go:UpdateCampaignSpec
UPDATE campaign_specs
SET (` + campaignSpecInsertCols + `) = ` + campaignSpecInsertColsFmt + `
WHERE id = %s
RETURNING ` + campaignSpecCols

func (s *Store) updateCampaignSpecQuery(c *campaigns.CampaignSpec) (*sqlf.Query, error) {
	spec, err := jsonbColumn(c.Spec)
	if err != nil {
		return nil, err
	}

	c.UpdatedAt = s.now()

	return sqlf.Sprintf(
		updateCampaignSpecQueryFmtstr,
		c.RandID,
		c.RawSpec,
		spec,
		nullInt32Column(c.NamespaceUserID),
		nullInt32Column(c.NamespaceOrgID),
		c.UserID,
		c.CreatedAt,
		c.UpdatedAt,
		c.ID,
	), nil
}

// DeleteCampaignSpec deletes the CampaignSpec with the given ID.
func (s *Store) DeleteCampaignSpec(ctx context.Context, id int64) error {
	q := sqlf.Sprintf(deleteCampaignSpecQueryFmtstr, id)

	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	return rows.Close()
}

var deleteCampaignSpecQueryFmtstr = `
-- source: enterprise/internal/campaigns/store_campaign_specs.go:DeleteCampaignSpec
DELETE FROM campaign_specs WHERE id = %s
`

// CountCampaignSpecs returns the number of code mods in the database.
func (s *Store) CountCampaignSpecs(ctx context.Context) (count int64, _ error) {
	q := sqlf.Sprintf(countCampaignSpecsQueryFmtstr)
	return count, s.exec(ctx, q, func(sc scanner) (_, _ int64, err error) {
		err = sc.Scan(&count)
		return 0, count, err
	})
}

var countCampaignSpecsQueryFmtstr = `
-- source: enterprise/internal/campaigns/store_campaign_specs.go:CountCampaignSpecs
SELECT COUNT(id)
FROM campaign_specs
`

// GetCampaignSpecOpts captures the query options needed for getting a CampaignSpec
type GetCampaignSpecOpts struct {
	ID     int64
	RandID string
}

// GetCampaignSpec gets a code mod matching the given options.
func (s *Store) GetCampaignSpec(ctx context.Context, opts GetCampaignSpecOpts) (*campaigns.CampaignSpec, error) {
	q := getCampaignSpecQuery(&opts)

	var c campaigns.CampaignSpec
	err := s.exec(ctx, q, func(sc scanner) (_, _ int64, err error) {
		return 0, 0, scanCampaignSpec(&c, sc)
	})
	if err != nil {
		return nil, err
	}

	if c.ID == 0 {
		return nil, ErrNoResults
	}

	return &c, nil
}

var getCampaignSpecsQueryFmtstr = `
-- source: enterprise/internal/campaigns/store_campaign_specs.go:GetCampaignSpec
SELECT ` + campaignSpecCols + `
FROM campaign_specs
WHERE %s
LIMIT 1
`

func getCampaignSpecQuery(opts *GetCampaignSpecOpts) *sqlf.Query {
	var preds []*sqlf.Query
	if opts.ID != 0 {
		preds = append(preds, sqlf.Sprintf("id = %s", opts.ID))
	}

	if opts.RandID != "" {
		preds = append(preds, sqlf.Sprintf("rand_id = %s", opts.RandID))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	return sqlf.Sprintf(getCampaignSpecsQueryFmtstr, sqlf.Join(preds, "\n AND "))
}

// ListCampaignSpecsOpts captures the query options needed for
// listing code mods.
type ListCampaignSpecsOpts struct {
	Cursor int64
	Limit  int
}

// ListCampaignSpecs lists CampaignSpecs with the given filters.
func (s *Store) ListCampaignSpecs(ctx context.Context, opts ListCampaignSpecsOpts) (cs []*campaigns.CampaignSpec, next int64, err error) {
	q := listCampaignSpecsQuery(&opts)

	cs = make([]*campaigns.CampaignSpec, 0, opts.Limit)
	_, _, err = s.query(ctx, q, func(sc scanner) (last, count int64, err error) {
		var c campaigns.CampaignSpec
		if err = scanCampaignSpec(&c, sc); err != nil {
			return 0, 0, err
		}
		cs = append(cs, &c)
		return c.ID, 1, err
	})

	if opts.Limit != 0 && len(cs) == opts.Limit {
		next = cs[len(cs)-1].ID
		cs = cs[:len(cs)-1]
	}

	return cs, next, err
}

var listCampaignSpecsQueryFmtstr = `
-- source: enterprise/internal/campaigns/store_campaign_specs.go:ListCampaignSpecs
SELECT ` + campaignSpecCols + ` FROM campaign_specs
WHERE %s
ORDER BY id ASC
`

func listCampaignSpecsQuery(opts *ListCampaignSpecsOpts) *sqlf.Query {
	if opts.Limit == 0 {
		opts.Limit = defaultListLimit
	}
	opts.Limit++

	var limitClause string
	if opts.Limit > 0 {
		limitClause = fmt.Sprintf("LIMIT %d", opts.Limit)
	}

	preds := []*sqlf.Query{
		sqlf.Sprintf("id >= %s", opts.Cursor),
	}

	return sqlf.Sprintf(
		listCampaignSpecsQueryFmtstr+limitClause,
		sqlf.Join(preds, "\n AND "),
	)
}

// DeleteExpiredCampaignSpecs deletes CampaignSpecs that have not been attached
// to a Campaign within CampaignSpecTTL.
func (s *Store) DeleteExpiredCampaignSpecs(ctx context.Context) error {
	expirationTime := s.now().Add(-campaigns.CampaignSpecTTL)
	q := sqlf.Sprintf(deleteExpiredCampaignSpecsQueryFmtstr, expirationTime)

	rows, err := s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return err
	}
	return rows.Close()
}

var deleteExpiredCampaignSpecsQueryFmtstr = `
-- source: enterprise/internal/campaigns/store.go:DeleteExpiredCampaignSpecs
DELETE FROM
  campaign_specs
WHERE
  created_at < %s
AND
NOT EXISTS (
  SELECT 1 FROM campaigns WHERE campaigns.campaign_spec_id = campaign_specs.id
);
`

func scanCampaignSpec(c *campaigns.CampaignSpec, s scanner) error {
	var spec json.RawMessage

	err := s.Scan(
		&c.ID,
		&c.RandID,
		&c.RawSpec,
		&spec,
		&dbutil.NullInt32{N: &c.NamespaceUserID},
		&dbutil.NullInt32{N: &c.NamespaceOrgID},
		&c.UserID,
		&c.CreatedAt,
		&c.UpdatedAt,
	)

	if err != nil {
		return errors.Wrap(err, "scanning campaign spec")
	}

	if err = json.Unmarshal(spec, &c.Spec); err != nil {
		return errors.Wrap(err, "scanCampaignSpec: failed to unmarshal spec")
	}

	return nil
}
