package database

import "github.com/keegancsmith/sqlf"

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
