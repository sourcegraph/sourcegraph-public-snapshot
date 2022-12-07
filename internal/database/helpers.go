package database

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/keegancsmith/sqlf"
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

// relayUnmarshalID is a best effort decoding of the ID from a marshalled
// graphql.ID
func relayUnmarshalID(s string) (id int32, ok bool) {
	err := relay.UnmarshalSpec(graphql.ID(s), &id)
	return id, err == nil
}
