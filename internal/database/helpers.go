package database

import (
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/lib/errors"
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

type QueryArgs struct {
	Where *sqlf.Query
	Order *sqlf.Query
	Limit *sqlf.Query
}

func (a *QueryArgs) AppendWhere(query *sqlf.Query) *sqlf.Query {
	if query == nil || a == nil {
		return nil
	}

	if a.Where == nil {
		return query
	}

	return sqlf.Sprintf("%v WHERE %v", query, a.Where)
}

func (a *QueryArgs) AppendOrder(query *sqlf.Query) *sqlf.Query {
	if query == nil || a == nil {
		return nil
	}

	if a.Order == nil {
		return query
	}

	return sqlf.Sprintf("%v ORDER BY %v", query, a.Order)
}

func (a *QueryArgs) AppendLimit(query *sqlf.Query) *sqlf.Query {
	if query == nil || a == nil {
		return nil
	}

	if a.Limit == nil {
		return query
	}

	return sqlf.Sprintf("%v %v", query, a.Limit)
}

func (a *QueryArgs) AppendAll(query *sqlf.Query) *sqlf.Query {
	query = a.AppendWhere(query)
	query = a.AppendOrder(query)
	query = a.AppendLimit(query)

	return query
}

type PaginationArgs struct {
	First  *int32
	Last   *int32
	After  *int32
	Before *int32
}

func (p *PaginationArgs) SQL() (queryArgs *QueryArgs, err error) {
	if p == nil {
		return nil, errors.New("pagination args is nil")
	}

	queryArgs = &QueryArgs{}

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
		err = errors.New("First or Last must be set")
	}

	return
}
