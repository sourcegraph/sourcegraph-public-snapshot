package store

import (
	"fmt"

	"github.com/keegancsmith/sqlf"
)

type CursorOpts struct {
	Limit  int
	Cursor int64
}

type CursorDirection int

const (
	CursorDirectionAscending CursorDirection = iota
	CursorDirectionDescending
)

func (dir CursorDirection) String() string {
	if dir == CursorDirectionDescending {
		return "DESC"
	}
	return "ASC"
}

func (o CursorOpts) DBLimit() int {
	if o.Limit == 0 {
		return o.Limit
	}
	// We always request one item more than actually requested, to determine the next ID for pagination.
	// The store should make sure to strip the last element in a result set, if len(rs) == o.DBLimit().
	return o.Limit + 1
}

func (o CursorOpts) LimitDB() *sqlf.Query {
	if o.Limit == 0 {
		return sqlf.Sprintf("")
	}
	return sqlf.Sprintf(fmt.Sprintf("LIMIT %d", o.Limit+1))
}

func (o CursorOpts) WhereDB(cursorField string, direction CursorDirection) *sqlf.Query {
	if o.Cursor == 0 {
		return sqlf.Sprintf("TRUE")
	}
	op := ">="
	if direction == CursorDirectionDescending {
		op = "<="
	}
	return sqlf.Sprintf("%s %s %s", sqlf.Sprintf(cursorField), sqlf.Sprintf(op), o.Cursor)
}

type Cursor interface {
	Cursor() int64
}

func CursorResultset[T Cursor](o CursorOpts, results []T, err error) ([]T, int64, error) {
	if err != nil {
		return nil, 0, err
	}
	if o.Limit == 0 {
		return results, 0, nil
	}

	if len(results) > o.Limit {
		return results[0:o.Limit], results[o.Limit].Cursor(), nil
	} else if len(results) > 0 {
		return results, 0, nil
	}

	return nil, 0, nil
}

func CursorIntResultset[T ~int | ~int8 | ~int16 | ~int32 | ~int64](o CursorOpts, results []T, err error) ([]T, int64, error) {
	if err != nil {
		return nil, 0, err
	}
	if o.Limit == 0 {
		return results, 0, nil
	}

	if len(results) > o.Limit {
		return results[0:o.Limit], int64(results[o.Limit]), nil
	} else if len(results) > 0 {
		return results, 0, nil
	}

	return nil, 0, nil
}
