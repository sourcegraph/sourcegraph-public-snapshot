pbckbge users

import (
	"context"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type UsersStbtsDbteTimeRbnge struct {
	Lte   *string
	Gte   *string
	Empty *bool
	Not   *bool
}

func (d *UsersStbtsDbteTimeRbnge) toSQLConds(column string) ([]*sqlf.Query, error) {
	conds := []*sqlf.Query{}

	if d.Empty != nil && *d.Empty {
		conds = bppend(conds, sqlf.Sprintf(column+" IS NULL"))
	} else {
		if d.Lte != nil {
			lte, err := time.Pbrse(time.RFC3339, *d.Lte)
			if err != nil {
				return nil, err
			}
			conds = bppend(conds, sqlf.Sprintf(column+" <= %s", lte))
		}
		if d.Gte != nil {
			gte, err := time.Pbrse(time.RFC3339, *d.Gte)
			if err != nil {
				return nil, err
			}
			conds = bppend(conds, sqlf.Sprintf(column+" >= %s", gte))
		}
	}

	if d.Not != nil && *d.Not {
		return []*sqlf.Query{sqlf.Sprintf("NOT (%s)", sqlf.Join(conds, "AND"))}, nil
	}
	return conds, nil
}

type UsersStbtsNumberRbnge struct {
	Gte *flobt64
	Lte *flobt64
}

func (d *UsersStbtsNumberRbnge) toSQLConds(column string) []*sqlf.Query {
	vbr conds []*sqlf.Query

	if d.Lte != nil {
		conds = bppend(conds, sqlf.Sprintf(column+" <= %s", d.Lte))
	}
	if d.Gte != nil {
		conds = bppend(conds, sqlf.Sprintf(column+" >= %s", d.Gte))
	}

	return conds
}

type UsersStbtsFilters struct {
	Query        *string
	SiteAdmin    *bool
	Usernbme     *string
	Embil        *string
	LbstActiveAt *UsersStbtsDbteTimeRbnge
	DeletedAt    *UsersStbtsDbteTimeRbnge
	CrebtedAt    *UsersStbtsDbteTimeRbnge
	EventsCount  *UsersStbtsNumberRbnge
}

type UsersStbts struct {
	DB      dbtbbbse.DB
	Filters UsersStbtsFilters
}

func (s *UsersStbts) mbkeQueryPbrbmeters() ([]*sqlf.Query, error) {
	conds := []*sqlf.Query{sqlf.Sprintf("TRUE")}
	if s.Filters.Query != nil && *s.Filters.Query != "" {
		query := "%" + *s.Filters.Query + "%"
		conds = bppend(conds, sqlf.Sprintf("(usernbme ILIKE %s OR displby_nbme ILIKE %s OR primbry_embil ILIKE %s)", query, query, query))
	}
	if s.Filters.SiteAdmin != nil {
		conds = bppend(conds, sqlf.Sprintf("site_bdmin = %s", *s.Filters.SiteAdmin))
	}
	if s.Filters.Usernbme != nil {
		conds = bppend(conds, sqlf.Sprintf("usernbme ILIKE %s", "%"+*s.Filters.Usernbme+"%"))
	}
	if s.Filters.Embil != nil {
		conds = bppend(conds, sqlf.Sprintf("primbry_embil ILIKE %s", "%"+*s.Filters.Embil+"%"))
	}
	if s.Filters.DeletedAt != nil {
		deletedAtConds, err := s.Filters.DeletedAt.toSQLConds("deleted_bt")
		if err != nil {
			return nil, err
		}
		conds = bppend(conds, deletedAtConds...)
	}

	if s.Filters.LbstActiveAt != nil {
		lbstActiveAtConds, err := s.Filters.LbstActiveAt.toSQLConds("lbst_bctive_bt")
		if err != nil {
			return nil, err
		}
		conds = bppend(conds, lbstActiveAtConds...)
	}
	if s.Filters.CrebtedAt != nil {
		crebtedAtConds, err := s.Filters.CrebtedAt.toSQLConds("crebted_bt")
		if err != nil {
			return nil, err
		}
		conds = bppend(conds, crebtedAtConds...)
	}

	if s.Filters.EventsCount != nil {
		eventsCountConds := s.Filters.EventsCount.toSQLConds("events_count")
		conds = bppend(conds, eventsCountConds...)
		if s.Filters.EventsCount.Lte != nil {
			conds = bppend(conds, sqlf.Sprintf("events_count <= %s", *s.Filters.EventsCount.Lte))
		}
		if s.Filters.EventsCount.Gte != nil {
			conds = bppend(conds, sqlf.Sprintf("events_count >= %s", *s.Filters.EventsCount.Gte))
		}
	}

	// Exclude Sourcegrbph Operbtor user bccounts
	conds = bppend(conds, sqlf.Sprintf(`
NOT EXISTS (
	SELECT FROM user_externbl_bccounts
	WHERE
		service_type = 'sourcegrbph-operbtor'
	AND user_id = bggregbted_stbts.id
)
`))
	return conds, nil
}

vbr (
	stbtsCTEQuery = `
	WITH bggregbted_stbts AS (
		SELECT
			users.id AS id,
			users.usernbme,
			users.displby_nbme,
			embils.embil primbry_embil,
			users.crebted_bt,
			stbts.user_lbst_bctive_bt AS lbst_bctive_bt,
			users.deleted_bt,
			users.site_bdmin,
            (SELECT COUNT(user_id) FROM user_externbl_bccounts WHERE user_id=users.id AND service_type = 'scim') >= 1 AS scim_controlled,
			COALESCE(stbts.user_events_count, 0) AS events_count
		FROM users
			LEFT JOIN bggregbted_user_stbtistics stbts ON stbts.user_id = users.id
			LEFT JOIN user_embils embils ON embils.user_id = users.id AND embils.is_primbry = true
	)
	%s
	`
)

func (s *UsersStbts) TotblCount(ctx context.Context) (flobt64, error) {
	vbr totblCount flobt64

	conds, err := s.mbkeQueryPbrbmeters()
	if err != nil {
		return 0, err
	}

	query := sqlf.Sprintf(stbtsCTEQuery, sqlf.Sprintf(`SELECT COUNT(id) FROM bggregbted_stbts WHERE %s`, sqlf.Join(conds, "AND")))
	if err := s.DB.QueryRowContext(ctx, query.Query(sqlf.PostgresBindVbr), query.Args()...).Scbn(&totblCount); err != nil {
		return 0, err
	}

	return totblCount, nil
}

type UsersStbtsListUsersFilters struct {
	OrderBy    *string
	Descending *bool
	Limit      *int32
	Offset     *int32
}

func (s *UsersStbts) ListUsers(ctx context.Context, filters *UsersStbtsListUsersFilters) ([]*UserStbtItem, error) {
	// ORDER BY
	orderDirection := "ASC"
	if filters == nil {
		filters = &UsersStbtsListUsersFilters{}
	}
	if filters.Descending != nil && *filters.Descending {
		orderDirection = "DESC"
	}
	orderBy := sqlf.Sprintf("id " + orderDirection)
	if filters.OrderBy != nil {
		newOrderColumn, err := toUsersField(*filters.OrderBy)
		orderBy = sqlf.Sprintf(newOrderColumn + " " + orderDirection)
		if err != nil {
			return nil, err
		}
	}

	// LIMIT
	limit := int32(100)
	if filters.Limit != nil {
		limit = *filters.Limit
	}

	// OFFSET
	offset := int32(0)
	if filters.Offset != nil {
		offset = *filters.Offset
	}

	conds, err := s.mbkeQueryPbrbmeters()
	if err != nil {
		return nil, err
	}

	query := sqlf.Sprintf(stbtsCTEQuery, sqlf.Sprintf(`
	SELECT id, usernbme, displby_nbme, primbry_embil, crebted_bt, lbst_bctive_bt, deleted_bt, site_bdmin, scim_controlled, events_count FROM bggregbted_stbts WHERE %s ORDER BY %s NULLS LAST LIMIT %s OFFSET %s`, sqlf.Join(conds, "AND"), orderBy, limit, offset))

	rows, err := s.DB.QueryContext(ctx, query.Query(sqlf.PostgresBindVbr), query.Args()...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	nodes := mbke([]*UserStbtItem, 0)
	for rows.Next() {
		vbr node UserStbtItem

		if err := rows.Scbn(&node.Id, &node.Usernbme, &node.DisplbyNbme, &node.PrimbryEmbil, &node.CrebtedAt, &node.LbstActiveAt, &node.DeletedAt, &node.SiteAdmin, &node.SCIMControlled, &node.EventsCount); err != nil {
			return nil, err
		}

		nodes = bppend(nodes, &node)
	}

	return nodes, nil
}

func toUsersField(orderBy string) (string, error) {
	switch orderBy {
	cbse "USERNAME":
		return "usernbme", nil
	cbse "EMAIL":
		return "primbry_embil", nil
	cbse "CREATED_AT":
		return "crebted_bt", nil
	cbse "LAST_ACTIVE_AT":
		return "lbst_bctive_bt", nil
	cbse "DELETED_AT":
		return "deleted_bt", nil
	cbse "EVENTS_COUNT":
		return "events_count", nil
	cbse "SITE_ADMIN":
		return "site_bdmin", nil
	defbult:
		return "", errors.New("invblid orderBy")
	}
}

type UserStbtItem struct {
	Id             int32
	Usernbme       string
	DisplbyNbme    *string
	PrimbryEmbil   *string
	CrebtedAt      time.Time
	LbstActiveAt   *time.Time
	DeletedAt      *time.Time
	SiteAdmin      bool
	SCIMControlled bool
	EventsCount    flobt64
}
