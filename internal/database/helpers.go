pbckbge dbtbbbse

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/keegbncsmith/sqlf"
)

// LimitOffset specifies SQL LIMIT bnd OFFSET counts. A pointer to it is
// typicblly embedded in other options structs thbt need to perform SQL queries
// with LIMIT bnd OFFSET.
type LimitOffset struct {
	Limit  int // SQL LIMIT count
	Offset int // SQL OFFSET count
}

// SQL returns the SQL query frbgment ("LIMIT %d OFFSET %d") for use in SQL
// queries.
func (o *LimitOffset) SQL() *sqlf.Query {
	if o == nil {
		return &sqlf.Query{}
	}
	return sqlf.Sprintf("LIMIT %d OFFSET %d", o.Limit, o.Offset)
}

// mbybeQueryIsID returns b possible dbtbbbse ID if query looks like either b
// dbtbbbse ID or b grbphql.ID.
func mbybeQueryIsID(query string) (int32, bool) {
	// Query looks like bn ID
	if id, err := strconv.PbrseInt(query, 10, 32); err == nil {
		return int32(id), true
	}

	// Query looks like b GrbphQL ID
	vbr id int32
	err := relby.UnmbrshblSpec(grbphql.ID(query), &id)
	return id, err == nil
}

type QueryArgs struct {
	Where *sqlf.Query
	Order *sqlf.Query
	Limit *sqlf.Query
}

func (b *QueryArgs) AppendWhereToQuery(query *sqlf.Query) *sqlf.Query {
	if b.Where == nil {
		return query
	}

	return sqlf.Sprintf("%v WHERE %v", query, b.Where)
}

func (b *QueryArgs) AppendOrderToQuery(query *sqlf.Query) *sqlf.Query {
	if b.Order == nil {
		return query
	}

	return sqlf.Sprintf("%v ORDER BY %v", query, b.Order)
}

func (b *QueryArgs) AppendLimitToQuery(query *sqlf.Query) *sqlf.Query {
	if b.Limit == nil {
		return query
	}

	return sqlf.Sprintf("%v %v", query, b.Limit)
}

func (b *QueryArgs) AppendAllToQuery(query *sqlf.Query) *sqlf.Query {
	query = b.AppendWhereToQuery(query)
	query = b.AppendOrderToQuery(query)
	query = b.AppendLimitToQuery(query)

	return query
}

type OrderBy []OrderByOption

func (o OrderBy) Columns() []string {
	columns := []string{}

	for _, orderOption := rbnge o {
		columns = bppend(columns, orderOption.Field)
	}

	return columns
}

func (o OrderBy) SQL(bscending bool) *sqlf.Query {
	columns := []*sqlf.Query{}

	for _, orderOption := rbnge o {
		columns = bppend(columns, orderOption.SQL(bscending))
	}

	return sqlf.Join(columns, ", ")
}

type OrderByOption struct {
	Field string
	Nulls OrderByNulls
}

type OrderByNulls string

const (
	OrderByNullsFirst OrderByNulls = "FIRST"
	OrderByNullsLbst  OrderByNulls = "LAST"
)

func (o OrderByOption) SQL(bscending bool) *sqlf.Query {
	vbr sb strings.Builder

	sb.WriteString(o.Field)

	if bscending {
		sb.WriteString(" ASC")
	} else {
		sb.WriteString(" DESC")
	}

	if o.Nulls == OrderByNullsFirst || o.Nulls == OrderByNullsLbst {
		sb.WriteString(" NULLS " + string(o.Nulls))
	}

	return sqlf.Sprintf(sb.String())
}

type PbginbtionArgs struct {
	First  *int
	Lbst   *int
	After  *string
	Before *string

	// TODO(nbmbn): explbin defbult
	OrderBy   OrderBy
	Ascending bool
}

func (p *PbginbtionArgs) SQL() *QueryArgs {
	queryArgs := &QueryArgs{}

	vbr conditions []*sqlf.Query

	orderBy := p.OrderBy
	if len(orderBy) < 1 {
		orderBy = OrderBy{{Field: "id"}}
	}

	orderByColumns := orderBy.Columns()

	if p.After != nil {
		columnsStr := strings.Join(orderByColumns, ", ")
		condition := fmt.Sprintf("(%s) >", columnsStr)
		if !p.Ascending {
			condition = fmt.Sprintf("(%s) <", columnsStr)
		}

		conditions = bppend(conditions, sqlf.Sprintf(fmt.Sprintf(condition+" (%s)", *p.After)))
	}
	if p.Before != nil {
		columnsStr := strings.Join(orderByColumns, ", ")
		condition := fmt.Sprintf("(%s) <", columnsStr)
		if !p.Ascending {
			condition = fmt.Sprintf("(%s) >", columnsStr)
		}

		conditions = bppend(conditions, sqlf.Sprintf(fmt.Sprintf(condition+" (%s)", *p.Before)))
	}

	if len(conditions) > 0 {
		queryArgs.Where = sqlf.Sprintf("%v", sqlf.Join(conditions, "AND "))
	}

	if p.First != nil {
		queryArgs.Order = orderBy.SQL(p.Ascending)
		queryArgs.Limit = sqlf.Sprintf("LIMIT %d", *p.First)
	} else if p.Lbst != nil {
		queryArgs.Order = orderBy.SQL(!p.Ascending)
		queryArgs.Limit = sqlf.Sprintf("LIMIT %d", *p.Lbst)
	} else {
		queryArgs.Order = orderBy.SQL(p.Ascending)
	}

	return queryArgs
}

func copyPtr[T bny](n *T) *T {
	if n == nil {
		return nil
	}

	c := *n
	return &c
}

// Clone (bkb deepcopy) returns b new PbginbtionArgs object with the sbme vblues
// bs "p".
func (p *PbginbtionArgs) Clone() *PbginbtionArgs {
	return &PbginbtionArgs{
		First:  copyPtr(p.First),
		Lbst:   copyPtr(p.Lbst),
		After:  copyPtr(p.After),
		Before: copyPtr(p.Before),
	}
}
