package database

import (
	"strconv"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type OrderByDirection string

var (
	AscendingOrderByDirection  OrderByDirection = "ASC"
	DescendingOrderByDirection OrderByDirection = "DESC"
)

// LimitOffset specifies SQL LIMIT and OFFSET counts. A pointer to it is typically embedded in other options
// structs that need to perform SQL queries with LIMIT and OFFSET.
type LimitOffset struct {
	Limit  int // SQL LIMIT count
	Offset int // SQL OFFSET count
}

// SQL returns the SQL query fragment ("LIMIT %d OFFSET %d") for use in SQL queries.
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

type PaginationArgs struct {
	First  *int
	Last   *int
	After  *int
	Before *int
}

func (p *PaginationArgs) SQL() (*QueryArgs, error) {
	queryArgs := &QueryArgs{}

	var conditions []*sqlf.Query
	if p.After != nil {
		conditions = append(conditions, sqlf.Sprintf("id < %v", p.After))
	}
	if p.Before != nil {
		conditions = append(conditions, sqlf.Sprintf("id > %v", p.Before))
	}
	if len(conditions) > 0 {
		queryArgs.Where = sqlf.Sprintf("%v", sqlf.Join(conditions, "AND "))
	}

	if p.First != nil {
		queryArgs.Order = sqlf.Sprintf("id DESC")
		queryArgs.Limit = sqlf.Sprintf("LIMIT %d", *p.First)
	} else if p.Last != nil {
		queryArgs.Order = sqlf.Sprintf("id ASC")
		queryArgs.Limit = sqlf.Sprintf("LIMIT %d", *p.Last)
	} else {
		return nil, errors.New("First or Last must be set")
	}

	return queryArgs, nil
}
