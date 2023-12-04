package database

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/keegancsmith/sqlf"
)

// LimitOffset specifies SQL LIMIT and OFFSET counts. A pointer to it is
// typically embedded in other options structs that need to perform SQL queries
// with LIMIT and OFFSET.
type LimitOffset struct {
	Limit  int // SQL LIMIT count
	Offset int // SQL OFFSET count
}

// SQL returns the SQL query fragment ("LIMIT %d OFFSET %d") for use in SQL
// queries.
func (o *LimitOffset) SQL() *sqlf.Query {
	if o == nil {
		return &sqlf.Query{}
	}
	return sqlf.Sprintf("LIMIT %d OFFSET %d", o.Limit, o.Offset)
}

// maybeQueryIsID returns a possible database ID if query looks like either a
// database ID or a graphql.ID.
func maybeQueryIsID(query string) (int32, bool) {
	// Query looks like an ID
	if id, err := strconv.ParseInt(query, 10, 32); err == nil {
		return int32(id), true
	}

	// Query looks like a GraphQL ID
	var id int32
	err := relay.UnmarshalSpec(graphql.ID(query), &id)
	return id, err == nil
}

type QueryArgs struct {
	Where *sqlf.Query
	Order *sqlf.Query
	Limit *sqlf.Query
}

func (a *QueryArgs) AppendWhereToQuery(query *sqlf.Query) *sqlf.Query {
	if a.Where == nil {
		return query
	}

	return sqlf.Sprintf("%v WHERE %v", query, a.Where)
}

func (a *QueryArgs) AppendOrderToQuery(query *sqlf.Query) *sqlf.Query {
	if a.Order == nil {
		return query
	}

	return sqlf.Sprintf("%v ORDER BY %v", query, a.Order)
}

func (a *QueryArgs) AppendLimitToQuery(query *sqlf.Query) *sqlf.Query {
	if a.Limit == nil {
		return query
	}

	return sqlf.Sprintf("%v %v", query, a.Limit)
}

func (a *QueryArgs) AppendAllToQuery(query *sqlf.Query) *sqlf.Query {
	query = a.AppendWhereToQuery(query)
	query = a.AppendOrderToQuery(query)
	query = a.AppendLimitToQuery(query)

	return query
}

type OrderBy []OrderByOption

func (o OrderBy) Columns() []string {
	columns := []string{}

	for _, orderOption := range o {
		columns = append(columns, orderOption.Field)
	}

	return columns
}

func (o OrderBy) SQL(ascending bool) *sqlf.Query {
	columns := []*sqlf.Query{}

	for _, orderOption := range o {
		columns = append(columns, orderOption.SQL(ascending))
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
	OrderByNullsLast  OrderByNulls = "LAST"
)

func (o OrderByOption) SQL(ascending bool) *sqlf.Query {
	var sb strings.Builder

	sb.WriteString(o.Field)

	if ascending {
		sb.WriteString(" ASC")
	} else {
		sb.WriteString(" DESC")
	}

	if o.Nulls == OrderByNullsFirst || o.Nulls == OrderByNullsLast {
		sb.WriteString(" NULLS " + string(o.Nulls))
	}

	return sqlf.Sprintf(sb.String())
}

type PaginationArgs struct {
	First  *int
	Last   *int
	After  *string
	Before *string

	// TODO(naman): explain default
	OrderBy   OrderBy
	Ascending bool
}

func (p *PaginationArgs) SQL() *QueryArgs {
	queryArgs := &QueryArgs{}

	var conditions []*sqlf.Query

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

		conditions = append(conditions, sqlf.Sprintf(fmt.Sprintf(condition+" (%s)", *p.After)))
	}
	if p.Before != nil {
		columnsStr := strings.Join(orderByColumns, ", ")
		condition := fmt.Sprintf("(%s) <", columnsStr)
		if !p.Ascending {
			condition = fmt.Sprintf("(%s) >", columnsStr)
		}

		conditions = append(conditions, sqlf.Sprintf(fmt.Sprintf(condition+" (%s)", *p.Before)))
	}

	if len(conditions) > 0 {
		queryArgs.Where = sqlf.Sprintf("%v", sqlf.Join(conditions, "AND "))
	}

	if p.First != nil {
		queryArgs.Order = orderBy.SQL(p.Ascending)
		queryArgs.Limit = sqlf.Sprintf("LIMIT %d", *p.First)
	} else if p.Last != nil {
		queryArgs.Order = orderBy.SQL(!p.Ascending)
		queryArgs.Limit = sqlf.Sprintf("LIMIT %d", *p.Last)
	} else {
		queryArgs.Order = orderBy.SQL(p.Ascending)
	}

	return queryArgs
}

func copyPtr[T any](n *T) *T {
	if n == nil {
		return nil
	}

	c := *n
	return &c
}

// Clone (aka deepcopy) returns a new PaginationArgs object with the same values
// as "p".
func (p *PaginationArgs) Clone() *PaginationArgs {
	return &PaginationArgs{
		First:  copyPtr(p.First),
		Last:   copyPtr(p.Last),
		After:  copyPtr(p.After),
		Before: copyPtr(p.Before),
	}
}
