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

func BuildLimitOffsetArgs(limit *int32, offset *int32) *LimitOffset {
	if limit == nil {
		return nil
	}

	if offset == nil {
		return &LimitOffset{
			int(*limit),
			0,
		}
	}

	return &LimitOffset{
		int(*limit),
		int(*offset),
	}
}

type PaginationArgs struct {
	First  *int32
	Last   *int32
	After  *string
	Before *string
}

func (p *PaginationArgs) SQL() (where *sqlf.Query, order *sqlf.Query, err error) {
	if p == nil {
		return nil, nil, errors.New("pagination args is nil")
	}

	/*
		10

		before 10 last 5

		5 6 7 8 9

		before 10 first 4

		1 2 3 4
	*/

	var conditions []*sqlf.Query
	if p.After != nil {
		conditions = append(conditions, sqlf.Sprintf("id > %d", p.After))
	}
	if p.Before != nil {
		conditions = append(conditions, sqlf.Sprintf("id < %d", p.Before))
	}

	where = sqlf.Sprintf("%v", sqlf.Join(conditions, "AND "))

	if p.First != nil {
		order = sqlf.Sprintf("id DESC LIMIT %d", *p.First)
	} else if p.Last != nil {
		order = sqlf.Sprintf("id ASC LIMIT %d", *p.Last)
	} else {
		err = errors.New("First or Last must be set")
	}

	return
}
